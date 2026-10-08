package postgres

import (
	"testing"
	"time"

	"awesome-chances/backend/internal/model"
)

func TestConnectionEscapesCredentialsAndConfiguresPool(t *testing.T) {
	settings := model.DatabaseConfig{Enabled: true, Host: "localhost", Port: 5432, Name: "learning space", User: "user@host", Password: "p:@/?#$'", SSLMode: "disable", MaxConns: 7, MinConns: 2, ConnectTimeout: time.Second, QueryTimeout: time.Second, MaxConnLifetime: time.Hour}
	cfg, err := poolConfig(settings)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ConnConfig.User != settings.User || cfg.ConnConfig.Password != settings.Password || cfg.ConnConfig.Database != settings.Name || cfg.ConnConfig.Port != 5432 || cfg.MaxConns != 7 || cfg.MinConns != 2 || cfg.ConnConfig.ConnectTimeout != time.Second {
		t.Fatal("connection parameters were not preserved")
	}
	settings.Enabled = false
	if _, err := poolConfig(settings); err == nil {
		t.Fatal("disabled database accepted")
	}
}
