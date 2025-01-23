package service

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/sirupsen/logrus"
	"github.com/vultisig/mobile-tss-lib/tss"

	"github.com/vultisig/vultisigner/common"
	"github.com/vultisig/vultisigner/internal/types"
	"github.com/vultisig/vultisigner/relay"
)

func (t *DKLSTssService) ProcessDKLSKeysign(req types.KeysignRequest) (map[string]tss.KeysignResponse, error) {
	result := map[string]tss.KeysignResponse{}
	keyFolder := t.cfg.Server.VaultsFilePath
	localStateAccessor, err := relay.NewLocalStateAccessorImp(keyFolder, req.PublicKey, req.VaultPassword, t.blockStorage)
	if err != nil {
		return nil, fmt.Errorf("failed to create localStateAccessor: %w", err)
	}
	t.localStateAccessor = localStateAccessor
	localPartyID := localStateAccessor.Vault.LocalPartyId
	relayClient := relay.NewRelayClient(t.cfg.Relay.Server)
	if err := relayClient.RegisterSession(req.SessionID, localPartyID); err != nil {
		return nil, fmt.Errorf("failed to start session: %w", err)
	}
	// wait longer for keysign start
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute+3*time.Second)
	defer cancel()

	partiesJoined, err := relayClient.WaitForSessionStart(ctx, req.SessionID)
	t.logger.WithFields(logrus.Fields{
		"session":        req.SessionID,
		"parties_joined": partiesJoined,
	}).Info("Session started")

	if err != nil {
		return nil, fmt.Errorf("failed to wait for session start: %w", err)
	}
	// start to do keysign
	for _, msg := range req.Messages {
		sig, err := t.keysignWithRetry(req.SessionID, req.HexEncryptionKey, req.PublicKey, !req.IsECDSA, msg, req.DerivePath, localPartyID, partiesJoined)
		if err != nil {
			return result, fmt.Errorf("failed to keysign: %w", err)
		}
		if sig == nil {
			return result, fmt.Errorf("failed to keysign: signature is nil")
		}
		result[msg] = *sig
	}
	if err := relayClient.CompleteSession(req.SessionID, localPartyID); err != nil {
		t.logger.WithFields(logrus.Fields{
			"session": req.SessionID,
			"error":   err,
		}).Error("Failed to complete session")
	}

	return result, nil
}
func (t *DKLSTssService) keysignWithRetry(sessionID string,
	hexEncryptionKey string,
	publicKey string,
	isEdDSA bool,
	message string,
	derivePath string,
	localPartyID string,
	keysignCommittee []string) (*tss.KeysignResponse, error) {
	for i := 0; i < 3; i++ {
		keysignResult, err := t.keysign(sessionID,
			hexEncryptionKey,
			publicKey,
			isEdDSA,
			message,
			derivePath,
			localPartyID,
			keysignCommittee, i)
		if err != nil {
			t.logger.WithFields(logrus.Fields{
				"session_id":        sessionID,
				"public_key_ecdsa":  publicKey,
				"message":           message,
				"derive_path":       derivePath,
				"local_party_id":    localPartyID,
				"keysign_committee": keysignCommittee,
				"attempt":           i,
			}).Error(err)
			time.Sleep(50 * time.Millisecond)
			continue
		} else {
			return keysignResult, nil
		}
	}
	return nil, fmt.Errorf("fail to keysign after max retry")
}

func (t *DKLSTssService) keysign(sessionID string,
	hexEncryptionKey string,
	publicKey string,
	isEdDSA bool,
	message string,
	derivePath string,
	localPartyID string,
	keysignCommittee []string,
	attempt int) (*tss.KeysignResponse, error) {
	if publicKey == "" {
		return nil, fmt.Errorf("public key is empty")
	}
	if message == "" {
		return nil, fmt.Errorf("message is empty")
	}
	if derivePath == "" {
		return nil, fmt.Errorf("derive path is empty")
	}
	if localPartyID == "" {
		return nil, fmt.Errorf("local party id is empty")
	}
	if len(keysignCommittee) == 0 {
		return nil, fmt.Errorf("keysign committee is empty")
	}

	relayClient := relay.NewRelayClient(t.cfg.Relay.Server)
	mpcWrapper := t.GetMPCKeygenWrapper(isEdDSA)
	t.logger.WithFields(logrus.Fields{
		"session_id":        sessionID,
		"public_key_ecdsa":  publicKey,
		"message":           message,
		"derive_path":       derivePath,
		"local_party_id":    localPartyID,
		"keysign_committee": keysignCommittee,
		"attempt":           attempt,
	}).Info("Keysign")

	// we need to get the shares
	keyshare, err := t.localStateAccessor.GetLocalState(publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get keyshare: %w", err)
	}
	keyshareBytes, err := base64.StdEncoding.DecodeString(keyshare)
	if err != nil {
		return nil, fmt.Errorf("failed to decode keyshare: %w", err)
	}
	keyshareHandle, err := mpcWrapper.KeyshareFromBytes(keyshareBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to create keyshare from bytes: %w", err)
	}
	defer func() {
		if err := mpcWrapper.KeyshareFree(keyshareHandle); err != nil {
			t.logger.Error("failed to free keyshare", "error", err)
		}
	}()

	msgHash := sha256.Sum256([]byte(message))
	messageID := hex.EncodeToString(msgHash[:])
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	// retrieve the setup Message
	encryptedEncodedSetupMsg, err := relayClient.WaitForPayload(ctx, sessionID, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get setup message: %w", err)
	}
	encryptedSetupMsg, err := base64.StdEncoding.DecodeString(encryptedEncodedSetupMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to decode setup message: %w", err)
	}
	// decrypt the setup message
	decryptedSetupMsg, err := common.DecryptGCM(encryptedSetupMsg, hexEncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt setup message: %w", err)
	}
	setupMessageBytes, err := base64.StdEncoding.DecodeString(string(decryptedSetupMsg))
	if err != nil {
		return nil, fmt.Errorf("failed to decode setup message: %w", err)
	}
	messageHashInSetupMsg, err := mpcWrapper.DecodeMessage(setupMessageBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decode message: %w", err)
	}
	msgRawBytes, err := hex.DecodeString(message)
	if err != nil {
		return nil, fmt.Errorf("failed to decode message: %w", err)
	}
	if !bytes.Equal(messageHashInSetupMsg, msgRawBytes) {
		return nil, fmt.Errorf("message hash in setup message is not equal to the message, stop keysign")
	}
	sessionHandle, err := mpcWrapper.SignSessionFromSetup(setupMessageBytes, []byte(localPartyID), keyshareHandle)
	if err != nil {
		return nil, fmt.Errorf("failed to create session from setup message: %w", err)
	}
	defer func() {
		if err := mpcWrapper.SignSessionFree(sessionHandle); err != nil {
			t.logger.Error("failed to free keysign session", "error", err)
		}
	}()
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		if err := t.processKeysignOutbound(sessionHandle, sessionID, hexEncryptionKey, keysignCommittee, localPartyID, wg, isEdDSA); err != nil {
			t.logger.Error("failed to process keygen outbound", "error", err)
		}
	}()
	sig, err := t.processKeysignInbound(sessionHandle, sessionID, hexEncryptionKey, localPartyID, isEdDSA, messageID, wg)
	wg.Wait()
	t.logger.Infoln("Keysign result is:", len(sig))
	resp := &tss.KeysignResponse{
		Msg: message,
		R:   hex.EncodeToString(sig[:32]),
		S:   hex.EncodeToString(sig[32:64]),
	}
	if isEdDSA {
		pubKeyBytes, err := hex.DecodeString(publicKey)
		if err != nil {
			return nil, fmt.Errorf("failed to decode public key: %w", err)
		}

		if ed25519.Verify(pubKeyBytes, msgRawBytes, sig) {
			t.logger.Infoln("Signature is valid")
		} else {
			t.logger.Error("Signature is invalid")
		}
	} else {
		childPublicKey, err := mpcWrapper.KeyshareDeriveChildPublicKey(keyshareHandle, []byte(strings.Replace(derivePath, "'", "", -1)))
		if err != nil {
			return nil, fmt.Errorf("failed to derive child public key: %w", err)
		}
		if len(sig) != 65 {
			return nil, fmt.Errorf("signature length is not 64")
		}
		r := sig[:32]
		s := sig[32:64]
		recovery := sig[64]
		resp.RecoveryID = hex.EncodeToString([]byte{recovery})
		publicKeyECDSA, err := secp256k1.ParsePubKey(childPublicKey)
		if err != nil {
			return nil, fmt.Errorf("failed to parse public key: %w", err)
		}

		if ecdsa.Verify(publicKeyECDSA.ToECDSA(), msgRawBytes, new(big.Int).SetBytes(r), new(big.Int).SetBytes(s)) {
			t.logger.Infoln("Signature is valid")
		} else {
			t.logger.Error("Signature is invalid")
		}
	}
	return resp, nil
}
func (t *DKLSTssService) processKeysignOutbound(handle Handle,
	sessionID string,
	hexEncryptionKey string,
	parties []string,
	localPartyID string,
	wg *sync.WaitGroup, isEdDSA bool) error {
	defer wg.Done()
	messenger := relay.NewMessenger(t.cfg.Relay.Server, sessionID, hexEncryptionKey, true)
	mpcWrapper := t.GetMPCKeygenWrapper(isEdDSA)
	for {
		outbound, err := mpcWrapper.SignSessionOutputMessage(handle)
		if err != nil {
			t.logger.Error("failed to get output message", "error", err)
		}
		if len(outbound) == 0 {
			if t.isKeysignFinished.Load() {
				// we are finished
				return nil
			}
			time.Sleep(time.Millisecond * 100)
			continue
		}
		fmt.Println("Outbound message is:", len(outbound))
		encodedOutbound := base64.StdEncoding.EncodeToString(outbound)
		for i := 0; i < len(parties); i++ {
			receiver, err := mpcWrapper.SignSessionMessageReceiver(handle, outbound, i)
			if err != nil {
				t.logger.Error("failed to get receiver message", "error", err)
			}
			if len(receiver) == 0 {
				break
			}

			t.logger.Infoln("Sending message to", string(receiver))
			// send the message to the receiver
			if err := messenger.Send(localPartyID, string(receiver), encodedOutbound); err != nil {
				t.logger.Errorf("failed to send message: %v", err)
			}
		}
	}
}

func (t *DKLSTssService) processKeysignInbound(handle Handle,
	sessionID string,
	hexEncryptionKey string,
	localPartyID string,
	isEdDSA bool,
	messageID string,
	wg *sync.WaitGroup) ([]byte, error) {
	defer wg.Done()
	var messageCache sync.Map
	mpcWrapper := t.GetMPCKeygenWrapper(isEdDSA)
	for {
		select {
		case <-time.After(time.Minute):
			// set isKeygenFinished to true , so the other go routine can be stopped
			t.isKeysignFinished.Store(true)
			return nil, TssKeyGenTimeout
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
				if messageID != "" {
					cacheKey = fmt.Sprintf("%s-%s-%s-%s", sessionID, localPartyID, messageID, message.Hash)
				}
				if _, found := messageCache.Load(cacheKey); found {
					t.logger.Infof("Message already applied, skipping,hash: %s", message.Hash)
					continue
				}
				encryptedBody, err := base64.StdEncoding.DecodeString(message.Body)
				if err != nil {
					t.logger.Error("fail to decode message", "error", err)
					continue
				}
				// decrypt the message
				stillEncodedBody, err := common.DecryptGCM(encryptedBody, hexEncryptionKey)
				if err != nil {
					t.logger.Error("fail to decrypt message", "error", err)
					continue
				}
				rawBody, err := base64.StdEncoding.DecodeString(string(stillEncodedBody))
				if err != nil {
					t.logger.Error("fail to decode inbound message", "error", err)
					continue
				}
				// decode to get raw message
				t.logger.Infoln("Received message from", message.From)
				isFinished, err := mpcWrapper.SignSessionInputMessage(handle, rawBody)
				if err != nil {
					t.logger.Error("fail to apply input message", "error", err)
					continue
				}
				messageCache.Store(cacheKey, true)
				hashStr := message.Hash
				if err := t.deleteMessageFromServer(sessionID, localPartyID, hashStr, messageID); err != nil {
					t.logger.Error("fail to delete message", "error", err)
				}
				if isFinished {
					t.logger.Infoln("keysign finished")
					result, err := mpcWrapper.SignSessionFinish(handle)
					if err != nil {
						t.logger.Error("fail to finish keygen", "error", err)
						return nil, err
					}
					encodedKeysignResult := base64.StdEncoding.EncodeToString(result)
					t.logger.Infof("Keysign result: %s", encodedKeysignResult)
					t.isKeysignFinished.Store(true)
					return result, nil
				}
			}
		}
	}
}
