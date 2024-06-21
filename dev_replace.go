//go:build dev
// +build dev

package main

import (
	_ "github.com/vultisig/mobile-tss-lib"
)

func init() {
	// This code is only included in the build when the 'dev' tag is specified.
	_ = `go:generate go mod edit -replace=github.com/vultisig/mobile-tss-lib=../mobile-tss-lib`
}
