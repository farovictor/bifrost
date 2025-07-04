package config

import (
	"os"
	"strconv"
)

// ServerPort returns the port the HTTP server should listen on.
// It reads the BIFROST_PORT environment variable and defaults to "3333" if unset.
// A leading colon is added if missing.
func ServerPort() string {
	port := os.Getenv("BIFROST_PORT")
	if port == "" {
		port = "3333"
	}
	if len(port) > 0 && port[0] != ':' {
		return ":" + port
	}
	return port
}

// RedisAddr returns the address of the redis server.
// It reads the REDIS_ADDR environment variable and defaults to "localhost:6379".
func RedisAddr() string {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	return addr
}

// RedisPassword returns the password for redis if set via REDIS_PASSWORD.
func RedisPassword() string {
	return os.Getenv("REDIS_PASSWORD")
}

// RedisDB returns the database index for redis, defaulting to 0.
func RedisDB() int {
	if v := os.Getenv("REDIS_DB"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return 0
}

// RedisProtocol returns the protocol version for redis connections, defaulting to 3.
func RedisProtocol() int {
	if v := os.Getenv("REDIS_PROTOCOL"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return 3
}

// MetricsEnabled determines whether Prometheus metrics should be exposed.
// It checks the BIFROST_ENABLE_METRICS environment variable for a truthy value.
func MetricsEnabled() bool {
	switch os.Getenv("BIFROST_ENABLE_METRICS") {
	case "1", "true", "TRUE", "True", "yes", "YES":
		return true
	default:
		return false
	}
}

// PostgresDSN returns the DSN string for connecting to Postgres.
// It reads the POSTGRES_DSN environment variable and may be empty if unset.
func PostgresDSN() string {
	return os.Getenv("POSTGRES_DSN")
}

// AdminAPIKey returns the API key for the initial admin user. When
// BIFROST_ADMIN_API_KEY is empty a random key should be generated by the caller.
func AdminAPIKey() string {
	return os.Getenv("BIFROST_ADMIN_API_KEY")
}

// AdminName returns the name to use for the initial admin user.
// It reads BIFROST_ADMIN_NAME and defaults to "Admin" when unset.
func AdminName() string {
	name := os.Getenv("BIFROST_ADMIN_NAME")
	if name == "" {
		name = "Admin"
	}
	return name
}

// AdminEmail returns the email to use for the initial admin user.
// It reads BIFROST_ADMIN_EMAIL and defaults to "admin@example.com" when unset.
func AdminEmail() string {
	email := os.Getenv("BIFROST_ADMIN_EMAIL")
	if email == "" {
		email = "admin@example.com"
	}
	return email
}

// AdminOrgName returns the name for the initial admin organization.
// It reads BIFROST_ADMIN_ORG_NAME and defaults to "Admin" when unset.
func AdminOrgName() string {
	name := os.Getenv("BIFROST_ADMIN_ORG_NAME")
	if name == "" {
		name = "Admin"
	}
	return name
}

// AdminOrgDomain returns the domain for the initial admin organization.
// It reads BIFROST_ADMIN_ORG_DOMAIN and defaults to "example.com" when unset.
func AdminOrgDomain() string {
	domain := os.Getenv("BIFROST_ADMIN_ORG_DOMAIN")
	if domain == "" {
		domain = "example.com"
	}
	return domain
}

// AdminOrgEmail returns the contact email for the initial admin organization.
// It reads BIFROST_ADMIN_ORG_EMAIL and defaults to "admin@example.com" when unset.
func AdminOrgEmail() string {
	email := os.Getenv("BIFROST_ADMIN_ORG_EMAIL")
	if email == "" {
		email = "admin@example.com"
	}
	return email
}

// AdminRole returns the membership role for the admin user.
// It reads BIFROST_ADMIN_ROLE and defaults to "owner" when unset.
func AdminRole() string {
	role := os.Getenv("BIFROST_ADMIN_ROLE")
	if role == "" {
		role = "owner"
	}
	return role
}
