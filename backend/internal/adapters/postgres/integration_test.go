package postgres

import (
	"context"
	"errors"
	"net/url"
	"os"
	"testing"
	"time"

	"awesome-chances/backend/internal/model"
	"awesome-chances/backend/internal/security"
	gormpostgres "gorm.io/driver/postgres"
)

// 专用 TEST_DATABASE_URL 才执行；在独立 schema 内测试 GORM 与既有版本迁移。
func TestPostgresMigrationAndAuthRepository(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL for PostgreSQL integration test")
	}
	parsed, err := url.Parse(dsn)
	if err != nil || (parsed.Scheme != "postgres" && parsed.Scheme != "postgresql") || parsed.Host == "" {
		t.Fatal("TEST_DATABASE_URL must be a PostgreSQL URL for a dedicated test database")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	base, err := openGORM(gormpostgres.Open(dsn))
	if err != nil {
		t.Fatal("invalid test database settings")
	}
	basePool, err := base.DB()
	if err != nil {
		t.Fatal("cannot open test database pool")
	}
	defer basePool.Close()
	id, err := security.RandomID()
	if err != nil {
		t.Fatal(err)
	}
	// schema 名仅来自随机十六进制 ID，不包含外部输入。
	schema := "test_auth_" + id
	quoted := `"` + schema + `"`
	if err := base.WithContext(ctx).Exec("CREATE SCHEMA " + quoted).Error; err != nil {
		t.Fatal("cannot create isolated test schema")
	}
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := base.WithContext(cleanup).Exec("DROP SCHEMA " + quoted + " CASCADE").Error; err != nil {
			t.Error("test schema cleanup failed")
		}
	}()
	query := parsed.Query()
	query.Set("search_path", schema)
	parsed.RawQuery = query.Encode()
	orm, err := openGORM(gormpostgres.Open(parsed.String()))
	if err != nil {
		t.Fatal("cannot open test ORM")
	}
	pool, err := orm.DB()
	if err != nil {
		t.Fatal("cannot open test connection pool")
	}
	db := &Database{orm: orm, sqlDB: pool, settings: model.DatabaseConfig{QueryTimeout: 10 * time.Second}}
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
	if _, err := repo.UserByEmail(ctx, "missing@example.com"); !errors.Is(err, model.ErrUserNotFound) {
		t.Fatal("missing user not mapped")
	}
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
	if err := repo.DeleteSession(ctx, session.TokenHash); err != nil {
		t.Fatal("repeated logout must remain idempotent")
	}
}
