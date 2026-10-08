package postgres

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"awesome-chances/backend/internal/model"
	"awesome-chances/backend/internal/security"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Opt-in real database test, using an isolated generated schema in a dedicated test DB.
func TestPostgresMigrationAndAuthRepository(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL for PostgreSQL integration test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	base, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal("invalid test database settings")
	}
	defer base.Close()
	id, err := security.RandomID()
	if err != nil {
		t.Fatal(err)
	}
	schema := "test_auth_" + id
	quoted := pgx.Identifier{schema}.Sanitize()
	if _, err := base.Exec(ctx, "CREATE SCHEMA "+quoted); err != nil {
		t.Fatal("cannot create isolated test schema")
	}
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := base.Exec(cleanup, "DROP SCHEMA "+quoted+" CASCADE"); err != nil {
			t.Error("test schema cleanup failed")
		}
	}()
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal("invalid test connection settings")
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal("cannot open test connection pool")
	}
	db := &Database{pool: pool, settings: model.DatabaseConfig{QueryTimeout: 10 * time.Second}}
	defer db.Close()
	for i := 0; i < 2; i++ {
		if err := db.Migrate(ctx); err != nil {
			t.Fatal(err)
		}
	}
	if err := db.CheckSchema(ctx); err != nil {
		t.Fatal(err)
	}
	repo := NewAuthRepository(db)
	hash, err := security.PasswordHash("correct horse battery staple")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	user := model.UserRecord{ID: id, Email: "learner@example.com", DisplayName: "x'); DROP users; --", PasswordHash: hash, CreatedAt: now, UpdatedAt: now}
	if _, err := repo.CreateUser(ctx, user); err != nil {
		t.Fatal("user insert failed")
	}
	duplicate := user
	duplicate.ID, _ = security.RandomID()
	if _, err := repo.CreateUser(ctx, duplicate); !errors.Is(err, model.ErrEmailExists) {
		t.Fatal("duplicate email not mapped")
	}
	got, err := repo.UserByEmail(ctx, user.Email)
	if err != nil || got.PasswordHash != hash || got.DisplayName != user.DisplayName {
		t.Fatal("user query/parameter binding failed")
	}
	token, _ := security.RandomToken()
	session := model.SessionRecord{TokenHash: security.TokenHash(token), UserID: id, CreatedAt: now, ExpiresAt: now.Add(time.Hour)}
	if err := repo.CreateSession(ctx, session); err != nil {
		t.Fatal("session insert failed")
	}
	if got, err := repo.SessionUser(ctx, session.TokenHash, now); err != nil || got.ID != id {
		t.Fatal("session lookup failed")
	}
	if _, err := repo.SessionUser(ctx, session.TokenHash, now.Add(2*time.Hour)); !errors.Is(err, model.ErrUnauthenticated) {
		t.Fatal("expired session accepted")
	}
	if err := repo.DeleteSession(ctx, session.TokenHash); err != nil {
		t.Fatal("session revoke failed")
	}
	if _, err := repo.SessionUser(ctx, session.TokenHash, now); !errors.Is(err, model.ErrUnauthenticated) {
		t.Fatal("revoked session accepted")
	}
}
