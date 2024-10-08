//go:build darwin
// +build darwin

package types

// #cgo CFLAGS: -I../include
// #cgo LDFLAGS: -L../libs/darwin -lTrustWalletCore -lwallet_core_rs -lprotobuf -lTrezorCrypto -lstdc++ -lm
import "C"
