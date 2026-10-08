package service

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"awesome-chances/backend/internal/model"
	"awesome-chances/backend/internal/security"
)

// AuthRepository 由 PostgreSQL 实现；测试使用替身，不提供运行时内存账号。
type AuthRepository interface {
	CreateUser(context.Context, model.UserRecord) (model.UserRecord, error)
	UserByEmail(context.Context, string) (model.UserRecord, error)
	CreateSession(context.Context, model.SessionRecord) error
	SessionUser(context.Context, string, time.Time) (model.UserRecord, error)
	DeleteSession(context.Context, string) error
}

type Auth struct {
	repo      AuthRepository
	settings  model.AuthConfig
	dummyHash string
	hashJobs  chan struct{}
	now       func() time.Time
}

func NewAuth(repo AuthRepository, settings model.AuthConfig) (*Auth, error) {
	if repo == nil || settings.MaxHashJobs < 1 || settings.MinPasswordLength < 8 || settings.SessionTTL <= 0 {
		return nil, errors.New("invalid auth configuration")
	}
	dummy, err := security.PasswordHash("dummy account verification password")
	if err != nil {
		return nil, err
	}
	return &Auth{repo: repo, settings: settings, dummyHash: dummy, hashJobs: make(chan struct{}, settings.MaxHashJobs), now: time.Now}, nil
}

func canonicalEmail(raw string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(raw))
	if len(email) > 254 || email == "" {
		return "", invalid("请填写有效邮箱")
	}
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email || !strings.Contains(email, "@") {
		return "", invalid("请填写有效邮箱")
	}
	return email, nil
}

func (a *Auth) validatePassword(password string) error {
	length := utf8.RuneCountInString(password)
	if length < a.settings.MinPasswordLength || length > 128 || len(password) > 512 {
		return invalid(fmt.Sprintf("密码需为 %d–128 个字符，最多 512 字节", a.settings.MinPasswordLength))
	}
	return nil
}

func (a *Auth) acquireHash(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	select {
	case a.hashJobs <- struct{}{}:
		return nil
	default:
		return model.ErrAuthBusy
	}
}

func userView(record model.UserRecord) model.UserView {
	return model.UserView{ID: record.ID, Email: record.Email, DisplayName: record.DisplayName, CreatedAt: record.CreatedAt}
}

func (a *Auth) Register(ctx context.Context, input model.RegisterRequest) (model.UserView, error) {
	email, err := canonicalEmail(input.Email)
	if err != nil {
		return model.UserView{}, err
	}
	if err := a.validatePassword(input.Password); err != nil {
		return model.UserView{}, err
	}
	name := strings.TrimSpace(input.DisplayName)
	if name == "" || utf8.RuneCountInString(name) > 20 {
		return model.UserView{}, invalid("昵称需为 1–20 个字符")
	}
	if err := a.acquireHash(ctx); err != nil {
		return model.UserView{}, err
	}
	hash, err := security.PasswordHash(input.Password)
	<-a.hashJobs
	if err != nil {
		return model.UserView{}, err
	}
	id, err := security.RandomID()
	if err != nil {
		return model.UserView{}, err
	}
	now := a.now().UTC()
	record, err := a.repo.CreateUser(ctx, model.UserRecord{ID: id, Email: email, PasswordHash: hash, DisplayName: name, CreatedAt: now, UpdatedAt: now})
	if err != nil {
		return model.UserView{}, err
	}
	return userView(record), nil
}

func (a *Auth) Login(ctx context.Context, input model.LoginRequest) (model.LoginResponse, string, error) {
	email, err := canonicalEmail(input.Email)
	if err != nil {
		return model.LoginResponse{}, "", model.ErrCredentials
	}
	// The login limit is independent of registration minimum, so policy changes do not lock out existing accounts.
	if input.Password == "" || utf8.RuneCountInString(input.Password) > 128 || len(input.Password) > 512 {
		return model.LoginResponse{}, "", model.ErrCredentials
	}
	if err := a.acquireHash(ctx); err != nil {
		return model.LoginResponse{}, "", err
	}
	defer func() { <-a.hashJobs }()
	user, err := a.repo.UserByEmail(ctx, email)
	missing := errors.Is(err, model.ErrUserNotFound)
	if err != nil && !missing {
		return model.LoginResponse{}, "", err
	}
	hash := user.PasswordHash
	if missing {
		hash = a.dummyHash
	}
	match, err := security.VerifyPassword(input.Password, hash)
	if err != nil {
		return model.LoginResponse{}, "", errors.New("stored password hash is invalid")
	}
	if !match || missing {
		return model.LoginResponse{}, "", model.ErrCredentials
	}
	token, err := security.RandomToken()
	if err != nil {
		return model.LoginResponse{}, "", err
	}
	now := a.now().UTC()
	expires := now.Add(a.settings.SessionTTL)
	if err := a.repo.CreateSession(ctx, model.SessionRecord{TokenHash: security.TokenHash(token), UserID: user.ID, CreatedAt: now, ExpiresAt: expires}); err != nil {
		return model.LoginResponse{}, "", err
	}
	return model.LoginResponse{User: userView(user), ExpiresAt: expires}, token, nil
}

func (a *Auth) CurrentUser(ctx context.Context, token string) (model.UserView, error) {
	if len(token) != 43 {
		return model.UserView{}, model.ErrUnauthenticated
	}
	user, err := a.repo.SessionUser(ctx, security.TokenHash(token), a.now().UTC())
	if err != nil {
		return model.UserView{}, err
	}
	return userView(user), nil
}

func (a *Auth) Logout(ctx context.Context, token string) error {
	if len(token) != 43 {
		return nil
	}
	return a.repo.DeleteSession(ctx, security.TokenHash(token))
}
