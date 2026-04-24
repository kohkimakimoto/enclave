package command

import "github.com/urfave/cli/v3"

func init() {
	cli.RootCommandHelpTemplate = `Usage: enclave <command> [options] [--] [args...]

{{if .Usage}}{{ .Usage }}{{end}}

Builtin commands:{{template "visibleCommandCategoryTemplate" .}}

Configuration:
   enclave looks for config files in the following order:

   1. $XDG_CONFIG_HOME/enclave/config.toml (or ~/.config/enclave/config.toml) (user-level)
   2. ./enclave.toml (project-level)
   3. ./enclave.local.toml (local overrides, gitignore-friendly)

   See: https://github.com/kohkimakimoto/enclave#configuration-file

Example Usage:
   # Run a command in a sandboxed environment
   $ enclave run claude --dangerously-skip-permissions
   $ enclave run copilot --allow-all

   # Run with a custom config file
   $ enclave run --config enclave-custom.toml copilot

   # Use -- to separate enclave options from command arguments
   $ enclave run --config enclave-custom.toml -- claude -p "hello"

   # Create project-specific config file
   $ enclave init

   # Create local override config file (not for version control)
   $ enclave init-local

   # Create user config file
   $ enclave init-user

   # Print the evaluated sandbox profile
   $ enclave profile

Version: {{ .Version }}
Commit: {{ index (ExtraInfo) "CommitHash" }}
{{template "copyrightTemplate" .}}
`

	cli.CommandHelpTemplate = `Usage: {{template "usageTemplate" .}}

{{if .Usage}}{{ .Usage }}{{end}}{{if .VisibleFlagCategories}}

Options:{{template "visibleFlagCategoryTemplate" .}}{{else if .VisibleFlags}}

Options:{{template "visibleFlagTemplate" .}}{{end}}{{if .VisiblePersistentFlags}}

Global Options:{{template "visiblePersistentFlagTemplate" .}}{{end}}
`
}
