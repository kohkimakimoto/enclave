package sandbox

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// GetWorkdir returns the working directory for sandbox execution.
// If configWorkdir is non-empty, it is used. Otherwise falls back to the current directory.
func GetWorkdir(configWorkdir string) string {
	if configWorkdir != "" {
		return configWorkdir
	}
	wd, _ := os.Getwd()
	return wd
}

// GetClaudeBin returns the path to the claude binary.
// If configClaudeBin is non-empty, it is used. Otherwise searches PATH, then falls back to ~/.claude/local/claude.
func GetClaudeBin(configClaudeBin string) string {
	if configClaudeBin != "" {
		return configClaudeBin
	}

	if p, err := exec.LookPath("claude"); err == nil {
		return p
	}

	home, _ := os.UserHomeDir()
	localClaude := filepath.Join(home, ".claude", "local", "claude")
	if _, err := os.Stat(localClaude); err == nil {
		return localClaude
	}

	return "claude"
}

// SocketPath returns the path for the daemon's Unix Domain Socket.
func SocketPath() string {
	return filepath.Join(os.TempDir(), fmt.Sprintf("enclave-unboxexec-%d.sock", os.Getpid()))
}
