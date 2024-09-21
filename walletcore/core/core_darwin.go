//go:build darwin

package core

// #cgo CFLAGS: -I../include
// #cgo LDFLAGS: -L../libs/darwin -lTrustWalletCore -lwallet_core_rs -lprotobuf -lTrezorCrypto -lstdc++ -lm
import "C"
