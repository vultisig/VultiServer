package relay

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/vultisig/vultisigner/common"
)

type Client struct {
	relayServer string
	client      http.Client
	logger      *logrus.Logger
}

func NewRelayClient(relayServer string) *Client {
	return &Client{
		relayServer: relayServer,
		client: http.Client{
			Timeout: 5 * time.Second,
		},
		logger: logrus.WithField("service", "relay-client").Logger,
	}
}
func (c *Client) bodyCloser(body io.ReadCloser) {
	if body != nil {
		if err := body.Close(); err != nil {
			c.logger.Error("Failed to close body,err:", err)
		}
	}
}

func (c *Client) StartSession(sessionID string, parties []string) error {
	sessionURL := c.relayServer + "/start/" + sessionID
	body, err := json.Marshal(parties)
	if err != nil {
		return fmt.Errorf("fail to start session: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, sessionURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("fail to start session: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("fail to start session: %w", err)
	}
	defer c.bodyCloser(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fail to start session: %s", resp.Status)
	}
	return nil
}

func (c *Client) RegisterSession(sessionID string, key string) error {
	sessionURL := c.relayServer + "/" + sessionID
	body := []byte("[\"" + key + "\"]")
	c.logger.WithFields(logrus.Fields{
		"session": sessionID,
		"key":     key,
		"body":    string(body),
	}).Info("Registering session")
	resp, err := c.client.Post(sessionURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("fail to register session: %w", err)
	}
	defer c.bodyCloser(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("fail to register session: %s", resp.Status)
	}

	return nil
}

func (c *Client) WaitForSessionStart(ctx context.Context, sessionID string) ([]string, error) {
	sessionURL := c.relayServer + "/start/" + sessionID
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			c.logger.WithFields(logrus.Fields{
				"session": sessionID,
			}).Info("Waiting for session start")
			resp, err := c.client.Get(sessionURL)
			if err != nil {
				return nil, fmt.Errorf("fail to get session: %w", err)
			}
			if resp.StatusCode != http.StatusOK {
				return nil, fmt.Errorf("fail to get session: %s", resp.Status)
			}
			var parties []string
			buff, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("fail to read session body: %w", err)
			}
			c.bodyCloser(resp.Body)
			if err := json.Unmarshal(buff, &parties); err != nil {
				return nil, fmt.Errorf("fail to unmarshal session body: %w", err)
			}
			//remove duplicates from parties
			distinctParties := make(map[string]struct{})
			for _, party := range parties {
				distinctParties[party] = struct{}{}
			}
			parties = make([]string, 0, len(distinctParties))
			for party := range distinctParties {
				parties = append(parties, party)
			}
			// We need to hold expected parties to start session
			if len(parties) > 1 {
				c.logger.WithFields(logrus.Fields{
					"session": sessionID,
					"parties": parties,
				}).Info("All parties joined")
				return parties, nil
			}

			c.logger.WithFields(logrus.Fields{
				"session": sessionID,
			}).Info("Waiting for someone to start session")

			// backoff
			time.Sleep(1 * time.Second)
		}
	}
}

func (c *Client) GetSession(sessionID string) ([]string, error) {
	sessionURL := c.relayServer + "/" + sessionID

	resp, err := c.client.Get(sessionURL)
	if err != nil {
		return nil, fmt.Errorf("fail to get session: %w", err)
	}
	defer c.bodyCloser(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fail to get session: %s", resp.Status)
	}
	var parties []string
	if err := json.NewDecoder(resp.Body).Decode(&parties); err != nil {
		return nil, fmt.Errorf("fail to unmarshal session body: %w", err)
	}

	return parties, nil
}

func (c *Client) CompleteSession(sessionID, localPartyID string) error {
	sessionURL := c.relayServer + "/complete/" + sessionID
	parties := []string{localPartyID}
	body, err := json.Marshal(parties)
	if err != nil {
		return fmt.Errorf("fail to complete session: %w", err)
	}
	bodyReader := bytes.NewReader(body)
	req, err := http.NewRequest(http.MethodPost, sessionURL, bodyReader)
	if err != nil {
		return fmt.Errorf("fail to complete session: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("fail to complete session: %w", err)
	}
	defer c.bodyCloser(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fail to complete session: %s", resp.Status)
	}
	return nil
}

func (c *Client) CheckCompletedParties(sessionID string, partiesJoined []string) (bool, error) {
	sessionURL := c.relayServer + "/complete/" + sessionID
	start := time.Now()
	timeout := time.Minute

	for {
		req, err := http.NewRequest(http.MethodGet, sessionURL, nil)
		if err != nil {
			return false, fmt.Errorf("fail to check completed parties: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		resp, err := c.client.Do(req)
		if err != nil {
			return false, fmt.Errorf("fail to check completed parties: %w", err)
		}
		if resp.StatusCode != http.StatusOK {
			return false, fmt.Errorf("fail to check completed parties: %s", resp.Status)
		}

		result, err := io.ReadAll(resp.Body)
		if err != nil {
			return false, fmt.Errorf("fail to fetch request: %w", err)
		}
		c.bodyCloser(resp.Body)

		if len(result) > 0 {
			var peers []string
			err := json.Unmarshal(result, &peers)
			if err != nil {
				c.logger.WithFields(logrus.Fields{
					"error": err,
				}).Error("Failed to decode response to JSON")
				continue
			}

			if common.IsSubset(partiesJoined, peers) {
				c.logger.Info("All parties have completed keygen successfully")
				return true, nil
			}
		}

		time.Sleep(time.Second)
		if time.Since(start) >= timeout {
			break
		}
	}

	return false, nil
}

func (c *Client) EndSession(sessionID string) error {
	sessionURL := c.relayServer + "/" + sessionID
	req, err := http.NewRequest(http.MethodDelete, sessionURL, nil)
	if err != nil {
		return fmt.Errorf("fail to end session: %w", err)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("fail to end session: %w", err)
	}
	defer c.bodyCloser(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fail to end session: %s", resp.Status)
	}
	return nil
}
