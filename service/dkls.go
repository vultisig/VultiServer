package service

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/vultisig/vultisigner/common"
	"github.com/vultisig/vultisigner/config"
	"github.com/vultisig/vultisigner/internal/types"
	"github.com/vultisig/vultisigner/relay"
	"github.com/vultisig/vultisigner/storage"
)

var TssKeyGenTimeout = errors.New("keygen timeout")

type DKLSTssService struct {
	cfg                config.Config
	messenger          *relay.MessengerImp
	logger             *logrus.Logger
	localStateAccessor *relay.LocalStateAccessorImp
	isKeygenFinished   *atomic.Bool
	isKeysignFinished  *atomic.Bool
	blockStorage       *storage.BlockStorage
	backup             Backup
}

func NewDKLSTssService(cfg config.Config,
	blockStorage *storage.BlockStorage, backupInterface Backup) (*DKLSTssService, error) {
	localStateAccessor, err := relay.NewLocalStateAccessorImp(cfg.Server.VaultsFilePath, "", "", blockStorage)
	if err != nil {
		return nil, fmt.Errorf("fail to create local state accessor: %w", err)
	}
	return &DKLSTssService{
		cfg:                cfg,
		logger:             logrus.WithField("service", "dkls").Logger,
		isKeygenFinished:   &atomic.Bool{},
		isKeysignFinished:  &atomic.Bool{},
		blockStorage:       blockStorage,
		localStateAccessor: localStateAccessor,
		backup:             backupInterface,
	}, nil
}

func (t *DKLSTssService) GetMPCKeygenWrapper(isEdDSA bool) *MPCWrapperImp {
	return NewMPCWrapperImp(isEdDSA)
}

func (t *DKLSTssService) ProceeDKLSKeygen(req types.VaultCreateRequest) (string, string, error) {
	serverURL := t.cfg.Relay.Server
	relayClient := relay.NewRelayClient(serverURL)

	// Let's register session here
	if err := relayClient.RegisterSession(req.SessionID, req.LocalPartyId); err != nil {
		return "", "", fmt.Errorf("failed to register session: %w", err)
	}
	// wait longer for keygen start
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	partiesJoined, err := relayClient.WaitForSessionStart(ctx, req.SessionID)
	t.logger.WithFields(logrus.Fields{
		"sessionID":      req.SessionID,
		"parties_joined": partiesJoined,
	}).Info("Session started")

	if err != nil {
		return "", "", fmt.Errorf("failed to wait for session start: %w", err)
	}
	// create ECDSA key
	publicKeyECDSA, chainCodeECDSA, err := t.keygenWithRetry(req.SessionID, req.HexEncryptionKey, req.LocalPartyId, false, partiesJoined)
	if err != nil {
		return "", "", fmt.Errorf("failed to keygen ECDSA: %w", err)
	}
	// create EdDSA key
	publicKeyEdDSA, _, err := t.keygenWithRetry(req.SessionID, req.HexEncryptionKey, req.LocalPartyId, true, partiesJoined)
	if err != nil {
		return "", "", fmt.Errorf("failed to keygen EdDSA: %w", err)
	}

	if err := relayClient.CompleteSession(req.SessionID, req.LocalPartyId); err != nil {
		t.logger.WithFields(logrus.Fields{
			"session": req.SessionID,
			"error":   err,
		}).Error("Failed to complete session")
	}

	if isCompleted, err := relayClient.CheckCompletedParties(req.SessionID, partiesJoined); err != nil || !isCompleted {
		t.logger.WithFields(logrus.Fields{
			"sessionID":   req.SessionID,
			"isCompleted": isCompleted,
			"error":       err,
		}).Error("Failed to check completed parties")
	}
	if t.backup == nil {
		return publicKeyECDSA, publicKeyEdDSA, nil
	}

	err = t.backup.BackupVault(req, partiesJoined, publicKeyECDSA, publicKeyEdDSA, chainCodeECDSA, t.localStateAccessor)
	if err != nil {
		return "", "", fmt.Errorf("failed to backup vault: %w", err)
	}
	return publicKeyECDSA, publicKeyEdDSA, nil
}

func (t *DKLSTssService) keygenWithRetry(sessionID string,
	hexEncryptionKey string,
	localPartyID string,
	isEdDSA bool,
	keygenCommittee []string) (string, string, error) {
	for i := 0; i < 3; i++ {
		publicKey, chainCode, err := t.keygen(sessionID, hexEncryptionKey, localPartyID, isEdDSA, keygenCommittee, i)
		if err != nil {
			t.logger.WithFields(logrus.Fields{
				"session_id":       sessionID,
				"local_party_id":   localPartyID,
				"keygen_committee": keygenCommittee,
				"attempt":          i,
			}).Error(err)
			time.Sleep(50 * time.Millisecond)
			continue
		} else {
			return publicKey, chainCode, nil
		}
	}
	return "", "", fmt.Errorf("fail to keygen after max retry")
}

func (t *DKLSTssService) keygen(sessionID string,
	hexEncryptionKey string,
	localPartyID string,
	isEdDSA bool,
	keygenCommittee []string,
	attempt int) (string, string, error) {
	t.logger.WithFields(logrus.Fields{
		"session_id":       sessionID,
		"local_party_id":   localPartyID,
		"keygen_committee": keygenCommittee,
		"attempt":          attempt,
	}).Info("Keygen")
	t.isKeygenFinished.Store(false)
	relayClient := relay.NewRelayClient(t.cfg.Relay.Server)
	mpcKeygenWrapper := t.GetMPCKeygenWrapper(isEdDSA)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	// retrieve the setup Message
	encryptedEncodedSetupMsg, err := relayClient.WaitForPayload(ctx, sessionID, "")
	if err != nil {
		return "", "", fmt.Errorf("failed to get setup message: %w", err)
	}
	encryptedSetupMsg, err := base64.StdEncoding.DecodeString(encryptedEncodedSetupMsg)
	if err != nil {
		return "", "", fmt.Errorf("failed to decode setup message: %w", err)
	}
	// decrypt the setup message
	decryptedSetupMsg, err := common.DecryptGCM(encryptedSetupMsg, hexEncryptionKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to decrypt setup message: %w", err)
	}
	setupMessageBytes, err := base64.StdEncoding.DecodeString(string(decryptedSetupMsg))
	if err != nil {
		return "", "", fmt.Errorf("failed to decode setup message: %w", err)
	}

	handle, err := mpcKeygenWrapper.KeygenSessionFromSetup(setupMessageBytes, []byte(localPartyID))
	if err != nil {
		return "", "", fmt.Errorf("failed to create session from setup message: %w", err)
	}
	defer func() {
		if err := mpcKeygenWrapper.KeygenSessionFree(handle); err != nil {
			t.logger.Error("failed to free keygen session", "error", err)
		}
	}()
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		if err := t.processKeygenOutbound(handle, sessionID, hexEncryptionKey, keygenCommittee, localPartyID, isEdDSA, wg); err != nil {
			t.logger.Error("failed to process keygen outbound", "error", err)
		}
	}()
	publicKey, chainCode, err := t.processKeygenInbound(handle, sessionID, hexEncryptionKey, isEdDSA, localPartyID, wg)
	wg.Wait()
	return publicKey, chainCode, err
}

func (t *DKLSTssService) processKeygenOutbound(handle Handle,
	sessionID string,
	hexEncryptionKey string,
	parties []string,
	localPartyID string,
	isEdDSA bool,
	wg *sync.WaitGroup) error {
	defer wg.Done()
	messenger := relay.NewMessenger(t.cfg.Relay.Server, sessionID, hexEncryptionKey, true)
	mpcKeygenWrapper := t.GetMPCKeygenWrapper(isEdDSA)
	for {
		outbound, err := mpcKeygenWrapper.KeygenSessionOutputMessage(handle)
		if err != nil {
			t.logger.Error("failed to get output message", "error", err)
		}
		if len(outbound) == 0 {
			if t.isKeygenFinished.Load() {
				// we are finished
				return nil
			}
			time.Sleep(time.Millisecond * 100)
			continue
		}
		encodedOutbound := base64.StdEncoding.EncodeToString(outbound)
		for i := 0; i < len(parties); i++ {
			receiver, err := mpcKeygenWrapper.KeygenSessionMessageReceiver(handle, outbound, i)
			if err != nil {
				t.logger.Error("failed to get receiver message", "error", err)
			}
			if len(receiver) == 0 {
				break
			}

			t.logger.Infoln("Sending message to", receiver)
			// send the message to the receiver
			if err := messenger.Send(localPartyID, string(receiver), encodedOutbound); err != nil {
				t.logger.Errorf("failed to send message: %v", err)
			}
		}
	}
}

func (t *DKLSTssService) processKeygenInbound(handle Handle,
	sessionID string,
	hexEncryptionKey string,
	isEdDSA bool,
	localPartyID string,
	wg *sync.WaitGroup) (string, string, error) {
	defer wg.Done()
	var messageCache sync.Map
	mpcKeygenWrapper := t.GetMPCKeygenWrapper(isEdDSA)
	for {
		select {
		case <-time.After(time.Minute):
			// set isKeygenFinished to true , so the other go routine can be stopped
			t.isKeygenFinished.Store(true)
			return "", "", TssKeyGenTimeout
		case <-time.After(time.Millisecond * 100):
			resp, err := http.Get(t.cfg.Relay.Server + "/message/" + sessionID + "/" + localPartyID)
			if err != nil {
				t.logger.Error("fail to get data from server", "error", err)
				continue
			}
			if resp.StatusCode != http.StatusOK {
				t.logger.Debug("fail to get data from server", "status", resp.Status)
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
				if err != io.EOF {
					t.logger.Error("fail to decode messages", "error", err)
				}
				continue
			}
			sort.Slice(messages, func(i, j int) bool {
				return messages[i].SequenceNo < messages[j].SequenceNo
			})
			for _, message := range messages {
				if message.From == localPartyID {
					continue
				}
				cacheKey := fmt.Sprintf("%s-%s-%s", sessionID, localPartyID, message.Hash)
				if _, found := messageCache.Load(cacheKey); found {
					t.logger.Infof("Message already applied, skipping,hash: %s", message.Hash)
					continue
				}
				decodedBody, err := base64.StdEncoding.DecodeString(message.Body)
				if err != nil {
					t.logger.Error("fail to decode message", "error", err)
					continue
				}
				rawBody, err := common.DecryptGCM(decodedBody, hexEncryptionKey)
				if err != nil {
					t.logger.Error("fail to decrypt message", "error", err)
					continue
				}
				inboundBody, err := base64.StdEncoding.DecodeString(string(rawBody))
				if err != nil {
					t.logger.Error("fail to decode inbound message", "error", err)
					continue
				}
				t.logger.Infoln("Received message from", message.From)
				isFinished, err := mpcKeygenWrapper.KeygenSessionInputMessage(handle, inboundBody)
				if err != nil {
					t.logger.Error("fail to apply input message", "error", err)
					continue
				}

				if err := t.deleteMessageFromServer(sessionID, localPartyID, message.Hash, ""); err != nil {
					t.logger.Error("fail to delete message", "error", err)
				}
				if isFinished {
					t.logger.Infoln("Keygen finished")
					result, err := mpcKeygenWrapper.KeygenSessionFinish(handle)
					if err != nil {
						t.logger.Error("fail to finish keygen", "error", err)
						return "", "", err
					}
					buf, err := mpcKeygenWrapper.KeyshareToBytes(result)
					if err != nil {
						t.logger.Error("fail to convert keyshare to bytes", "error", err)
						return "", "", err
					}
					encodedShare := base64.StdEncoding.EncodeToString(buf)
					publicKeyECDSABytes, err := mpcKeygenWrapper.KeysharePublicKey(result)
					if err != nil {
						t.logger.Error("fail to get public key", "error", err)
						return "", "", err
					}
					encodedPublicKey := hex.EncodeToString(publicKeyECDSABytes)
					t.logger.Infof("Public key: %s", encodedPublicKey)
					chainCode := ""
					if !isEdDSA {
						chainCodeBytes, err := mpcKeygenWrapper.KeyshareChainCode(result)
						if err != nil {
							t.logger.Error("fail to get chain code", "error", err)
							return "", "", err
						}
						chainCode = hex.EncodeToString(chainCodeBytes)
					}
					// This sleep give the local party a chance to send last message to others
					t.isKeygenFinished.Store(true)
					err = t.localStateAccessor.SaveLocalState(encodedPublicKey, encodedShare)
					return encodedPublicKey, chainCode, err
				}
			}
		}
	}
}
func (t *DKLSTssService) deleteMessageFromServer(sessionID, localPartyID, hash, messageID string) error {
	client := http.Client{}
	req, err := http.NewRequest(http.MethodDelete, t.cfg.Relay.Server+"/message/"+sessionID+"/"+localPartyID+"/"+hash, nil)
	if err != nil {
		return fmt.Errorf("fail to delete message: %w", err)
	}
	if messageID != "" {
		req.Header.Add("message_id", messageID)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("fail to delete message: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fail to delete message: status %s", resp.Status)
	}
	return nil
}
func (t *DKLSTssService) convertKeygenCommitteeToBytes(paries []string) ([]byte, error) {
	if len(paries) == 0 {
		return nil, fmt.Errorf("no parties provided")
	}
	result := make([]byte, 0)
	for _, party := range paries {
		result = append(result, []byte(party)...)
		result = append(result, byte(0))
	}
	// remove the last 0
	if len(result) > 0 {
		result = result[:len(result)-1]
	}
	return result, nil
}
