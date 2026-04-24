package command

import (
	"context"
	"fmt"
	"os"

	"github.com/kohkimakimoto/enclave/v3/internal/config"
	"github.com/kohkimakimoto/enclave/v3/internal/sandbox"
	"github.com/urfave/cli/v3"
)

func ProfileCommand() *cli.Command {
	return &cli.Command{
		Name:   "profile",
		Usage:  "Print evaluated profile and exit",
		Action: profileAction,
	}
}

func profileAction(ctx context.Context, cmd *cli.Command) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	profilePath, cleanup, err := sandbox.BuildProfile(cfg.Sandbox.Profile)
	if err != nil {
		return err
	}
	defer cleanup()

	content, err := os.ReadFile(profilePath)
	if err != nil {
		return fmt.Errorf("failed to read profile: %w", err)
	}

	_, err = cmd.Root().Writer.Write(content)
	return err
}
