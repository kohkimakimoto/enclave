package command

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kohkimakimoto/enclave/v3/internal/config"
	"github.com/kohkimakimoto/enclave/v3/internal/sandbox"
	"github.com/urfave/cli/v3"
)

// userConfigTemplate generates the template for user-level config.toml.
func userConfigTemplate() string {
	return `# User-level configuration for enclave.
# See https://github.com/kohkimakimoto/enclave

# Sandbox profile for sandbox-exec.
# If not set, the built-in default profile is used.
# sandbox_profile = '''
` + sandbox.CommentedDefaultProfile() + `
# '''

# Regex patterns for allowed commands in unboxexec.
# The command + args joined by spaces is matched against each pattern.
# If any pattern matches, the command is allowed.
# If empty or not configured, all commands are rejected.
unboxexec_allowed_commands = [
    # "^playwright-cli",
]
`
}

// InitUserCommand creates the user-level config.toml (~/.config/enclave/config.toml).
func InitUserCommand() *cli.Command {
	return &cli.Command{
		Name:   "init-user",
		Usage:  "Create $XDG_CONFIG_HOME/enclave/config.toml (or ~/.config/enclave/config.toml) if it doesn't exist",
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
	configDir := config.UserConfigDir()
	configFile := filepath.Join(configDir, "config.toml")

	if _, err := os.Stat(configFile); err == nil {
		return fmt.Errorf("user config file already exists: %s", configFile)
	}

	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(configFile, []byte(userConfigTemplate()), 0o644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	fmt.Fprintf(cmd.Root().Writer, "Created user config file: %s\n", configFile)
	return nil
}
