package command

import (
	"context"

	"github.com/kohkimakimoto/enclave/v3/internal/config"
	"github.com/kohkimakimoto/enclave/v3/internal/version"
	"github.com/urfave/cli/v3"
)

func Run(args []string) error {
	return newApp().Run(context.Background(), args)
}

func newApp() *cli.Command {
	return &cli.Command{
		Name:            "enclave",
		Usage:           "A wrapper around the claude command to run it in a sandboxed environment.",
		Copyright:       "Copyright (c) Kohki Makimoto",
		HideVersion:     true,
		Version:         version.Version,
		ExtraInfo:       func() map[string]string { return map[string]string{"CommitHash": version.CommitHash} },
		SkipFlagParsing: true,
		Commands: []*cli.Command{
			InitCommand(),
			InitLocalCommand(),
			InitUserCommand(),
			InitGlobalCommand(),
			ConfigCommand(),
			SkillCommand(),
			ProfileCommand(),
			VersionCommand(),
			ClaudeCommand(),
			UnboxexecCommand(),
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Args().Present() {
				first := cmd.Args().First()
				if first == "help" || first == "--help" || first == "-h" {
					return cli.ShowAppHelp(cmd)
				}
				if first == "-v" || first == "--version" {
					return versionAction(ctx, cmd)
				}
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}
			// If args are present and not a builtin command, run claude with all args
			return RunClaudeAction(ctx, cmd, cmd.Args().Slice(), cfg)
		},
	}
}
