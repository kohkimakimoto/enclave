package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/BurntSushi/toml"
)

// Config represents the enclave configuration.
type Config struct {
	Sandbox   SandboxConfig   `toml:"sandbox"`
	Unboxexec UnboxexecConfig `toml:"unboxexec"`
}

// SandboxConfig holds settings for the sandbox environment.
type SandboxConfig struct {
	// Profile is the sandbox-exec profile content.
	// If empty, the built-in default profile is used.
	Profile string `toml:"profile"`
	// Workdir overrides the working directory for sandbox execution.
	// If empty, the current directory is used.
	Workdir string `toml:"workdir"`
	// ClaudeBin overrides the path to the claude binary.
	// If empty, PATH search is used.
	ClaudeBin string `toml:"claude_bin"`
}

// UnboxexecConfig holds settings for the unboxexec daemon.
type UnboxexecConfig struct {
	AllowedCommands []string `toml:"allowed_commands"`
}

// mergeInto merges non-zero fields of src into dst.
func mergeInto(dst, src *Config) {
	if src.Sandbox.Profile != "" {
		dst.Sandbox.Profile = src.Sandbox.Profile
	}
	if src.Sandbox.Workdir != "" {
		dst.Sandbox.Workdir = src.Sandbox.Workdir
	}
	if src.Sandbox.ClaudeBin != "" {
		dst.Sandbox.ClaudeBin = src.Sandbox.ClaudeBin
	}
	if len(src.Unboxexec.AllowedCommands) > 0 {
		dst.Unboxexec.AllowedCommands = src.Unboxexec.AllowedCommands
	}
}

// ConfigPaths holds the resolved paths for each config scope.
type ConfigPaths struct {
	// User is the user-level config path (~/.claude/sandbox.toml).
	User string
	// Project is the project-level config path (.claude/sandbox.toml in workdir).
	Project string
	// Local is the local override config path (.claude/sandbox.local.toml in workdir).
	Local string
}

// ResolveConfigPaths returns the paths for all three config scopes.
// Each field is set to the path only if the file exists; otherwise it is empty.
func ResolveConfigPaths() ConfigPaths {
	wd, _ := os.Getwd()
	home, _ := os.UserHomeDir()

	paths := ConfigPaths{}

	userConfig := filepath.Join(home, ".claude", "sandbox.toml")
	if _, err := os.Stat(userConfig); err == nil {
		paths.User = userConfig
	}

	projectConfig := filepath.Join(wd, ".claude", "sandbox.toml")
	if _, err := os.Stat(projectConfig); err == nil {
		paths.Project = projectConfig
	}

	localConfig := filepath.Join(wd, ".claude", "sandbox.local.toml")
	if _, err := os.Stat(localConfig); err == nil {
		paths.Local = localConfig
	}

	return paths
}

// Load loads and merges configs from all scopes in order:
//  1. user   (~/.claude/sandbox.toml)
//  2. project (.claude/sandbox.toml in workdir)
//  3. local   (.claude/sandbox.local.toml in workdir)
//
// Each scope overrides the previous one for any field that is explicitly set.
func Load() (*Config, error) {
	paths := ResolveConfigPaths()

	merged := &Config{}

	for _, path := range []string{paths.User, paths.Project, paths.Local} {
		if path == "" {
			continue
		}
		cfg, err := LoadFile(path)
		if err != nil {
			return nil, err
		}
		mergeInto(merged, cfg)
	}

	return merged, nil
}

// LoadFile reads and parses a TOML config file at the given path.
// If the path is empty or the file does not exist, it returns an empty Config without error.
func LoadFile(path string) (*Config, error) {
	cfg := &Config{}

	if path == "" {
		return cfg, nil
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil
	}

	if _, err := toml.DecodeFile(path, cfg); err != nil {
		return nil, fmt.Errorf("failed to load config %s: %w", path, err)
	}

	return cfg, nil
}

// CompileAllowedCommands compiles a list of regex pattern strings into []*regexp.Regexp.
func CompileAllowedCommands(patterns []string) ([]*regexp.Regexp, error) {
	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		re, err := regexp.Compile(p)
		if err != nil {
			return nil, fmt.Errorf("invalid allowed_commands pattern %q: %w", p, err)
		}
		compiled = append(compiled, re)
	}
	return compiled, nil
}
