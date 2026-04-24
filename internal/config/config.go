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
	// SandboxProfile is the sandbox-exec profile content.
	// If empty, the built-in default profile is used.
	SandboxProfile string `toml:"sandbox_profile"`
	// UnboxexecAllowedCommands is a list of regex patterns for allowed commands.
	UnboxexecAllowedCommands []string `toml:"unboxexec_allowed_commands"`
}

// mergeInto merges non-zero fields of src into dst.
func mergeInto(dst, src *Config) {
	if src.SandboxProfile != "" {
		dst.SandboxProfile = src.SandboxProfile
	}
	if len(src.UnboxexecAllowedCommands) > 0 {
		dst.UnboxexecAllowedCommands = src.UnboxexecAllowedCommands
	}
}

// ConfigPaths holds the resolved paths for each config scope.
type ConfigPaths struct {
	// User is the user-level config path (~/.config/enclave/config.toml).
	User string
	// Project is the project-level config path (./enclave.toml in workdir).
	Project string
	// Local is the local override config path (./enclave.local.toml in workdir).
	Local string
}

// UserConfigDir returns the XDG-compliant user config directory for enclave.
// Uses $XDG_CONFIG_HOME if set, otherwise falls back to ~/.config.
func UserConfigDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "enclave")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "enclave")
}

// ResolveConfigPaths returns the paths for all three config scopes.
// Each field is set to the path only if the file exists; otherwise it is empty.
func ResolveConfigPaths() ConfigPaths {
	wd, _ := os.Getwd()

	paths := ConfigPaths{}

	userConfig := filepath.Join(UserConfigDir(), "config.toml")
	if _, err := os.Stat(userConfig); err == nil {
		paths.User = userConfig
	}

	projectConfig := filepath.Join(wd, "enclave.toml")
	if _, err := os.Stat(projectConfig); err == nil {
		paths.Project = projectConfig
	}

	localConfig := filepath.Join(wd, "enclave.local.toml")
	if _, err := os.Stat(localConfig); err == nil {
		paths.Local = localConfig
	}

	return paths
}

// Load loads and merges configs from all scopes in order:
//  1. user    (~/.config/enclave/config.toml)
//  2. project (./enclave.toml in workdir)
//  3. local   (./enclave.local.toml in workdir)
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

// DumpFile writes cfg to the file at path in TOML format.
func DumpFile(path string, cfg *Config) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create config dump %s: %w", path, err)
	}
	defer f.Close()
	if err := toml.NewEncoder(f).Encode(cfg); err != nil {
		return fmt.Errorf("failed to write config dump %s: %w", path, err)
	}
	return nil
}

// CompileAllowedCommands compiles a list of regex pattern strings into []*regexp.Regexp.
func CompileAllowedCommands(patterns []string) ([]*regexp.Regexp, error) {
	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		re, err := regexp.Compile(p)
		if err != nil {
			return nil, fmt.Errorf("invalid unboxexec_allowed_commands pattern %q: %w", p, err)
		}
		compiled = append(compiled, re)
	}
	return compiled, nil
}
