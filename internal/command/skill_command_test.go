package command

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/kohkimakimoto/claude-sandbox/v2/internal/skill"
)

func TestSkillCommand(t *testing.T) {
	t.Run("prints skill content to stdout", func(t *testing.T) {
		testChdirTemp(t)

		buf := &bytes.Buffer{}
		cmd := SkillCommand()
		cmd.Writer = buf

		if err := cmd.Run(context.Background(), []string{"skill"}); err != nil {
			t.Fatalf("skill failed: %v", err)
		}

		if got := buf.Bytes(); !bytes.Equal(got, skill.Content) {
			t.Errorf("expected skill content:\n%s\ngot:\n%s", skill.Content, got)
		}
	})

	t.Run("-i installs SKILL.md to .claude/skills/claude-sandbox/SKILL.md", func(t *testing.T) {
		dir := testChdirTemp(t)

		buf := &bytes.Buffer{}
		cmd := SkillCommand()
		cmd.Writer = buf

		if err := cmd.Run(context.Background(), []string{"skill", "-i"}); err != nil {
			t.Fatalf("skill -i failed: %v", err)
		}

		dest := filepath.Join(dir, ".claude", "skills", "claude-sandbox", "SKILL.md")
		got, err := os.ReadFile(dest)
		if err != nil {
			t.Fatalf("failed to read installed SKILL.md: %v", err)
		}
		if !bytes.Equal(got, skill.Content) {
			t.Errorf("installed content mismatch:\nexpected:\n%s\ngot:\n%s", skill.Content, got)
		}
		assertContains(t, buf.String(), filepath.Join(".claude", "skills", "claude-sandbox", "SKILL.md"))
	})

	t.Run("-i creates intermediate directories", func(t *testing.T) {
		dir := testChdirTemp(t)

		cmd := SkillCommand()
		cmd.Writer = &bytes.Buffer{}

		if err := cmd.Run(context.Background(), []string{"skill", "-i"}); err != nil {
			t.Fatalf("skill -i failed: %v", err)
		}

		skillsDir := filepath.Join(dir, ".claude", "skills", "claude-sandbox")
		if _, err := os.Stat(skillsDir); err != nil {
			t.Errorf("expected directory to exist: %s", skillsDir)
		}
	})

	t.Run("--install is equivalent to -i", func(t *testing.T) {
		dir := testChdirTemp(t)

		cmd := SkillCommand()
		cmd.Writer = &bytes.Buffer{}

		if err := cmd.Run(context.Background(), []string{"skill", "--install"}); err != nil {
			t.Fatalf("skill --install failed: %v", err)
		}

		dest := filepath.Join(dir, ".claude", "skills", "claude-sandbox", "SKILL.md")
		if _, err := os.Stat(dest); err != nil {
			t.Errorf("expected SKILL.md to exist at %s", dest)
		}
	})
}
