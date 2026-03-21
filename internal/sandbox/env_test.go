package sandbox

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestGetWorkdir(t *testing.T) {
	t.Run("uses configWorkdir when set", func(t *testing.T) {
		configured := "/configured/path"
		got := GetWorkdir(configured)
		if got != configured {
			t.Errorf("expected %q, got %q", configured, got)
		}
	})

	t.Run("falls back to current directory when empty", func(t *testing.T) {
		dir := testChdirTemp(t)
		// Resolve symlinks to handle macOS /var -> /private/var
		realDir, err := filepath.EvalSymlinks(dir)
		if err != nil {
			t.Fatalf("failed to eval symlinks: %v", err)
		}
		got := GetWorkdir("")
		realGot, err := filepath.EvalSymlinks(got)
		if err != nil {
			t.Fatalf("failed to eval symlinks of result: %v", err)
		}
		if realGot != realDir {
			t.Errorf("expected %q, got %q", realDir, realGot)
		}
	})
}

func TestGetClaudeBin(t *testing.T) {
	t.Run("uses configClaudeBin when set", func(t *testing.T) {
		configured := "/custom/bin/claude"
		got := GetClaudeBin(configured)
		if got != configured {
			t.Errorf("expected %q, got %q", configured, got)
		}
	})

	t.Run("returns ~/.claude/local/claude when it exists", func(t *testing.T) {
		fakeHome := testSetupFakeHome(t)
		// Clear PATH so exec.LookPath("claude") won't find a system claude
		origPath := os.Getenv("PATH")
		if err := os.Setenv("PATH", ""); err != nil {
			t.Fatalf("failed to clear PATH: %v", err)
		}
		t.Cleanup(func() {
			_ = os.Setenv("PATH", origPath)
		})

		localClaude := filepath.Join(fakeHome, ".claude", "local", "claude")
		if err := os.MkdirAll(filepath.Dir(localClaude), 0o755); err != nil {
			t.Fatalf("failed to create dirs: %v", err)
		}
		if err := os.WriteFile(localClaude, []byte("#!/bin/sh"), 0o755); err != nil {
			t.Fatalf("failed to create fake claude binary: %v", err)
		}

		got := GetClaudeBin("")
		if got != localClaude {
			t.Errorf("expected %q, got %q", localClaude, got)
		}
	})

	t.Run("falls back to 'claude' when nothing found", func(t *testing.T) {
		// Set HOME to a dir without ~/.claude/local/claude, and PATH to empty
		testSetupFakeHome(t)
		origPath := os.Getenv("PATH")
		if err := os.Setenv("PATH", ""); err != nil {
			t.Fatalf("failed to clear PATH: %v", err)
		}
		t.Cleanup(func() {
			_ = os.Setenv("PATH", origPath)
		})

		got := GetClaudeBin("")
		if got != "claude" {
			t.Errorf("expected %q, got %q", "claude", got)
		}
	})
}

func TestSocketPath(t *testing.T) {
	got := SocketPath()
	expected := filepath.Join(os.TempDir(), fmt.Sprintf("claude-sandbox-unboxexec-%d.sock", os.Getpid()))
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}
