package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"awesome-chances/backend/internal/model"
	"awesome-chances/backend/migrations"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"
	gormpostgres "gorm.io/driver/postgres"
)

// 执行真实 GORM 查询构建/映射和事务逻辑，只有数据库驱动响应使用替身。
func mockDatabase(t *testing.T) (*Database, sqlmock.Sqlmock) {
	t.Helper()
	pool, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatal(err)
	}
	orm, err := openGORM(gormpostgres.New(gormpostgres.Config{Conn: pool}))
	if err != nil {
		_ = pool.Close()
		t.Fatal(err)
	}
	settings := connectionSettings()
	configurePool(pool, settings)
	db := &Database{orm: orm, sqlDB: pool, settings: settings}
	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
		mock.ExpectClose()
		if err := pool.Close(); err != nil {
			t.Error(err)
		}
	})
	return db, mock
}

var userColumns = []string{"id", "email", "password_hash", "display_name", "created_at", "updated_at"}

func storedUser() model.UserRecord {
	now := time.Date(2026, 10, 8, 1, 0, 0, 0, time.UTC)
	return model.UserRecord{ID: "learner-id", Email: "learner@example.com", PasswordHash: "$argon2id$test-hash", DisplayName: "x'); DROP users; --", CreatedAt: now, UpdatedAt: now}
}

func userRows(user model.UserRecord) *sqlmock.Rows {
	return sqlmock.NewRows(userColumns).AddRow(user.ID, user.Email, user.PasswordHash, user.DisplayName, user.CreatedAt, user.UpdatedAt)
}

func TestGORMCreateUserPreservesFieldsAndEmailConflictMapping(t *testing.T) {
	for _, test := range []struct {
		name string
		err  error
		want error
	}{
		{name: "success"},
		{name: "email-conflict", err: &pgconn.PgError{Code: "23505", ConstraintName: "users_email_key"}, want: model.ErrEmailExists},
		{name: "other-conflict", err: &pgconn.PgError{Code: "23505", ConstraintName: "users_pkey"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			db, mock := mockDatabase(t)
			user := storedUser()
			insert := mock.ExpectExec(`INSERT INTO "users"`).WithArgs(user.ID, user.Email, user.PasswordHash, user.DisplayName, user.CreatedAt, user.UpdatedAt)
			if test.err != nil {
				insert.WillReturnError(test.err)
			} else {
				insert.WillReturnResult(sqlmock.NewResult(0, 1))
			}
			got, err := NewAuthRepository(db).CreateUser(context.Background(), user)
			switch {
			case test.want != nil:
				if !errors.Is(err, test.want) {
					t.Fatal("email uniqueness error not mapped")
				}
			case test.err != nil:
				if !errors.Is(err, test.err) || errors.Is(err, model.ErrEmailExists) {
					t.Fatal("non-email conflict was incorrectly mapped")
				}
			default:
				if err != nil || got != user {
					t.Fatal("ORM failed to preserve application assigned fields")
				}
			}
		})
	}
}

func TestGORMUserLookupAndMissingUser(t *testing.T) {
	for _, missing := range []bool{false, true} {
		db, mock := mockDatabase(t)
		user := storedUser()
		rows := sqlmock.NewRows(userColumns)
		if !missing {
			rows = userRows(user)
		}
		mock.ExpectQuery(`SELECT .* FROM "users" WHERE email = \$1`).WithArgs(user.Email, 1).WillReturnRows(rows)
		got, err := NewAuthRepository(db).UserByEmail(context.Background(), user.Email)
		if missing {
			if !errors.Is(err, model.ErrUserNotFound) {
				t.Fatal("missing user error not mapped")
			}
		} else if err != nil || got != user {
			t.Fatal("user columns not mapped correctly")
		}
	}
}

func TestGORMSessionLookupRequiresHashAndUnexpiredSession(t *testing.T) {
	for _, valid := range []bool{true, false} {
		db, mock := mockDatabase(t)
		user := storedUser()
		rows := sqlmock.NewRows(userColumns)
		if valid {
			rows = userRows(user)
		}
		mock.ExpectQuery(`SELECT users\.\* FROM "users" JOIN auth_sessions ON auth_sessions.user_id = users.id WHERE auth_sessions.token_hash = \$1 AND auth_sessions.expires_at > \$2`).WithArgs("token-hash", user.CreatedAt, 1).WillReturnRows(rows)
		got, err := NewAuthRepository(db).SessionUser(context.Background(), "token-hash", user.CreatedAt)
		if valid {
			if err != nil || got != user {
				t.Fatal("session user mapping failed")
			}
		} else if !errors.Is(err, model.ErrUnauthenticated) {
			t.Fatal("invalid/expired session must require authentication")
		}
	}
}

func TestGORMSessionCreationIsAtomic(t *testing.T) {
	for _, failure := range []string{"", "cleanup", "insert"} {
		t.Run("failure/"+failure, func(t *testing.T) {
			db, mock := mockDatabase(t)
			user := storedUser()
			session := model.SessionRecord{TokenHash: "token-hash", UserID: user.ID, CreatedAt: user.CreatedAt, ExpiresAt: user.CreatedAt.Add(time.Hour)}
			writeError := errors.New("write failed")
			mock.ExpectBegin()
			cleanup := mock.ExpectExec(`DELETE FROM "auth_sessions" WHERE user_id = \$1 AND expires_at <= \$2`).WithArgs(session.UserID, session.CreatedAt)
			if failure == "cleanup" {
				cleanup.WillReturnError(writeError)
			} else {
				cleanup.WillReturnResult(sqlmock.NewResult(0, 2))
				insert := mock.ExpectExec(`INSERT INTO "auth_sessions"`).WithArgs(session.TokenHash, session.UserID, session.CreatedAt, session.ExpiresAt)
				if failure == "insert" {
					insert.WillReturnError(writeError)
				} else {
					insert.WillReturnResult(sqlmock.NewResult(0, 1))
				}
			}
			if failure == "" {
				mock.ExpectCommit()
			} else {
				mock.ExpectRollback()
			}
			err := NewAuthRepository(db).CreateSession(context.Background(), session)
			if failure == "" && err != nil || failure != "" && !errors.Is(err, writeError) {
				t.Fatal("session transaction did not preserve atomicity")
			}
		})
	}
}

func TestGORMLogoutRemainsIdempotent(t *testing.T) {
	db, mock := mockDatabase(t)
	for _, affected := range []int64{1, 0} {
		mock.ExpectExec(`DELETE FROM "auth_sessions" WHERE token_hash = \$1`).WithArgs("token-hash").WillReturnResult(sqlmock.NewResult(0, affected))
		if err := NewAuthRepository(db).DeleteSession(context.Background(), "token-hash"); err != nil {
			t.Fatal("repeated logout must succeed")
		}
	}
}

func TestGORMVersionedMigrationAndChecksumRollback(t *testing.T) {
	body, err := migrations.Files.ReadFile("001_auth.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(body)
	checksum := hex.EncodeToString(sum[:])
	for _, existing := range []string{"", checksum, "modified-checksum"} {
		t.Run("stored/"+existing, func(t *testing.T) {
			db, mock := mockDatabase(t)
			mock.ExpectBegin()
			mock.ExpectExec(`SELECT pg_advisory_xact_lock`).WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectExec(`CREATE TABLE IF NOT EXISTS schema_migrations`).WillReturnResult(sqlmock.NewResult(0, 0))
			rows := sqlmock.NewRows([]string{"version", "checksum"})
			if existing != "" {
				rows.AddRow("001_auth.up.sql", existing)
			}
			mock.ExpectQuery(`SELECT .* FROM "schema_migrations" WHERE version = \$1`).WithArgs("001_auth.up.sql", 1).WillReturnRows(rows)
			if existing == "" {
				mock.ExpectExec(`CREATE TABLE users`).WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectExec(`INSERT INTO "schema_migrations"`).WithArgs("001_auth.up.sql", checksum).WillReturnResult(sqlmock.NewResult(0, 1))
			}
			if existing == "modified-checksum" {
				mock.ExpectRollback()
			} else {
				mock.ExpectCommit()
			}
			err := db.Migrate(context.Background())
			if existing == "modified-checksum" && err == nil || existing != "modified-checksum" && err != nil {
				t.Fatal("migration version/checksum contract was broken")
			}
		})
	}
}

func TestGORMPoolPingAndSchemaCheck(t *testing.T) {
	db, mock := mockDatabase(t)
	if db.sqlDB.Stats().MaxOpenConnections != 7 {
		t.Fatal("maximum pool size not configured")
	}
	mock.ExpectPing()
	if err := db.Ping(context.Background()); err != nil {
		t.Fatal(err)
	}
	for _, rows := range []*sqlmock.Rows{
		sqlmock.NewRows([]string{"version", "checksum"}),
		sqlmock.NewRows([]string{"version", "checksum"}).AddRow("001_auth.up.sql", "changed-checksum"),
	} {
		mock.ExpectQuery(`SELECT .* FROM "schema_migrations" WHERE version = \$1`).WithArgs("001_auth.up.sql", 1).WillReturnRows(rows)
		if err := db.CheckSchema(context.Background()); err == nil {
			t.Fatal("schema check accepted missing or changed migration")
		}
	}
}
