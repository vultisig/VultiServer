package relay

import (
	"bytes"
	"crypto/md5"
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
	Server    string
	SessionID string
}

func (m *MessengerImp) Send(from, to, body string) error {
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
		// "body": body,
	}).Info("message sent")

	return nil
}

func DownloadMessage(server, session, key string, tssServerImp tss.Service, endCh chan struct{}, wg *sync.WaitGroup) {
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
				}).Error("fail to get data from server")
				continue
			}
			if resp.StatusCode != http.StatusOK {
				logging.Logger.WithFields(logrus.Fields{
					"session": session,
					"key":     key,
				}).Error("fail to get data from server")
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
					}).Error("fail to decode messages")
				}
				continue
			}
			for _, message := range messages {
				if message.From == key {
					continue
				}

				// @TODO: decode message.Body when encrypted
				client := http.Client{}
				req, err := http.NewRequest(http.MethodDelete, server+"/message/"+session+"/"+key+"/"+message.Hash, nil)
				if err != nil {
					// log.Println("fail to delete message:", err)
					logging.Logger.WithFields(logrus.Fields{
						"session": session,
						"key":     key,
					}).Error("fail to delete message")
					continue
				}
				resp, err := client.Do(req)
				if err != nil {
					logging.Logger.WithFields(logrus.Fields{
						"session": session,
						"key":     key,
					}).Error("fail to delete message")
					continue
				}
				if resp.StatusCode != http.StatusOK {
					logging.Logger.WithFields(logrus.Fields{
						"session": session,
						"key":     key,
					}).Error("fail to delete message")
					continue
				}

				if err := tssServerImp.ApplyData(message.Body); err != nil {
					logging.Logger.WithFields(logrus.Fields{
						"session": session,
						"key":     key,
					}).Error("fail to apply data")
				}

			}
		}
	}
}
