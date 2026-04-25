package sandbox

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestSocketPath(t *testing.T) {
	got := SocketPath()
	expected := filepath.Join(os.TempDir(), fmt.Sprintf("enclave-unboxexec-%d.sock", os.Getpid()))
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}
