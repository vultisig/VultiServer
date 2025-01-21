package service

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/sirupsen/logrus"
	"github.com/vultisig/mobile-tss-lib/coordinator"

	"github.com/vultisig/vultisigner/common"
	"github.com/vultisig/vultisigner/relay"
)

var TssKeyGenTimeout = errors.New("keygen timeout")

type DKLSTssService struct {
	relayServer        string
	messenger          *relay.MessengerImp
	logger             *logrus.Logger
	localStateAccessor *relay.LocalStateAccessorImp
	isKeygenFinished   *atomic.Bool
	isKeysignFinished  *atomic.Bool
	isEdDSA            bool
}

func NewDKLSTssService(server string, localStateAccessor *relay.LocalStateAccessorImp, isEdDSA bool) (*DKLSTssService, error) {
	return &DKLSTssService{
		relayServer:        server,
		messenger:          nil,
		localStateAccessor: localStateAccessor,
		logger:             logrus.WithField("service", "tss").Logger,
		isKeygenFinished:   &atomic.Bool{},
		isKeysignFinished:  &atomic.Bool{},
		isEdDSA:            isEdDSA,
	}, nil
}
func (t *DKLSTssService) GetMPCKeygenWrapper() *MPCWrapperImp {
	return NewMPCWrapperImp(t.isEdDSA)
}
func (t *DKLSTssService) Keygen(sessionID string,
	chainCode string,
	hexEncryptionKey string,
	localPartyID string,
	keygenCommittee []string,
	isInitiateDevice bool) error {
	t.logger.WithFields(logrus.Fields{
		"session_id":         sessionID,
		"chain_code":         chainCode,
		"local_party_id":     localPartyID,
		"keygen_committee":   keygenCommittee,
		"is_initiate_device": isInitiateDevice,
	}).Info("Keygen")
	relayClient := relay.NewRelayClient(t.relayServer)
	if err := relayClient.RegisterSession(sessionID, localPartyID); err != nil {
		return fmt.Errorf("failed to register session: %w", err)
	}
	mpcKeygenWrapper := t.GetMPCKeygenWrapper()
	var encodedSetupMsg string = ""
	if isInitiateDevice {
		if coordinator.WaitAllParties(keygenCommittee, t.relayServer, sessionID) != nil {
			return fmt.Errorf("failed to wait for all parties to join")
		}
		fmt.Println("I am the leader , construct the setup message")
		keygenCommitteeBytes, err := t.convertKeygenCommitteeToBytes(keygenCommittee)
		if err != nil {
			return fmt.Errorf("failed to get keygen committee: %v", err)
		}
		threshold, err := common.GetThreshold(len(keygenCommittee))
		if err != nil {
			return fmt.Errorf("failed to get threshold: %v", err)
		}
		t.logger.Infof("Threshold is %v", threshold+1)
		setupMsg, err := mpcKeygenWrapper.KeygenSetupMsgNew(threshold+1, nil, keygenCommitteeBytes)
		if err != nil {
			return fmt.Errorf("failed to create setup message: %v", err)
		}
		encodedSetupMsg = base64.StdEncoding.EncodeToString(setupMsg)
		t.logger.Infoln("setup message is:", encodedSetupMsg)
		if err := relayClient.UploadPayload(sessionID, encodedSetupMsg); err != nil {
			return fmt.Errorf("failed to upload setup message: %v", err)
		}

		if err := relayClient.StartSession(sessionID, keygenCommittee); err != nil {
			return fmt.Errorf("failed to start session: %w", err)
		}
	} else {
		// wait longer for keysign start
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel()
		// wait for the keygen to start
		_, err := relayClient.WaitForSessionStart(ctx, sessionID)
		if err != nil {
			return fmt.Errorf("failed to wait for session to start: %w", err)
		}
		// retrieve the setup Message
		encodedSetupMsg, err = relayClient.GetPayload(sessionID)
	}
	setupMessageBytes, err := base64.StdEncoding.DecodeString(encodedSetupMsg)
	if err != nil {
		return fmt.Errorf("failed to decode setup message: %w", err)
	}

	handle, err := mpcKeygenWrapper.KeygenSessionFromSetup(setupMessageBytes, []byte(localPartyID))
	if err != nil {
		return fmt.Errorf("failed to create session from setup message: %w", err)
	}
	defer func() {
		if err := mpcKeygenWrapper.KeygenSessionFree(handle); err != nil {
			t.logger.Error("failed to free keygen session", "error", err)
		}
	}()
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		if err := t.processKeygenOutbound(handle, sessionID, hexEncryptionKey, keygenCommittee, localPartyID, wg); err != nil {
			t.logger.Error("failed to process keygen outbound", "error", err)
		}
	}()
	err = t.processKeygenInbound(handle, sessionID, hexEncryptionKey, localPartyID, wg)
	wg.Wait()
	return err
}

func (t *DKLSTssService) processKeygenOutbound(handle Handle,
	sessionID string,
	hexEncryptionKey string,
	parties []string,
	localPartyID string,
	wg *sync.WaitGroup) error {
	defer wg.Done()
	messenger := relay.NewMessenger(t.relayServer, sessionID, hexEncryptionKey, true)
	mpcKeygenWrapper := t.GetMPCKeygenWrapper()
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

			t.logger.Infoln("Sending message to", string(receiver))
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
	localPartyID string,
	wg *sync.WaitGroup) error {
	defer wg.Done()
	cache := make(map[string]bool)
	mpcKeygenWrapper := t.GetMPCKeygenWrapper()
	for {
		select {
		case <-time.After(time.Minute):
			// set isKeygenFinished to true , so the other go routine can be stopped
			t.isKeygenFinished.Store(true)
			return TssKeyGenTimeout
		case <-time.After(time.Millisecond * 100):
			resp, err := http.Get(t.relayServer + "/message/" + sessionID + "/" + localPartyID)
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
				SessionID string   `json:"session_id,omitempty"`
				From      string   `json:"from,omitempty"`
				To        []string `json:"to,omitempty"`
				Body      string   `json:"body,omitempty"`
			}
			if err := decoder.Decode(&messages); err != nil {
				if err != io.EOF {
					t.logger.Error("fail to decode messages", "error", err)
				}
				continue
			}
			for _, message := range messages {
				if message.From == localPartyID {
					continue
				}

				hash := md5.Sum([]byte(message.Body))
				hashStr := hex.EncodeToString(hash[:])

				client := http.Client{}
				req, err := http.NewRequest(http.MethodDelete, t.relayServer+"/message/"+sessionID+"/"+localPartyID+"/"+hashStr, nil)
				if err != nil {
					t.logger.Error("fail to delete message", "error", err)
					continue
				}
				resp, err := client.Do(req)
				if err != nil {
					t.logger.Error("fail to delete message", "error", err)
					continue
				}
				if resp.StatusCode != http.StatusOK {
					t.logger.Error("fail to delete message", "status", resp.Status)
					continue
				}
				if _, ok := cache[hashStr]; ok {
					continue
				}
				cache[hashStr] = true
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
				t.logger.Infoln("Received message from", message.From)
				isFinished, err := mpcKeygenWrapper.KeygenSessionInputMessage(handle, rawBody)
				if err != nil {
					t.logger.Error("fail to apply input message", "error", err)
					continue
				}
				if isFinished {
					t.logger.Infoln("Keygen finished")
					result, err := mpcKeygenWrapper.KeygenSessionFinish(handle)
					if err != nil {
						t.logger.Error("fail to finish keygen", "error", err)
						return err
					}
					buf, err := mpcKeygenWrapper.KeyshareToBytes(result)
					if err != nil {
						t.logger.Error("fail to convert keyshare to bytes", "error", err)
						return err
					}
					encodedShare := base64.StdEncoding.EncodeToString(buf)
					publicKeyECDSABytes, err := mpcKeygenWrapper.KeysharePublicKey(result)
					if err != nil {
						t.logger.Error("fail to get public key", "error", err)
						return err
					}
					encodedPublicKey := hex.EncodeToString(publicKeyECDSABytes)
					t.logger.Infof("Public key: %s", encodedPublicKey)
					// This sleep give the local party a chance to send last message to others
					t.isKeygenFinished.Store(true)
					return t.localStateAccessor.SaveLocalState(encodedPublicKey, encodedShare)
				}
			}
		}
	}
}

func (t *DKLSTssService) Keysign(sessionID string,
	hexEncryptionKey string,
	publicKeyECDSA string,
	message string,
	derivePath string,
	localPartyID string,
	keysignCommittee []string,
	isInitiateDevice bool) error {
	if publicKeyECDSA == "" {
		return fmt.Errorf("public key is empty")
	}
	if message == "" {
		return fmt.Errorf("message is empty")
	}
	if derivePath == "" {
		return fmt.Errorf("derive path is empty")
	}
	if localPartyID == "" {
		return fmt.Errorf("local party id is empty")
	}
	if len(keysignCommittee) == 0 {
		return fmt.Errorf("keysign committee is empty")
	}
	relayClient := relay.NewRelayClient(t.relayServer)
	mpcWrapper := t.GetMPCKeygenWrapper()
	t.logger.WithFields(logrus.Fields{
		"session_id":         sessionID,
		"public_key_ecdsa":   publicKeyECDSA,
		"message":            message,
		"derive_path":        derivePath,
		"local_party_id":     localPartyID,
		"keysign_committee":  keysignCommittee,
		"is_initiate_device": isInitiateDevice,
	}).Info("Keysign")

	if err := relayClient.RegisterSession(sessionID, localPartyID); err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}
	// we need to get the shares
	keyshare, err := t.localStateAccessor.GetLocalState(publicKeyECDSA)
	if err != nil {
		return fmt.Errorf("failed to get keyshare: %w", err)
	}
	keyshareBytes, err := base64.StdEncoding.DecodeString(keyshare)
	if err != nil {
		return fmt.Errorf("failed to decode keyshare: %w", err)
	}
	keyshareHandle, err := mpcWrapper.KeyshareFromBytes(keyshareBytes)
	if err != nil {
		return fmt.Errorf("failed to create keyshare from bytes: %w", err)
	}
	defer func() {
		if err := mpcWrapper.KeyshareFree(keyshareHandle); err != nil {
			t.logger.Error("failed to free keyshare", "error", err)
		}
	}()
	msgHash := sha256.Sum256([]byte(message))
	var encodedSetupMsg string = ""
	if isInitiateDevice {
		if coordinator.WaitAllParties(keysignCommittee, t.relayServer, sessionID) != nil {
			return fmt.Errorf("failed to wait for all parties to join")
		}
		keyID, err := mpcWrapper.KeyshareKeyID(keyshareHandle)
		if err != nil {
			return fmt.Errorf("failed to get key id: %w", err)
		}
		keysignCommitteeBytes, err := t.convertKeygenCommitteeToBytes(keysignCommittee)
		if err != nil {
			return fmt.Errorf("failed to get keysign committee: %w", err)
		}
		intialMsg, err := mpcWrapper.SignSetupMsgNew(keyID, []byte("m/44/931/0/0/0"), msgHash[:], keysignCommitteeBytes)
		if err != nil {
			return fmt.Errorf("failed to create initial message: %w", err)
		}
		encodedInitialMsg := base64.StdEncoding.EncodeToString(intialMsg)
		t.logger.Infoln("initial message is:", encodedInitialMsg)
		if err := relayClient.UploadPayload(sessionID, encodedInitialMsg); err != nil {
			return fmt.Errorf("failed to upload initial message: %w", err)
		}
		encodedSetupMsg = encodedInitialMsg
		if err := relayClient.StartSession(sessionID, keysignCommittee); err != nil {
			return fmt.Errorf("failed to start session: %w", err)
		}
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel()
		_, err := relayClient.WaitForSessionStart(ctx, sessionID)
		if err != nil {
			return fmt.Errorf("failed to wait for session to start: %w", err)
		}
		// retrieve the setup Message
		encodedSetupMsg, err = relayClient.GetPayload(sessionID)
	}
	setupMessageBytes, err := base64.StdEncoding.DecodeString(encodedSetupMsg)
	if err != nil {
		return fmt.Errorf("failed to decode setup message: %w", err)
	}
	messageHashInSetupMsg, err := mpcWrapper.DecodeMessage(setupMessageBytes)
	if err != nil {
		return fmt.Errorf("failed to decode message: %w", err)
	}
	if !bytes.Equal(messageHashInSetupMsg, msgHash[:]) {
		return fmt.Errorf("message hash in setup message is not equal to the message, stop keysign")
	}
	sessionHandle, err := mpcWrapper.SignSessionFromSetup(setupMessageBytes, []byte(localPartyID), keyshareHandle)
	if err != nil {
		return fmt.Errorf("failed to create session from setup message: %w", err)
	}
	defer func() {
		if err := mpcWrapper.SignSessionFree(sessionHandle); err != nil {
			t.logger.Error("failed to free keysign session", "error", err)
		}
	}()
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		if err := t.processKeysignOutbound(sessionHandle, sessionID, hexEncryptionKey, keysignCommittee, localPartyID, message, wg); err != nil {
			t.logger.Error("failed to process keygen outbound", "error", err)
		}
	}()
	sig, err := t.processKeysignInbound(sessionHandle, sessionID, localPartyID, wg)
	wg.Wait()
	t.logger.Infoln("Keysign result is:", len(sig))
	if t.isEdDSA {
		pubKeyBytes, err := hex.DecodeString(publicKeyECDSA)
		if err != nil {
			return fmt.Errorf("failed to decode public key: %w", err)
		}

		if ed25519.Verify(pubKeyBytes, msgHash[:], sig) {
			t.logger.Infoln("Signature is valid")
		} else {
			t.logger.Error("Signature is invalid")
		}
	} else {
		if len(sig) != 65 {
			return fmt.Errorf("signature length is not 64")
		}
		r := sig[:32]
		s := sig[32:64]
		// recovery := sig[64]
		pubKeyBytes, err := hex.DecodeString(publicKeyECDSA)
		if err != nil {
			return fmt.Errorf("failed to decode public key: %w", err)
		}
		publicKey, err := secp256k1.ParsePubKey(pubKeyBytes)
		if err != nil {
			return fmt.Errorf("failed to parse public key: %w", err)
		}

		if ecdsa.Verify(publicKey.ToECDSA(), msgHash[:], new(big.Int).SetBytes(r), new(big.Int).SetBytes(s)) {
			t.logger.Infoln("Signature is valid")
		} else {
			t.logger.Error("Signature is invalid")
		}
	}
	return nil
}
func (t *DKLSTssService) processKeysignOutbound(handle Handle,
	sessionID string,
	hexEncryptionKey string,
	parties []string,
	localPartyID string,
	message string,
	wg *sync.WaitGroup) error {
	defer wg.Done()
	messenger := relay.NewMessenger(t.relayServer, sessionID, hexEncryptionKey, true)
	mpcWrapper := t.GetMPCKeygenWrapper()
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
	localPartyID string,
	wg *sync.WaitGroup) ([]byte, error) {
	defer wg.Done()
	cache := make(map[string]bool)
	mpcWrapper := t.GetMPCKeygenWrapper()
	for {
		select {
		case <-time.After(time.Minute):
			// set isKeygenFinished to true , so the other go routine can be stopped
			t.isKeysignFinished.Store(true)
			return nil, TssKeyGenTimeout
		case <-time.After(time.Millisecond * 100):
			resp, err := http.Get(t.relayServer + "/message/" + sessionID + "/" + localPartyID)
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
				SessionID string   `json:"session_id,omitempty"`
				From      string   `json:"from,omitempty"`
				To        []string `json:"to,omitempty"`
				Body      string   `json:"body,omitempty"`
			}
			if err := decoder.Decode(&messages); err != nil {
				if err != io.EOF {
					t.logger.Error("fail to decode messages", "error", err)
				}
				continue
			}
			for _, message := range messages {
				if message.From == localPartyID {
					continue
				}

				hash := md5.Sum([]byte(message.Body))
				hashStr := hex.EncodeToString(hash[:])

				client := http.Client{}
				req, err := http.NewRequest(http.MethodDelete, t.relayServer+"/message/"+sessionID+"/"+localPartyID+"/"+hashStr, nil)
				if err != nil {
					t.logger.Error("fail to delete message", "error", err)
					continue
				}
				resp, err := client.Do(req)
				if err != nil {
					t.logger.Error("fail to delete message", "error", err)
					continue
				}
				if resp.StatusCode != http.StatusOK {
					t.logger.Error("fail to delete message", "status", resp.Status)
					continue
				}
				if _, ok := cache[hashStr]; ok {
					continue
				}
				cache[hashStr] = true
				decodedBody, err := base64.StdEncoding.DecodeString(message.Body)
				if err != nil {
					t.logger.Error("fail to decode message", "error", err)
					continue
				}
				t.logger.Infoln("Received message from", message.From)
				isFinished, err := mpcWrapper.SignSessionInputMessage(handle, decodedBody)
				if err != nil {
					t.logger.Error("fail to apply input message", "error", err)
					continue
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
