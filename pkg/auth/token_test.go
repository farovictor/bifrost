package auth

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/farovictor/bifrost/pkg/logging"
	"github.com/rs/zerolog"
)

func TestLoadKeyWarnInvalid(t *testing.T) {
	oldKey := os.Getenv("BIFROST_SIGNING_KEY")
	oldMode := os.Getenv("BIFROST_MODE")
	oldDB := os.Getenv("BIFROST_DB")
	t.Cleanup(func() {
		os.Setenv("BIFROST_SIGNING_KEY", oldKey)
		os.Setenv("BIFROST_MODE", oldMode)
		os.Setenv("BIFROST_DB", oldDB)
	})
	os.Setenv("BIFROST_SIGNING_KEY", "invalid")
	os.Setenv("BIFROST_MODE", "")
	os.Setenv("BIFROST_DB", "postgres")

	var buf bytes.Buffer
	logging.Logger = zerolog.New(&buf)

	loadKey()

	if !strings.Contains(buf.String(), "invalid BIFROST_SIGNING_KEY") {
		t.Fatalf("warning not logged: %s", buf.String())
	}
}

func TestLoadKeyWarnEmpty(t *testing.T) {
	oldKey := os.Getenv("BIFROST_SIGNING_KEY")
	oldMode := os.Getenv("BIFROST_MODE")
	oldDB := os.Getenv("BIFROST_DB")
	t.Cleanup(func() {
		os.Setenv("BIFROST_SIGNING_KEY", oldKey)
		os.Setenv("BIFROST_MODE", oldMode)
		os.Setenv("BIFROST_DB", oldDB)
	})
	os.Unsetenv("BIFROST_SIGNING_KEY")
	os.Setenv("BIFROST_MODE", "")
	os.Setenv("BIFROST_DB", "postgres")

	var buf bytes.Buffer
	logging.Logger = zerolog.New(&buf)

	loadKey()

	if !strings.Contains(buf.String(), "BIFROST_SIGNING_KEY not set") {
		t.Fatalf("warning not logged: %s", buf.String())
	}
}
