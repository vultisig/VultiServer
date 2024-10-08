package relay

import (
	"fmt"
	"os"
	"path/filepath"

	vaultType "github.com/vultisig/commondata/go/vultisig/vault/v1"

	"github.com/vultisig/vultisigner/common"
)

type LocalStateAccessorImp struct {
	localPartyID string
	Folder       string
	Vault        *vaultType.Vault
	cache        map[string]string
}

func NewLocalStateAccessorImp(localPartyID, folder, vaultFileName, vaultPasswd string) (*LocalStateAccessorImp, error) {
	localStateAccessor := &LocalStateAccessorImp{
		localPartyID: localPartyID,
		Folder:       folder,
		Vault:        nil,
		cache:        make(map[string]string),
	}

	var err error
	if localStateAccessor.Folder == "" {
		localStateAccessor.Folder, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	if vaultFileName != "" {
		fileName := filepath.Join(localStateAccessor.Folder, vaultFileName+".bak")
		if _, err := os.Stat(fileName); os.IsNotExist(err) {
			return nil, fmt.Errorf("file %s does not exist", fileName)
		}

		buf, err := os.ReadFile(fileName)
		if err != nil {
			return nil, fmt.Errorf("fail to read file %s: %w", fileName, err)
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
	return l.cache[pubKey], nil
}

func (l *LocalStateAccessorImp) SaveLocalState(pubKey, localState string) error {
	l.cache[pubKey] = localState
	return nil
}
