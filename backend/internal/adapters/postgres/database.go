package postgres

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"net"
	"net/url"
	"strconv"

	"awesome-chances/backend/internal/model"
	"awesome-chances/backend/migrations"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database 通过 GORM 操作 PostgreSQL，database/sql 管理连接池和生命周期。
// orm 和 sqlDB 不向业务层暴露；配置模型在 model/，不记录连接凭据。
type Database struct {
	orm      *gorm.DB
	sqlDB    *sql.DB
	settings model.DatabaseConfig
}

func databaseDSN(settings model.DatabaseConfig) (string, error) {
	if !settings.Enabled || settings.Host == "" || settings.Port < 1 || settings.Port > 65535 || settings.Name == "" || settings.User == "" || settings.Password == "" || settings.MaxConns < 1 || settings.MaxIdleConns < 0 || settings.MaxIdleConns > settings.MaxConns || settings.ConnectTimeout <= 0 || settings.QueryTimeout <= 0 || settings.MaxConnLifetime <= 0 {
		return "", errors.New("数据库配置不完整")
	}
	switch settings.SSLMode {
	case "disable", "require", "verify-ca", "verify-full":
	default:
		return "", errors.New("数据库 SSL 模式无效")
	}
	u := url.URL{Scheme: "postgres", Host: net.JoinHostPort(settings.Host, strconv.Itoa(settings.Port)), Path: "/" + settings.Name, User: url.UserPassword(settings.User, settings.Password)}
	// PostgreSQL connect_timeout 以整秒计；首次 Ping 同时受精确的 ConnectTimeout 约束。
	query := url.Values{
		"sslmode":         {settings.SSLMode},
		"connect_timeout": {strconv.FormatInt(int64(math.Ceil(settings.ConnectTimeout.Seconds())), 10)},
	}
	u.RawQuery = query.Encode()
	return u.String(), nil
}

func openGORM(dialector gorm.Dialector) (*gorm.DB, error) {
	return gorm.Open(dialector, &gorm.Config{
		// 首次探测由 Connect 中带超时的 Ping 执行，避免 GORM 的无上下文自动 Ping。
		DisableAutomaticPing: true,
		// 单条增删已原子执行；会话清理/创建和迁移显式使用事务。
		SkipDefaultTransaction: true,
		// 不输出 SQL 参数、密码哈希、会话令牌哈希或初始化错误中的连接配置。
		Logger: logger.Default.LogMode(logger.Silent),
	})
}

func configurePool(pool *sql.DB, settings model.DatabaseConfig) {
	pool.SetMaxOpenConns(int(settings.MaxConns))
	pool.SetMaxIdleConns(int(settings.MaxIdleConns))
	pool.SetConnMaxLifetime(settings.MaxConnLifetime)
}

// Connect 用 GORM 创建数据库句柄，配置连接池后立即 Ping；未探测成功不返回连接。
func Connect(ctx context.Context, settings model.DatabaseConfig) (*Database, error) {
	dsn, err := databaseDSN(settings)
	if err != nil {
		return nil, err
	}
	orm, err := openGORM(gormpostgres.Open(dsn))
	if err != nil {
		return nil, errors.New("初始化 GORM 数据库连接失败")
	}
	pool, err := orm.DB()
	if err != nil {
		return nil, errors.New("获取数据库连接池失败")
	}
	configurePool(pool, settings)
	db := &Database{orm: orm, sqlDB: pool, settings: settings}
	connectCtx, cancel := context.WithTimeout(ctx, settings.ConnectTimeout)
	defer cancel()
	if err := db.Ping(connectCtx); err != nil {
		db.Close()
		return nil, errors.New("数据库连接失败，请检查 PG 配置、网络和 PostgreSQL 服务")
	}
	return db, nil
}

func (d *Database) Ping(ctx context.Context) error {
	queryCtx, cancel := context.WithTimeout(ctx, d.settings.QueryTimeout)
	defer cancel()
	return d.sqlDB.PingContext(queryCtx)
}

func (d *Database) Close() {
	if d != nil && d.sqlDB != nil {
		_ = d.sqlDB.Close()
	}
}

// Migrate 通过 GORM 事务执行版本 SQL，保留 advisory lock、校验和及已有表结构。
// 不使用 AutoMigrate 替代版本迁移；业务 CRUD 使用 ORM，DDL 与迁移锁保留必要 SQL。
func (d *Database) Migrate(ctx context.Context) error {
	queryCtx, cancel := context.WithTimeout(ctx, d.settings.QueryTimeout)
	defer cancel()
	// 只返回我们生成的脱敏诊断，不将数据库原始错误输出到启动日志。
	var diagnostic error
	err := d.orm.WithContext(queryCtx).Transaction(func(tx *gorm.DB) error {
		fail := func(message string) error {
			diagnostic = errors.New(message)
			return diagnostic
		}
		if err := tx.Exec(`SELECT pg_advisory_xact_lock(73510426)`).Error; err != nil {
			return fail("获取数据库迁移锁失败")
		}
		if err := tx.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY, checksum TEXT NOT NULL, applied_at TIMESTAMPTZ NOT NULL DEFAULT now())`).Error; err != nil {
			return fail("创建数据库迁移记录失败")
		}
		entries, err := migrations.Files.ReadDir(".")
		if err != nil {
			return fail("读取数据库迁移文件失败")
		}
		for _, entry := range entries {
			body, err := migrations.Files.ReadFile(entry.Name())
			if err != nil {
				return fail("读取数据库迁移内容失败")
			}
			sum := sha256.Sum256(body)
			checksum := hex.EncodeToString(sum[:])
			var stored model.MigrationRecord
			err = tx.Where("version = ?", entry.Name()).Take(&stored).Error
			if err == nil {
				if stored.Checksum != checksum {
					return fail(fmt.Sprintf("已执行迁移 %s 的校验和不一致", entry.Name()))
				}
				continue
			}
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return fail("读取数据库迁移版本失败")
			}
			if err := tx.Exec(string(body)).Error; err != nil {
				return fail(fmt.Sprintf("执行迁移 %s 失败，请检查数据库权限与现有表", entry.Name()))
			}
			record := model.MigrationRecord{Version: entry.Name(), Checksum: checksum}
			if err := tx.Create(&record).Error; err != nil {
				return fail("记录数据库迁移版本失败")
			}
		}
		return nil
	})
	if diagnostic != nil {
		return diagnostic
	}
	if err != nil {
		return errors.New("数据库迁移事务失败")
	}
	return nil
}

// CheckSchema 确认已执行当前迁移，避免服务带着缺失的用户表启动。
func (d *Database) CheckSchema(ctx context.Context) error {
	queryCtx, cancel := context.WithTimeout(ctx, d.settings.QueryTimeout)
	defer cancel()
	entries, err := migrations.Files.ReadDir(".")
	if err != nil {
		return errors.New("读取数据库迁移文件失败")
	}
	for _, entry := range entries {
		var stored model.MigrationRecord
		if err := d.orm.WithContext(queryCtx).Where("version = ?", entry.Name()).Take(&stored).Error; err != nil {
			return errors.New("数据库迁移未完成，请先运行 go run ./cmd/migrate")
		}
		body, err := migrations.Files.ReadFile(entry.Name())
		if err != nil {
			return errors.New("读取数据库迁移内容失败")
		}
		sum := sha256.Sum256(body)
		if stored.Checksum != hex.EncodeToString(sum[:]) {
			return errors.New("数据库迁移校验和不一致")
		}
	}
	return nil
}
