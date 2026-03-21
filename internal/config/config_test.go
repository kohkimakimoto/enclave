package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadValidConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	content := `
[unboxexec]
allowed_commands = [
    "^playwright",
    "^echo hello",
]
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile failed: %v", err)
	}

	if len(cfg.Unboxexec.AllowedCommands) != 2 {
		t.Fatalf("expected 2 allowed_commands, got %d", len(cfg.Unboxexec.AllowedCommands))
	}
	if cfg.Unboxexec.AllowedCommands[0] != "^playwright" {
		t.Errorf("expected %q, got %q", "^playwright", cfg.Unboxexec.AllowedCommands[0])
	}
	if cfg.Unboxexec.AllowedCommands[1] != "^echo hello" {
		t.Errorf("expected %q, got %q", "^echo hello", cfg.Unboxexec.AllowedCommands[1])
	}
}

func TestLoadConfigWithSandboxSection(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	content := `
[sandbox]
profile = "(version 1)\n(allow default)"
workdir = "/tmp/myworkdir"
claude_bin = "/usr/local/bin/claude"

[unboxexec]
allowed_commands = [
    "^playwright-cli",
]
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile failed: %v", err)
	}

	if cfg.Sandbox.Profile != "(version 1)\n(allow default)" {
		t.Errorf("expected profile %q, got %q", "(version 1)\n(allow default)", cfg.Sandbox.Profile)
	}
	if cfg.Sandbox.Workdir != "/tmp/myworkdir" {
		t.Errorf("expected workdir %q, got %q", "/tmp/myworkdir", cfg.Sandbox.Workdir)
	}
	if cfg.Sandbox.ClaudeBin != "/usr/local/bin/claude" {
		t.Errorf("expected claude_bin %q, got %q", "/usr/local/bin/claude", cfg.Sandbox.ClaudeBin)
	}
	if len(cfg.Unboxexec.AllowedCommands) != 1 {
		t.Fatalf("expected 1 allowed_commands, got %d", len(cfg.Unboxexec.AllowedCommands))
	}
}

func TestLoadConfigWithMultilineProfile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	content := `
[sandbox]
profile = '''
(version 1)
(allow default)
(deny file-write*)
'''
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile failed: %v", err)
	}

	// TOML multiline literal strings (''') strip the first newline
	expected := "(version 1)\n(allow default)\n(deny file-write*)\n"
	if cfg.Sandbox.Profile != expected {
		t.Errorf("expected profile %q, got %q", expected, cfg.Sandbox.Profile)
	}
}

func TestLoadConfigSandboxDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	content := `
[unboxexec]
allowed_commands = ["^echo"]
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile failed: %v", err)
	}

	// Sandbox fields should be zero values when not specified
	if cfg.Sandbox.Profile != "" {
		t.Errorf("expected empty profile, got %q", cfg.Sandbox.Profile)
	}
	if cfg.Sandbox.Workdir != "" {
		t.Errorf("expected empty workdir, got %q", cfg.Sandbox.Workdir)
	}
	if cfg.Sandbox.ClaudeBin != "" {
		t.Errorf("expected empty claude_bin, got %q", cfg.Sandbox.ClaudeBin)
	}
}

func TestLoadEmptyPath(t *testing.T) {
	cfg, err := LoadFile("")
	if err != nil {
		t.Fatalf("expected no error for empty path, got: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if len(cfg.Unboxexec.AllowedCommands) != 0 {
		t.Errorf("expected empty allowed_commands, got %d", len(cfg.Unboxexec.AllowedCommands))
	}
}

func TestLoadMissingFile(t *testing.T) {
	cfg, err := LoadFile("/nonexistent/path/config.toml")
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if len(cfg.Unboxexec.AllowedCommands) != 0 {
		t.Errorf("expected empty allowed_commands, got %d", len(cfg.Unboxexec.AllowedCommands))
	}
}

func TestLoadInvalidTOML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte("invalid [[[ toml"), 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	_, err := LoadFile(path)
	if err == nil {
		t.Fatal("expected error for invalid TOML")
	}
}

func TestCompileAllowedCommands(t *testing.T) {
	patterns := []string{"^playwright", "^echo .*"}
	compiled, err := CompileAllowedCommands(patterns)
	if err != nil {
		t.Fatalf("CompileAllowedCommands failed: %v", err)
	}
	if len(compiled) != 2 {
		t.Fatalf("expected 2 compiled patterns, got %d", len(compiled))
	}

	// Test matching
	if !compiled[0].MatchString("playwright install chromium") {
		t.Error("expected pattern to match 'playwright install chromium'")
	}
	if compiled[0].MatchString("echo playwright") {
		t.Error("expected pattern not to match 'echo playwright'")
	}
	if !compiled[1].MatchString("echo hello world") {
		t.Error("expected pattern to match 'echo hello world'")
	}
}

func TestCompileAllowedCommandsEmpty(t *testing.T) {
	compiled, err := CompileAllowedCommands(nil)
	if err != nil {
		t.Fatalf("CompileAllowedCommands failed: %v", err)
	}
	if len(compiled) != 0 {
		t.Errorf("expected 0 compiled patterns, got %d", len(compiled))
	}
}

func TestCompileAllowedCommandsInvalidPattern(t *testing.T) {
	patterns := []string{"[invalid"}
	_, err := CompileAllowedCommands(patterns)
	if err == nil {
		t.Fatal("expected error for invalid regex pattern")
	}
}

// --- mergeInto tests ---

func TestMergeIntoOverridesNonEmpty(t *testing.T) {
	dst := &Config{
		Sandbox: SandboxConfig{
			Profile:   "old-profile",
			Workdir:   "/old",
			ClaudeBin: "/old/claude",
		},
		Unboxexec: UnboxexecConfig{
			AllowedCommands: []string{"^old"},
		},
	}
	src := &Config{
		Sandbox: SandboxConfig{
			Profile:   "new-profile",
			Workdir:   "/new",
			ClaudeBin: "/new/claude",
		},
		Unboxexec: UnboxexecConfig{
			AllowedCommands: []string{"^new"},
		},
	}
	mergeInto(dst, src)

	if dst.Sandbox.Profile != "new-profile" {
		t.Errorf("expected profile %q, got %q", "new-profile", dst.Sandbox.Profile)
	}
	if dst.Sandbox.Workdir != "/new" {
		t.Errorf("expected workdir %q, got %q", "/new", dst.Sandbox.Workdir)
	}
	if dst.Sandbox.ClaudeBin != "/new/claude" {
		t.Errorf("expected claude_bin %q, got %q", "/new/claude", dst.Sandbox.ClaudeBin)
	}
	if len(dst.Unboxexec.AllowedCommands) != 1 || dst.Unboxexec.AllowedCommands[0] != "^new" {
		t.Errorf("expected allowed_commands [^new], got %v", dst.Unboxexec.AllowedCommands)
	}
}

func TestMergeIntoKeepsDstWhenSrcEmpty(t *testing.T) {
	dst := &Config{
		Sandbox: SandboxConfig{
			Profile:   "kept-profile",
			Workdir:   "/kept",
			ClaudeBin: "/kept/claude",
		},
		Unboxexec: UnboxexecConfig{
			AllowedCommands: []string{"^kept"},
		},
	}
	src := &Config{}
	mergeInto(dst, src)

	if dst.Sandbox.Profile != "kept-profile" {
		t.Errorf("expected profile to be kept, got %q", dst.Sandbox.Profile)
	}
	if dst.Sandbox.Workdir != "/kept" {
		t.Errorf("expected workdir to be kept, got %q", dst.Sandbox.Workdir)
	}
	if dst.Sandbox.ClaudeBin != "/kept/claude" {
		t.Errorf("expected claude_bin to be kept, got %q", dst.Sandbox.ClaudeBin)
	}
	if len(dst.Unboxexec.AllowedCommands) != 1 || dst.Unboxexec.AllowedCommands[0] != "^kept" {
		t.Errorf("expected allowed_commands to be kept, got %v", dst.Unboxexec.AllowedCommands)
	}
}

// --- Load tests ---

// setupHome temporarily overrides HOME and creates the .claude dir.
// Returns cleanup func.
func setupHome(t *testing.T, dir string) func() {
	t.Helper()
	orig := os.Getenv("HOME")
	if err := os.Setenv("HOME", dir); err != nil {
		t.Fatalf("failed to set HOME: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".claude"), 0o755); err != nil {
		t.Fatalf("failed to create home .claude dir: %v", err)
	}
	return func() {
		if err := os.Setenv("HOME", orig); err != nil {
			t.Errorf("failed to restore HOME: %v", err)
		}
	}
}

func TestLoadUserOnly(t *testing.T) {
	tmpHome := t.TempDir()
	cleanup := setupHome(t, tmpHome)
	defer cleanup()

	// Write user config
	userCfgPath := filepath.Join(tmpHome, ".claude", "sandbox.toml")
	if err := os.WriteFile(userCfgPath, []byte(`
[sandbox]
workdir = "/from-user"
[unboxexec]
allowed_commands = ["^user-cmd"]
`), 0o644); err != nil {
		t.Fatal(err)
	}

	// Change to a temp dir with no project/local configs
	wd := t.TempDir()
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(wd); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(origWd); err != nil {
			t.Errorf("failed to restore workdir: %v", err)
		}
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Sandbox.Workdir != "/from-user" {
		t.Errorf("expected workdir from user config, got %q", cfg.Sandbox.Workdir)
	}
	if len(cfg.Unboxexec.AllowedCommands) != 1 || cfg.Unboxexec.AllowedCommands[0] != "^user-cmd" {
		t.Errorf("expected allowed_commands from user config, got %v", cfg.Unboxexec.AllowedCommands)
	}
}

func TestLoadProjectOverridesUser(t *testing.T) {
	tmpHome := t.TempDir()
	cleanup := setupHome(t, tmpHome)
	defer cleanup()

	// Write user config
	userCfgPath := filepath.Join(tmpHome, ".claude", "sandbox.toml")
	if err := os.WriteFile(userCfgPath, []byte(`
[sandbox]
workdir = "/from-user"
claude_bin = "/user/claude"
[unboxexec]
allowed_commands = ["^user-cmd"]
`), 0o644); err != nil {
		t.Fatal(err)
	}

	// Set up workdir with project config
	wd := t.TempDir()
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(wd); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(origWd); err != nil {
			t.Errorf("failed to restore workdir: %v", err)
		}
	}()

	if err := os.MkdirAll(filepath.Join(wd, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	projectCfgPath := filepath.Join(wd, ".claude", "sandbox.toml")
	if err := os.WriteFile(projectCfgPath, []byte(`
[sandbox]
workdir = "/from-project"
[unboxexec]
allowed_commands = ["^project-cmd"]
`), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	// project overrides user workdir and allowed_commands
	if cfg.Sandbox.Workdir != "/from-project" {
		t.Errorf("expected workdir from project config, got %q", cfg.Sandbox.Workdir)
	}
	if len(cfg.Unboxexec.AllowedCommands) != 1 || cfg.Unboxexec.AllowedCommands[0] != "^project-cmd" {
		t.Errorf("expected allowed_commands from project config, got %v", cfg.Unboxexec.AllowedCommands)
	}
	// user-only field is kept
	if cfg.Sandbox.ClaudeBin != "/user/claude" {
		t.Errorf("expected claude_bin from user config, got %q", cfg.Sandbox.ClaudeBin)
	}
}

func TestLoadLocalOverridesAll(t *testing.T) {
	tmpHome := t.TempDir()
	cleanup := setupHome(t, tmpHome)
	defer cleanup()

	// Write user config
	userCfgPath := filepath.Join(tmpHome, ".claude", "sandbox.toml")
	if err := os.WriteFile(userCfgPath, []byte(`
[sandbox]
workdir = "/from-user"
[unboxexec]
allowed_commands = ["^user-cmd"]
`), 0o644); err != nil {
		t.Fatal(err)
	}

	// Set up workdir with project and local configs
	wd := t.TempDir()
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(wd); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(origWd); err != nil {
			t.Errorf("failed to restore workdir: %v", err)
		}
	}()

	if err := os.MkdirAll(filepath.Join(wd, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	projectCfgPath := filepath.Join(wd, ".claude", "sandbox.toml")
	if err := os.WriteFile(projectCfgPath, []byte(`
[sandbox]
workdir = "/from-project"
[unboxexec]
allowed_commands = ["^project-cmd"]
`), 0o644); err != nil {
		t.Fatal(err)
	}
	localCfgPath := filepath.Join(wd, ".claude", "sandbox.local.toml")
	if err := os.WriteFile(localCfgPath, []byte(`
[unboxexec]
allowed_commands = ["^local-cmd"]
`), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	// local overrides allowed_commands
	if len(cfg.Unboxexec.AllowedCommands) != 1 || cfg.Unboxexec.AllowedCommands[0] != "^local-cmd" {
		t.Errorf("expected allowed_commands from local config, got %v", cfg.Unboxexec.AllowedCommands)
	}
	// project workdir is kept (local didn't set it)
	if cfg.Sandbox.Workdir != "/from-project" {
		t.Errorf("expected workdir from project config, got %q", cfg.Sandbox.Workdir)
	}
}

func TestLoadNoConfigsReturnsEmpty(t *testing.T) {
	tmpHome := t.TempDir()
	cleanup := setupHome(t, tmpHome)
	defer cleanup()

	wd := t.TempDir()
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(wd); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(origWd); err != nil {
			t.Errorf("failed to restore workdir: %v", err)
		}
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Sandbox.Profile != "" || cfg.Sandbox.Workdir != "" || cfg.Sandbox.ClaudeBin != "" {
		t.Errorf("expected all sandbox fields empty, got %+v", cfg.Sandbox)
	}
	if len(cfg.Unboxexec.AllowedCommands) != 0 {
		t.Errorf("expected empty allowed_commands, got %v", cfg.Unboxexec.AllowedCommands)
	}
}
