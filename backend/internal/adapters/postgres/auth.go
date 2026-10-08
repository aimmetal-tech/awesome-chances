package postgres

import (
	"context"
	"errors"
	"time"

	"awesome-chances/backend/internal/model"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type AuthRepository struct{ db *Database }

func NewAuthRepository(db *Database) *AuthRepository { return &AuthRepository{db: db} }

func (r *AuthRepository) CreateUser(ctx context.Context, user model.UserRecord) (model.UserRecord, error) {
	ctx, cancel := context.WithTimeout(ctx, r.db.settings.QueryTimeout)
	defer cancel()
	err := r.db.orm.WithContext(ctx).Create(&user).Error
	// GORM 的 PostgreSQL 驱动底层使用 pgx；只映射邮箱唯一约束，其他冲突保留原错误。
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "users_email_key" {
		return model.UserRecord{}, model.ErrEmailExists
	}
	if err != nil {
		return model.UserRecord{}, err
	}
	return user, nil
}

func (r *AuthRepository) UserByEmail(ctx context.Context, email string) (model.UserRecord, error) {
	ctx, cancel := context.WithTimeout(ctx, r.db.settings.QueryTimeout)
	defer cancel()
	var user model.UserRecord
	err := r.db.orm.WithContext(ctx).Where("email = ?", email).Take(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.UserRecord{}, model.ErrUserNotFound
	}
	return user, err
}

func (r *AuthRepository) CreateSession(ctx context.Context, session model.SessionRecord) error {
	ctx, cancel := context.WithTimeout(ctx, r.db.settings.QueryTimeout)
	defer cancel()
	return r.db.orm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ? AND expires_at <= ?", session.UserID, session.CreatedAt).Delete(&model.SessionRecord{}).Error; err != nil {
			return err
		}
		return tx.Create(&session).Error
	})
}

func (r *AuthRepository) SessionUser(ctx context.Context, tokenHash string, now time.Time) (model.UserRecord, error) {
	ctx, cancel := context.WithTimeout(ctx, r.db.settings.QueryTimeout)
	defer cancel()
	var user model.UserRecord
	err := r.db.orm.WithContext(ctx).Model(&model.UserRecord{}).
		Select("users.*").
		Joins("JOIN auth_sessions ON auth_sessions.user_id = users.id").
		Where("auth_sessions.token_hash = ? AND auth_sessions.expires_at > ?", tokenHash, now).
		Take(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.UserRecord{}, model.ErrUnauthenticated
	}
	return user, err
}

func (r *AuthRepository) DeleteSession(ctx context.Context, tokenHash string) error {
	ctx, cancel := context.WithTimeout(ctx, r.db.settings.QueryTimeout)
	defer cancel()
	return r.db.orm.WithContext(ctx).Where("token_hash = ?", tokenHash).Delete(&model.SessionRecord{}).Error
}
