package postgres

import (
	"context"
	"errors"
	"time"

	"awesome-chances/backend/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type AuthRepository struct{ db *Database }

func NewAuthRepository(db *Database) *AuthRepository { return &AuthRepository{db: db} }

func scanUser(row pgx.Row) (model.UserRecord, error) {
	var user model.UserRecord
	err := row.Scan(&user.ID, &user.Email, &user.DisplayName, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	return user, err
}

func (r *AuthRepository) CreateUser(ctx context.Context, user model.UserRecord) (model.UserRecord, error) {
	ctx, cancel := context.WithTimeout(ctx, r.db.settings.QueryTimeout)
	defer cancel()
	created, err := scanUser(r.db.pool.QueryRow(ctx, `INSERT INTO users(id,email,display_name,password_hash,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6) RETURNING id,email,display_name,password_hash,created_at,updated_at`, user.ID, user.Email, user.DisplayName, user.PasswordHash, user.CreatedAt, user.UpdatedAt))
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "users_email_key" {
		return model.UserRecord{}, model.ErrEmailExists
	}
	return created, err
}

func (r *AuthRepository) UserByEmail(ctx context.Context, email string) (model.UserRecord, error) {
	ctx, cancel := context.WithTimeout(ctx, r.db.settings.QueryTimeout)
	defer cancel()
	user, err := scanUser(r.db.pool.QueryRow(ctx, `SELECT id,email,display_name,password_hash,created_at,updated_at FROM users WHERE email=$1`, email))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.UserRecord{}, model.ErrUserNotFound
	}
	return user, err
}

func (r *AuthRepository) CreateSession(ctx context.Context, session model.SessionRecord) error {
	ctx, cancel := context.WithTimeout(ctx, r.db.settings.QueryTimeout)
	defer cancel()
	tx, err := r.db.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `DELETE FROM auth_sessions WHERE user_id=$1 AND expires_at<=$2`, session.UserID, session.CreatedAt); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO auth_sessions(token_hash,user_id,created_at,expires_at) VALUES($1,$2,$3,$4)`, session.TokenHash, session.UserID, session.CreatedAt, session.ExpiresAt); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *AuthRepository) SessionUser(ctx context.Context, tokenHash string, now time.Time) (model.UserRecord, error) {
	ctx, cancel := context.WithTimeout(ctx, r.db.settings.QueryTimeout)
	defer cancel()
	user, err := scanUser(r.db.pool.QueryRow(ctx, `SELECT u.id,u.email,u.display_name,u.password_hash,u.created_at,u.updated_at FROM auth_sessions s JOIN users u ON u.id=s.user_id WHERE s.token_hash=$1 AND s.expires_at>$2`, tokenHash, now))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.UserRecord{}, model.ErrUnauthenticated
	}
	return user, err
}

func (r *AuthRepository) DeleteSession(ctx context.Context, tokenHash string) error {
	ctx, cancel := context.WithTimeout(ctx, r.db.settings.QueryTimeout)
	defer cancel()
	_, err := r.db.pool.Exec(ctx, `DELETE FROM auth_sessions WHERE token_hash=$1`, tokenHash)
	return err
}
