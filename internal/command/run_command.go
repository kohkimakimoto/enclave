package command

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/kohkimakimoto/enclave/v3/internal/config"
	"github.com/kohkimakimoto/enclave/v3/internal/sandbox"
	"github.com/kohkimakimoto/enclave/v3/internal/unboxexec"
	"github.com/urfave/cli/v3"
)

func RunCommand() *cli.Command {
	return &cli.Command{
		Name:      "run",
		Usage:     "Run a command in a sandboxed environment",
		UsageText: "enclave run [options] [--] <command> [args...]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Path to a config file (overrides automatic config resolution)",
			},
		},
		Action: runAction,
	}
}

func runAction(ctx context.Context, cmd *cli.Command) error {
	var cfg *config.Config
	var err error

	if configPath := cmd.String("config"); configPath != "" {
		cfg, err = config.LoadFile(configPath)
		if err != nil {
			return err
		}
	} else {
		cfg, err = config.Load()
		if err != nil {
			return err
		}
	}

	args := cmd.Args().Slice()
	if len(args) == 0 {
		return fmt.Errorf("no command specified\n\nUsage: enclave run [options] [--] <command> [args...]")
	}

	return runSandboxed(ctx, args, cfg)
}

// runSandboxed executes the given command inside a macOS sandbox using sandbox-exec.
// It starts an internal daemon for sandbox-external command execution,
// then runs sandbox-exec as a child process.
func runSandboxed(ctx context.Context, args []string, cfg *config.Config) error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	profilePath, cleanup, err := sandbox.BuildProfile(cfg.SandboxProfile)
	if err != nil {
		return err
	}
	defer cleanup()

	wd, _ := os.Getwd()
	home, _ := os.UserHomeDir()

	// Compile allowed command patterns
	allowedCommands, err := config.CompileAllowedCommands(cfg.UnboxexecAllowedCommands)
	if err != nil {
		return fmt.Errorf("failed to compile unboxexec_allowed_commands: %w", err)
	}

	// Start the daemon for sandbox-external command execution
	sockPath := sandbox.SocketPath()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if err := unboxexec.StartDaemon(ctx, sockPath, allowedCommands); err != nil {
		return fmt.Errorf("failed to start daemon: %w", err)
	}

	// Build sandbox-exec command arguments
	sandboxExecArgs := []string{
		"-D", "WORKDIR=" + wd,
		"-D", "HOME=" + home,
		"-f", profilePath,
	}
	sandboxExecArgs = append(sandboxExecArgs, args...)

	// Run sandbox-exec as a child process
	eCmd := exec.CommandContext(ctx, "sandbox-exec", sandboxExecArgs...)
	eCmd.Env = append(os.Environ(),
		"ENCLAVE_SANDBOX=1",
		"ENCLAVE_UNBOXEXEC_SOCK="+sockPath,
	)
	eCmd.Stdin = os.Stdin
	eCmd.Stdout = os.Stdout
	eCmd.Stderr = os.Stderr

	err = eCmd.Run()
	cancel() // shut down daemon

	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return cli.Exit("", exitErr.ExitCode())
		}
		return err
	}

	return nil
}
