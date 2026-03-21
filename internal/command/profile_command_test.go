package command

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestProfileCommand(t *testing.T) {
	// TOML multiline literal strings (''') strip the leading newline,
	// so the actual profile value starts on the line after the opening '''.
	const userProfile = "(version 1)\n(allow default)\n(deny file-write*)\n(allow file-write* (subpath \"/tmp\"))\n"
	const projectProfile = "(version 1)\n(allow default)\n(deny file-write*)\n"
	const localProfile = "(version 1)\n(allow default)\n"

	tomlProfile := func(profile string) string {
		return "[sandbox]\nprofile = '''\n" + profile + "'''\n"
	}

	t.Run("outputs profile from user config", func(t *testing.T) {
		fakeHome := testSetupFakeHome(t)
		testChdirTemp(t)

		// Write user-level config with a custom profile
		if err := os.MkdirAll(filepath.Join(fakeHome, ".claude"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(fakeHome, ".claude", "sandbox.toml"), []byte(tomlProfile(userProfile)), 0o644); err != nil {
			t.Fatal(err)
		}

		buf := &bytes.Buffer{}
		cmd := ProfileCommand()
		cmd.Writer = buf

		if err := cmd.Run(context.Background(), []string{"profile"}); err != nil {
			t.Fatalf("profile failed: %v", err)
		}

		if got := buf.String(); got != userProfile {
			t.Errorf("expected profile:\n%q\ngot:\n%q", userProfile, got)
		}
	})

	t.Run("project config overrides user config profile", func(t *testing.T) {
		fakeHome := testSetupFakeHome(t)
		dir := testChdirTemp(t)

		// User config sets one profile
		if err := os.MkdirAll(filepath.Join(fakeHome, ".claude"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(fakeHome, ".claude", "sandbox.toml"), []byte(tomlProfile(userProfile)), 0o644); err != nil {
			t.Fatal(err)
		}

		// Project config overrides with a different profile
		if err := os.MkdirAll(filepath.Join(dir, ".claude"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, ".claude", "sandbox.toml"), []byte(tomlProfile(projectProfile)), 0o644); err != nil {
			t.Fatal(err)
		}

		buf := &bytes.Buffer{}
		cmd := ProfileCommand()
		cmd.Writer = buf

		if err := cmd.Run(context.Background(), []string{"profile"}); err != nil {
			t.Fatalf("profile failed: %v", err)
		}

		if got := buf.String(); got != projectProfile {
			t.Errorf("expected profile from project config:\n%q\ngot:\n%q", projectProfile, got)
		}
	})

	t.Run("local config overrides project config profile", func(t *testing.T) {
		testSetupFakeHome(t)
		dir := testChdirTemp(t)

		if err := os.MkdirAll(filepath.Join(dir, ".claude"), 0o755); err != nil {
			t.Fatal(err)
		}

		// Project config sets one profile
		if err := os.WriteFile(filepath.Join(dir, ".claude", "sandbox.toml"), []byte(tomlProfile(projectProfile)), 0o644); err != nil {
			t.Fatal(err)
		}

		// Local config overrides with a different profile
		if err := os.WriteFile(filepath.Join(dir, ".claude", "sandbox.local.toml"), []byte(tomlProfile(localProfile)), 0o644); err != nil {
			t.Fatal(err)
		}

		buf := &bytes.Buffer{}
		cmd := ProfileCommand()
		cmd.Writer = buf

		if err := cmd.Run(context.Background(), []string{"profile"}); err != nil {
			t.Fatalf("profile failed: %v", err)
		}

		if got := buf.String(); got != localProfile {
			t.Errorf("expected profile from local config:\n%q\ngot:\n%q", localProfile, got)
		}
	})
}
