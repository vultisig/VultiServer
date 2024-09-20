package core

// #cgo CFLAGS: -I../../../wallet-core/include
// #cgo LDFLAGS: -L../../../wallet-core/build -L../../../wallet-core/build/local/lib -L../../../wallet-core/build/trezor-crypto -lTrustWalletCore -lwallet_core_rs -lprotobuf -lTrezorCrypto -lstdc++ -lm
// #include <TrustWalletCore/TWCoinType.h>
// #include <TrustWalletCore/TWCoinTypeConfiguration.h>
import "C"

import "github.com/vultisig/vultisigner/walletcore/types"

type CoinType uint32

const (
	CoinTypeBitcoin     CoinType = C.TWCoinTypeBitcoin
	CoinTypeBitcoinCash CoinType = C.TWCoinTypeBitcoinCash
	CoinTypeLitecoin    CoinType = C.TWCoinTypeLitecoin
	CoinTypeDash        CoinType = C.TWCoinTypeDash
	CoinTypeDogecoin    CoinType = C.TWCoinTypeDogecoin
	CoinTypeZcash       CoinType = C.TWCoinTypeZcash
	CoinTypeKujira      CoinType = C.TWCoinTypeKujira
	CoinTypeBinance     CoinType = C.TWCoinTypeBinance
	CoinTypeEthereum    CoinType = C.TWCoinTypeEthereum
	CoinTypeTron        CoinType = C.TWCoinTypeTron
	CoinTypeTHORChain   CoinType = C.TWCoinTypeTHORChain
	CoinTypeCosmos      CoinType = C.TWCoinTypeCosmos
	CoinTypeSmartChain  CoinType = C.TWCoinTypeSmartChain
	CoinTypeSolana      CoinType = C.TWCoinTypeSolana
	CoinTypePolkadot    CoinType = C.TWCoinTypePolkadot
	CoinTypePolygon     CoinType = C.TWCoinTypePolygon
	CoinTypeSui         CoinType = C.TWCoinTypeSui
)

func (c CoinType) GetName() string {
	name := C.TWCoinTypeConfigurationGetName(C.enum_TWCoinType(c))
	defer C.TWStringDelete(name)
	return types.TWStringGoString(name)
}

func (c CoinType) Decimals() int {
	return int(C.TWCoinTypeConfigurationGetDecimals(C.enum_TWCoinType(c)))
}
func (c CoinType) ChainID() string {
	chainID := C.TWCoinTypeChainId(C.enum_TWCoinType(c))
	defer C.TWStringDelete(chainID)
	return types.TWStringGoString(chainID)
}
