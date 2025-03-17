package main

import (
	"context"
	"crypto/ecdsa"
	"flag"
	"fmt"
	"math/big"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	gcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/vultisig/vultisigner/common"
	"github.com/vultisig/vultisigner/config"
)

const (
	derivePath   = "m/44'/60'/0'/0/0"
	hexChainCode = "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
)

var (
	privateKey = os.Getenv("PRIVATE_KEY")
)

var vaultName string
var stateDir string

func main() {
	flag.StringVar(&vaultName, "vault", "", "vault name")
	flag.StringVar(&stateDir, "state-dir", "", "state directory")
	flag.Parse()

	if vaultName == "" {
		panic("vault name is required")
	}

	if stateDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}

		stateDir = filepath.Join(homeDir, ".verifier", "vaults")
	}

	keyPath := filepath.Join("vaults", vaultName, "public_key")
	rawKeyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		panic(err)
	}
	compressedPubKeyHex := string(rawKeyBytes)

	vaultAddress, err := common.DeriveAddress(compressedPubKeyHex, hexChainCode, derivePath)
	if err != nil {
		panic(err)
	}
	fmt.Println("To vault address:", vaultAddress.Hex())

	pluginConfig, err := config.ReadConfig("config-plugin")
	if err != nil {
		panic(err)
	}

	rpcClient, err := ethclient.Dial(pluginConfig.Server.Plugin.Eth.Rpc)
	if err != nil {
		panic(err)
	}

	signerPrivateKey, _, signerAddress, err := toKeysAndAddress(privateKey)
	if err != nil {
		panic(err)
	}
	// 1 eth
	amount := big.NewInt(9e18)

	gasPrice, err := rpcClient.SuggestGasPrice(context.Background())
	if err != nil {
		panic(err)
	}
	nonce, err := rpcClient.PendingNonceAt(context.Background(), signerAddress)
	if err != nil {
		panic(err)
	}
	gasLimit, err := rpcClient.EstimateGas(context.Background(), ethereum.CallMsg{})
	if err != nil {
		panic(err)
	}
	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       vaultAddress,
		Value:    amount,
		Gas:      gasLimit,
		GasPrice: gasPrice,
		Data:     nil,
	})

	chainID, err := rpcClient.NetworkID(context.Background())
	if err != nil {
		panic(err)
	}
	signer := types.NewEIP155Signer(chainID)
	signedTx, err := types.SignTx(tx, signer, signerPrivateKey)
	if err != nil {
		panic(err)
	}

	sender, err := signer.Sender(signedTx)
	if err != nil {
		panic("failed to get sender: " + err.Error())
	}
	fmt.Println("From sender address:", sender.Hex())

	err = rpcClient.SendTransaction(context.Background(), signedTx)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Transaction sent: %s", signedTx.Hash().Hex())

	receipt, err := bind.WaitMined(context.Background(), rpcClient, signedTx)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Transaction receipt status: %v", receipt.Status)
}

func toKeysAndAddress(privateKeyHex string) (*ecdsa.PrivateKey, *ecdsa.PublicKey, gcommon.Address, error) {
	if privateKeyHex == "" {
		return nil, nil, gcommon.Address{}, fmt.Errorf("private key is not set")
	}
	ecdsaPrivateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, nil, gcommon.Address{}, fmt.Errorf("failed to get private key: %v", err)
	}
	ecdsaPublicKey := ecdsaPrivateKey.Public().(*ecdsa.PublicKey)

	address := crypto.PubkeyToAddress(*ecdsaPublicKey)

	return ecdsaPrivateKey, ecdsaPublicKey, address, nil
}
