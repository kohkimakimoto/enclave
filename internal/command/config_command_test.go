package command

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigCommand(t *testing.T) {
	t.Run("no config files", func(t *testing.T) {
		testSetupFakeXDGConfig(t)
		testChdirTemp(t)

		buf := &bytes.Buffer{}
		cmd := ConfigCommand()
		cmd.Writer = buf

		if err := cmd.Run(context.Background(), []string{"config"}); err != nil {
			t.Fatalf("config failed: %v", err)
		}

		got := buf.String()
		assertContains(t, got, "# Loaded config files:")
		assertContains(t, got, "#   user:    (none)")
		assertContains(t, got, "#   project: (none)")
		assertContains(t, got, "#   local:   (none)")
		assertContains(t, got, `sandbox_profile = ""`)
		assertContains(t, got, "unboxexec_allowed_commands = []")
	})

	t.Run("user config with unboxexec_allowed_commands", func(t *testing.T) {
		configDir := testSetupFakeXDGConfig(t)
		testChdirTemp(t)

		if err := os.MkdirAll(configDir, 0o755); err != nil {
			t.Fatal(err)
		}
		toml := "unboxexec_allowed_commands = [\"^playwright-cli\", \"^my-tool\"]\n"
		if err := os.WriteFile(filepath.Join(configDir, "config.toml"), []byte(toml), 0o644); err != nil {
			t.Fatal(err)
		}

		buf := &bytes.Buffer{}
		cmd := ConfigCommand()
		cmd.Writer = buf

		if err := cmd.Run(context.Background(), []string{"config"}); err != nil {
			t.Fatalf("config failed: %v", err)
		}

		got := buf.String()
		assertContains(t, got, configDir)
		assertContains(t, got, `"^playwright-cli",`)
		assertContains(t, got, `"^my-tool",`)
	})

	t.Run("sandbox_profile without triple-single-quote", func(t *testing.T) {
		configDir := testSetupFakeXDGConfig(t)
		testChdirTemp(t)

		profile := "(version 1)\n(allow default)\n(deny file-write*)\n"
		toml := "sandbox_profile = '''\n" + profile + "'''\n"
		if err := os.MkdirAll(configDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(configDir, "config.toml"), []byte(toml), 0o644); err != nil {
			t.Fatal(err)
		}

		buf := &bytes.Buffer{}
		cmd := ConfigCommand()
		cmd.Writer = buf

		if err := cmd.Run(context.Background(), []string{"config"}); err != nil {
			t.Fatalf("config failed: %v", err)
		}

		got := buf.String()
		assertContains(t, got, "sandbox_profile = '''\n")
		assertContains(t, got, "(version 1)")
		assertContains(t, got, "(allow default)")
		assertContains(t, got, "'''")
	})

	t.Run("sandbox_profile containing triple-single-quote falls back to basic multiline", func(t *testing.T) {
		configDir := testSetupFakeXDGConfig(t)
		testChdirTemp(t)

		profileContent := "line1\n'''\nline3\n"
		toml := "sandbox_profile = \"line1\\n'''\\nline3\\n\"\n"
		if err := os.MkdirAll(configDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(configDir, "config.toml"), []byte(toml), 0o644); err != nil {
			t.Fatal(err)
		}

		buf := &bytes.Buffer{}
		cmd := ConfigCommand()
		cmd.Writer = buf

		if err := cmd.Run(context.Background(), []string{"config"}); err != nil {
			t.Fatalf("config failed: %v", err)
		}

		got := buf.String()
		assertContains(t, got, "sandbox_profile = \"\"\"\n")
		assertContains(t, got, profileContent)
		assertContains(t, got, "\"\"\"")
	})

	t.Run("config files from all three scopes are reported", func(t *testing.T) {
		configDir := testSetupFakeXDGConfig(t)
		dir := testChdirTemp(t)

		if err := os.MkdirAll(configDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(configDir, "config.toml"), []byte(""), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "enclave.toml"), []byte(""), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "enclave.local.toml"), []byte(""), 0o644); err != nil {
			t.Fatal(err)
		}

		buf := &bytes.Buffer{}
		cmd := ConfigCommand()
		cmd.Writer = buf

		if err := cmd.Run(context.Background(), []string{"config"}); err != nil {
			t.Fatalf("config failed: %v", err)
		}

		got := buf.String()
		assertContains(t, got, configDir)
		assertContains(t, got, dir)
		assertNotContains(t, got, "(none)")
	})
}

func TestTomlMultilineString(t *testing.T) {
	t.Run("no triple-single-quote uses literal multiline", func(t *testing.T) {
		s := "(version 1)\n(allow default)\n"
		got := tomlMultilineString(s)
		if !strings.HasPrefix(got, "'''") {
			t.Errorf("expected literal multiline, got: %s", got)
		}
		assertContains(t, got, s)
	})

	t.Run("contains triple-single-quote uses basic multiline", func(t *testing.T) {
		s := "line1\n'''\nline3\n"
		got := tomlMultilineString(s)
		if !strings.HasPrefix(got, `"""`) {
			t.Errorf("expected basic multiline fallback, got: %s", got)
		}
		assertContains(t, got, "line1")
		assertContains(t, got, "line3")
	})

	t.Run("content without trailing newline gets one added", func(t *testing.T) {
		s := "no newline"
		got := tomlMultilineString(s)
		assertContains(t, got, "no newline\n")
	})

	t.Run("triple-double-quote in content is escaped in basic multiline", func(t *testing.T) {
		s := "line1\n'''\n\"\"\"\nline4\n"
		got := tomlMultilineString(s)
		if !strings.HasPrefix(got, `"""`) {
			t.Errorf("expected basic multiline fallback, got: %s", got)
		}
		assertContains(t, got, `""\"`)
	})
}

func assertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Errorf("expected output to contain %q\ngot:\n%s", substr, s)
	}
}

func assertNotContains(t *testing.T, s, substr string) {
	t.Helper()
	if strings.Contains(s, substr) {
		t.Errorf("expected output NOT to contain %q\ngot:\n%s", substr, s)
	}
}
