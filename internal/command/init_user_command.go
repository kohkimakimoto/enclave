package command

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kohkimakimoto/enclave/v3/internal/sandbox"
	"github.com/urfave/cli/v3"
)

// userConfigTemplate generates the template for user-level sandbox.toml.
func userConfigTemplate() string {
	return `# User-level configuration for enclave.
# See https://github.com/kohkimakimoto/enclave

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

// InitUserCommand creates the user-level sandbox.toml (~/.claude/sandbox.toml).
func InitUserCommand() *cli.Command {
	return &cli.Command{
		Name:   "init-user",
		Usage:  "Create $HOME/.claude/sandbox.toml file if it doesn't exist",
		Action: initUserAction,
	}
}

// InitGlobalCommand is kept as a backward-compatible alias for InitUserCommand.
func InitGlobalCommand() *cli.Command {
	return &cli.Command{
		Name:   "init-global",
		Usage:  "Alias for init-user (deprecated)",
		Action: initUserAction,
		Hidden: true,
	}
}

func initUserAction(ctx context.Context, cmd *cli.Command) error {
	home, _ := os.UserHomeDir()
	configFile := filepath.Join(home, ".claude", "sandbox.toml")

	if _, err := os.Stat(configFile); err == nil {
		return fmt.Errorf("user config file already exists: %s", configFile)
	}

	// Create .claude directory if it doesn't exist
	if err := os.MkdirAll(filepath.Join(home, ".claude"), 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(configFile, []byte(userConfigTemplate()), 0o644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	fmt.Fprintf(cmd.Root().Writer, "Created user config file: %s\n", configFile)
	return nil
}
