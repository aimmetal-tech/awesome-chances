package config

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"awesome-chances/backend/internal/model"
	"github.com/joho/godotenv"
)

// Load 从 ENV_FILE 或当前/上层的 .env 读取，已有进程环境变量优先。
func Load() (model.AppConfig, error) {
	if explicit := os.Getenv("ENV_FILE"); explicit != "" {
		if err := godotenv.Load(explicit); err != nil {
			return model.AppConfig{}, fmt.Errorf("无法读取 ENV_FILE")
		}
	} else {
		for _, file := range []string{filepath.Join(".", ".env"), filepath.Join("..", ".env")} {
			if _, err := os.Stat(file); err == nil {
				if err := godotenv.Load(file); err != nil {
					return model.AppConfig{}, fmt.Errorf("无法解析 .env")
				}
				break
			}
		}
	}
	return Parse(os.Getenv)
}

// Parse 独立于文件/全局环境，便于验证配置和复用进程环境。
func Parse(get func(string) string) (model.AppConfig, error) {
	var cfg model.AppConfig
	var parseErr error
	value := func(key, fallback string) string {
		if v := get(key); v != "" {
			return v
		}
		return fallback
	}
	integer := func(key string, fallback, min, max int) int {
		v, err := strconv.Atoi(value(key, strconv.Itoa(fallback)))
		if err != nil || v < min || v > max {
			parseErr = fmt.Errorf("%s 必须为 %d–%d 的整数", key, min, max)
		}
		return v
	}
	boolean := func(key string) bool {
		v, err := strconv.ParseBool(value(key, "false"))
		if err != nil {
			parseErr = fmt.Errorf("%s 必须为 true 或 false", key)
		}
		return v
	}
	duration := func(key, fallback string) time.Duration {
		v, err := time.ParseDuration(value(key, fallback))
		if err != nil || v <= 0 || v > 720*time.Hour {
			parseErr = fmt.Errorf("%s 必须为大于 0 且不超过 720h 的时长", key)
		}
		return v
	}
	cfg.BackendHost = strings.TrimSpace(get("BACKEND_HOST"))
	if cfg.BackendHost == "" {
		return cfg, fmt.Errorf("请在 .env 设置 BACKEND_HOST")
	}
	cfg.BackendPort = integer("BACKEND_PORT", 0, 1, 65535)
	cfg.Auth = model.AuthConfig{
		CookieName: value("AUTH_COOKIE_NAME", "ac_session"), CookieSecure: boolean("AUTH_COOKIE_SECURE"),
		SessionTTL: duration("AUTH_SESSION_TTL", "24h"), MinPasswordLength: integer("AUTH_PASSWORD_MIN_LENGTH", 15, 8, 128),
		MaxHashJobs: integer("AUTH_MAX_HASH_JOBS", 2, 1, 8), RateLimitPerMinute: integer("AUTH_RATE_LIMIT_PER_MINUTE", 30, 1, 1000),
	}
	if err := (&http.Cookie{Name: cfg.Auth.CookieName, Value: "check"}).Valid(); err != nil {
		return cfg, fmt.Errorf("AUTH_COOKIE_NAME 无效")
	}
	if cfg.Auth.SessionTTL < time.Minute {
		return cfg, fmt.Errorf("AUTH_SESSION_TTL 不能少于 1m")
	}
	if origins := get("AUTH_ALLOWED_ORIGINS"); origins != "" {
		for _, raw := range strings.Split(origins, ",") {
			u, err := url.Parse(strings.TrimSpace(raw))
			if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
				return cfg, fmt.Errorf("AUTH_ALLOWED_ORIGINS 必须为逗号分隔的 HTTP/HTTPS origin")
			}
			cfg.Auth.AllowedOrigins = append(cfg.Auth.AllowedOrigins, u.Scheme+"://"+u.Host)
		}
	} else if webHost, webPort := get("WEB_HOST"), get("WEB_PORT"); webHost != "" && webPort != "" {
		cfg.Auth.AllowedOrigins = append(cfg.Auth.AllowedOrigins, "http://"+net.JoinHostPort(webHost, webPort))
		if ip := net.ParseIP(webHost); (ip != nil && ip.IsLoopback()) || webHost == "localhost" {
			cfg.Auth.AllowedOrigins = append(cfg.Auth.AllowedOrigins, "http://"+net.JoinHostPort("localhost", webPort))
		}
	}
	cfg.Database.Enabled = boolean("DATABASE_ENABLED")
	cfg.Database.AutoMigrate = boolean("DATABASE_AUTO_MIGRATE")
	cfg.Database.MaxConns = int32(integer("PG_MAX_CONNS", 10, 1, 100))
	// database/sql 没有最小池容量；新配置控制最大空闲数，兼容旧 PG_MIN_CONNS。
	idleKey := "PG_MAX_IDLE_CONNS"
	if get(idleKey) == "" && get("PG_MIN_CONNS") != "" {
		idleKey = "PG_MIN_CONNS"
	}
	cfg.Database.MaxIdleConns = int32(integer(idleKey, 1, 0, 100))
	cfg.Database.ConnectTimeout = duration("PG_CONNECT_TIMEOUT", "5s")
	cfg.Database.QueryTimeout = duration("PG_QUERY_TIMEOUT", "5s")
	cfg.Database.MaxConnLifetime = duration("PG_MAX_CONN_LIFETIME", "1h")
	if cfg.Database.MaxIdleConns > cfg.Database.MaxConns {
		return cfg, fmt.Errorf("%s 不能超过 PG_MAX_CONNS", idleKey)
	}
	if cfg.Database.Enabled {
		cfg.Database.Host = strings.TrimSpace(get("PGHOST"))
		cfg.Database.Port = integer("PGPORT", 0, 1, 65535)
		cfg.Database.Name = strings.TrimSpace(get("PGDATABASE"))
		cfg.Database.User = strings.TrimSpace(get("PGUSER"))
		cfg.Database.Password = get("PGPASSWORD")
		cfg.Database.SSLMode = strings.TrimSpace(get("PGSSLMODE"))
		for _, item := range []struct{ key, value string }{{"PGHOST", cfg.Database.Host}, {"PGDATABASE", cfg.Database.Name}, {"PGUSER", cfg.Database.User}, {"PGPASSWORD", cfg.Database.Password}} {
			if item.value == "" {
				return cfg, fmt.Errorf("数据库已启用，请填写 %s", item.key)
			}
		}
		switch cfg.Database.SSLMode {
		case "disable", "require", "verify-ca", "verify-full":
		default:
			return cfg, fmt.Errorf("请设置 PGSSLMODE 为 disable、require、verify-ca 或 verify-full")
		}
	}
	if parseErr != nil {
		return cfg, parseErr
	}
	return cfg, nil
}

func Address(cfg model.AppConfig) string {
	return net.JoinHostPort(cfg.BackendHost, strconv.Itoa(cfg.BackendPort))
}
