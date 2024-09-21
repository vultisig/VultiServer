package core

// #include <TrustWalletCore/TWAnyAddress.h>
// #include <TrustWalletCore/TWCoinType.h>
import "C"
import (
	"github.com/vultisig/vultisigner/walletcore/types"
)

func GetAnyAddressData(addr string, coinType CoinType) []byte {
	cInputStr := types.TWStringCreateWithGoString(addr)
	defer C.TWStringDelete(cInputStr)
	cAddr := C.TWAnyAddressCreateWithString(cInputStr, C.enum_TWCoinType(coinType))
	defer C.TWAnyAddressDelete(cAddr)
	cData := C.TWAnyAddressData(cAddr)
	defer C.TWDataDelete(cData)
	return types.TWDataGoBytes(cData)
}
