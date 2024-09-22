package core

// #include <TrustWalletCore/TWMnemonic.h>
import "C"

import "github.com/vultisig/vultisigner/walletcore/types"

func IsMnemonicValid(mn string) bool {
	str := types.TWStringCreateWithGoString(mn)
	defer C.TWStringDelete(str)
	return bool(C.TWMnemonicIsValid(str))
}
