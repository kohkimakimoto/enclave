package command

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestInitLocalCommand(t *testing.T) {
	t.Run("creates enclave.local.toml in the working directory", func(t *testing.T) {
		dir := testChdirTemp(t)

		cmd := InitLocalCommand()
		cmd.Writer = &bytes.Buffer{}

		if err := cmd.Run(context.Background(), []string{"init-local"}); err != nil {
			t.Fatalf("init-local failed: %v", err)
		}

		configFile := filepath.Join(dir, "enclave.local.toml")
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			t.Errorf("expected local config file to exist: %s", configFile)
		}
	})

	t.Run("fails if local config file already exists", func(t *testing.T) {
		dir := testChdirTemp(t)

		existing := filepath.Join(dir, "enclave.local.toml")
		if err := os.WriteFile(existing, []byte(""), 0o644); err != nil {
			t.Fatal(err)
		}

		cmd := InitLocalCommand()
		cmd.Writer = &bytes.Buffer{}

		if err := cmd.Run(context.Background(), []string{"init-local"}); err == nil {
			t.Error("expected error when local config file already exists, got nil")
		}
	})

	t.Run("generated file uses flat schema (no sections)", func(t *testing.T) {
		dir := testChdirTemp(t)

		cmd := InitLocalCommand()
		cmd.Writer = &bytes.Buffer{}

		if err := cmd.Run(context.Background(), []string{"init-local"}); err != nil {
			t.Fatalf("init-local failed: %v", err)
		}

		content, err := os.ReadFile(filepath.Join(dir, "enclave.local.toml"))
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
