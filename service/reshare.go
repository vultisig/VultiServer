package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"
	vaultType "github.com/vultisig/commondata/go/vultisig/vault/v1"
	mtss "github.com/vultisig/mobile-tss-lib/tss"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/vultisig/vultisigner/common"
	"github.com/vultisig/vultisigner/internal/tasks"
	"github.com/vultisig/vultisigner/internal/types"
	"github.com/vultisig/vultisigner/relay"
)

func (s *WorkerService) Reshare(vault *vaultType.Vault,
	sessionID,
	hexEncryptionKey,
	serverURL string,
	encryptionPassword string, email string) error {
	if vault.Name == "" {
		return fmt.Errorf("vault name is empty")
	}
	if vault.LocalPartyId == "" {
		return fmt.Errorf("local party id is empty")
	}
	if vault.HexChainCode == "" {
		return fmt.Errorf("hex chain code is empty")
	}
	if serverURL == "" {
		return fmt.Errorf("serverURL is empty")
	}
	client := relay.NewRelayClient(serverURL)
	// Let's register session here
	if err := client.RegisterSession(sessionID, vault.LocalPartyId); err != nil {
		return fmt.Errorf("failed to register session: %w", err)
	}
	// wait longer for keygen start
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	partiesJoined, err := client.WaitForSessionStart(ctx, sessionID)
	s.logger.WithFields(logrus.Fields{
		"session":        sessionID,
		"parties_joined": partiesJoined,
	}).Info("Session started")
	if err != nil {
		return fmt.Errorf("failed to wait for session start: %w", err)
	}
	keyShares := make(map[string]string)
	for _, share := range vault.KeyShares {
		keyShares[share.PublicKey] = share.Keyshare
	}
	localStateAccessor, err := relay.NewLocalStateAccessorImp(vault.LocalPartyId, s.cfg.Server.VaultsFilePath, vault.PublicKeyEcdsa, encryptionPassword)
	if err != nil {
		return fmt.Errorf("failed to create localStateAccessor: %w", err)
	}

	tssServerImp, err := s.createTSSService(serverURL, sessionID, hexEncryptionKey, localStateAccessor, true)
	if err != nil {
		return fmt.Errorf("failed to create TSS service: %w", err)
	}
	localPartyID := vault.LocalPartyId
	endCh, wg := s.startMessageDownload(serverURL, sessionID, localPartyID, hexEncryptionKey, tssServerImp, "")
	ecdsaPubkey, eddsaPubkey, newResharePrefix := "", "", ""
	for attempt := 0; attempt < 3; attempt++ {
		ecdsaPubkey, eddsaPubkey, newResharePrefix, err = s.reshareWithRetry(
			tssServerImp,
			vault,
			partiesJoined,
		)
		if err == nil {
			break
		}
		s.logger.WithFields(logrus.Fields{
			"session": sessionID,
			"attempt": attempt,
		}).Error(err)
	}
	close(endCh)
	wg.Wait()

	if err != nil {
		return err
	}

	if err := client.CompleteSession(sessionID, localPartyID); err != nil {
		s.logger.WithFields(logrus.Fields{
			"session": sessionID,
			"error":   err,
		}).Error("Failed to complete session")
	}

	if isCompleted, err := client.CheckCompletedParties(sessionID, partiesJoined); err != nil || !isCompleted {
		s.logger.WithFields(logrus.Fields{
			"session":     sessionID,
			"isCompleted": isCompleted,
			"error":       err,
		}).Error("Failed to check completed parties")
	}

	ecdsaKeyShare, err := localStateAccessor.GetLocalState(ecdsaPubkey)
	if err != nil {
		return fmt.Errorf("failed to get local sate: %w", err)
	}

	eddsaKeyShare, err := localStateAccessor.GetLocalState(eddsaPubkey)
	if err != nil {
		return fmt.Errorf("failed to get local sate: %w", err)
	}

	newVault := &vaultType.Vault{
		Name:           vault.Name,
		PublicKeyEcdsa: ecdsaPubkey,
		PublicKeyEddsa: eddsaPubkey,
		Signers:        partiesJoined,
		CreatedAt:      timestamppb.Now(),
		HexChainCode:   vault.HexChainCode,
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
		LocalPartyId:  vault.LocalPartyId,
		ResharePrefix: newResharePrefix,
	}
	return s.SaveVaultAndScheduleEmail(newVault, encryptionPassword, email)
}
func (s *WorkerService) SaveVaultAndScheduleEmail(vault *vaultType.Vault,
	encryptionPassword string,
	email string) error {
	vaultData, err := proto.Marshal(vault)
	if err != nil {
		return fmt.Errorf("failed to Marshal vault: %w", err)
	}

	vaultData, err = common.EncryptVault(encryptionPassword, vaultData)
	if err != nil {
		return fmt.Errorf("common.EncryptVault failed: %w", err)
	}

	vaultBackup := &vaultType.VaultContainer{
		Version:     1,
		Vault:       base64.StdEncoding.EncodeToString(vaultData),
		IsEncrypted: true,
	}
	filePathName := filepath.Join(s.cfg.Server.VaultsFilePath, vault.PublicKeyEcdsa+".bak")
	file, err := os.Create(filePathName)

	if err != nil {
		return fmt.Errorf("fail to create file, err: %w", err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			s.logger.Errorf("fail to close file, err: %v", err)
		}
	}()

	vaultBackupData, err := proto.Marshal(vaultBackup)
	if err != nil {
		return fmt.Errorf("failed to Marshal vaultBackup: %w", err)
	}
	base64VaultContent := base64.StdEncoding.EncodeToString(vaultBackupData)
	if _, err := file.Write([]byte(base64VaultContent)); err != nil {
		return fmt.Errorf("fail to write file, err: %w", err)
	}

	emailRequest := types.EmailRequest{
		Email:       email,
		FileName:    common.GetVaultName(vault),
		FileContent: base64VaultContent,
		VaultName:   vault.Name,
	}
	buf, err := json.Marshal(emailRequest)
	if err != nil {
		return fmt.Errorf("json.Marshal failed: %w", err)
	}
	taskInfo, err := s.queueClient.Enqueue(asynq.NewTask(tasks.TypeEmailVaultBackup, buf),
		asynq.Retention(10*time.Minute),
		asynq.Queue(tasks.EMAIL_QUEUE_NAME))
	if err != nil {
		s.logger.Errorf("fail to enqueue email task: %v", err)
	}
	s.logger.Info("Email task enqueued: ", taskInfo.ID)
	return nil
}
func (s *WorkerService) getOldParties(newParties []string, oldSignerCommittee []string) []string {
	oldParties := make([]string, 0)
	for _, party := range oldSignerCommittee {
		if slices.Contains(newParties, party) {
			oldParties = append(oldParties, party)
		}
	}
	return oldParties
}

func (s *WorkerService) reshareWithRetry(tssService *mtss.ServiceImpl,
	vault *vaultType.Vault,
	newParties []string,
) (string, string, string, error) {
	oldParties := s.getOldParties(newParties, vault.Signers)
	resp, err := s.reshareECDSAKey(tssService, vault.PublicKeyEcdsa, vault.LocalPartyId, vault.HexChainCode, vault.ResharePrefix,
		newParties, oldParties)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to reshare ECDSA key: %w", err)
	}
	newResharePrefix := resp.ResharePrefix
	ecdsaPubkey := resp.PubKey
	resp, err = s.reshareEDDSAKey(tssService, vault.PublicKeyEddsa, vault.LocalPartyId, vault.HexChainCode, vault.ResharePrefix,
		newParties, oldParties, newResharePrefix)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to reshare EDDSA key: %w", err)
	}
	eddsaPubkey := resp.PubKey
	return ecdsaPubkey, eddsaPubkey, newResharePrefix, nil
}

func (s *WorkerService) reshareECDSAKey(tssService *mtss.ServiceImpl,
	publicKey string,
	localPartyID, hexChainCode string,
	resharePrefix string,
	partiesJoined []string,
	oldParties []string) (*mtss.ReshareResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"public_key":         publicKey,
		"localPartyID":       localPartyID,
		"chain_code":         hexChainCode,
		"reshare_prefix":     resharePrefix,
		"parties_joined":     partiesJoined,
		"old_parties":        oldParties,
		"new_reshare_prefix": "",
	}).Info("Start ECDSA reshare...")

	resp, err := tssService.ReshareECDSA(&mtss.ReshareRequest{
		PubKey:           publicKey,
		LocalPartyID:     localPartyID,
		NewParties:       strings.Join(partiesJoined, ","),
		OldParties:       strings.Join(oldParties, ","),
		ChainCodeHex:     hexChainCode,
		ResharePrefix:    resharePrefix,
		NewResharePrefix: "",
	})
	if err != nil {
		return nil, fmt.Errorf("fail to reshare ECDSA key: %w", err)
	}
	s.logger.WithFields(logrus.Fields{
		"key":     localPartyID,
		"pub_key": resp.PubKey,
	}).Info("ECDSA keygen response")

	return resp, nil
}

func (s *WorkerService) reshareEDDSAKey(tssService *mtss.ServiceImpl,
	publicKey string,
	localPartyID, hexChainCode string,
	resharePrefix string,
	partiesJoined []string,
	oldParties []string,
	newResharePrefix string) (*mtss.ReshareResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"public_key":         publicKey,
		"localPartyID":       localPartyID,
		"chain_code":         hexChainCode,
		"reshare_prefix":     resharePrefix,
		"parties_joined":     partiesJoined,
		"old_parties":        oldParties,
		"new_reshare_prefix": newResharePrefix,
	}).Info("Start EdDSA keygen...")
	resp, err := tssService.ResharingEdDSA(&mtss.ReshareRequest{
		PubKey:           publicKey,
		LocalPartyID:     localPartyID,
		NewParties:       strings.Join(partiesJoined, ","),
		ChainCodeHex:     hexChainCode,
		OldParties:       strings.Join(oldParties, ","),
		ResharePrefix:    resharePrefix,
		NewResharePrefix: newResharePrefix,
	})
	if err != nil {
		return nil, fmt.Errorf("fail to reshare EdDSA key: %w", err)
	}
	s.logger.WithFields(logrus.Fields{
		"localPartyID": localPartyID,
		"pub_key":      resp.PubKey,
	}).Info("EdDSA reshare response")
	return resp, nil
}
