package command

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kohkimakimoto/claude-sandbox/v2/internal/skill"
	"github.com/urfave/cli/v3"
)

func SkillCommand() *cli.Command {
	return &cli.Command{
		Name:  "skill",
		Usage: "Print the Claude Code skill definition and exit",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "install",
				Aliases: []string{"i"},
				Usage:   "Install the skill to .claude/skills/claude-sandbox/SKILL.md",
			},
		},
		Action: skillAction,
	}
}

func skillAction(ctx context.Context, cmd *cli.Command) error {
	if cmd.Bool("install") {
		return skillInstall(cmd)
	}
	_, err := cmd.Root().Writer.Write(skill.Content)
	return err
}

func skillInstall(cmd *cli.Command) error {
	dest := filepath.Join(".claude", "skills", "claude-sandbox", "SKILL.md")

	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(dest, skill.Content, 0644); err != nil {
		return fmt.Errorf("failed to write skill: %w", err)
	}

	fmt.Fprintf(cmd.Root().Writer, "Installed skill to: %s\n", dest)
	return nil
}
