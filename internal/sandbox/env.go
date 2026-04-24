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
