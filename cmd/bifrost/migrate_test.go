package main

import (
	"bytes"
	"testing"
)

func TestMigrateSkipsSQLite(t *testing.T) {
	t.Setenv("BIFROST_DB", "sqlite")
	t.Setenv("DATABASE_DSN", "file::memory:?cache=shared")
	buf := &bytes.Buffer{}
	migrateCmd.SetOut(buf)
	migrateCmd.SetErr(buf)
	if err := migrateCmd.RunE(migrateCmd, []string{}); err != nil {
		t.Fatalf("migrate returned error: %v", err)
	}
	got := buf.String()
	want := "sqlite selected; skipping migrations\n"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
