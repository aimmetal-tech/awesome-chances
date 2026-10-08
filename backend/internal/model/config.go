package model

import "time"

// AppConfig 是进程运行配置；端口与数据库设置从 .env / 环境变量读取，兼容上层 .env。
type AppConfig struct {
	BackendHost string
	BackendPort int
	Database    DatabaseConfig
	Auth        AuthConfig
}

// DatabaseConfig 不返回到 HTTP，不记录完整配置或连接字符串。
type DatabaseConfig struct {
	Enabled         bool
	AutoMigrate     bool
	Host            string
	Port            int
	Name            string
	User            string
	Password        string `json:"-"`
	SSLMode         string
	MaxConns        int32
	MaxIdleConns    int32
	ConnectTimeout  time.Duration
	QueryTimeout    time.Duration
	MaxConnLifetime time.Duration
}

type AuthConfig struct {
	CookieName         string
	CookieSecure       bool
	SessionTTL         time.Duration
	MinPasswordLength  int
	MaxHashJobs        int
	RateLimitPerMinute int
	AllowedOrigins     []string
}
