package command

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestInitLocalCommand(t *testing.T) {
	t.Run("creates .claude/sandbox.local.toml in the working directory", func(t *testing.T) {
		dir := testChdirTemp(t)

		cmd := InitLocalCommand()
		cmd.Writer = &bytes.Buffer{}

		if err := cmd.Run(context.Background(), []string{"init-local"}); err != nil {
			t.Fatalf("init-local failed: %v", err)
		}

		configFile := filepath.Join(dir, ".claude", "sandbox.local.toml")
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			t.Errorf("expected local config file to exist: %s", configFile)
		}
	})

	t.Run("fails if local config file already exists", func(t *testing.T) {
		dir := testChdirTemp(t)

		// Pre-create the local config file
		if err := os.MkdirAll(filepath.Join(dir, ".claude"), 0o755); err != nil {
			t.Fatal(err)
		}
		existing := filepath.Join(dir, ".claude", "sandbox.local.toml")
		if err := os.WriteFile(existing, []byte(""), 0o644); err != nil {
			t.Fatal(err)
		}

		cmd := InitLocalCommand()
		cmd.Writer = &bytes.Buffer{}

		if err := cmd.Run(context.Background(), []string{"init-local"}); err == nil {
			t.Error("expected error when local config file already exists, got nil")
		}
	})
}
