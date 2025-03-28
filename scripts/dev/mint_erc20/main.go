// Package main provides a CLI tool for USDC operations: deposit, balance check, and transfer
package main

import (
	"context"
	"crypto/ecdsa"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	gcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Constants
const (
	// Default USDC contract address on Ethereum mainnet

	derivePath   = "m/44'/60'/0'/0/0"
	hexChainCode = "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

	// USDC decimals (USDC uses 6 decimals unlike most ERC20 tokens that use 18)
	WETHDecimals = 18

	// Gas limits
	DefaultGasLimit = 1000000
	GasLimitBuffer  = 50000

	DeadlineDuration = 15 * time.Minute
)

// Contract ABIs
const (
	// WETH ABI with deposit and transfer functions
	WETHABI = `[
		{
			"constant": false,
			"inputs": [],
			"name": "deposit",
			"outputs": [],
			"payable": true,
			"stateMutability": "payable",
			"type": "function"
		},
		{
			"constant": false,
			"inputs": [
				{
					"name": "dst",
					"type": "address"
				},
				{
					"name": "wad",
					"type": "uint256"
				}
			],
			"name": "transfer",
			"outputs": [
				{
					"name": "",
					"type": "bool"
				}
			],
			"payable": false,
			"stateMutability": "nonpayable",
			"type": "function"
		},
		{
			"constant": true,
			"inputs": [
				{
					"name": "",
					"type": "address"
				}
			],
			"name": "balanceOf",
			"outputs": [
				{
					"name": "",
					"type": "uint256"
				}
			],
			"payable": false,
			"stateMutability": "view",
			"type": "function"
		}
	]`
)

var (
	uniswapV2RouterAddress = gcommon.HexToAddress("0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D")
	WETHAddr               = gcommon.HexToAddress("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2")
	swapAmountIn           = big.NewInt(9e18)
	tokenAddress           string
	vaultAddressHex        string
)

func main() {

	flag.StringVar(&tokenAddress, "token", "", "token address")
	flag.StringVar(&vaultAddressHex, "vault-address", "", "vault address")
	flag.Parse()

	if vaultAddressHex == "" {
		panic("vault address is required")
	}

	if tokenAddress == "" {
		fmt.Println("Token Address is defaulting to WETH")
		tokenAddress = WETHAddr.Hex()
	}
	fmt.Println("To vault address:", vaultAddressHex)

	// Get private key from environment variable
	privateKey := os.Getenv("PRIVATE_KEY")
	if privateKey == "" {
		panic("PRIVATE_KEY environment variable not set")
	}

	// Get RPC URL from environment variable
	rpcURL := os.Getenv("RPC_URL")
	if rpcURL == "" {
		panic("RPC_URL not set, using default")
	}

	// Connect to Ethereum client
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		fatalError("Failed to connect to Ethereum node", err)
	}

	// Get signer's keys and address
	signerPrivateKey, signerAddress, err := getKeysAndAddress(privateKey)
	if err != nil {
		fatalError("Failed to get keys from private key", err)
	}

	fmt.Println("Connected to Ethereum network")
	fmt.Printf("From address: %s\n", signerAddress.Hex())
	fmt.Printf("To receive WETH and then transfer to: %s\n", vaultAddressHex)
	fmt.Printf("WETH contract: %s\n", WETHAddr)
	fmt.Printf("Operation: Depositing %s ETH to get WETH\n", swapAmountIn.String())

	// Step 1: Call deposit function to receive USDC
	fmt.Println("\n=== Step 1: MINTING WETH ===")
	err = MintWETH(
		client,
		signerAddress,
		signerPrivateKey,
		swapAmountIn,
	)
	if err != nil {
		fatalError("Failed to MINT WETH", err)
	}

	// Step 2: Check USDC balance
	fmt.Println("\n=== Step 2: Checking WETH Balance ===")
	balance, err := GetTokenBalance(WETHAddr, signerAddress, client)
	if err != nil {
		fatalError("Failed to get USDC balance", err)
	}
	fmt.Printf("WETH Balance: %s (with %d decimals)\n", balance.String(), WETHDecimals)

	if tokenAddress == WETHAddr.Hex() {
		fmt.Println("\n===Transferring TOKEN to Recipient ===")
		transferTxHash, err := transferWETH(
			client,
			signerPrivateKey,
			signerAddress,
			gcommon.HexToAddress(vaultAddressHex),
			WETHAddr.Hex(),
			balance, // Transfer entire balance
		)
		if err != nil {
			fatalError("Failed to transfer USDC", err)
		}
		fmt.Printf("Transfer transaction sent: %s\n", transferTxHash)

		// Give blockchain time to process transaction
		fmt.Println("Waiting for transfer transaction to be mined...")
		time.Sleep(15 * time.Second)

		// Step 4: Check balances after transfer
		fmt.Println("\n=== Step 4: Checking Balances After Transfer ===")
		senderBalance, err := GetTokenBalance(WETHAddr, signerAddress, client)
		if err != nil {
			fmt.Printf("Warning: Failed to get sender balance: %v\n", err)
		} else {
			fmt.Printf("Sender WETH Balance: %s\n", senderBalance.String())
		}

		recipientBalance, err := GetTokenBalance(WETHAddr, gcommon.HexToAddress(vaultAddressHex), client)
		if err != nil {
			fmt.Printf("Warning: Failed to get recipient balance: %v\n", err)
		} else {
			fmt.Printf("Recipient WETH Balance: %s\n", recipientBalance.String())
		}

		fmt.Println("\nOperation completed successfully!")
		os.Exit(0)
	}

	// If token address is not nil we will swap the WETH for the desired token and then transfer it to the vault address
	tokenAddr := gcommon.HexToAddress(tokenAddress)

	fmt.Println("Approving Uniswap Router to spend ", tokenAddr.Hex())
	err = ApproveERC20Token(WETHAddr, uniswapV2RouterAddress, swapAmountIn, client, signerAddress, signerPrivateKey)
	if err != nil {
		fatalError("Failed to approve uniswap v2 router", err)
	}

	fmt.Println("SWAP TOKENS")
	tokensPair := []gcommon.Address{WETHAddr, tokenAddr}
	expectedAmountOut, err := GetExpectedAmountOut(swapAmountIn, tokensPair, client)
	if err != nil {
		log.Fatalf("Failed to get expected amount out: %v", err)
	}
	log.Println("Expected amount out:", expectedAmountOut.String())

	// calculate output amount with slippage
	slippagePercentage := 5.0
	amountOutMin := CalculateAmountOutMin(expectedAmountOut, slippagePercentage)

	err = SwapTokens(swapAmountIn, amountOutMin, tokensPair, client, signerAddress, signerPrivateKey)
	if err != nil {
		fatalError("Failed to swap tokens", err)
	}
	fmt.Println("Balance of Token after swap")

	fmt.Println("Transfer Tokens to VAULT Address")
	tokenBalance, err := GetTokenBalance(tokenAddr, signerAddress, client)
	if err != nil {
		fatalError("Failed to get token balance", err)
	}
	fmt.Println("Signer Token Balance Before SWAP: ", tokenBalance.String())
	vaultTokenBalance, err := GetTokenBalance(tokenAddr, gcommon.HexToAddress(vaultAddressHex), client)
	if err != nil {
		fatalError("Failed to get Vault token balance", err)
	}
	fmt.Println("Vault token balance BEFORE SWAP : ", vaultTokenBalance.String())
	if err != nil {
		fatalError("Failed to approve USDC to vault", err)
	}

	err = TransferERC20Token(tokenAddr, tokenBalance, gcommon.HexToAddress(vaultAddressHex), client, signerAddress, signerPrivateKey)
	if err != nil {
		fatalError("Failed to transfer ERC20 to vault", err)
	}

	vaultTokenBalance, err = GetTokenBalance(tokenAddr, gcommon.HexToAddress(vaultAddressHex), client)
	if err != nil {
		fatalError("Failed to get Vault token balance", err)
	}
	fmt.Println("Vault token balance AFTER SWAP: ", vaultTokenBalance.String())
	tokenBalance, err = GetTokenBalance(tokenAddr, signerAddress, client)
	if err != nil {
		fatalError("Failed to get Signer token balance", err)
	}
	fmt.Println("Signer Token Balance AFTER SWAP: ", tokenBalance.String())
}
func CalculateAmountOutMin(expectedAmountOut *big.Int, slippagePercentage float64) *big.Int {
	slippageFactor := big.NewFloat(1 - slippagePercentage/100)
	expectedAmountOutFloat := new(big.Float).SetInt(expectedAmountOut)
	amountOutMinFloat := new(big.Float).Mul(expectedAmountOutFloat, slippageFactor)
	amountOutMin, _ := amountOutMinFloat.Int(nil)
	return amountOutMin
}

func GetExpectedAmountOut(amountIn *big.Int, path []gcommon.Address, client *ethclient.Client) (*big.Int, error) {
	routerABI := `[
		{
			"name": "getAmountsOut",
			"type": "function",
			"inputs": [
				{
					"name": "amountIn",
					"type": "uint256"
				},
				{
					"name": "path",
					"type": "address[]"
				}
			],
			"outputs": [
				{
					"name": "",
					"type": "uint256[]"
				}
			]
		}
	]`
	parsedABI, err := abi.JSON(strings.NewReader(routerABI))
	if err != nil {
		return nil, err
	}

	callData, err := parsedABI.Pack("getAmountsOut", amountIn, path)
	if err != nil {
		return nil, err
	}

	msg := ethereum.CallMsg{
		To:   &uniswapV2RouterAddress,
		Data: callData,
	}

	result, err := client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return nil, err
	}

	var amountsOut []*big.Int
	err = parsedABI.UnpackIntoInterface(&amountsOut, "getAmountsOut", result)
	if err != nil {
		return nil, err
	}

	if len(amountsOut) < 2 {
		return nil, fmt.Errorf("unexpected result length")
	}

	return amountsOut[len(amountsOut)-1], nil
}

// depositUSDC calls the deposit function on the USDC contract
func MintWETH(
	client *ethclient.Client,
	signerAddress gcommon.Address,
	privateKey *ecdsa.PrivateKey,
	amount *big.Int,
) error {
	wethABI := `[{"name":"deposit","type":"function","payable":true}]`
	parsedABI, err := abi.JSON(strings.NewReader(wethABI))
	if err != nil {
		return err
	}

	data, err := parsedABI.Pack("deposit")
	if err != nil {
		return err
	}
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return err
	}
	nonce, err := client.PendingNonceAt(context.Background(), signerAddress)
	if err != nil {
		return err
	}
	gasLimit, err := client.EstimateGas(context.Background(), ethereum.CallMsg{
		To:   &WETHAddr,
		Data: data,
	})
	if err != nil {
		return err
	}
	gasLimit += GasLimitBuffer
	tx := types.NewTransaction(nonce, WETHAddr, amount, gasLimit, gasPrice, data)
	return sendTransaction(tx, client, privateKey)
}

// getTokenBalance retrieves the token balance for an address
func GetTokenBalance(tokenAddress gcommon.Address, accountAddress gcommon.Address, client *ethclient.Client) (*big.Int, error) {
	tokenABI := `[
		{
			"name": "balanceOf",
			"type": "function",
			"inputs": [
				{
					"name": "account",
					"type": "address"
				}
			],
			"outputs": [
				{
					"name": "",
					"type": "uint256"
				}
			]
		}
	]`
	parsedABI, err := abi.JSON(strings.NewReader(tokenABI))
	if err != nil {
		return nil, err
	}
	callData, err := parsedABI.Pack("balanceOf", accountAddress)
	if err != nil {
		return nil, err
	}

	msg := ethereum.CallMsg{
		To:   &tokenAddress,
		Data: callData,
	}

	result, err := client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return nil, err
	}

	balance := new(big.Int)
	balance.SetBytes(result)
	return balance, nil
}

// transferUSDC transfers USDC tokens to a recipient
func transferWETH(
	client *ethclient.Client,
	privateKey *ecdsa.PrivateKey,
	from gcommon.Address,
	to gcommon.Address,
	tokenContract string,
	amount *big.Int,
) (string, error) {
	// Parse transfer ABI
	parsedABI, err := abi.JSON(strings.NewReader(WETHABI))
	if err != nil {
		return "", fmt.Errorf("failed to parse transfer ABI: %w", err)
	}

	// Create token address from string
	tokenAddress := gcommon.HexToAddress(tokenContract)

	// Pack transfer function with recipient and amount
	data, err := parsedABI.Pack("transfer", to, amount)
	if err != nil {
		return "", fmt.Errorf("failed to pack transfer data: %w", err)
	}

	// Get chain ID
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return "", fmt.Errorf("failed to get network ID: %w", err)
	}

	// Get nonce
	nonce, err := client.PendingNonceAt(context.Background(), from)
	if err != nil {
		return "", fmt.Errorf("failed to get nonce: %w", err)
	}

	// Get gas price
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", fmt.Errorf("failed to get gas price: %w", err)
	}

	// Estimate gas
	gasLimit, err := client.EstimateGas(context.Background(), ethereum.CallMsg{
		From: from,
		To:   &tokenAddress,
		Data: data,
	})

	// Use default gas limit if estimation fails
	if err != nil {
		fmt.Printf("Gas estimation failed: %v. Using default gas limit.\n", err)
		gasLimit = DefaultGasLimit
	}

	// Add buffer to gas limit
	gasLimit += GasLimitBuffer

	// Create transaction
	tx := types.NewTransaction(
		nonce,
		tokenAddress,
		big.NewInt(0), // No ETH value is being sent
		gasLimit,
		gasPrice,
		data,
	)

	// Sign transaction
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Send transaction
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", fmt.Errorf("failed to send transaction: %w", err)
	}

	return signedTx.Hash().Hex(), nil
}

// getKeysAndAddress converts a private key hex string to keys and address
func getKeysAndAddress(privateKeyHex string) (*ecdsa.PrivateKey, gcommon.Address, error) {
	ecdsaPrivateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, gcommon.Address{}, fmt.Errorf("failed to parse private key: %w", err)
	}

	address := crypto.PubkeyToAddress(ecdsaPrivateKey.PublicKey)

	return ecdsaPrivateKey, address, nil
}

// fatalError prints an error message and exits
func fatalError(message string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s: %v\n", message, err)
	} else {
		fmt.Fprintf(os.Stderr, "Error: %s\n", message)
	}
	os.Exit(1)
}

func ApproveERC20Token(
	tokenAddress,
	spenderAddress gcommon.Address,
	amount *big.Int,
	client *ethclient.Client,
	signerAddress gcommon.Address,
	signerPrivateKey *ecdsa.PrivateKey) error {
	tokenABI := `[
		{
			"name": "approve",
			"type": "function",
			"inputs": [
				{
					"name": "spender",
					"type": "address"
				},
				{
					"name": "value",
					"type": "uint256"
				}
			],
			"outputs": [
				{
					"name": "",
					"type": "bool"
				}
			]
		}
	]`
	parsedABI, err := abi.JSON(strings.NewReader(tokenABI))
	if err != nil {
		return err
	}
	approveData, err := parsedABI.Pack("approve", spenderAddress, amount)
	if err != nil {
		return err
	}
	nonce, err := client.PendingNonceAt(context.Background(), signerAddress)
	if err != nil {
		return err
	}
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return err
	}
	fmt.Println("TOKEN APPROVING: ", tokenAddress)
	gasLimit, err := client.EstimateGas(context.Background(), ethereum.CallMsg{
		From: signerAddress,
		To:   &tokenAddress,
		Data: approveData,
	})
	if err != nil {
		return err
	}
	gasLimit += GasLimitBuffer
	tx := types.NewTransaction(nonce, tokenAddress, big.NewInt(0), gasLimit, gasPrice, approveData)
	return sendTransaction(tx, client, signerPrivateKey)
}

func SwapTokens(
	amountIn,
	amountOutMin *big.Int,
	path []gcommon.Address,
	client *ethclient.Client,
	signerAddress gcommon.Address,
	signerPrivateKey *ecdsa.PrivateKey) error {

	fmt.Println("SWAPPING TOKENS")
	routerABI := `[
		{
			"name": "swapExactTokensForTokens",
			"type": "function",
			"inputs": [
				{
					"name": "amountIn",
					"type": "uint256"
				},
				{
					"name": "amountOutMin",
					"type": "uint256"
				},
				{
					"name": "path",
					"type": "address[]"
				},
				{
					"name": "to",
					"type": "address"
				},
				{
					"name": "deadline",
					"type": "uint256"
				}
			]
		}
	]`
	parsedRouterABI, err := abi.JSON(strings.NewReader(routerABI))
	if err != nil {
		return err
	}
	deadline := big.NewInt(time.Now().Add(DeadlineDuration).Unix())

	swapData, err := parsedRouterABI.Pack("swapExactTokensForTokens", amountIn, amountOutMin, path, signerAddress, deadline)
	if err != nil {
		return err
	}
	nonce, err := client.PendingNonceAt(context.Background(), signerAddress)
	if err != nil {
		return err
	}
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return err
	}
	tx := types.NewTransaction(nonce, uniswapV2RouterAddress, big.NewInt(0), DefaultGasLimit, gasPrice, swapData)
	return sendTransaction(tx, client, signerPrivateKey)
}

func TransferERC20Token(
	tokenAddress gcommon.Address,
	amount *big.Int,
	vaultAddress gcommon.Address,
	client *ethclient.Client,
	signerAddress gcommon.Address,
	signerPrivateKey *ecdsa.PrivateKey) error {
	tokenABI := `[
        {
            "constant": false,
            "inputs": [
                {
                    "name": "to",
                    "type": "address"
                },
                {
                    "name": "value",
                    "type": "uint256"
                }
            ],
            "name": "transfer",
            "outputs": [
                {
                    "name": "",
                    "type": "bool"
                }
            ],
            "payable": false,
            "stateMutability": "nonpayable",
            "type": "function"
        }
    ]`

	// Parse transfer ABI
	parsedABI, err := abi.JSON(strings.NewReader(tokenABI))
	if err != nil {
		return fmt.Errorf("failed to parse transfer ABI: %w", err)
	}

	// Pack transfer function with recipient and amount
	data, err := parsedABI.Pack("transfer", vaultAddress, amount)
	if err != nil {
		return fmt.Errorf("failed to pack transfer data: %w", err)
	}

	// Get nonce
	nonce, err := client.PendingNonceAt(context.Background(), signerAddress)
	if err != nil {
		return fmt.Errorf("failed to get nonce: %w", err)
	}

	// Get gas price
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get gas price: %w", err)
	}

	// Estimate gas
	gasLimit, err := client.EstimateGas(context.Background(), ethereum.CallMsg{
		From: signerAddress,
		To:   &tokenAddress,
		Data: data,
	})

	// Use default gas limit if estimation fails
	if err != nil {
		fmt.Printf("Gas estimation failed: %v. Using default gas limit.\n", err)
		gasLimit = DefaultGasLimit
	}

	// Add buffer to gas limit
	gasLimit += GasLimitBuffer

	// Create transaction
	tx := types.NewTransaction(
		nonce,
		tokenAddress,
		big.NewInt(0), // No ETH value is being sent
		gasLimit,
		gasPrice,
		data,
	)

	return sendTransaction(tx, client, signerPrivateKey)
}

func sendTransaction(tx *types.Transaction, client *ethclient.Client, signerPrivateKey *ecdsa.PrivateKey) error {
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return err
	}
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), signerPrivateKey)
	if err != nil {
		return err
	}
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return err
	}
	fmt.Println("Transaction sent: ", signedTx.Hash().Hex())

	receipt, err := bind.WaitMined(context.Background(), client, signedTx)
	if err != nil {
		return err
	}
	fmt.Println("Transaction receipt status: ", receipt.Status)
	return nil

}
