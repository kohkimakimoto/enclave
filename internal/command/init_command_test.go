package command

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestInitCommand(t *testing.T) {
	t.Run("creates enclave.toml in the working directory", func(t *testing.T) {
		dir := testChdirTemp(t)

		cmd := InitCommand()
		cmd.Writer = &bytes.Buffer{}

		if err := cmd.Run(context.Background(), []string{"init"}); err != nil {
			t.Fatalf("init failed: %v", err)
		}

		configFile := filepath.Join(dir, "enclave.toml")
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			t.Errorf("expected config file to exist: %s", configFile)
		}
	})

	t.Run("fails if config file already exists", func(t *testing.T) {
		dir := testChdirTemp(t)

		existing := filepath.Join(dir, "enclave.toml")
		if err := os.WriteFile(existing, []byte(""), 0o644); err != nil {
			t.Fatal(err)
		}

		cmd := InitCommand()
		cmd.Writer = &bytes.Buffer{}

		if err := cmd.Run(context.Background(), []string{"init"}); err == nil {
			t.Error("expected error when config file already exists, got nil")
		}
	})

	t.Run("generated file uses flat schema (no sections)", func(t *testing.T) {
		dir := testChdirTemp(t)

		cmd := InitCommand()
		cmd.Writer = &bytes.Buffer{}

		if err := cmd.Run(context.Background(), []string{"init"}); err != nil {
			t.Fatalf("init failed: %v", err)
		}

		content, err := os.ReadFile(filepath.Join(dir, "enclave.toml"))
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
