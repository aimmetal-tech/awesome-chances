package config

import (
	"strings"
	"testing"
)

func TestConfigurationValidationAndCustomPort(t *testing.T) {
	base := map[string]string{"BACKEND_HOST": "127.0.0.1", "BACKEND_PORT": "9091", "WEB_HOST": "127.0.0.1", "WEB_PORT": "3091", "DATABASE_ENABLED": "false"}
	cfg, err := Parse(func(key string) string { return base[key] })
	if err != nil || Address(cfg) != "127.0.0.1:9091" || cfg.Database.Enabled || cfg.Auth.MinPasswordLength != 15 {
		t.Fatal(cfg.BackendPort, err)
	}
	for _, test := range []struct{ key, value string }{
		{"BACKEND_PORT", "0"}, {"BACKEND_PORT", "65536"}, {"BACKEND_PORT", "abc"}, {"DATABASE_ENABLED", "maybe"},
		{"PG_MIN_CONNS", "99"}, {"PG_MAX_IDLE_CONNS", "99"}, {"PG_MAX_IDLE_CONNS", "-1"}, {"PG_QUERY_TIMEOUT", "0s"}, {"AUTH_COOKIE_NAME", "invalid;name"}, {"AUTH_SESSION_TTL", "1s"},
		{"AUTH_ALLOWED_ORIGINS", "https://user:pass@example.com"}, {"DATABASE_ENABLED", "true"},
	} {
		t.Run(test.key+"/"+test.value, func(t *testing.T) {
			settings := make(map[string]string)
			for k, v := range base {
				settings[k] = v
			}
			settings[test.key] = test.value
			if _, err := Parse(func(k string) string { return settings[k] }); err == nil {
				t.Fatal("invalid settings accepted")
			}
		})
	}
	for k, v := range map[string]string{"DATABASE_ENABLED": "true", "PGHOST": "localhost", "PGPORT": "5432", "PGDATABASE": "test", "PGUSER": "test", "PGPASSWORD": "secret-do-not-print", "PGSSLMODE": "disable"} {
		base[k] = v
	}
	if _, err := Parse(func(k string) string { return base[k] }); err != nil {
		t.Fatal(err)
	}
	base["PGPORT"] = "wrong-secret-do-not-print"
	if _, err := Parse(func(k string) string { return base[k] }); err == nil || strings.Contains(err.Error(), "secret-do-not-print") {
		t.Fatal("missing/sensitive config error")
	}
}

func TestIdleConnectionConfigSupportsLegacyAndNewPrecedence(t *testing.T) {
	settings := map[string]string{"BACKEND_HOST": "127.0.0.1", "BACKEND_PORT": "8080", "PG_MIN_CONNS": "2"}
	get := func(key string) string { return settings[key] }
	cfg, err := Parse(get)
	if err != nil || cfg.Database.MaxIdleConns != 2 {
		t.Fatal("legacy pool configuration was not retained")
	}
	settings["PG_MAX_IDLE_CONNS"] = "0"
	cfg, err = Parse(get)
	if err != nil || cfg.Database.MaxIdleConns != 0 {
		t.Fatal("new idle configuration must override the legacy value, including zero")
	}
}
