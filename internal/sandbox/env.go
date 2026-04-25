package sandbox

import (
	"fmt"
	"os"
	"path/filepath"
)

// SocketPath returns the path for the daemon's Unix Domain Socket.
func SocketPath() string {
	return filepath.Join(os.TempDir(), fmt.Sprintf("enclave-unboxexec-%d.sock", os.Getpid()))
}

// ConfigDumpPath returns the path for the effective config dump file.
func ConfigDumpPath() string {
	return filepath.Join(os.TempDir(), fmt.Sprintf("enclave-config-%d.toml", os.Getpid()))
}
