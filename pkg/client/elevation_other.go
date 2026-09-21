//go:build !windows

package client

import "os"

func isProcessElevated() bool {
	return os.Geteuid() == 0
}
