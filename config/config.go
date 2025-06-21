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
