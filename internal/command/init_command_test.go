package command

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestInitCommand(t *testing.T) {
	t.Run("creates .claude/sandbox.toml in the working directory", func(t *testing.T) {
		dir := testChdirTemp(t)

		cmd := InitCommand()
		cmd.Writer = &bytes.Buffer{}

		if err := cmd.Run(context.Background(), []string{"init"}); err != nil {
			t.Fatalf("init failed: %v", err)
		}

		configFile := filepath.Join(dir, ".claude", "sandbox.toml")
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			t.Errorf("expected config file to exist: %s", configFile)
		}
	})

	t.Run("fails if config file already exists", func(t *testing.T) {
		dir := testChdirTemp(t)

		// Pre-create the config file
		if err := os.MkdirAll(filepath.Join(dir, ".claude"), 0o755); err != nil {
			t.Fatal(err)
		}
		existing := filepath.Join(dir, ".claude", "sandbox.toml")
		if err := os.WriteFile(existing, []byte(""), 0o644); err != nil {
			t.Fatal(err)
		}

		cmd := InitCommand()
		cmd.Writer = &bytes.Buffer{}

		if err := cmd.Run(context.Background(), []string{"init"}); err == nil {
			t.Error("expected error when config file already exists, got nil")
		}
	})
}
