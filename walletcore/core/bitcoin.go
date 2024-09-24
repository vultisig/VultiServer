package core

// #include <TrustWalletCore/TWCoinType.h>
// #include <TrustWalletCore/TWBitcoinScript.h>
// #include <TrustWalletCore/TWBitcoinSigHashType.h>
import "C"

import (
	"github.com/vultisig/vultisigner/walletcore/types"
)

const (
	BitcoinSigHashTypeAll          = C.TWBitcoinSigHashTypeAll
	BitcoinSigHashTypeNone         = C.TWBitcoinSigHashTypeNone
	BitcoinSigHashTypeSingle       = C.TWBitcoinSigHashTypeSingle
	BitcoinSigHashTypeFork         = C.TWBitcoinSigHashTypeFork
	BitcoinSigHashTypeForkBTG      = C.TWBitcoinSigHashTypeForkBTG
	BitcoinSigHashTypeAnyoneCanPay = C.TWBitcoinSigHashTypeAnyoneCanPay
)

func BitcoinScriptLockScriptForAddress(addr string, ct CoinType) []byte {
	address := types.TWStringCreateWithGoString(addr)
	defer C.TWStringDelete(address)

	script := C.TWBitcoinScriptLockScriptForAddress(address, C.enum_TWCoinType(ct))
	scriptData := C.TWBitcoinScriptData(script)
	defer C.TWBitcoinScriptDelete(script)
	defer C.TWDataDelete(scriptData)

	return types.TWDataGoBytes(scriptData)
}

func BitcoinScriptBuildPayToPublicKeyHash(hash []byte) []byte {
	hashData := types.TWDataCreateWithGoBytes(hash)
	defer C.TWDataDelete(hashData)

	script := C.TWBitcoinScriptBuildPayToPublicKeyHash(hashData)
	scriptData := C.TWBitcoinScriptData(script)
	defer C.TWBitcoinScriptDelete(script)
	defer C.TWDataDelete(scriptData)

	return types.TWDataGoBytes(scriptData)
}

func BitcoinScriptBuildPayToWitnessPublicKeyHash(hash []byte) []byte {
	hashData := types.TWDataCreateWithGoBytes(hash)
	defer C.TWDataDelete(hashData)

	script := C.TWBitcoinScriptBuildPayToWitnessPubkeyHash(hashData)
	scriptData := C.TWBitcoinScriptData(script)
	defer C.TWBitcoinScriptDelete(script)
	defer C.TWDataDelete(scriptData)

	return types.TWDataGoBytes(scriptData)
}

func BitcoinScriptMatchPayToWitnessPublicKeyHash(script []byte) []byte {
	scriptData := types.TWDataCreateWithGoBytes(script)
	defer C.TWDataDelete(scriptData)
	scriptObj := C.TWBitcoinScriptCreateWithData(scriptData)
	defer C.TWBitcoinScriptDelete(scriptObj)

	hash := C.TWBitcoinScriptMatchPayToWitnessPublicKeyHash(scriptObj)
	defer C.TWDataDelete(hash)
	return types.TWDataGoBytes(hash)
}

func BitcoinScriptMatchPayToPublicKeyHash(script []byte) []byte {
	scriptData := types.TWDataCreateWithGoBytes(script)
	defer C.TWDataDelete(scriptData)
	scriptObj := C.TWBitcoinScriptCreateWithData(scriptData)
	defer C.TWBitcoinScriptDelete(scriptObj)

	hash := C.TWBitcoinScriptMatchPayToPubkeyHash(scriptObj)
	defer C.TWDataDelete(hash)
	return types.TWDataGoBytes(hash)
}

func HashTypeForCoin(coinType CoinType) uint32 {
	return uint32(C.TWBitcoinScriptHashTypeForCoin(C.enum_TWCoinType(coinType)))
}
