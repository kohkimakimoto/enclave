package command

import (
	"context"
	"fmt"
	"os"

	"github.com/kohkimakimoto/enclave/v3/internal/sandbox"
	"github.com/urfave/cli/v3"
)

// localConfigTemplate generates the template for local override enclave.local.toml.
func localConfigTemplate() string {
	return `# Local override configuration for enclave.
# This file is intended for personal, machine-specific settings that should
# not be committed to version control. Add it to .gitignore.
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

func InitLocalCommand() *cli.Command {
	return &cli.Command{
		Name:   "init-local",
		Usage:  "Create enclave.local.toml file if it doesn't exist",
		Action: initLocalAction,
	}
}

func initLocalAction(ctx context.Context, cmd *cli.Command) error {
	wd, _ := os.Getwd()
	configFile := wd + "/enclave.local.toml"

	if _, err := os.Stat(configFile); err == nil {
		return fmt.Errorf("local config file already exists: %s", configFile)
	}

	if err := os.WriteFile(configFile, []byte(localConfigTemplate()), 0o644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	fmt.Fprintf(cmd.Root().Writer, "Created local config file: %s\n", configFile)
	return nil
}
