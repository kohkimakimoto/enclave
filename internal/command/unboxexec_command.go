package command

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/kohkimakimoto/enclave/v3/internal/unboxexec"
	"github.com/urfave/cli/v3"
)

func UnboxexecCommand() *cli.Command {
	return &cli.Command{
		Name:  "unboxexec",
		Usage: "Execute a command outside the sandbox via the unboxexec daemon",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "dir",
				Aliases: []string{"C"},
				Usage:   "Working directory for the command",
			},
			&cli.IntFlag{
				Name:    "timeout",
				Aliases: []string{"t"},
				Usage:   "Timeout in seconds (default: 60)",
			},
			&cli.StringSliceFlag{
				Name:    "env",
				Aliases: []string{"e"},
				Usage:   "Environment variable in KEY=VALUE format (can be specified multiple times)",
			},
		},
		Action: unboxexecAction,
	}
}

func unboxexecAction(ctx context.Context, cmd *cli.Command) error {
	sockPath := os.Getenv("ENCLAVE_UNBOXEXEC_SOCK")
	if sockPath == "" {
		return fmt.Errorf("ENCLAVE_UNBOXEXEC_SOCK is not set")
	}

	args := cmd.Args().Slice()
	if len(args) == 0 {
		return fmt.Errorf("no command specified")
	}

	// Build environment variables map from --env flags
	var envMap map[string]string
	envSlice := cmd.StringSlice("env")
	if len(envSlice) > 0 {
		envMap = make(map[string]string, len(envSlice))
		for _, e := range envSlice {
			k, v, ok := strings.Cut(e, "=")
			if !ok {
				return fmt.Errorf("invalid env format (expected KEY=VALUE): %s", e)
			}
			envMap[k] = v
		}
	}

	// Default working directory to current directory
	dir := cmd.String("dir")
	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	req := &unboxexec.ExecRequest{
		Command: args[0],
		Args:    args[1:],
		Env:     envMap,
		Dir:     dir,
		Timeout: int(cmd.Int("timeout")),
	}

	resp, err := unboxexec.SendRequest(sockPath, req)
	if err != nil {
		return err
	}

	if resp.Stdout != "" {
		fmt.Fprint(os.Stdout, resp.Stdout)
	}
	if resp.Stderr != "" {
		fmt.Fprint(os.Stderr, resp.Stderr)
	}

	if resp.Error != "" {
		return fmt.Errorf("daemon error: %s", resp.Error)
	}

	if resp.ExitCode != 0 {
		return cli.Exit("", resp.ExitCode)
	}

	return nil
}
