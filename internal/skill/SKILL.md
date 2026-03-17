---
name: claude-sandbox
description: A sandbox environment information. You should use this skill when you need to check if the sandbox is enabled or not, or when you need to execute commands outside the sandbox.
---

# Claude Sandbox

Claude Code is executed within a sandboxed environment using `claude-sandbox`.
This uses Apple's `sandbox-exec` to restrict access to unnecessary system resources.

## Checking the Sandbox Status

Verify whether you are running inside the sandbox:

```bash
echo $CLAUDE_SANDBOX
# => 1 (inside sandbox), or empty (not sandboxed)
```

## Checking the Sandbox Profile

To inspect the full `sandbox-exec` profile in effect (useful for troubleshooting permission errors):

```bash
claude-sandbox profile
```

## Checking the Effective Configuration

Use the `config` subcommand to see the current effective configuration, including which config files are loaded and what commands are allowed for sandbox-bypass execution:

```bash
claude-sandbox config
```

Example output:

```toml
# Loaded config files:
#   user:    /Users/yourname/.claude/sandbox.toml
#   project: .claude/sandbox.toml
#   local:   (none)

[sandbox]
workdir    = ""
claude_bin = ""
profile    = ""

[unboxexec]
allowed_commands = [
  "^playwright-cli",
]
```

## Executing Commands Outside the Sandbox

The `claude-sandbox unboxexec` subcommand executes commands outside the sandbox.

**Important:** `claude-sandbox unboxexec` bypasses sandbox protections. You MUST ask for explicit user approval before using it, UNLESS the command is listed in `allowed_commands`.

### Basic Usage

```bash
claude-sandbox unboxexec [<options>] -- <command> [<args...>]
```

### Options

| Flag | Short | Description |
|------|-------|-------------|
| `--dir` | `-C` | Specify the working directory for the command |
| `--timeout` | `-t` | Timeout in seconds (default: 60 seconds) |
| `--env` | `-e` | Specify environment variables in `KEY=VALUE` format (can be specified multiple times) |

### Examples

```bash
# Execute a command outside the sandbox
claude-sandbox unboxexec -- echo "hello from outside"

# Execute with a specified working directory
claude-sandbox unboxexec --dir /tmp -- ls -la

# Execute with an extended timeout
claude-sandbox unboxexec --timeout 300 -- long-running-command

# Execute with environment variables
claude-sandbox unboxexec --env API_KEY=secret --env DEBUG=1 -- my-command
```

## Troubleshooting: When a Command Fails

When you need to run a command that may fail due to sandbox restrictions:

1. Try running the command normally inside the sandbox
2. If it fails with a permission error, try `claude-sandbox unboxexec -- <command>`
3. If that also fails, run `claude-sandbox config` and check `allowed_commands` under `[unboxexec]`
4. If the command matches an allowed pattern → it should be permitted; investigate the error further
5. If the command does NOT match → ask the user for explicit approval before retrying
