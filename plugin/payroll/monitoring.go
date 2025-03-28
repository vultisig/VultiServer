package payroll

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	gcommon "github.com/ethereum/go-ethereum/common"
	gtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/vultisig/vultisigner/internal/types"
)

func (p *PayrollPlugin) handleBroadcastError(err error, sender gcommon.Address) error {
	errMsg := err.Error()

	switch {
	case strings.Contains(errMsg, "insufficient funds"):
		// this is for ETH balance for gas - immediate failure, what to do?
		//goal : retry only when we dectect user send funds
		// for now : we can skip this trigger and wait for next one
		return &types.TransactionError{
			Code:    types.ErrInsufficientFunds,
			Message: fmt.Sprintf("Account %s has insufficient gas", sender.Hex()),
			Err:     err,
		}

	case strings.Contains(errMsg, "nonce too low"):
	case strings.Contains(errMsg, "nonce too high"):
	case strings.Contains(errMsg, "gas price too low"):
	case strings.Contains(errMsg, "gas limit reached"):
		// these are retriable errors - the caller should retry with updated parameters
		//we should not skip this trigger and retry immediately
		return &types.TransactionError{
			Code:    types.ErrRetriable,
			Message: err.Error(),
			Err:     err,
		}

	default:
		return &types.TransactionError{
			Code:    types.ErrRPCConnectionFailed,
			Message: "Unknown RPC error",
			Err:     err,
		}
	}
	return nil
}

func (p *PayrollPlugin) monitorTransaction(tx *gtypes.Transaction) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute) //how much time should we monitor the tx?
	defer cancel()

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	txHash := tx.Hash()
	for {
		select {
		case <-ctx.Done():
			return &types.TransactionError{
				Code:    types.ErrTxTimeout,
				Message: fmt.Sprintf("Transaction monitoring timed out for tx: %s", txHash.Hex()),
			}

		case <-ticker.C:
			// check tx status
			_, isPending, err := p.rpcClient.TransactionByHash(ctx, txHash)
			if err != nil {
				if err == ethereum.NotFound {
					return &types.TransactionError{
						Code:    types.ErrTxDropped,
						Message: fmt.Sprintf("Transaction dropped from mempool: %s", txHash.Hex()),
					}
				}
				continue // keep trying on other RPC errors
			}

			if !isPending {
				receipt, err := p.rpcClient.TransactionReceipt(ctx, txHash)
				if err != nil {
					continue
				}

				if receipt.Status == 0 {
					reason := p.getRevertReason(ctx, tx, receipt.BlockNumber)

					// Check if it's a permanent failure (like insufficient token balance)
					if !p.isRetriableError(reason) {
						return &types.TransactionError{
							Code:    types.ErrPermanentFailure,
							Message: fmt.Sprintf("Transaction permanently failed: %s", reason),
						}
					}

					// It's a retriable error
					return &types.TransactionError{
						Code:    types.ErrRetriable,
						Message: fmt.Sprintf("Transaction failed with retriable error: %s", reason),
					}
				}

				// Transaction successful
				return nil
			}
		}
	}
}

func (p *PayrollPlugin) isRetriableError(reason string) bool {
	// implement logic to determine if the error is retriable based on the reason
	return strings.Contains(reason, "insufficient funds") || strings.Contains(reason, "nonce too low") || strings.Contains(reason, "nonce too high") || strings.Contains(reason, "gas price too low") || strings.Contains(reason, "gas limit reached")
}

func (p *PayrollPlugin) getRevertReason(ctx context.Context, tx *gtypes.Transaction, blockNum *big.Int) string {
	callMsg := ethereum.CallMsg{
		To:       tx.To(),
		Data:     tx.Data(),
		Gas:      tx.Gas(),
		GasPrice: tx.GasPrice(),
		Value:    tx.Value(),
	}

	_, err := p.rpcClient.CallContract(ctx, callMsg, blockNum)
	if err != nil {
		// try to parse standard revert reason
		if strings.Contains(err.Error(), "execution reverted:") {
			parts := strings.Split(err.Error(), "execution reverted:")
			if len(parts) > 1 {
				return strings.TrimSpace(parts[1])
			}
		}
		return err.Error()
	}
	return "Unknown revert reason"
}
