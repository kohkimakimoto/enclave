package command

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestInitUserCommand(t *testing.T) {
	t.Run("creates ~/.claude/sandbox.toml", func(t *testing.T) {
		fakeHome := testSetupFakeHome(t)

		buf := &bytes.Buffer{}
		cmd := InitUserCommand()
		cmd.Writer = buf

		if err := cmd.Run(context.Background(), []string{"init-user"}); err != nil {
			t.Fatalf("init-user failed: %v", err)
		}

		configFile := filepath.Join(fakeHome, ".claude", "sandbox.toml")
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			t.Errorf("expected config file to exist: %s", configFile)
		}
	})

	t.Run("fails if user config file already exists", func(t *testing.T) {
		fakeHome := testSetupFakeHome(t)

		// Pre-create the config file
		existing := filepath.Join(fakeHome, ".claude", "sandbox.toml")
		if err := os.MkdirAll(filepath.Dir(existing), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(existing, []byte(""), 0o644); err != nil {
			t.Fatal(err)
		}

		cmd := InitUserCommand()
		cmd.Writer = &bytes.Buffer{}

		if err := cmd.Run(context.Background(), []string{"init-user"}); err == nil {
			t.Error("expected error when user config file already exists, got nil")
		}
	})
}
