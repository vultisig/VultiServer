package service

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
	"github.com/vultisig/mobile-tss-lib/tss"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/vultisig/vultisigner/common"
	"github.com/vultisig/vultisigner/config"
	"github.com/vultisig/vultisigner/internal/types"
	"github.com/vultisig/vultisigner/relay"

	vaultType "github.com/vultisig/commondata/go/vultisig/vault/v1"
)

func (s *WorkerService) JoinKeyGeneration(req types.VaultCreateRequest) (string, string, error) {
	keyFolder := s.cfg.Server.VaultsFilePath
	serverURL := s.cfg.Relay.Server
	relayClient := relay.NewRelayClient(serverURL)

	// Let's register session here
	if err := relayClient.RegisterSession(req.Session, kg.Key); err != nil {
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

	localStateAccessor, err := relay.NewLocalStateAccessorImp(kg.Key, keyFolder, "", "")
	if err != nil {
		return "", "", fmt.Errorf("failed to create localStateAccessor: %w", err)
	}

	tssServerImp, err := createTSSService(serverURL, kg.Session, kg.HexEncryptionKey, localStateAccessor, true)
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

	if isCompleted, err := server.CheckCompletedParties(kg.Session, partiesJoined); err != nil || !isCompleted {
		logging.Logger.WithFields(logrus.Fields{
			"session":     kg.Session,
			"isCompleted": isCompleted,
			"error":       err,
		}).Error("Failed to check completed parties")
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

	err = localStateAccessor.RemoveLocalState(ecdsaPubkey)
	if err != nil {
		return "", "", fmt.Errorf("failed to remove local state: %w", err)
	}

	err = localStateAccessor.RemoveLocalState(eddsaPubkey)
	if err != nil {
		return "", "", fmt.Errorf("failed to remove local state: %w", err)
	}

	return ecdsaPubkey, eddsaPubkey, nil
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

func createTSSService(serverURL, Session, HexEncryptionKey string, localStateAccessor tss.LocalStateAccessor, createPreParam bool) (tss.Service, error) {
	messenger := &relay.MessengerImp{
		Server:           serverURL,
		SessionID:        Session,
		HexEncryptionKey: HexEncryptionKey,
	}

	tssService, err := tss.NewService(messenger, localStateAccessor, createPreParam)
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

func JoinKeySign(ks *types.KeysignRequest) (map[string]tss.KeysignResponse, error) {
	result := map[string]tss.KeysignResponse{}
	keyFolder := config.AppConfig.Server.VaultsFilePath
	serverURL := config.AppConfig.Relay.Server
	localStateAccessor, err := relay.NewLocalStateAccessorImp("", keyFolder, ks.PublicKey, ks.VaultPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to create localStateAccessor: %w", err)
	}

	localPartyId := localStateAccessor.Vault.LocalPartyId
	server := relay.NewServer(serverURL)

	// Let's register session here
	if err := server.RegisterSession(ks.Session, localPartyId); err != nil {
		return nil, fmt.Errorf("failed to register session: %w", err)
	}
	// wait longer for keysign start
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute+3*time.Second)
	defer cancel()

	partiesJoined, err := server.WaitForSessionStart(ctx, ks.Session)
	logging.Logger.WithFields(logrus.Fields{
		"session":        ks.Session,
		"parties_joined": partiesJoined,
	}).Info("Session started")

	if err != nil {
		return nil, fmt.Errorf("failed to wait for session start: %w", err)
	}

	tssServerImp, err := createTSSService(serverURL, ks.Session, ks.HexEncryptionKey, localStateAccessor, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create TSS service: %w", err)
	}

	for _, message := range ks.Messages {
		var signature *tss.KeysignResponse
		for attempt := 0; attempt < 3; attempt++ {
			signature, err = keysignWithRetry(serverURL, localPartyId, ks, partiesJoined, tssServerImp, message)
			if err == nil {
				break
			}
		}
		if err != nil {
			return result, err
		}
		result[message] = *signature
	}

	if err := server.CompleteSession(ks.Session, localPartyId); err != nil {
		logging.Logger.WithFields(logrus.Fields{
			"session": ks.Session,
			"error":   err,
		}).Error("Failed to complete session")
	}

	return result, nil
}

func keysignWithRetry(serverURL, localPartyId string, ks *types.KeysignRequest, partiesJoined []string, tssService tss.Service, msg string) (*tss.KeysignResponse, error) {
	endCh, wg := startMessageDownload(serverURL, ks.Session, localPartyId, ks.HexEncryptionKey, tssService)

	var signature *tss.KeysignResponse
	var err error

	if ks.IsECDSA {
		signature, err = tssService.KeysignECDSA(&tss.KeysignRequest{
			PubKey:               ks.PublicKey,
			MessageToSign:        msg,
			LocalPartyKey:        localPartyId,
			KeysignCommitteeKeys: strings.Join(partiesJoined, ","),
			DerivePath:           ks.DerivePath,
		})
	} else {
		signature, err = tssService.KeysignEdDSA(&tss.KeysignRequest{
			PubKey:               ks.PublicKey,
			MessageToSign:        msg,
			LocalPartyKey:        localPartyId,
			KeysignCommitteeKeys: strings.Join(partiesJoined, ","),
			DerivePath:           ks.DerivePath,
		})
	}

	if err != nil {
		return nil, fmt.Errorf("fail to key sign: %w", err)
	}

	close(endCh)
	wg.Wait()

	return signature, nil
}
