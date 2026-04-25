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
unboxexec_allowed_commands = [
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

	if len(cfg.UnboxexecAllowedCommands) != 2 {
		t.Fatalf("expected 2 unboxexec_allowed_commands, got %d", len(cfg.UnboxexecAllowedCommands))
	}
	if cfg.UnboxexecAllowedCommands[0] != "^playwright" {
		t.Errorf("expected %q, got %q", "^playwright", cfg.UnboxexecAllowedCommands[0])
	}
	if cfg.UnboxexecAllowedCommands[1] != "^echo hello" {
		t.Errorf("expected %q, got %q", "^echo hello", cfg.UnboxexecAllowedCommands[1])
	}
}

func TestLoadConfigWithSandboxProfile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	content := `
sandbox_profile = "(version 1)\n(allow default)"

unboxexec_allowed_commands = [
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

	if cfg.SandboxProfile != "(version 1)\n(allow default)" {
		t.Errorf("expected profile %q, got %q", "(version 1)\n(allow default)", cfg.SandboxProfile)
	}
	if len(cfg.UnboxexecAllowedCommands) != 1 {
		t.Fatalf("expected 1 unboxexec_allowed_commands, got %d", len(cfg.UnboxexecAllowedCommands))
	}
}

func TestLoadConfigWithMultilineProfile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	content := `
sandbox_profile = '''
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
	if cfg.SandboxProfile != expected {
		t.Errorf("expected profile %q, got %q", expected, cfg.SandboxProfile)
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	content := `
unboxexec_allowed_commands = ["^echo"]
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile failed: %v", err)
	}

	if cfg.SandboxProfile != "" {
		t.Errorf("expected empty sandbox_profile, got %q", cfg.SandboxProfile)
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
	if len(cfg.UnboxexecAllowedCommands) != 0 {
		t.Errorf("expected empty unboxexec_allowed_commands, got %d", len(cfg.UnboxexecAllowedCommands))
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
	if len(cfg.UnboxexecAllowedCommands) != 0 {
		t.Errorf("expected empty unboxexec_allowed_commands, got %d", len(cfg.UnboxexecAllowedCommands))
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
		SandboxProfile:           "old-profile",
		UnboxexecAllowedCommands: []string{"^old"},
	}
	src := &Config{
		SandboxProfile:           "new-profile",
		UnboxexecAllowedCommands: []string{"^new"},
	}
	mergeInto(dst, src)

	if dst.SandboxProfile != "new-profile" {
		t.Errorf("expected profile %q, got %q", "new-profile", dst.SandboxProfile)
	}
	if len(dst.UnboxexecAllowedCommands) != 1 || dst.UnboxexecAllowedCommands[0] != "^new" {
		t.Errorf("expected unboxexec_allowed_commands [^new], got %v", dst.UnboxexecAllowedCommands)
	}
}

func TestMergeIntoKeepsDstWhenSrcEmpty(t *testing.T) {
	dst := &Config{
		SandboxProfile:           "kept-profile",
		UnboxexecAllowedCommands: []string{"^kept"},
	}
	src := &Config{}
	mergeInto(dst, src)

	if dst.SandboxProfile != "kept-profile" {
		t.Errorf("expected profile to be kept, got %q", dst.SandboxProfile)
	}
	if len(dst.UnboxexecAllowedCommands) != 1 || dst.UnboxexecAllowedCommands[0] != "^kept" {
		t.Errorf("expected unboxexec_allowed_commands to be kept, got %v", dst.UnboxexecAllowedCommands)
	}
}

// --- Load tests ---

// setupXDGConfig temporarily overrides XDG_CONFIG_HOME to a temp directory.
func setupXDGConfig(t *testing.T) (configDir string) {
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
	// UserConfigDir() returns filepath.Join(XDG_CONFIG_HOME, "enclave")
	configDir = filepath.Join(tmpDir, "enclave")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	return configDir
}

func TestLoadUserOnly(t *testing.T) {
	configDir := setupXDGConfig(t)

	userCfgPath := filepath.Join(configDir, "config.toml")
	if err := os.WriteFile(userCfgPath, []byte(`
sandbox_profile = "from-user"
unboxexec_allowed_commands = ["^user-cmd"]
`), 0o644); err != nil {
		t.Fatal(err)
	}

	wd := t.TempDir()
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(wd); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origWd) }()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.SandboxProfile != "from-user" {
		t.Errorf("expected sandbox_profile from user config, got %q", cfg.SandboxProfile)
	}
	if len(cfg.UnboxexecAllowedCommands) != 1 || cfg.UnboxexecAllowedCommands[0] != "^user-cmd" {
		t.Errorf("expected unboxexec_allowed_commands from user config, got %v", cfg.UnboxexecAllowedCommands)
	}
}

func TestLoadProjectOverridesUser(t *testing.T) {
	configDir := setupXDGConfig(t)

	userCfgPath := filepath.Join(configDir, "config.toml")
	if err := os.WriteFile(userCfgPath, []byte(`
sandbox_profile = "from-user"
unboxexec_allowed_commands = ["^user-cmd"]
`), 0o644); err != nil {
		t.Fatal(err)
	}

	wd := t.TempDir()
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(wd); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origWd) }()

	projectCfgPath := filepath.Join(wd, "enclave.toml")
	if err := os.WriteFile(projectCfgPath, []byte(`
unboxexec_allowed_commands = ["^project-cmd"]
`), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	// project overrides allowed_commands
	if len(cfg.UnboxexecAllowedCommands) != 1 || cfg.UnboxexecAllowedCommands[0] != "^project-cmd" {
		t.Errorf("expected unboxexec_allowed_commands from project config, got %v", cfg.UnboxexecAllowedCommands)
	}
	// user-only field is kept
	if cfg.SandboxProfile != "from-user" {
		t.Errorf("expected sandbox_profile from user config, got %q", cfg.SandboxProfile)
	}
}

func TestLoadLocalOverridesAll(t *testing.T) {
	configDir := setupXDGConfig(t)

	userCfgPath := filepath.Join(configDir, "config.toml")
	if err := os.WriteFile(userCfgPath, []byte(`
unboxexec_allowed_commands = ["^user-cmd"]
`), 0o644); err != nil {
		t.Fatal(err)
	}

	wd := t.TempDir()
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(wd); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origWd) }()

	if err := os.WriteFile(filepath.Join(wd, "enclave.toml"), []byte(`
unboxexec_allowed_commands = ["^project-cmd"]
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wd, "enclave.local.toml"), []byte(`
unboxexec_allowed_commands = ["^local-cmd"]
`), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(cfg.UnboxexecAllowedCommands) != 1 || cfg.UnboxexecAllowedCommands[0] != "^local-cmd" {
		t.Errorf("expected unboxexec_allowed_commands from local config, got %v", cfg.UnboxexecAllowedCommands)
	}
}

func TestLoadNoConfigsReturnsEmpty(t *testing.T) {
	setupXDGConfig(t) // sets XDG_CONFIG_HOME to empty temp dir (no config.toml inside)

	wd := t.TempDir()
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(wd); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origWd) }()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.SandboxProfile != "" {
		t.Errorf("expected empty sandbox_profile, got %q", cfg.SandboxProfile)
	}
	if len(cfg.UnboxexecAllowedCommands) != 0 {
		t.Errorf("expected empty unboxexec_allowed_commands, got %v", cfg.UnboxexecAllowedCommands)
	}
}

func TestUserConfigDir_XDG(t *testing.T) {
	orig := os.Getenv("XDG_CONFIG_HOME")
	defer func() { _ = os.Setenv("XDG_CONFIG_HOME", orig) }()

	_ = os.Setenv("XDG_CONFIG_HOME", "/custom/xdg")
	got := UserConfigDir()
	expected := "/custom/xdg/enclave"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestUserConfigDir_Fallback(t *testing.T) {
	orig := os.Getenv("XDG_CONFIG_HOME")
	defer func() { _ = os.Setenv("XDG_CONFIG_HOME", orig) }()

	_ = os.Setenv("XDG_CONFIG_HOME", "")
	home, _ := os.UserHomeDir()
	got := UserConfigDir()
	expected := filepath.Join(home, ".config", "enclave")
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}
