package command

import (
	"bytes"
	"context"
	"testing"

	"github.com/kohkimakimoto/enclave/v3/internal/version"
)

func TestVersionCommand(t *testing.T) {
	t.Run("via version subcommand", func(t *testing.T) {
		buf := &bytes.Buffer{}
		app := newApp()
		app.Writer = buf

		if err := app.Run(context.Background(), []string{"enclave", "version"}); err != nil {
			t.Fatalf("version command failed: %v", err)
		}

		expected := "enclave version " + version.Version + " (commit: " + version.CommitHash + ")\n"
		if got := buf.String(); got != expected {
			t.Errorf("expected %q, got %q", expected, got)
		}
	})
}
