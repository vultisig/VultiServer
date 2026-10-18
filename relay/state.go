package relay

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"sync"

	vaultType "github.com/vultisig/commondata/go/vultisig/vault/v1"

	"github.com/vultisig/vultiserver/common"
	"github.com/vultisig/vultiserver/storage"
)

type LocalStateAccessorImp struct {
	Folder       string
	Vault        *vaultType.Vault
	cache        sync.Map
	blockStorage *storage.BlockStorage
}

func NewLocalStateAccessorImp(folder, vaultFileName, vaultPasswd string,
	storage *storage.BlockStorage) (*LocalStateAccessorImp, error) {
	localStateAccessor := &LocalStateAccessorImp{
		Folder:       folder,
		Vault:        nil,
		blockStorage: storage,
	}

	var err error
	if localStateAccessor.Folder == "" {
		localStateAccessor.Folder, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	if vaultFileName != "" {
		buf, err := storage.GetFile(vaultFileName + ".bak")
		if err != nil {
			return nil, fmt.Errorf("fail to get vault file: %w", err)
		}
		localStateAccessor.Vault, err = common.DecryptVaultFromBackup(vaultPasswd, buf)
		if err != nil {
			return nil, fmt.Errorf("fail to decrypt vault from the backup, err: %w", err)
		}
	}

	return localStateAccessor, nil
}

func (l *LocalStateAccessorImp) GetLocalState(pubKey string) (string, error) {
	if l.Vault != nil {
		for _, item := range l.Vault.KeyShares {
			if item.PublicKey == pubKey {
				return item.Keyshare, nil
			}
		}
		return "", fmt.Errorf("%s keyshare does not exist", pubKey)
	}
	val, ok := l.cache.Load(pubKey)
	if !ok {
		return "", nil
	}
	state, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("invalid cached local state type for key %s", pubKey)
	}
	return state, nil
}

// GetChainLocalState returns the keyshare that was imported for a specific chain of a
// KeyImport vault.
//
// Several chains can derive the same key (every EVM chain uses m/44'/60'/0'/0/0, the
// Cosmos-SDK chains share coin type 118, ...). Each chain still runs its own key-import
// ceremony, so the vault holds one distinct keyshare per chain, all under the same public
// key. GetLocalState returns the first match only, which pairs with the client's share for
// just one chain of each group; every other chain then fails with "Invalid final_session_id".
//
// Keyshares are stored root keys first, then one per entry of ChainPublicKeys in the same
// order, so the n-th chain using a public key owns the n-th keyshare stored under it (after
// any root keyshare with that same public key).
func (l *LocalStateAccessorImp) GetChainLocalState(chain string) (string, string, error) {
	if l.Vault == nil {
		return "", "", fmt.Errorf("vault is not loaded")
	}
	idx := slices.IndexFunc(l.Vault.ChainPublicKeys, func(cp *vaultType.Vault_ChainPublicKey) bool {
		return strings.EqualFold(cp.Chain, chain)
	})
	if idx < 0 {
		return "", "", fmt.Errorf("public key for chain %s not found in vault", chain)
	}
	pubKey := l.Vault.ChainPublicKeys[idx].PublicKey
	// earlier chains that share this public key own the earlier keyshares
	occurrence := 0
	for _, cp := range l.Vault.ChainPublicKeys[:idx] {
		if cp.PublicKey == pubKey {
			occurrence++
		}
	}
	// root keyshares are stored ahead of the chain keyshares
	if l.Vault.PublicKeyEcdsa == pubKey {
		occurrence++
	}
	if l.Vault.PublicKeyEddsa == pubKey {
		occurrence++
	}
	var matches []string
	for _, item := range l.Vault.KeyShares {
		if item.PublicKey == pubKey {
			matches = append(matches, item.Keyshare)
		}
	}
	if len(matches) == 0 {
		return "", "", fmt.Errorf("%s keyshare does not exist", pubKey)
	}
	if occurrence >= len(matches) {
		// vault does not hold one keyshare per chain (e.g. chains imported once per key)
		return pubKey, matches[0], nil
	}
	return pubKey, matches[occurrence], nil
}

func (l *LocalStateAccessorImp) SaveLocalState(pubKey, localState string) error {
	l.cache.Store(pubKey, localState)
	return nil
}

func (l *LocalStateAccessorImp) GetLocalCacheState(pubKey string) (string, error) {
	val, ok := l.cache.Load(pubKey)
	if !ok {
		return "", nil
	}
	state, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("invalid cached state type for key %s", pubKey)
	}
	return state, nil
}
