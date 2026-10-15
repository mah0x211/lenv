//go:build !linux && !darwin && !freebsd
// +build !linux,!darwin,!freebsd

package main

import (
	"os"
	"strings"
)

func isInvalidCurrentLink(err error) bool {
	return os.IsNotExist(err) || strings.Contains(err.Error(), "EvalSymlinks: too many links")
}
