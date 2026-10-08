package postgres

import (
	"context"
	"net/url"
	"strings"
	"testing"
	"time"

	"awesome-chances/backend/internal/model"
)

func connectionSettings() model.DatabaseConfig {
	return model.DatabaseConfig{Enabled: true, Host: "localhost", Port: 5432, Name: "learning space", User: "user@host", Password: "p:@/?#$'", SSLMode: "disable", MaxConns: 7, MaxIdleConns: 2, ConnectTimeout: 1500 * time.Millisecond, QueryTimeout: time.Second, MaxConnLifetime: time.Hour}
}

func TestConnectionEscapesCredentials(t *testing.T) {
	settings := connectionSettings()
	dsn, err := databaseDSN(settings)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatal("invalid encoded DSN")
	}
	password, _ := parsed.User.Password()
	if parsed.User.Username() != settings.User || password != settings.Password || parsed.Path != "/"+settings.Name || parsed.Port() != "5432" || parsed.Query().Get("sslmode") != "disable" || parsed.Query().Get("connect_timeout") != "2" {
		t.Fatal("connection parameters were not preserved")
	}
	for _, mutate := range []func(*model.DatabaseConfig){
		func(c *model.DatabaseConfig) { c.Enabled = false },
		func(c *model.DatabaseConfig) { c.SSLMode = "wrong-secret-do-not-print" },
		func(c *model.DatabaseConfig) { c.MaxIdleConns = c.MaxConns + 1 },
	} {
		invalid := settings
		mutate(&invalid)
		if _, err := databaseDSN(invalid); err == nil || strings.Contains(err.Error(), "secret-do-not-print") {
			t.Fatal("invalid or sensitive configuration accepted")
		}
	}
}

func TestCancelledConnectionDoesNotLeakCredentials(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	db, err := Connect(ctx, connectionSettings())
	if db != nil || err == nil || strings.Contains(err.Error(), connectionSettings().Password) {
		t.Fatal("cancelled connection was accepted or leaked credentials")
	}
}
