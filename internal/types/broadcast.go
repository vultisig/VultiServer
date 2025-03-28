package types

import "fmt"

type BroadcastStrategy string

const (
	BroadcastImmediate BroadcastStrategy = "IMMEDIATE"
	BroadcastPrivate   BroadcastStrategy = "PRIVATE_MEMPOOL"
	BroadcastManual    BroadcastStrategy = "MANUAL"
)

type TransactionError struct {
	Code    string
	Message string
	Err     error
}

func (e *TransactionError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

const (
	// nonce related
	ErrNonce = "NONCE_TOO_LOW"

	// gas related
	ErrGasTooLow           = "GAS_TOO_LOW"
	ErrGasTooHigh          = "GAS_TOO_HIGH"
	ErrGasPriceUnderpriced = "GAS_UNDERPRICED"

	// balance related
	ErrInsufficientFunds = "INSUFFICIENT_FUNDS"

	// network/RPC related
	ErrRPCConnectionFailed = "RPC_CONNECTION_FAILED"

	// transaction state
	ErrTxDropped = "TX_DROPPED"
	ErrTxTimeout = "TX_TIMEOUT"

	// contract related
	ErrExecutionReverted = "EXECUTION_REVERTED"

	// retriable errors
	ErrRetriable = "RETRIABLE_ERROR"

	// permanent failure
	ErrPermanentFailure = "PERMANENT_FAILURE"

	// other
	ErrUnknown = "UNKNOWN_ERROR"
)
