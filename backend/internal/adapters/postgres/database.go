package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"

	"awesome-chances/backend/internal/model"
	"awesome-chances/backend/migrations"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Database 管理连接池和生命周期；配置模型在 model/，不暴露连接凭据。
type Database struct {
	pool     *pgxpool.Pool
	settings model.DatabaseConfig
}

func poolConfig(settings model.DatabaseConfig) (*pgxpool.Config, error) {
	if !settings.Enabled || settings.Host == "" || settings.Port < 1 || settings.Port > 65535 || settings.Name == "" || settings.User == "" || settings.Password == "" || settings.MaxConns < 1 || settings.MinConns < 0 || settings.MinConns > settings.MaxConns || settings.ConnectTimeout <= 0 || settings.QueryTimeout <= 0 || settings.MaxConnLifetime <= 0 {
		return nil, errors.New("数据库配置不完整")
	}
	u := url.URL{Scheme: "postgres", Host: net.JoinHostPort(settings.Host, strconv.Itoa(settings.Port)), Path: "/" + settings.Name, User: url.UserPassword(settings.User, settings.Password)}
	query := url.Values{"sslmode": {settings.SSLMode}}
	u.RawQuery = query.Encode()
	parsed, err := pgxpool.ParseConfig(u.String())
	if err != nil {
		return nil, errors.New("数据库配置无效")
	}
	parsed.MaxConns = settings.MaxConns
	parsed.MinConns = settings.MinConns
	parsed.MaxConnLifetime = settings.MaxConnLifetime
	parsed.ConnConfig.ConnectTimeout = settings.ConnectTimeout
	return parsed, nil
}

// Connect 创建池后立即 Ping，不能把仅分配连接池当作连接成功。
func Connect(ctx context.Context, settings model.DatabaseConfig) (*Database, error) {
	cfg, err := poolConfig(settings)
	if err != nil {
		return nil, err
	}
	connectCtx, cancel := context.WithTimeout(ctx, settings.ConnectTimeout)
	defer cancel()
	pool, err := pgxpool.NewWithConfig(connectCtx, cfg)
	if err != nil {
		return nil, errors.New("创建数据库连接池失败")
	}
	db := &Database{pool: pool, settings: settings}
	if err := db.Ping(connectCtx); err != nil {
		db.Close()
		return nil, errors.New("数据库连接失败，请检查 PG 配置、网络和 PostgreSQL 服务")
	}
	return db, nil
}

func (d *Database) Ping(ctx context.Context) error {
	queryCtx, cancel := context.WithTimeout(ctx, d.settings.QueryTimeout)
	defer cancel()
	return d.pool.Ping(queryCtx)
}

func (d *Database) Close() {
	if d != nil && d.pool != nil {
		d.pool.Close()
	}
}

// Migrate 在事务和数据库锁内执行嵌入的版本 SQL，记录校验和，避免重复或并发建表。
func (d *Database) Migrate(ctx context.Context) error {
	queryCtx, cancel := context.WithTimeout(ctx, d.settings.QueryTimeout)
	defer cancel()
	tx, err := d.pool.Begin(queryCtx)
	if err != nil {
		return errors.New("无法开启数据库迁移事务")
	}
	defer tx.Rollback(queryCtx)
	if _, err = tx.Exec(queryCtx, `SELECT pg_advisory_xact_lock(73510426)`); err != nil {
		return errors.New("获取数据库迁移锁失败")
	}
	if _, err = tx.Exec(queryCtx, `CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY, checksum TEXT NOT NULL, applied_at TIMESTAMPTZ NOT NULL DEFAULT now())`); err != nil {
		return errors.New("创建数据库迁移记录失败")
	}
	entries, err := migrations.Files.ReadDir(".")
	if err != nil {
		return err
	}
	for _, entry := range entries {
		body, err := migrations.Files.ReadFile(entry.Name())
		if err != nil {
			return err
		}
		sum := sha256.Sum256(body)
		checksum := hex.EncodeToString(sum[:])
		var stored string
		err = tx.QueryRow(queryCtx, `SELECT checksum FROM schema_migrations WHERE version=$1`, entry.Name()).Scan(&stored)
		if err == nil {
			if stored != checksum {
				return fmt.Errorf("已执行迁移 %s 的校验和不一致", entry.Name())
			}
			continue
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return errors.New("读取数据库迁移版本失败")
		}
		if _, err = tx.Exec(queryCtx, string(body)); err != nil {
			return fmt.Errorf("执行迁移 %s 失败，请检查数据库权限与现有表", entry.Name())
		}
		if _, err = tx.Exec(queryCtx, `INSERT INTO schema_migrations(version,checksum) VALUES($1,$2)`, entry.Name(), checksum); err != nil {
			return errors.New("记录数据库迁移版本失败")
		}
	}
	if err = tx.Commit(queryCtx); err != nil {
		return errors.New("提交数据库迁移失败")
	}
	return nil
}

// CheckSchema 确认已执行当前迁移，避免服务带着缺失的用户表启动。
func (d *Database) CheckSchema(ctx context.Context) error {
	queryCtx, cancel := context.WithTimeout(ctx, d.settings.QueryTimeout)
	defer cancel()
	entries, err := migrations.Files.ReadDir(".")
	if err != nil {
		return err
	}
	for _, entry := range entries {
		var stored string
		if err := d.pool.QueryRow(queryCtx, `SELECT checksum FROM schema_migrations WHERE version=$1`, entry.Name()).Scan(&stored); err != nil {
			return errors.New("数据库迁移未完成，请先运行 go run ./cmd/migrate")
		}
		body, _ := migrations.Files.ReadFile(entry.Name())
		sum := sha256.Sum256(body)
		if stored != hex.EncodeToString(sum[:]) {
			return errors.New("数据库迁移校验和不一致")
		}
	}
	return nil
}
