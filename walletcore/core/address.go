package core

// #cgo CFLAGS: -I../../../wallet-core/include
// #cgo LDFLAGS: -L../../../wallet-core/build -L../../../wallet-core/build/local/lib -L../../../wallet-core/build/trezor-crypto -lTrustWalletCore -lwallet_core_rs -lprotobuf -lTrezorCrypto -lstdc++ -lm
// #include <TrustWalletCore/TWAnyAddress.h>
import "C"
import (
	"github.com/vultisig/vultisigner/walletcore/types"
)

func GetAnyAddressData(addr string, coinType CoinType) []byte {
	cInputStr := types.TWStringCreateWithGoString(addr)
	defer C.TWStringDelete(cInputStr)
	cAddr := C.TWAnyAddressCreateWithString(cInputStr, C.TWCoinType(coinType))
	defer C.TWAnyAddressDelete(cAddr)
	cData := C.TWAnyAddressData(cAddr)
	defer C.TWDataDelete(cData)
	return types.TWDataGoBytes(cData)
}
