package relay

import (
	"fmt"
	"os"
	"path/filepath"

	vaultType "github.com/vultisig/commondata/go/vultisig/vault/v1"
	"github.com/vultisig/vultisigner/common"
)

type LocalStateAccessorImp struct {
	Key    string
	Folder string
}

func (l *LocalStateAccessorImp) ensureFolder() error {
	if l.Folder == "" {
		var err error
		l.Folder, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	return nil
}

func (l *LocalStateAccessorImp) GetLocalState(pubKey string) (string, error) {
	if err := l.ensureFolder(); err != nil {
		return "", err
	}

	fileName := filepath.Join(l.Folder, pubKey+"-"+l.Key+".json")
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return "", fmt.Errorf("file %s does not exist", fileName)
	}

	buf, err := os.ReadFile(fileName)
	if err != nil {
		return "", fmt.Errorf("fail to read file %s: %w", fileName, err)
	}

	return string(buf), nil
}

func (l *LocalStateAccessorImp) SaveLocalState(pubKey, localState string) error {
	if err := l.ensureFolder(); err != nil {
		return err
	}

	fileName := filepath.Join(l.Folder, pubKey+"-"+l.Key+".json")

	return os.WriteFile(fileName, []byte(localState), 0644)
}

func (l *LocalStateAccessorImp) RemoveLocalState(pubKey string) error {
	if err := l.ensureFolder(); err != nil {
		return err
	}

	fileName := filepath.Join(l.Folder, pubKey+"-"+l.Key+".json")

	return os.Remove(fileName)
}

func (l *LocalStateAccessorImp) GetVault(pubKey, passwd string) (*vaultType.Vault, error) {
	if err := l.ensureFolder(); err != nil {
		return nil, err
	}

	fileName := filepath.Join(l.Folder, pubKey+".bak")
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return nil, fmt.Errorf("file %s does not exist", fileName)
	}

	buf, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("fail to read file %s: %w", fileName, err)
	}

	vault, err := common.DecryptVaultFromBackup(passwd, buf)
	if err != nil {
		return nil, fmt.Errorf("fail to decrypt vault from the backup, err: %w", err)
	}

	return vault, nil
}
