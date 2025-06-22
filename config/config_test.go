package config

import (
	"testing"
)

func TestServerPort(t *testing.T) {
	t.Setenv("BIFROST_PORT", "8080")
	if got := ServerPort(); got != ":8080" {
		t.Errorf("expected :8080, got %s", got)
	}
	t.Setenv("BIFROST_PORT", "")
	if got := ServerPort(); got != ":3333" {
		t.Errorf("expected :3333, got %s", got)
	}
}

func TestRedisAddr(t *testing.T) {
	t.Setenv("REDIS_ADDR", "redis:1234")
	if got := RedisAddr(); got != "redis:1234" {
		t.Errorf("expected redis:1234, got %s", got)
	}
	t.Setenv("REDIS_ADDR", "")
	if got := RedisAddr(); got != "localhost:6379" {
		t.Errorf("expected localhost:6379, got %s", got)
	}
}

func TestRedisPassword(t *testing.T) {
	t.Setenv("REDIS_PASSWORD", "secret")
	if got := RedisPassword(); got != "secret" {
		t.Errorf("expected secret, got %s", got)
	}
	t.Setenv("REDIS_PASSWORD", "")
	if got := RedisPassword(); got != "" {
		t.Errorf("expected empty string, got %s", got)
	}
}

func TestRedisDB(t *testing.T) {
	t.Setenv("REDIS_DB", "5")
	if got := RedisDB(); got != 5 {
		t.Errorf("expected 5, got %d", got)
	}
	t.Setenv("REDIS_DB", "")
	if got := RedisDB(); got != 0 {
		t.Errorf("expected 0, got %d", got)
	}
}

func TestRedisProtocol(t *testing.T) {
	t.Setenv("REDIS_PROTOCOL", "2")
	if got := RedisProtocol(); got != 2 {
		t.Errorf("expected 2, got %d", got)
	}
	t.Setenv("REDIS_PROTOCOL", "")
	if got := RedisProtocol(); got != 3 {
		t.Errorf("expected 3, got %d", got)
	}
}

func TestMetricsEnabled(t *testing.T) {
	t.Setenv("BIFROST_ENABLE_METRICS", "true")
	if !MetricsEnabled() {
		t.Errorf("expected metrics enabled")
	}
	t.Setenv("BIFROST_ENABLE_METRICS", "")
	if MetricsEnabled() {
		t.Errorf("expected metrics disabled by default")
	}
}

func TestPostgresDSN(t *testing.T) {
	t.Setenv("POSTGRES_DSN", "postgres://user:pass@localhost/db")
	if got := PostgresDSN(); got != "postgres://user:pass@localhost/db" {
		t.Errorf("expected postgres://user:pass@localhost/db, got %s", got)
	}
	t.Setenv("POSTGRES_DSN", "")
	if got := PostgresDSN(); got != "" {
		t.Errorf("expected empty string, got %s", got)
	}
}

func TestAdminAPIKey(t *testing.T) {
	t.Setenv("BIFROST_ADMIN_API_KEY", "apikey")
	if got := AdminAPIKey(); got != "apikey" {
		t.Errorf("expected apikey, got %s", got)
	}
	t.Setenv("BIFROST_ADMIN_API_KEY", "")
	if got := AdminAPIKey(); got != "" {
		t.Errorf("expected empty string, got %s", got)
	}
}

func TestAdminName(t *testing.T) {
	t.Setenv("BIFROST_ADMIN_NAME", "Root")
	if got := AdminName(); got != "Root" {
		t.Errorf("expected Root, got %s", got)
	}
	t.Setenv("BIFROST_ADMIN_NAME", "")
	if got := AdminName(); got != "Admin" {
		t.Errorf("expected Admin, got %s", got)
	}
}

func TestAdminEmail(t *testing.T) {
	t.Setenv("BIFROST_ADMIN_EMAIL", "root@example.com")
	if got := AdminEmail(); got != "root@example.com" {
		t.Errorf("expected root@example.com, got %s", got)
	}
	t.Setenv("BIFROST_ADMIN_EMAIL", "")
	if got := AdminEmail(); got != "admin@example.com" {
		t.Errorf("expected admin@example.com, got %s", got)
	}
}

func TestAdminOrgName(t *testing.T) {
	t.Setenv("BIFROST_ADMIN_ORG_NAME", "RootOrg")
	if got := AdminOrgName(); got != "RootOrg" {
		t.Errorf("expected RootOrg, got %s", got)
	}
	t.Setenv("BIFROST_ADMIN_ORG_NAME", "")
	if got := AdminOrgName(); got != "Admin" {
		t.Errorf("expected Admin, got %s", got)
	}
}

func TestAdminOrgDomain(t *testing.T) {
	t.Setenv("BIFROST_ADMIN_ORG_DOMAIN", "example.org")
	if got := AdminOrgDomain(); got != "example.org" {
		t.Errorf("expected example.org, got %s", got)
	}
	t.Setenv("BIFROST_ADMIN_ORG_DOMAIN", "")
	if got := AdminOrgDomain(); got != "example.com" {
		t.Errorf("expected example.com, got %s", got)
	}
}

func TestAdminOrgEmail(t *testing.T) {
	t.Setenv("BIFROST_ADMIN_ORG_EMAIL", "root@example.org")
	if got := AdminOrgEmail(); got != "root@example.org" {
		t.Errorf("expected root@example.org, got %s", got)
	}
	t.Setenv("BIFROST_ADMIN_ORG_EMAIL", "")
	if got := AdminOrgEmail(); got != "admin@example.com" {
		t.Errorf("expected admin@example.com, got %s", got)
	}
}

func TestAdminRole(t *testing.T) {
	t.Setenv("BIFROST_ADMIN_ROLE", "member")
	if got := AdminRole(); got != "member" {
		t.Errorf("expected member, got %s", got)
	}
	t.Setenv("BIFROST_ADMIN_ROLE", "")
	if got := AdminRole(); got != "owner" {
		t.Errorf("expected owner, got %s", got)
	}
}
