package model

import (
	"errors"
	"time"
)

var (
	ErrEmailExists     = errors.New("邮箱已注册")
	ErrCredentials     = errors.New("邮箱或密码错误")
	ErrUnauthenticated = errors.New("请先登录")
	ErrAuthUnavailable = errors.New("数据库未启用，用户认证暂不可用")
	ErrAuthBusy        = errors.New("认证请求较多，请稍后重试")
	ErrUserNotFound    = errors.New("user not found")
)

type RegisterRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"displayName"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// UserView 是唯一对外用户结构，不包含密码哈希或会话凭据。
type UserView struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"displayName"`
	CreatedAt   time.Time `json:"createdAt"`
}

type LoginResponse struct {
	User      UserView  `json:"user"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type SessionRecord struct {
	TokenHash string    `db:"token_hash" json:"-"`
	UserID    string    `db:"user_id"`
	CreatedAt time.Time `db:"created_at"`
	ExpiresAt time.Time `db:"expires_at"`
}
