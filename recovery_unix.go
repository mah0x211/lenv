//go:build linux || darwin || freebsd
// +build linux darwin freebsd

package main

import (
	"errors"
	"os"
	"syscall"
)

// These errors establish an invalid reference. Permission and I/O errors do not.
func isInvalidCurrentLink(err error) bool {
	return os.IsNotExist(err) || errors.Is(err, syscall.ENOTDIR) ||
		errors.Is(err, syscall.ELOOP) || err.Error() == "EvalSymlinks: too many links"
}
