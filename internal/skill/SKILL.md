---
name: enclave
description: A sandbox environment information. You should use this skill when you need to check if the sandbox is enabled or not, or when you need to execute commands outside the sandbox.
---

# Enclave

The current command is executed within a sandboxed environment using `enclave`.
This uses Apple's `sandbox-exec` to restrict access to unnecessary system resources.

## Checking the Sandbox Status

Verify whether you are running inside the sandbox:

```bash
echo $ENCLAVE_SANDBOX
# => 1 (inside sandbox), or empty (not sandboxed)
```

## Checking the Sandbox Profile

To inspect the full `sandbox-exec` profile in effect (useful for troubleshooting permission errors):

```bash
enclave profile
```

## Checking the Effective Configuration

Use the `config` subcommand to see the current effective configuration, including which config files are loaded and what commands are allowed for sandbox-bypass execution:

```bash
enclave config
```

Example output:

```toml
# Loaded config files:
#   user:    /Users/yourname/.config/enclave/config.toml
#   project: ./enclave.toml
#   local:   (none)

sandbox_profile = ""

unboxexec_allowed_commands = [
  "^playwright-cli",
]
```

## Executing Commands Outside the Sandbox

The `enclave unboxexec` subcommand executes commands outside the sandbox.

**Important:** `enclave unboxexec` bypasses sandbox protections. You MUST ask for explicit user approval before using it, UNLESS the command is listed in `unboxexec_allowed_commands`.

### Basic Usage

```bash
enclave unboxexec [<options>] -- <command> [<args...>]
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
enclave unboxexec -- echo "hello from outside"

# Execute with a specified working directory
enclave unboxexec --dir /tmp -- ls -la

# Execute with an extended timeout
enclave unboxexec --timeout 300 -- long-running-command

# Execute with environment variables
enclave unboxexec --env API_KEY=secret --env DEBUG=1 -- my-command
```

## Troubleshooting: When a Command Fails

When you need to run a command that may fail due to sandbox restrictions:

1. Try running the command normally inside the sandbox
2. If it fails with a permission error, try `enclave unboxexec -- <command>`
3. If that also fails, run `enclave config` and check `unboxexec_allowed_commands`
4. If the command matches an allowed pattern → it should be permitted; investigate the error further
5. If the command does NOT match → ask the user for explicit approval before retrying
