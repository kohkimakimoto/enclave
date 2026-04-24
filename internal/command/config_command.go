package command

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/kohkimakimoto/enclave/v3/internal/config"
	"github.com/urfave/cli/v3"
)

func ConfigCommand() *cli.Command {
	return &cli.Command{
		Name:   "config",
		Usage:  "Print the effective configuration and exit",
		Action: configAction,
	}
}

func configAction(ctx context.Context, cmd *cli.Command) error {
	paths := config.ResolveConfigPaths()
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	w := cmd.Root().Writer
	printConfig(w, paths, cfg)
	return nil
}

func printConfig(w io.Writer, paths config.ConfigPaths, cfg *config.Config) {
	noneIfEmpty := func(s string) string {
		if s == "" {
			return "(none)"
		}
		return s
	}

	fmt.Fprintf(w, "# Loaded config files:\n")
	fmt.Fprintf(w, "#   user:    %s\n", noneIfEmpty(paths.User))
	fmt.Fprintf(w, "#   project: %s\n", noneIfEmpty(paths.Project))
	fmt.Fprintf(w, "#   local:   %s\n", noneIfEmpty(paths.Local))
	fmt.Fprintf(w, "\n")

	fmt.Fprintf(w, "[sandbox]\n")
	fmt.Fprintf(w, "workdir    = %s\n", tomlString(cfg.Sandbox.Workdir))
	fmt.Fprintf(w, "claude_bin = %s\n", tomlString(cfg.Sandbox.ClaudeBin))
	if cfg.Sandbox.Profile == "" {
		fmt.Fprintf(w, "profile    = \"\"\n")
	} else {
		fmt.Fprintf(w, "profile    = %s\n", tomlMultilineString(cfg.Sandbox.Profile))
	}
	fmt.Fprintf(w, "\n")

	fmt.Fprintf(w, "[unboxexec]\n")
	if len(cfg.Unboxexec.AllowedCommands) == 0 {
		fmt.Fprintf(w, "allowed_commands = []\n")
	} else {
		fmt.Fprintf(w, "allowed_commands = [\n")
		for _, cmd := range cfg.Unboxexec.AllowedCommands {
			fmt.Fprintf(w, "  %s,\n", tomlString(cmd))
		}
		fmt.Fprintf(w, "]\n")
	}
}

// tomlString formats s as a TOML basic string (double-quoted, with escaping).
func tomlString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	s = strings.ReplaceAll(s, "\t", `\t`)
	return `"` + s + `"`
}

// tomlMultilineString formats s as a TOML multiline literal string (''' delimited).
// If s contains ''', falls back to a TOML multiline basic string (""" delimited).
func tomlMultilineString(s string) string {
	// Ensure the content ends with a newline for clean closing delimiter placement.
	if !strings.HasSuffix(s, "\n") {
		s += "\n"
	}

	if !strings.Contains(s, "'''") {
		return "'''\n" + s + "'''"
	}

	// Fallback: basic multiline string with escaping for special characters.
	// In TOML basic multiline strings, only \ and " need escaping (not ''').
	escaped := strings.ReplaceAll(s, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"""`, `""\"`)
	return "\"\"\"\n" + escaped + "\"\"\""
}
