package core

// #include <TrustWalletCore/TWPublicKey.h>
import "C"

import "github.com/vultisig/vultisigner/walletcore/types"

type PublicKeyType uint32

const (
	PublicKeyTypeSECP256k1         PublicKeyType = C.TWPublicKeyTypeSECP256k1
	PublicKeyTypeSECP256k1Extended PublicKeyType = C.TWPublicKeyTypeSECP256k1Extended
)

func PublicKeyVerify(key []byte, keyType PublicKeyType, signature []byte, message []byte) bool {
	keyData := types.TWDataCreateWithGoBytes(key)
	defer C.TWDataDelete(keyData)
	publicKey := C.TWPublicKeyCreateWithData(keyData, C.enum_TWPublicKeyType(keyType))
	defer C.TWPublicKeyDelete(publicKey)
	sig := types.TWDataCreateWithGoBytes(signature)
	defer C.TWDataDelete(sig)
	msg := types.TWDataCreateWithGoBytes(message)
	defer C.TWDataDelete(msg)

	return bool(C.TWPublicKeyVerify(publicKey, sig, msg))
}

func PublicKeyVerifyAsDER(key []byte, keyType PublicKeyType, signature []byte, message []byte) bool {
	keyData := types.TWDataCreateWithGoBytes(key)
	defer C.TWDataDelete(keyData)
	publicKey := C.TWPublicKeyCreateWithData(keyData, C.enum_TWPublicKeyType(keyType))
	defer C.TWPublicKeyDelete(publicKey)
	sig := types.TWDataCreateWithGoBytes(signature)
	defer C.TWDataDelete(sig)
	msg := types.TWDataCreateWithGoBytes(message)
	defer C.TWDataDelete(msg)

	return bool(C.TWPublicKeyVerifyAsDER(publicKey, sig, msg))
}
