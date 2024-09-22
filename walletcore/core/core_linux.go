//go:build linux

package core

// #cgo CFLAGS: -I../include
// #cgo LDFLAGS: -L../libs/linux -lTrustWalletCore -lwallet_core_rs -lprotobuf -lTrezorCrypto -lstdc++ -lm
import "C"
