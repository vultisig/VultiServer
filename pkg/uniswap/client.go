package uniswap

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

type Client struct {
	cfg *Config
}

func NewClient(cfg *Config) (*Client, error) {
	return &Client{cfg}, nil
}

func (uc *Client) GetRouterAddress() *common.Address {
	return uc.cfg.routerAddress
}

func (uc *Client) ApproveERC20Token(chainID *big.Int, signerAddress *common.Address, tokenAddress, spenderAddress common.Address, amount *big.Int, nonceOffset uint64) ([]byte, []byte, error) {
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
		return nil, nil, err
	}
	approveData, err := parsedABI.Pack("approve", spenderAddress, amount)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to pack approve data: %w", err)
	}
	nonce, err := uc.cfg.rpcClient.PendingNonceAt(context.Background(), *signerAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get nonce: %w", err)
	}
	nonce += nonceOffset

	gasPrice, err := uc.cfg.rpcClient.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get gas price: %w", err)
	}
	gasLimit, err := uc.cfg.rpcClient.EstimateGas(context.Background(), ethereum.CallMsg{
		From: *signerAddress, //This field is needed when there is approve on USDC token.
		To:   &tokenAddress,
		Data: approveData,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to estimate gas limit: %w", err)
	}
	gasLimit += uc.cfg.gasLimitBuffer
	tx := types.NewTransaction(nonce, tokenAddress, big.NewInt(0), gasLimit, gasPrice, approveData)
	hash, rawTx, err := uc.rlpUnsignedTxAndHash(tx, chainID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed rlp hash tx data: %w", err)
	}

	return hash, rawTx, err
}

func (uc *Client) GetAllowance(signerAddress common.Address, tokenAddress common.Address) (*big.Int, error) {
	tokenABI := `[{
        "constant": true,
        "inputs": [
            {
                "name": "_owner",
                "type": "address"
            },
            {
                "name": "_spender",
                "type": "address"
            }
        ],
        "name": "allowance",
        "outputs": [
            {
                "name": "",
                "type": "uint256"
            }
        ],
        "payable": false,
        "stateMutability": "view",
        "type": "function"
    }]`

	parsedABI, err := abi.JSON(strings.NewReader(tokenABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse allowance ABI: %w", err)
	}

	data, err := parsedABI.Pack("allowance", signerAddress, *uc.cfg.routerAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to pack allowance data: %w", err)
	}

	result, err := uc.cfg.rpcClient.CallContract(context.Background(), ethereum.CallMsg{
		To:   &tokenAddress,
		Data: data,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call allowance: %w", err)
	}

	var allowance *big.Int
	err = parsedABI.UnpackIntoInterface(&allowance, "allowance", result)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack allowance: %w", err)
	}
	return allowance, nil
}

func (uc *Client) SwapTokens(chainID *big.Int, signerAddress *common.Address, amountIn, amountOutMin *big.Int, path []common.Address, nonceOffset uint64) ([]byte, []byte, error) {
	log.Println("Swapping tokens...")
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
		return nil, nil, err
	}

	deadline := big.NewInt(time.Now().Add(uc.cfg.deadlineDuration).Unix())

	swapData, err := parsedRouterABI.Pack("swapExactTokensForTokens", amountIn, amountOutMin, path, *signerAddress, deadline)
	if err != nil {
		return nil, nil, err
	}
	nonce, err := uc.cfg.rpcClient.PendingNonceAt(context.Background(), *signerAddress)
	if err != nil {
		return nil, nil, err
	}
	nonce += nonceOffset

	gasPrice, err := uc.cfg.rpcClient.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, nil, err
	}

	tx := types.NewTransaction(nonce, *uc.cfg.routerAddress, big.NewInt(0), uc.cfg.swapGasLimit, gasPrice, swapData)
	hash, rawTx, err := uc.rlpUnsignedTxAndHash(tx, chainID)
	if err != nil {
		return nil, nil, err
	}

	return hash, rawTx, err
}

func (uc *Client) GetTokenBalance(signerAddress *common.Address, tokenAddress common.Address) (*big.Int, error) {
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
	callData, err := parsedABI.Pack("balanceOf", *signerAddress)
	if err != nil {
		return nil, err
	}

	msg := ethereum.CallMsg{
		To:   &tokenAddress,
		Data: callData,
	}

	result, err := uc.cfg.rpcClient.CallContract(context.Background(), msg, nil)
	if err != nil {
		return nil, err
	}

	balance := new(big.Int)
	balance.SetBytes(result)
	return balance, nil
}

func (uc *Client) GetExpectedAmountOut(amountIn *big.Int, path []common.Address) (*big.Int, error) {
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
		To:   uc.cfg.routerAddress,
		Data: callData,
	}

	result, err := uc.cfg.rpcClient.CallContract(context.Background(), msg, nil)
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

func (uc *Client) rlpUnsignedTxAndHash(tx *types.Transaction, chainID *big.Int) ([]byte, []byte, error) {
	// post EIP-155 transaction
	V := new(big.Int).Set(chainID)
	V = V.Mul(V, big.NewInt(2))
	V = V.Add(V, big.NewInt(35))
	rawTx, err := rlp.EncodeToBytes([]interface{}{
		tx.Nonce(),
		tx.GasPrice(),
		tx.Gas(),
		tx.To(),
		tx.Value(),
		tx.Data(),
		V,       // chain id
		uint(0), // r
		uint(0), // s
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to rlp encode transaction: %v", err)
	}

	signer := types.NewEIP155Signer(chainID)
	txHash := signer.Hash(tx).Bytes()

	return txHash, rawTx, nil
}

func (uc *Client) CalculateAmountOutMin(expectedAmountOut *big.Int, slippagePercentage float64) *big.Int {
	slippageFactor := big.NewFloat(1 - slippagePercentage/100)
	expectedAmountOutFloat := new(big.Float).SetInt(expectedAmountOut)
	amountOutMinFloat := new(big.Float).Mul(expectedAmountOutFloat, slippageFactor)
	amountOutMin, _ := amountOutMinFloat.Int(nil)
	return amountOutMin
}
