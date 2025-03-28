package plugin

import (
	"context"
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type NonceManager struct {
	rpcClient *ethclient.Client
	nonceMap  sync.Map
	mu        sync.Mutex
}

func NewNonceManager(rpcClient *ethclient.Client) *NonceManager {
	return &NonceManager{
		rpcClient: rpcClient,
	}
}

func (n *NonceManager) GetNextNonce(address string) (uint64, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	nonce, err := n.rpcClient.PendingNonceAt(context.Background(), common.HexToAddress(address))
	if err != nil {
		return 0, fmt.Errorf("failed to get nonce from network: %w", err)
	}
	return nonce, nil
}

func (n *NonceManager) ResetNonce(address string) {
	n.nonceMap.Delete(address)
}
