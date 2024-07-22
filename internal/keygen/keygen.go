package keygen

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	vaultType "github.com/vultisig/commondata/go/vultisig/vault/v1"
	"github.com/vultisig/mobile-tss-lib/tss"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/vultisig/vultisigner/common"
	"github.com/vultisig/vultisigner/config"
	"github.com/vultisig/vultisigner/internal/logging"
	"github.com/vultisig/vultisigner/internal/types"
	"github.com/vultisig/vultisigner/relay"
)

func JoinKeyGeneration(kg *types.KeyGeneration) (string, string, error) {
	keyFolder := config.AppConfig.Server.VaultsFilePath
	serverURL := config.AppConfig.Relay.Server

	server := relay.NewServer(serverURL)

	// Let's register session here
	if err := server.RegisterSession(kg.Session, kg.Key); err != nil {
		return "", "", fmt.Errorf("failed to register session: %w", err)
	}
	// wait longer for keygen start
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	partiesJoined, err := server.WaitForSessionStart(ctx, kg.Session)
	logging.Logger.WithFields(logrus.Fields{
		"session":        kg.Session,
		"parties_joined": partiesJoined,
	}).Info("Session started")

	if err != nil {
		return "", "", fmt.Errorf("failed to wait for session start: %w", err)
	}

	localStateAccessor := &relay.LocalStateAccessorImp{
		Key:    kg.Key,
		Folder: keyFolder,
	}
	tssServerImp, err := createTSSService(serverURL, localStateAccessor, kg)
	if err != nil {
		return "", "", fmt.Errorf("failed to create TSS service: %w", err)
	}

	ecdsaPubkey, eddsaPubkey := "", ""
	for attempt := 0; attempt < 3; attempt++ {
		ecdsaPubkey, eddsaPubkey, err = keygenWithRetry(serverURL, kg, partiesJoined, tssServerImp)
		if err == nil {
			break
		}
	}

	if err != nil {
		return "", "", err
	}

	if err := server.CompleteSession(kg.Session, kg.Key); err != nil {
		logging.Logger.WithFields(logrus.Fields{
			"session": kg.Session,
			"error":   err,
		}).Error("Failed to complete session")
	}

	if err := server.EndSession(kg.Session); err != nil {
		logging.Logger.WithFields(logrus.Fields{
			"session": kg.Session,
			"error":   err,
		}).Error("Failed to end session")
	}

	err = BackupVault(kg, partiesJoined, ecdsaPubkey, eddsaPubkey, localStateAccessor)

	if err != nil {
		return "", "", fmt.Errorf("failed to backup vault: %w", err)
	}

	return ecdsaPubkey, eddsaPubkey, nil
}

func createTSSService(serverURL string, localStateAccessor tss.LocalStateAccessor, kg *types.KeyGeneration) (tss.Service, error) {
	messenger := &relay.MessengerImp{
		Server:           serverURL,
		SessionID:        kg.Session,
		HexEncryptionKey: kg.HexEncryptionKey,
	}

	tssService, err := tss.NewService(messenger, localStateAccessor, true)
	if err != nil {
		return nil, fmt.Errorf("create TSS service: %w", err)
	}
	return tssService, nil
}

func startMessageDownload(serverURL, session, key, hexEncryptionKey string, tssService tss.Service) (chan struct{}, *sync.WaitGroup) {
	logging.Logger.WithFields(logrus.Fields{
		"session": session,
		"key":     key,
	}).Info("Start downloading messages")

	endCh := make(chan struct{})
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go relay.DownloadMessage(serverURL, session, key, hexEncryptionKey, tssService, endCh, wg)
	return endCh, wg
}

func keygenWithRetry(serverURL string, kg *types.KeyGeneration, partiesJoined []string, tssService tss.Service) (string, string, error) {
	endCh, wg := startMessageDownload(serverURL, kg.Session, kg.Key, kg.HexEncryptionKey, tssService)

	resp, err := generateECDSAKey(tssService, kg, partiesJoined)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate ECDSA key: %w", err)
	}

	respEDDSA, err := generateEDDSAKey(tssService, kg, partiesJoined)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate EDDSA key: %w", err)
	}

	close(endCh)
	wg.Wait()

	return resp.PubKey, respEDDSA.PubKey, nil
}

func generateECDSAKey(tssService tss.Service, kg *types.KeyGeneration, partiesJoined []string) (*tss.KeygenResponse, error) {
	logging.Logger.WithFields(logrus.Fields{
		"key":            kg.Key,
		"chain_code":     kg.ChainCode,
		"parties_joined": partiesJoined,
	}).Info("Start ECDSA keygen...")
	resp, err := tssService.KeygenECDSA(&tss.KeygenRequest{
		LocalPartyID: kg.Key,
		AllParties:   strings.Join(partiesJoined, ","),
		ChainCodeHex: kg.ChainCode,
	})
	if err != nil {
		return nil, fmt.Errorf("generate ECDSA key: %w", err)
	}
	logging.Logger.WithFields(logrus.Fields{
		"key":     kg.Key,
		"pub_key": resp.PubKey,
	}).Info("ECDSA keygen response")
	time.Sleep(time.Second)
	return resp, nil
}

func generateEDDSAKey(tssService tss.Service, kg *types.KeyGeneration, partiesJoined []string) (*tss.KeygenResponse, error) {
	logging.Logger.WithFields(logrus.Fields{
		"key":            kg.Key,
		"chain_code":     kg.ChainCode,
		"parties_joined": partiesJoined,
	}).Info("Start EDDSA keygen...")
	resp, err := tssService.KeygenEdDSA(&tss.KeygenRequest{
		LocalPartyID: kg.Key,
		AllParties:   strings.Join(partiesJoined, ","),
		ChainCodeHex: kg.ChainCode,
	})
	if err != nil {
		return nil, fmt.Errorf("generate EDDSA key: %w", err)
	}
	logging.Logger.WithFields(logrus.Fields{
		"key":     kg.Key,
		"pub_key": resp.PubKey,
	}).Info("EDDSA keygen response")
	time.Sleep(time.Second)
	return resp, nil
}

func BackupVault(kg *types.KeyGeneration, partiesJoined []string, ecdsaPubkey, eddsaPubkey string, localStateAccessor *relay.LocalStateAccessorImp) error {
	ecdsaKeyShare, err := localStateAccessor.GetLocalState(ecdsaPubkey)
	if err != nil {
		return fmt.Errorf("failed to get local sate: %w", err)
	}

	eddsaKeyShare, err := localStateAccessor.GetLocalState(eddsaPubkey)
	if err != nil {
		return fmt.Errorf("failed to get local sate: %w", err)
	}

	vault := &vaultType.Vault{
		Name:           kg.Name,
		PublicKeyEcdsa: ecdsaPubkey,
		PublicKeyEddsa: eddsaPubkey,
		Signers:        partiesJoined,
		CreatedAt:      timestamppb.New(time.Now()),
		HexChainCode:   kg.ChainCode,
		KeyShares: []*vaultType.Vault_KeyShare{
			{
				PublicKey: ecdsaPubkey,
				Keyshare:  ecdsaKeyShare,
			},
			{
				PublicKey: eddsaPubkey,
				Keyshare:  eddsaKeyShare,
			},
		},
		LocalPartyId:  kg.Key,
		ResharePrefix: "",
	}

	isEncrypted := kg.EncryptionPassword != ""
	vaultData, err := proto.Marshal(vault)
	if err != nil {
		return fmt.Errorf("failed to Marshal vault: %w", err)
	}

	if isEncrypted {
		vaultData, err = common.EncryptVault(kg.EncryptionPassword, vaultData)
		if err != nil {
			return fmt.Errorf("common.EncryptVault failed: %w", err)
		}
	}
	vaultBackup := &vaultType.VaultContainer{
		Version:     1,
		Vault:       base64.StdEncoding.EncodeToString(vaultData),
		IsEncrypted: isEncrypted,
	}
	filePathName := filepath.Join(config.AppConfig.Server.VaultsFilePath, ecdsaPubkey+".bak")
	file, err := os.Create(filePathName)

	if err != nil {
		return fmt.Errorf("fail to create file, err: %w", err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			logging.Logger.Errorf("fail to close file, err: %v", err)
		}
	}()

	vaultBackupData, err := proto.Marshal(vaultBackup)
	if err != nil {
		return fmt.Errorf("failed to Marshal vaultBackup: %w", err)
	}

	if _, err := file.Write([]byte(base64.StdEncoding.EncodeToString(vaultBackupData))); err != nil {
		return fmt.Errorf("fail to write file, err: %w", err)
	}

	return nil
}
