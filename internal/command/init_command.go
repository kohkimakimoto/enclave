package command

import (
	"context"
	"fmt"
	"os"

	"github.com/kohkimakimoto/enclave/v3/internal/sandbox"
	"github.com/urfave/cli/v3"
)

// projectConfigTemplate generates the template for project-specific enclave.toml.
func projectConfigTemplate() string {
	return `# Project-specific configuration for enclave.
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

func InitCommand() *cli.Command {
	return &cli.Command{
		Name:   "init",
		Usage:  "Create enclave.toml file if it doesn't exist",
		Action: initAction,
	}
}

func initAction(ctx context.Context, cmd *cli.Command) error {
	wd, _ := os.Getwd()
	configFile := wd + "/enclave.toml"

	if _, err := os.Stat(configFile); err == nil {
		return fmt.Errorf("config file already exists: %s", configFile)
	}

	if err := os.WriteFile(configFile, []byte(projectConfigTemplate()), 0o644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	fmt.Fprintf(cmd.Root().Writer, "Created config file: %s\n", configFile)
	return nil
}
