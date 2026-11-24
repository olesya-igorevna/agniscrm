package config

import (
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("DB_TIMEOUT", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Addr != ":8080" {
		t.Errorf("Addr = %q, want :8080", cfg.Addr)
	}
	if cfg.DatabaseURL != "postgres://example" {
		t.Errorf("DatabaseURL = %q", cfg.DatabaseURL)
	}
	if cfg.DBTimeout != 3*time.Second {
		t.Errorf("DBTimeout = %v, want 3s", cfg.DBTimeout)
	}
}

func TestLoadCustomValues(t *testing.T) {
	t.Setenv("HTTP_ADDR", ":9090")
	t.Setenv("DATABASE_URL", "postgres://user:pw@host/db")
	t.Setenv("DB_TIMEOUT", "500ms")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Addr != ":9090" {
		t.Errorf("Addr = %q, want :9090", cfg.Addr)
	}
	if cfg.DBTimeout != 500*time.Millisecond {
		t.Errorf("DBTimeout = %v, want 500ms", cfg.DBTimeout)
	}
}

func TestLoadMissingDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	if _, err := Load(); err == nil {
		t.Fatal("Load: expected error for missing DATABASE_URL")
	}
}

func TestLoadInvalidTimeout(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("DB_TIMEOUT", "не-время")
	if _, err := Load(); err == nil {
		t.Fatal("Load: expected error for invalid DB_TIMEOUT")
	}
}
