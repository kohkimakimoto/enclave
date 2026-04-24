package command

import (
	"context"

	"github.com/kohkimakimoto/enclave/v3/internal/version"
	"github.com/urfave/cli/v3"
)

func Run(args []string) error {
	return newApp().Run(context.Background(), args)
}

func newApp() *cli.Command {
	return &cli.Command{
		Name:        "enclave",
		Usage:       "Run any command in a sandboxed environment.",
		Copyright:   "Copyright (c) Kohki Makimoto",
		HideVersion: true,
		Version:     version.Version,
		ExtraInfo:   func() map[string]string { return map[string]string{"CommitHash": version.CommitHash} },
		Commands: []*cli.Command{
			RunCommand(),
			InitCommand(),
			InitLocalCommand(),
			InitUserCommand(),
			InitGlobalCommand(),
			ConfigCommand(),
			SkillCommand(),
			ProfileCommand(),
			VersionCommand(),
			UnboxexecCommand(),
		},
	}
}
