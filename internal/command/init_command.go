package command

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kohkimakimoto/claude-sandbox/v2/internal/sandbox"
	"github.com/urfave/cli/v3"
)

// projectConfigTemplate generates the template for project-specific sandbox.toml.
func projectConfigTemplate() string {
	return `# Project-specific configuration for claude-sandbox.
# See https://github.com/kohkimakimoto/claude-sandbox

[sandbox]
# Sandbox profile for sandbox-exec.
# If not set, the built-in default profile is used.
# profile = '''
` + sandbox.CommentedDefaultProfile() + `
# '''

# Override working directory (optional).
# workdir = "/path/to/workdir"

# Override claude binary path (optional).
# claude_bin = "/path/to/claude"

[unboxexec]
# Regex patterns for allowed commands.
# The command + args joined by spaces is matched against each pattern.
# If any pattern matches, the command is allowed.
# If empty or not configured, all commands are rejected.
allowed_commands = [
    # "^playwright-cli",
]
`
}

func InitCommand() *cli.Command {
	return &cli.Command{
		Name:   "init",
		Usage:  "Create .claude/sandbox.toml file if it doesn't exist",
		Action: initAction,
	}
}

func initAction(ctx context.Context, cmd *cli.Command) error {
	workdir := sandbox.GetWorkdir("")
	configFile := filepath.Join(workdir, ".claude", "sandbox.toml")

	if _, err := os.Stat(configFile); err == nil {
		return fmt.Errorf("config file already exists: %s", configFile)
	}

	// Create .claude directory if it doesn't exist
	if err := os.MkdirAll(filepath.Join(workdir, ".claude"), 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(configFile, []byte(projectConfigTemplate()), 0o644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	fmt.Fprintf(cmd.Root().Writer, "Created config file: %s\n", configFile)
	return nil
}
