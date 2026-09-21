//go:build windows

package client

import (
	"golang.org/x/sys/windows"
)

func isProcessElevated() bool {
	token := windows.GetCurrentProcessToken()
	return token.IsElevated()
}
