package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/vultisig/mobile-tss-lib/tss"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/vultisig/vultisigner/internal/types"
	"github.com/vultisig/vultisigner/relay"

	vaultType "github.com/vultisig/commondata/go/vultisig/vault/v1"
)

func (s *WorkerService) JoinKeyGeneration(req types.VaultCreateRequest) (string, string, error) {
	keyFolder := s.cfg.Server.VaultsFilePath
	serverURL := s.cfg.Relay.Server
	relayClient := relay.NewRelayClient(serverURL)

	if req.StartSession {
		if err := relayClient.StartSession(req.SessionID, req.Parties); err != nil {
			return "", "", fmt.Errorf("failed to start session: %w", err)
		}
	} else {
		// Let's register session here
		if err := relayClient.RegisterSession(req.SessionID, req.LocalPartyId); err != nil {
			return "", "", fmt.Errorf("failed to register session: %w", err)
		}
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

	localStateAccessor, err := relay.NewLocalStateAccessorImp(req.LocalPartyId, keyFolder, "", "", s.blockStorage)
	if err != nil {
		return "", "", fmt.Errorf("failed to create localStateAccessor: %w", err)
	}

	tssServerImp, err := s.createTSSService(serverURL, req.SessionID, req.HexEncryptionKey, localStateAccessor, true)
	if err != nil {
		return "", "", fmt.Errorf("failed to create TSS service: %w", err)
	}

	ecdsaPubkey, eddsaPubkey := "", ""
	endCh, wg := s.startMessageDownload(serverURL, req.SessionID, req.LocalPartyId, req.HexEncryptionKey, tssServerImp, "")
	for attempt := 0; attempt < 3; attempt++ {
		ecdsaPubkey, eddsaPubkey, err = s.keygenWithRetry(req, partiesJoined, tssServerImp)
		if err == nil {
			break
		}
	}
	close(endCh)
	wg.Wait()
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

	err = s.BackupVault(req, partiesJoined, ecdsaPubkey, eddsaPubkey, localStateAccessor)
	if err != nil {
		return "", "", fmt.Errorf("failed to backup vault: %w", err)
	}

	return ecdsaPubkey, eddsaPubkey, nil
}

func (s *WorkerService) keygenWithRetry(req types.VaultCreateRequest, partiesJoined []string, tssService tss.Service) (string, string, error) {
	resp, err := s.generateECDSAKey(tssService, req, partiesJoined)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate ECDSA key: %w", err)
	}

	respEDDSA, err := s.generateEDDSAKey(tssService, req, partiesJoined)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate EDDSA key: %w", err)
	}
	return resp.PubKey, respEDDSA.PubKey, nil
}

func (s *WorkerService) generateECDSAKey(tssService tss.Service, req types.VaultCreateRequest, partiesJoined []string) (*tss.KeygenResponse, error) {
	defer s.measureTime("worker.vault.create.ECDSA.latency", time.Now(), []string{})
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
	defer s.measureTime("worker.vault.create.EDDSA.latency", time.Now(), []string{})
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

func (s *WorkerService) BackupVault(req types.VaultCreateRequest,
	partiesJoined []string,
	ecdsaPubkey, eddsaPubkey string,
	localStateAccessor *relay.LocalStateAccessorImp) error {
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
	return s.SaveVaultAndScheduleEmail(vault, req.EncryptionPassword, req.Email)
}

func (s *WorkerService) createTSSService(serverURL, Session, HexEncryptionKey string, localStateAccessor tss.LocalStateAccessor, createPreParam bool) (*tss.ServiceImpl, error) {
	messenger := relay.NewMessenger(serverURL, Session, HexEncryptionKey)
	tssService, err := tss.NewService(messenger, localStateAccessor, createPreParam)
	if err != nil {
		return nil, fmt.Errorf("create TSS service: %w", err)
	}
	return tssService, nil
}

func (s *WorkerService) startMessageDownload(serverURL, session, key, hexEncryptionKey string, tssService tss.Service, messageID string) (chan struct{}, *sync.WaitGroup) {
	s.logger.WithFields(logrus.Fields{
		"session": session,
		"key":     key,
	}).Info("Start downloading messages")

	endCh := make(chan struct{})
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go s.downloadMessages(serverURL, session, key, hexEncryptionKey, tssService, endCh, messageID, wg)
	return endCh, wg
}

func (s *WorkerService) downloadMessages(server, session, localPartyID, hexEncryptionKey string, tssServerImp tss.Service, endCh chan struct{}, messageID string, wg *sync.WaitGroup) {
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
			logger.Info("Stop downloading messages")
			return
		case <-time.After(time.Second):

			req, err := http.NewRequest(http.MethodGet, server+"/message/"+session+"/"+localPartyID, nil)
			if err != nil {
				logger.Errorf("Fail to create request,err: %s", err)
				return
			}
			if messageID != "" {
				req.Header.Set("message_id", messageID)
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				logger.Errorf("Failed to get data from server,err: %s", err)
				continue
			}
			if resp.StatusCode != http.StatusOK {
				logger.Errorf("Failed to get data from server, status code is not 200 OK")
				continue
			}
			decoder := json.NewDecoder(resp.Body)
			var messages []struct {
				SessionID  string   `json:"session_id,omitempty"`
				From       string   `json:"from,omitempty"`
				To         []string `json:"to,omitempty"`
				Body       string   `json:"body,omitempty"`
				Hash       string   `json:"hash,omitempty"`
				SequenceNo int64    `json:"sequence_no,omitempty"`
			}
			if err := decoder.Decode(&messages); err != nil {
				logger.Errorf("Failed to decode data: %v", err)
				continue
			}
			sort.Slice(messages, func(i, j int) bool {
				return messages[i].SequenceNo < messages[j].SequenceNo
			})
			for _, message := range messages {
				cacheKey := fmt.Sprintf("%s-%s-%s", session, localPartyID, message.Hash)
				if messageID != "" {
					cacheKey = fmt.Sprintf("%s-%s-%s-%s", session, localPartyID, messageID, message.Hash)
				}
				if _, found := messageCache.Load(cacheKey); found {
					logger.Infof("Message already applied, skipping,hash: %s", message.Hash)
					continue
				}

				decodedBody, err := base64.StdEncoding.DecodeString(message.Body)
				if err != nil {
					logger.Errorf("Failed to decode data: %v", err)
					continue
				}

				decryptedBody, err := decrypt(string(decodedBody), hexEncryptionKey)
				if err != nil {
					logger.Errorf("Failed to decrypt data: %v", err)
					continue
				}

				if err := tssServerImp.ApplyData(decryptedBody); err != nil {
					logger.Errorf("Failed to apply data: %v", err)
					continue
				}

				messageCache.Store(cacheKey, true)
				client := http.Client{}
				reqDel, err := http.NewRequest(http.MethodDelete, server+"/message/"+session+"/"+localPartyID+"/"+message.Hash, nil)
				if err != nil {
					logger.Errorf("Failed to delete message: %v", err)
					continue
				}
				if messageID != "" {
					reqDel.Header.Set("message_id", messageID)
				}
				resp, err := client.Do(reqDel)
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
	localStateAccessor, err := relay.NewLocalStateAccessorImp("", keyFolder, req.PublicKey, req.VaultPassword, s.blockStorage)
	if err != nil {
		return nil, fmt.Errorf("failed to create localStateAccessor: %w", err)
	}

	localPartyId := localStateAccessor.Vault.LocalPartyId
	server := relay.NewRelayClient(serverURL)

	// Let's register session here
	if req.StartSession {
		if err := server.StartSession(req.SessionID, req.Parties); err != nil {
			return nil, fmt.Errorf("failed to start session: %w", err)
		}
	} else {
		if err := server.RegisterSessionWithRetry(req.SessionID, localPartyId); err != nil {
			return nil, fmt.Errorf("failed to register session: %w", err)
		}
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
			signature, err = s.keysignWithRetry(serverURL,
				localPartyId,
				req,
				partiesJoined,
				tssServerImp,
				message,
				localStateAccessor.Vault.PublicKeyEddsa)
			if err == nil {
				break
			}
		}
		if err != nil {
			return result, err
		}
		if signature == nil {
			return result, fmt.Errorf("signature is nil")
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

func (s *WorkerService) keysignWithRetry(serverURL, localPartyId string,
	req types.KeysignRequest,
	partiesJoined []string,
	tssService tss.Service,
	msg string,
	publicKeyEdDSA string) (*tss.KeysignResponse, error) {
	messageID := hex.EncodeToString(md5.New().Sum([]byte(msg)))
	msgBuf, err := hex.DecodeString(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to decode message: %w", err)
	}
	messageToSign := base64.StdEncoding.EncodeToString(msgBuf)
	endCh, wg := s.startMessageDownload(serverURL, req.SessionID, localPartyId, req.HexEncryptionKey, tssService, messageID)

	var signature *tss.KeysignResponse
	if req.IsECDSA {
		signature, err = tssService.KeysignECDSA(&tss.KeysignRequest{
			PubKey:               req.PublicKey,
			MessageToSign:        messageToSign,
			LocalPartyKey:        localPartyId,
			KeysignCommitteeKeys: strings.Join(partiesJoined, ","),
			DerivePath:           req.DerivePath,
		})
	} else {
		signature, err = tssService.KeysignEdDSA(&tss.KeysignRequest{
			PubKey:               publicKeyEdDSA, // request public key should be EdDSA public key
			MessageToSign:        messageToSign,
			LocalPartyKey:        localPartyId,
			KeysignCommitteeKeys: strings.Join(partiesJoined, ","),
			DerivePath:           req.DerivePath,
		})
	}

	client := relay.NewRelayClient(serverURL)
	if err == nil {
		if err := client.MarkKeysignComplete(req.SessionID, messageID, *signature); err != nil {
			s.logger.Errorf("fail to mark keysign complete: %v", err)
		}
	} else {
		s.logger.Errorf("fail to key sign: %v", err)
		sigResp, err := client.CheckKeysignComplete(req.SessionID, messageID)
		if err == nil && sigResp != nil {
			signature = sigResp
		}
	}
	close(endCh)
	wg.Wait()
	return signature, nil
}
