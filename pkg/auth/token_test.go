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
	old := os.Getenv("BIFROST_SIGNING_KEY")
	t.Cleanup(func() {
		os.Setenv("BIFROST_SIGNING_KEY", old)
	})
	os.Setenv("BIFROST_SIGNING_KEY", "invalid")

	var buf bytes.Buffer
	logging.Logger = zerolog.New(&buf)

	loadKey()

	if !strings.Contains(buf.String(), "invalid BIFROST_SIGNING_KEY") {
		t.Fatalf("warning not logged: %s", buf.String())
	}
}

func TestLoadKeyWarnEmpty(t *testing.T) {
	old := os.Getenv("BIFROST_SIGNING_KEY")
	t.Cleanup(func() {
		os.Setenv("BIFROST_SIGNING_KEY", old)
	})
	os.Unsetenv("BIFROST_SIGNING_KEY")

	var buf bytes.Buffer
	logging.Logger = zerolog.New(&buf)

	loadKey()

	if !strings.Contains(buf.String(), "BIFROST_SIGNING_KEY not set") {
		t.Fatalf("warning not logged: %s", buf.String())
	}
}
