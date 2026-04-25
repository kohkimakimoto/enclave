package command

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
)

// testSetupFakeXDGConfig sets XDG_CONFIG_HOME to a temp directory and returns
// the enclave config directory path. The original value is restored on cleanup.
func testSetupFakeXDGConfig(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	orig := os.Getenv("XDG_CONFIG_HOME")
	if err := os.Setenv("XDG_CONFIG_HOME", tmpDir); err != nil {
		t.Fatalf("failed to set XDG_CONFIG_HOME: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Setenv("XDG_CONFIG_HOME", orig); err != nil {
			t.Errorf("failed to restore XDG_CONFIG_HOME: %v", err)
		}
	})
	return filepath.Join(tmpDir, "enclave")
}

func TestInitUserCommand(t *testing.T) {
	t.Run("creates ~/.config/enclave/config.toml", func(t *testing.T) {
		configDir := testSetupFakeXDGConfig(t)

		buf := &bytes.Buffer{}
		cmd := InitUserCommand()
		cmd.Writer = buf

		if err := cmd.Run(context.Background(), []string{"init-user"}); err != nil {
			t.Fatalf("init-user failed: %v", err)
		}

		configFile := filepath.Join(configDir, "config.toml")
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			t.Errorf("expected config file to exist: %s", configFile)
		}
	})

	t.Run("fails if user config file already exists", func(t *testing.T) {
		configDir := testSetupFakeXDGConfig(t)

		if err := os.MkdirAll(configDir, 0o755); err != nil {
			t.Fatal(err)
		}
		existing := filepath.Join(configDir, "config.toml")
		if err := os.WriteFile(existing, []byte(""), 0o644); err != nil {
			t.Fatal(err)
		}

		cmd := InitUserCommand()
		cmd.Writer = &bytes.Buffer{}

		if err := cmd.Run(context.Background(), []string{"init-user"}); err == nil {
			t.Error("expected error when user config file already exists, got nil")
		}
	})

	t.Run("generated file uses flat schema (no sections)", func(t *testing.T) {
		configDir := testSetupFakeXDGConfig(t)

		cmd := InitUserCommand()
		cmd.Writer = &bytes.Buffer{}

		if err := cmd.Run(context.Background(), []string{"init-user"}); err != nil {
			t.Fatalf("init-user failed: %v", err)
		}

		content, err := os.ReadFile(filepath.Join(configDir, "config.toml"))
		if err != nil {
			t.Fatal(err)
		}
		got := string(content)
		assertContains(t, got, "sandbox_profile")
		assertContains(t, got, "unboxexec_allowed_commands")
		assertNotContains(t, got, "[sandbox]")
		assertNotContains(t, got, "[unboxexec]")
	})
}
