package relay

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/vultisig/mobile-tss-lib/tss"
	"github.com/vultisig/vultisigner/internal/logging"
)

type MessengerImp struct {
	Server           string
	SessionID        string
	hexEncryptionKey string
}

var messageCache sync.Map

func (m *MessengerImp) SetHexEncryptionKey(key string) {
	m.hexEncryptionKey = key
}

func (m *MessengerImp) Send(from, to, body string) error {
	if m.hexEncryptionKey != "" {
		encryptedBody, err := encrypt(body, m.hexEncryptionKey)
		fmt.Println("!!!!!!!!!!!!!!!!!encryptedBody", encryptedBody)
		if err != nil {
			return fmt.Errorf("failed to encrypt body: %w", err)
		}
		body = encryptedBody
	}

	hash := md5.New()
	hash.Write([]byte(body))
	hashStr := hex.EncodeToString(hash.Sum(nil))

	if hashStr == "" {
		return fmt.Errorf("hash is empty")
	}

	buf, err := json.MarshalIndent(struct {
		SessionID string   `json:"session_id,omitempty"`
		From      string   `json:"from,omitempty"`
		To        []string `json:"to,omitempty"`
		Body      string   `json:"body,omitempty"`
		Hash      string   `json:"hash,omitempty"`
	}{
		SessionID: m.SessionID,
		From:      from,
		To:        []string{to},
		Body:      body,
		Hash:      hashStr,
	}, "", "  ")
	if err != nil {
		return fmt.Errorf("fail to marshal message: %w", err)
	}

	url := fmt.Sprintf("%s/message/%s", m.Server, m.SessionID)
	req, err := http.NewRequest("POST", url, bytes.NewReader(buf))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if body == "" {
		return fmt.Errorf("body is empty")
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.Status != "202 Accepted" {
		return fmt.Errorf("fail to send message, response code is not 202 Accepted: %s", resp.Status)
	}

	logging.Logger.WithFields(logrus.Fields{
		"from": from,
		"to":   to,
		"hash": hashStr,
	}).Info("Message sent")

	return nil
}

func DownloadMessage(server, session, key, hexEncryptionKey string, tssServerImp tss.Service, endCh chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-endCh: // we are done
			return
		case <-time.After(time.Second):
			resp, err := http.Get(server + "/message/" + session + "/" + key)
			if err != nil {
				logging.Logger.WithFields(logrus.Fields{
					"session": session,
					"key":     key,
					"error":   err,
				}).Error("Failed to get data from server")
				continue
			}
			if resp.StatusCode != http.StatusOK {
				logging.Logger.WithFields(logrus.Fields{
					"session": session,
					"key":     key,
				}).Error("Failed to get data from server, status code is not 200 OK")
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
				if err != io.EOF {
					logging.Logger.WithFields(logrus.Fields{
						"session": session,
						"key":     key,
						"error":   err,
					}).Error("Failed to decode data")
				}
				continue
			}
			for _, message := range messages {
				if message.From == key {
					continue
				}

				cacheKey := fmt.Sprintf("%s-%s-%s", session, key, message.Hash)
				if _, found := messageCache.Load(cacheKey); found {
					logging.Logger.WithFields(logrus.Fields{
						"session": session,
						"key":     key,
						"hash":    message.Hash,
					}).Info("Message already applied, skipping")
					continue
				}

				decryptedBody := message.Body
				if hexEncryptionKey != "" {
					// The entire body we get is base64 encoded
					decodedBody, err := base64.StdEncoding.DecodeString(message.Body)
					if err != nil {
						logging.Logger.WithFields(logrus.Fields{
							"session": session,
							"key":     key,
							"hash":    message.Hash,
							// "body":    message.Body,
							"error": err,
						}).Error("Failed to decode data")
						continue
					}

					// The decoded body contains wire_bytes which is also base64 encoded
					var decodedBodyMap map[string]interface{}
					err = json.Unmarshal(decodedBody, &decodedBodyMap)
					if err != nil {
						logging.Logger.WithFields(logrus.Fields{
							"session": session,
							"key":     key,
							"hash":    message.Hash,
							// "body":    message.Body,
							"error": err,
						}).Error("Failed to unmarshal data")
						continue
					}

					// Checks if the wire_bytes is present in the decoded body
					wireBytes, ok := decodedBodyMap["wire_bytes"].(string)
					if !ok {
						logging.Logger.WithFields(logrus.Fields{
							"session": session,
							"key":     key,
							"hash":    message.Hash,
							// "body":    message.Body,
							"error": err,
						}).Error("Failed to get wire_bytes")
						continue
					}

					// Decodes the wire_bytes which is base64 encoded
					decodedBody, err = base64.StdEncoding.DecodeString(wireBytes)
					if err != nil {
						logging.Logger.WithFields(logrus.Fields{
							"session": session,
							"key":     key,
							"hash":    message.Hash,
							// "body":    message.Body,
							"error": err,
						}).Error("Failed to decode wire_bytes")
						continue
					}
					_ = string(decodedBody)

					// Check if it is a valid encrypted message
					// If it is not a valid encrypted message, we will just continue and treat it as a normal message, it should be aes encrypted
					// does aes have a prefix?

					// Decrypts the wire_bytes using the encryption key
					// decryptedBody, err = decrypt(decodedBodyStr, hexEncryptionKey)
					// if err != nil {
					// 	logging.Logger.WithFields(logrus.Fields{
					// 		"session": session,
					// 		"key":     key,
					// 		"hash":    message.Hash,
					// 		// "wireBytes": wireBytes,
					// 		// "body":    message.Body,
					// 		"error": err,
					// 	}).Error("Failed to decrypt data, possibly not encrypted (or wrong key)")
					// 	continue
					// }
				}

				if err := tssServerImp.ApplyData(decryptedBody); err != nil {
					logging.Logger.WithFields(logrus.Fields{
						"session": session,
						"key":     key,
						"error":   err,
					}).Error("Failed to apply data")
					continue
				}

				messageCache.Store(cacheKey, true)
				client := http.Client{}
				req, err := http.NewRequest(http.MethodDelete, server+"/message/"+session+"/"+key+"/"+message.Hash, nil)
				if err != nil {
					logging.Logger.WithFields(logrus.Fields{
						"session": session,
						"key":     key,
						"error":   err,
					}).Error("Failed to delete message")
					continue
				}

				resp, err := client.Do(req)
				if err != nil {
					logging.Logger.WithFields(logrus.Fields{
						"session": session,
						"key":     key,
						"error":   err,
					}).Error("Failed to delete message")
					continue
				}

				if resp.StatusCode != http.StatusOK {
					logging.Logger.WithFields(logrus.Fields{
						"session": session,
						"key":     key,
					}).Error("Failed to delete message, status code is not 200 OK")
					continue
				}
			}
		}
	}
}

func encrypt(plainText, hexKey string) (string, error) {
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return "", fmt.Errorf("invalid hex key: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}
	nonce := make([]byte, aesGCM.NonceSize())
	cipherText := aesGCM.Seal(nonce, nonce, []byte(plainText), nil)
	return hex.EncodeToString(cipherText), nil
}

func decrypt(cipherText, hexKey string) (string, error) {
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return "", fmt.Errorf("invalid hex key: %w", err)
	}
	cipherTextBytes, err := hex.DecodeString(cipherText)
	if err != nil {
		return "", fmt.Errorf("invalid cipher text: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}
	nonceSize := aesGCM.NonceSize()
	if len(cipherTextBytes) < nonceSize {
		return "", fmt.Errorf("cipher text too short")
	}
	nonce, cipherTextBytes := cipherTextBytes[:nonceSize], cipherTextBytes[nonceSize:]
	plainText, err := aesGCM.Open(nil, nonce, cipherTextBytes, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}
	return string(plainText), nil
}
