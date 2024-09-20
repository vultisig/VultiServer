package core

// #cgo CFLAGS: -I../../../wallet-core/include
// #cgo LDFLAGS: -L../../../wallet-core/build -L../../../wallet-core/build/local/lib -L../../../wallet-core/build/trezor-crypto -lTrustWalletCore -lwallet_core_rs -lprotobuf -lTrezorCrypto -lstdc++ -lm
// #include <TrustWalletCore/TWMnemonic.h>
import "C"

import "github.com/vultisig/vultisigner/walletcore/types"

func IsMnemonicValid(mn string) bool {
	str := types.TWStringCreateWithGoString(mn)
	defer C.TWStringDelete(str)
	return bool(C.TWMnemonicIsValid(str))
}
