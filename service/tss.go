package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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
	"github.com/vultisig/vultisigner/internal/types"
	"github.com/vultisig/vultisigner/relay"

	vaultType "github.com/vultisig/commondata/go/vultisig/vault/v1"
)

func (s *WorkerService) JoinKeyGeneration(req types.VaultCreateRequest) (string, string, error) {
	keyFolder := s.cfg.Server.VaultsFilePath
	serverURL := s.cfg.Relay.Server
	relayClient := relay.NewRelayClient(serverURL)

	// Let's register session here
	if err := relayClient.RegisterSession(req.SessionID, req.LocalPartyId); err != nil {
		return "", "", fmt.Errorf("failed to register session: %w", err)
	}
	// wait longer for keygen start
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	partiesJoined, err := relayClient.WaitForSessionStart(ctx, req.SessionID)
	s.logger.WithFields(logrus.Fields{
		"sessionID":      req.SessionID,
		"parties_joined": partiesJoined,
	}).Info("Session started")

	if err != nil {
		return "", "", fmt.Errorf("failed to wait for session start: %w", err)
	}

	localStateAccessor, err := relay.NewLocalStateAccessorImp(req.LocalPartyId, keyFolder, "", "")
	if err != nil {
		return "", "", fmt.Errorf("failed to create localStateAccessor: %w", err)
	}

	tssServerImp, err := s.createTSSService(serverURL, req.SessionID, req.HexEncryptionKey, localStateAccessor, true)
	if err != nil {
		return "", "", fmt.Errorf("failed to create TSS service: %w", err)
	}

	ecdsaPubkey, eddsaPubkey := "", ""
	for attempt := 0; attempt < 3; attempt++ {
		ecdsaPubkey, eddsaPubkey, err = s.keygenWithRetry(serverURL, req, partiesJoined, tssServerImp)
		if err == nil {
			break
		}
	}

	if err != nil {
		return "", "", err
	}

	if err := relayClient.CompleteSession(req.SessionID, req.LocalPartyId); err != nil {
		s.logger.WithFields(logrus.Fields{
			"session": req.SessionID,
			"error":   err,
		}).Error("Failed to complete session")
	}

	if isCompleted, err := relayClient.CheckCompletedParties(req.SessionID, partiesJoined); err != nil || !isCompleted {
		s.logger.WithFields(logrus.Fields{
			"sessionID":   req.SessionID,
			"isCompleted": isCompleted,
			"error":       err,
		}).Error("Failed to check completed parties")
	}

	if err := relayClient.EndSession(req.SessionID); err != nil {
		s.logger.WithFields(logrus.Fields{
			"sessionID": req.SessionID,
			"error":     err,
		}).Error("Failed to end session")
	}

	err = s.BackupVault(req, partiesJoined, ecdsaPubkey, eddsaPubkey, localStateAccessor)
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

func (s *WorkerService) keygenWithRetry(serverURL string, req types.VaultCreateRequest, partiesJoined []string, tssService tss.Service) (string, string, error) {
	endCh, wg := s.startMessageDownload(serverURL, req.SessionID, req.LocalPartyId, req.HexEncryptionKey, tssService)

	resp, err := s.generateECDSAKey(tssService, req, partiesJoined)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate ECDSA key: %w", err)
	}

	respEDDSA, err := s.generateEDDSAKey(tssService, req, partiesJoined)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate EDDSA key: %w", err)
	}

	close(endCh)
	wg.Wait()

	return resp.PubKey, respEDDSA.PubKey, nil
}

func (s *WorkerService) generateECDSAKey(tssService tss.Service, req types.VaultCreateRequest, partiesJoined []string) (*tss.KeygenResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"local_party_id": req.LocalPartyId,
		"chain_code":     req.HexChainCode,
		"parties_joined": partiesJoined,
	}).Info("Start ECDSA keygen...")
	resp, err := tssService.KeygenECDSA(&tss.KeygenRequest{
		LocalPartyID: req.LocalPartyId,
		AllParties:   strings.Join(partiesJoined, ","),
		ChainCodeHex: req.HexChainCode,
	})
	if err != nil {
		return nil, fmt.Errorf("generate ECDSA key: %w", err)
	}
	s.logger.WithFields(logrus.Fields{
		"local_party_id": req.LocalPartyId,
		"pub_key":        resp.PubKey,
	}).Info("ECDSA keygen response")
	time.Sleep(time.Second)
	return resp, nil
}

func (s *WorkerService) generateEDDSAKey(tssService tss.Service, req types.VaultCreateRequest, partiesJoined []string) (*tss.KeygenResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"local_party_id": req.LocalPartyId,
		"chain_code":     req.HexChainCode,
		"parties_joined": partiesJoined,
	}).Info("Start EDDSA keygen...")
	resp, err := tssService.KeygenEdDSA(&tss.KeygenRequest{
		LocalPartyID: req.LocalPartyId,
		AllParties:   strings.Join(partiesJoined, ","),
		ChainCodeHex: req.HexChainCode,
	})
	if err != nil {
		return nil, fmt.Errorf("generate EDDSA key: %w", err)
	}
	s.logger.WithFields(logrus.Fields{
		"local_party_id": req.LocalPartyId,
		"pub_key":        resp.PubKey,
	}).Info("EDDSA keygen response")
	time.Sleep(time.Second)
	return resp, nil
}

func (s *WorkerService) BackupVault(req types.VaultCreateRequest, partiesJoined []string, ecdsaPubkey, eddsaPubkey string, localStateAccessor *relay.LocalStateAccessorImp) error {
	ecdsaKeyShare, err := localStateAccessor.GetLocalState(ecdsaPubkey)
	if err != nil {
		return fmt.Errorf("failed to get local sate: %w", err)
	}

	eddsaKeyShare, err := localStateAccessor.GetLocalState(eddsaPubkey)
	if err != nil {
		return fmt.Errorf("failed to get local sate: %w", err)
	}

	vault := &vaultType.Vault{
		Name:           req.Name,
		PublicKeyEcdsa: ecdsaPubkey,
		PublicKeyEddsa: eddsaPubkey,
		Signers:        partiesJoined,
		CreatedAt:      timestamppb.New(time.Now()),
		HexChainCode:   req.HexChainCode,
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
		LocalPartyId:  req.LocalPartyId,
		ResharePrefix: "",
	}

	isEncrypted := req.EncryptionPassword != ""
	vaultData, err := proto.Marshal(vault)
	if err != nil {
		return fmt.Errorf("failed to Marshal vault: %w", err)
	}

	if isEncrypted {
		vaultData, err = common.EncryptVault(req.EncryptionPassword, vaultData)
		if err != nil {
			return fmt.Errorf("common.EncryptVault failed: %w", err)
		}
	}
	vaultBackup := &vaultType.VaultContainer{
		Version:     1,
		Vault:       base64.StdEncoding.EncodeToString(vaultData),
		IsEncrypted: isEncrypted,
	}
	filePathName := filepath.Join(s.cfg.Server.VaultsFilePath, ecdsaPubkey+".bak")
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

	if _, err := file.Write([]byte(base64.StdEncoding.EncodeToString(vaultBackupData))); err != nil {
		return fmt.Errorf("fail to write file, err: %w", err)
	}

	return nil
}

func (s *WorkerService) createTSSService(serverURL, Session, HexEncryptionKey string, localStateAccessor tss.LocalStateAccessor, createPreParam bool) (tss.Service, error) {
	messenger := relay.NewMessenger(serverURL, Session, HexEncryptionKey)
	tssService, err := tss.NewService(messenger, localStateAccessor, createPreParam)
	if err != nil {
		return nil, fmt.Errorf("create TSS service: %w", err)
	}
	return tssService, nil
}

func (s *WorkerService) startMessageDownload(serverURL, session, key, hexEncryptionKey string, tssService tss.Service) (chan struct{}, *sync.WaitGroup) {
	s.logger.WithFields(logrus.Fields{
		"session": session,
		"key":     key,
	}).Info("Start downloading messages")

	endCh := make(chan struct{})
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go s.downloadMessages(serverURL, session, key, hexEncryptionKey, tssService, endCh, wg)
	return endCh, wg
}

func (s *WorkerService) downloadMessages(server, session, localPartyID, hexEncryptionKey string, tssServerImp tss.Service, endCh chan struct{}, wg *sync.WaitGroup) {
	var messageCache sync.Map
	defer wg.Done()
	logger := s.logger.WithFields(logrus.Fields{
		"session":        session,
		"local_party_id": localPartyID,
	})
	logger.Info("Start downloading messages from : ", server)

	for {
		select {
		case <-endCh: // we are done
			return
		case <-time.After(time.Second):
			resp, err := http.Get(server + "/message/" + session + "/" + localPartyID)
			if err != nil {
				logger.Errorf("Failed to get data from server: %v", err)
				continue
			}
			if resp.StatusCode != http.StatusOK {
				logger.Errorf("Failed to get data from server, status code is not 200 OK")
				continue
			}
			decoder := json.NewDecoder(resp.Body)
			var messages []struct {
				SessionID string   `json:"session_id,omitempty"`
				From      string   `json:"from,omitempty"`
				To        []string `json:"to,omitempty"`
				Body      string   `json:"body,omitempty"`
				Hash      string   `json:"hash,omitempty"`
			}
			if err := decoder.Decode(&messages); err != nil {
				logger.Errorf("Failed to decode data: %v", err)
				continue
			}
			for _, message := range messages {
				if message.From == localPartyID {
					continue
				}

				cacheKey := fmt.Sprintf("%s-%s-%s", session, localPartyID, message.Hash)
				if _, found := messageCache.Load(cacheKey); found {
					logger.Infof("Message already applied, skipping,hash: %s", message.Hash)
					continue
				}

				decryptedBody := message.Body
				if hexEncryptionKey != "" {
					decodedBody, err := base64.StdEncoding.DecodeString(message.Body)
					if err != nil {
						logger.Errorf("Failed to decode data: %v", err)
						continue
					}

					decryptedBody, err = decrypt(string(decodedBody), hexEncryptionKey)
					if err != nil {
						logger.Errorf("Failed to decrypt data: %v", err)
						continue
					}
				}

				if err := tssServerImp.ApplyData(decryptedBody); err != nil {
					logger.Errorf("Failed to apply data: %v", err)
					continue
				}

				messageCache.Store(cacheKey, true)
				client := http.Client{}
				req, err := http.NewRequest(http.MethodDelete, server+"/message/"+session+"/"+localPartyID+"/"+message.Hash, nil)
				if err != nil {
					logger.Errorf("Failed to delete message: %v", err)
					continue
				}
				resp, err := client.Do(req)
				if err != nil {
					logger.Errorf("Failed to delete message: %v", err)
					continue
				}

				if resp.StatusCode != http.StatusOK {
					logger.Errorf("Failed to delete message, status code is not 200 OK")
					continue
				}
			}
		}
	}
}

func decrypt(cipherText, hexKey string) (string, error) {
	var block cipher.Block
	var err error
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return "", err
	}
	cipherByte := []byte(cipherText)

	if block, err = aes.NewCipher(key); err != nil {
		return "", err
	}

	if len(cipherByte) < aes.BlockSize {
		fmt.Printf("ciphertext too short")
		return "", err
	}

	iv := cipherByte[:aes.BlockSize]
	cipherByte = cipherByte[aes.BlockSize:]

	cbc := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(cipherByte))
	cbc.CryptBlocks(plaintext, cipherByte)
	plaintext, err = unpad(plaintext)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func unpad(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("unpad: input data is empty")
	}

	paddingLen := int(data[length-1])
	if paddingLen > length || paddingLen == 0 {
		return nil, errors.New("unpad: invalid padding length")
	}

	for i := 0; i < paddingLen; i++ {
		if data[length-1-i] != byte(paddingLen) {
			return nil, errors.New("unpad: invalid padding")
		}
	}

	return data[:length-paddingLen], nil
}

func (s *WorkerService) JoinKeySign(req types.KeysignRequest) (map[string]tss.KeysignResponse, error) {
	result := map[string]tss.KeysignResponse{}
	keyFolder := s.cfg.Server.VaultsFilePath
	serverURL := s.cfg.Relay.Server
	localStateAccessor, err := relay.NewLocalStateAccessorImp("", keyFolder, req.PublicKey, req.VaultPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to create localStateAccessor: %w", err)
	}

	localPartyId := localStateAccessor.Vault.LocalPartyId
	server := relay.NewRelayClient(serverURL)

	// Let's register session here
	if err := server.RegisterSession(req.SessionID, localPartyId); err != nil {
		return nil, fmt.Errorf("failed to register session: %w", err)
	}
	// wait longer for keysign start
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute+3*time.Second)
	defer cancel()

	partiesJoined, err := server.WaitForSessionStart(ctx, req.SessionID)
	s.logger.WithFields(logrus.Fields{
		"session":        req.SessionID,
		"parties_joined": partiesJoined,
	}).Info("Session started")

	if err != nil {
		return nil, fmt.Errorf("failed to wait for session start: %w", err)
	}

	tssServerImp, err := s.createTSSService(serverURL, req.SessionID, req.HexEncryptionKey, localStateAccessor, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create TSS service: %w", err)
	}

	for _, message := range req.Messages {
		var signature *tss.KeysignResponse
		for attempt := 0; attempt < 3; attempt++ {
			signature, err = s.keysignWithRetry(serverURL, localPartyId, req, partiesJoined, tssServerImp, message)
			if err == nil {
				break
			}
		}
		if err != nil {
			return result, err
		}
		result[message] = *signature
	}

	if err := server.CompleteSession(req.SessionID, localPartyId); err != nil {
		s.logger.WithFields(logrus.Fields{
			"session": req.SessionID,
			"error":   err,
		}).Error("Failed to complete session")
	}

	return result, nil
}

func (s *WorkerService) keysignWithRetry(serverURL, localPartyId string, req types.KeysignRequest, partiesJoined []string, tssService tss.Service, msg string) (*tss.KeysignResponse, error) {
	endCh, wg := s.startMessageDownload(serverURL, req.SessionID, localPartyId, req.HexEncryptionKey, tssService)
	var signature *tss.KeysignResponse
	var err error
	if req.IsECDSA {
		signature, err = tssService.KeysignECDSA(&tss.KeysignRequest{
			PubKey:               req.PublicKey,
			MessageToSign:        msg,
			LocalPartyKey:        localPartyId,
			KeysignCommitteeKeys: strings.Join(partiesJoined, ","),
			DerivePath:           req.DerivePath,
		})
	} else {
		signature, err = tssService.KeysignEdDSA(&tss.KeysignRequest{
			PubKey:               req.PublicKey,
			MessageToSign:        msg,
			LocalPartyKey:        localPartyId,
			KeysignCommitteeKeys: strings.Join(partiesJoined, ","),
			DerivePath:           req.DerivePath,
		})
	}

	if err != nil {
		return nil, fmt.Errorf("fail to key sign: %w", err)
	}

	close(endCh)
	wg.Wait()

	return signature, nil
}
