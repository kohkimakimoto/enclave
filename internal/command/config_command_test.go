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
		testSetupFakeHome(t)
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
		assertContains(t, got, `workdir    = ""`)
		assertContains(t, got, `claude_bin = ""`)
		assertContains(t, got, `profile    = ""`)
		assertContains(t, got, "allowed_commands = []")
	})

	t.Run("user config with allowed_commands", func(t *testing.T) {
		fakeHome := testSetupFakeHome(t)
		testChdirTemp(t)

		toml := "[unboxexec]\nallowed_commands = [\"^playwright-cli\", \"^my-tool\"]\n"
		if err := os.MkdirAll(filepath.Join(fakeHome, ".claude"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(fakeHome, ".claude", "sandbox.toml"), []byte(toml), 0o644); err != nil {
			t.Fatal(err)
		}

		buf := &bytes.Buffer{}
		cmd := ConfigCommand()
		cmd.Writer = buf

		if err := cmd.Run(context.Background(), []string{"config"}); err != nil {
			t.Fatalf("config failed: %v", err)
		}

		got := buf.String()
		assertContains(t, got, fakeHome)
		assertContains(t, got, `"^playwright-cli",`)
		assertContains(t, got, `"^my-tool",`)
	})

	t.Run("profile without triple-single-quote", func(t *testing.T) {
		fakeHome := testSetupFakeHome(t)
		testChdirTemp(t)

		profile := "(version 1)\n(allow default)\n(deny file-write*)\n"
		toml := "[sandbox]\nprofile = '''\n" + profile + "'''\n"
		if err := os.MkdirAll(filepath.Join(fakeHome, ".claude"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(fakeHome, ".claude", "sandbox.toml"), []byte(toml), 0o644); err != nil {
			t.Fatal(err)
		}

		buf := &bytes.Buffer{}
		cmd := ConfigCommand()
		cmd.Writer = buf

		if err := cmd.Run(context.Background(), []string{"config"}); err != nil {
			t.Fatalf("config failed: %v", err)
		}

		got := buf.String()
		// Should use literal multiline string
		assertContains(t, got, "profile    = '''\n")
		assertContains(t, got, "(version 1)")
		assertContains(t, got, "(allow default)")
		assertContains(t, got, "'''")
	})

	t.Run("profile containing triple-single-quote falls back to basic multiline", func(t *testing.T) {
		fakeHome := testSetupFakeHome(t)
		testChdirTemp(t)

		// Write the profile value directly via basic string in TOML to embed '''
		profileContent := "line1\n'''\nline3\n"
		toml := "[sandbox]\nprofile = \"line1\\n'''\\nline3\\n\"\n"
		if err := os.MkdirAll(filepath.Join(fakeHome, ".claude"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(fakeHome, ".claude", "sandbox.toml"), []byte(toml), 0o644); err != nil {
			t.Fatal(err)
		}

		buf := &bytes.Buffer{}
		cmd := ConfigCommand()
		cmd.Writer = buf

		if err := cmd.Run(context.Background(), []string{"config"}); err != nil {
			t.Fatalf("config failed: %v", err)
		}

		got := buf.String()
		// Should fall back to basic multiline string (""")
		assertContains(t, got, "profile    = \"\"\"\n")
		assertContains(t, got, profileContent)
		assertContains(t, got, "\"\"\"")
	})

	t.Run("config files from all three scopes are reported", func(t *testing.T) {
		fakeHome := testSetupFakeHome(t)
		dir := testChdirTemp(t)

		if err := os.MkdirAll(filepath.Join(fakeHome, ".claude"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(fakeHome, ".claude", "sandbox.toml"), []byte(""), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(dir, ".claude"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, ".claude", "sandbox.toml"), []byte(""), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, ".claude", "sandbox.local.toml"), []byte(""), 0o644); err != nil {
			t.Fatal(err)
		}

		buf := &bytes.Buffer{}
		cmd := ConfigCommand()
		cmd.Writer = buf

		if err := cmd.Run(context.Background(), []string{"config"}); err != nil {
			t.Fatalf("config failed: %v", err)
		}

		got := buf.String()
		assertContains(t, got, fakeHome)
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
		// """ should be escaped as ""\"; verify the escaped form is present
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
