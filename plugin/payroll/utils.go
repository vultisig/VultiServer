package payroll

import (
	"encoding/hex"

	"github.com/btcsuite/btcd/btcec/v2"
	gcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/vultisig/mobile-tss-lib/tss"
)

func DeriveAddress(compressedPubKeyHex, hexChainCode, derivePath string) (*gcommon.Address, error) {
	derivedPubKeyHex, err := tss.GetDerivedPubKey(compressedPubKeyHex, hexChainCode, derivePath, false)
	if err != nil {
		return nil, err
	}

	derivedPubKeyBytes, err := hex.DecodeString(derivedPubKeyHex)
	if err != nil {
		return nil, err
	}

	derivedPubKey, err := btcec.ParsePubKey(derivedPubKeyBytes)
	if err != nil {
		return nil, err
	}

	uncompressedPubKeyBytes := derivedPubKey.SerializeUncompressed()
	pubKeyBytesWithoutPrefix := uncompressedPubKeyBytes[1:]
	hash := crypto.Keccak256(pubKeyBytesWithoutPrefix)
	address := gcommon.BytesToAddress(hash[12:])

	return &address, nil
}
