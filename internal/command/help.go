package command

import "github.com/urfave/cli/v3"

func init() {
	cli.RootCommandHelpTemplate = `Usage: enclave [<command>]|[claude [<args of claude command...>]]

{{if .Usage}}{{ .Usage }}{{end}}

Builtin commands:{{template "visibleCommandCategoryTemplate" .}}

Configuration:
   enclave looks for config files in the following order:

   1. $HOME/.claude/sandbox.toml (user-level)
   2. .claude/sandbox.toml (project-level)
   3. .claude/sandbox.local.toml (local overrides, gitignore-friendly)

   See: https://github.com/kohkimakimoto/enclave#configuration-file

Example Usage:
   # Create project-specific config file
   $ enclave init

   # Create local override config file (not for version control)
   $ enclave init-local

   # Create user config file
   $ enclave init-user

   # Print the evaluated sandbox profile
   $ enclave profile

   # Run Claude Code in a sandboxed environment
   $ enclave claude

   # Run Claude Code with arguments in a sandboxed environment
   $ enclave claude --dangerously-skip-permissions

   # You can also run Claude Code without the 'claude' command prefix.
   $ enclave
   $ enclave --dangerously-skip-permissions

   Commands or options that conflict with enclave can be used with the claude command prefix.
   For example, the following command shows the claude help, not the enclave help.
   $ enclave claude -h

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
