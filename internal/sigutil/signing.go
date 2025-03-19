package sigutil

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/eager7/dogd/btcec"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/vultisig/mobile-tss-lib/tss"
)

func SignLegacyTx(keysignResponse tss.KeysignResponse, txHash string, rawTx string, chainID *big.Int) (*types.Transaction, *common.Address, error) {
	unsignedTxBytes, err := hex.DecodeString(rawTx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode raw transaction: %w", err)
	}

	unsignedTx := new(types.Transaction)
	if err := unsignedTx.UnmarshalBinary(unsignedTxBytes); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal unsigned transaction: %w", err)
	}

	r, ok := new(big.Int).SetString(keysignResponse.R, 16)
	if !ok {
		return nil, nil, fmt.Errorf("failed to parse R")
	}

	s, ok := new(big.Int).SetString(keysignResponse.S, 16)
	if !ok {
		return nil, nil, fmt.Errorf("failed to parse S")
	}

	recID, err := strconv.ParseInt(keysignResponse.RecoveryID, 10, 8)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse recovery ID: %w", err)
	}
	recoveryID := uint8(recID) // 0 or 1

	// Manually reconstruct the unsigned transaction to ensure consistency
	tx := types.NewTransaction(
		unsignedTx.Nonce(),
		*unsignedTx.To(),
		unsignedTx.Value(),
		unsignedTx.Gas(),
		unsignedTx.GasPrice(),
		unsignedTx.Data(),
	)

	signer := types.NewEIP155Signer(chainID)
	fmt.Println("Raw signature hex:", hex.EncodeToString(rawSignature(r, s, recoveryID)))
	signedTx, err := tx.WithSignature(signer, rawSignature(r, s, recoveryID))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to attach signature: %w", err)
	}

	// recover the sender's address
	sender, err := signer.Sender(signedTx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to recover sender: %w", err)
	}

	return signedTx, &sender, nil
}

func VerifySignature(vaultPublicKey string, chainCodeHex string, derivePath string, messageHex []byte, signature []byte) (bool, error) {
	msgHash := crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(messageHex), messageHex)))

	derivedPubKeyHex, err := tss.GetDerivedPubKey(strings.TrimPrefix(vaultPublicKey, "0x"), chainCodeHex, derivePath, false)
	if err != nil {
		return false, err
	}

	publicKeyBytes, err := hex.DecodeString(derivedPubKeyHex)
	if err != nil {
		return false, err
	}

	pk, err := btcec.ParsePubKey(publicKeyBytes, btcec.S256())
	if err != nil {
		return false, err
	}

	ecdsaPubKey := ecdsa.PublicKey{
		Curve: btcec.S256(),
		X:     pk.X,
		Y:     pk.Y,
	}
	R := new(big.Int).SetBytes(signature[:32])
	S := new(big.Int).SetBytes(signature[32:64])

	return ecdsa.Verify(&ecdsaPubKey, msgHash, R, S), nil
}

func rawSignature(r *big.Int, s *big.Int, recoveryID uint8) []byte {
	var signature [65]byte
	copy(signature[0:32], r.Bytes())
	copy(signature[32:64], s.Bytes())
	signature[64] = byte(recoveryID)
	return signature[:]
}
