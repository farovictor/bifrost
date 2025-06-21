package logging

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Logger is the application-wide logger instance.
var Logger zerolog.Logger

// Setup initializes the global logger based on environment variables.
// BIFROST_LOG_LEVEL controls the log level (debug, info, warn, error).
// BIFROST_LOG_FORMAT controls the output format ("json" or "console").
func Setup() {
	lvl := strings.ToLower(os.Getenv("BIFROST_LOG_LEVEL"))
	if lvl == "" {
		lvl = "info"
	}
	level, err := zerolog.ParseLevel(lvl)
	if err != nil {
		level = zerolog.InfoLevel
	}

	var w io.Writer = os.Stdout
	format := strings.ToLower(os.Getenv("BIFROST_LOG_FORMAT"))
	if format == "console" {
		w = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	}

	Logger = zerolog.New(w).Level(level).With().Timestamp().Logger()
	log.Logger = Logger
}
