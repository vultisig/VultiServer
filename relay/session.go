package relay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Server struct {
	vultisigRelay string
}

func NewServer(vultisigRelay string) *Server {
	return &Server{
		vultisigRelay: vultisigRelay,
	}
}

func (s *Server) StartSession(sessionID string) error {
	sessionURL := s.vultisigRelay + "/start/" + sessionID
	body := []byte("[]")
	bodyReader := bytes.NewReader(body)
	resp, err := http.Post(sessionURL, "application/json", bodyReader)
	if err != nil {
		return fmt.Errorf("fail to start session: %w", err)
	}
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("fail to start session: %s", resp.Status)
	}
	// session registered , and device can use this session id to join the session
	return nil
}

func (s *Server) RegisterSession(sessionID string, key string) error {
	sessionURL := s.vultisigRelay + "/" + sessionID
	fmt.Println("Registering session with url: ", sessionURL)
	body := []byte("[\"" + key + "\"]")
	fmt.Println("Registering session with body: ", string(body))
	bodyReader := bytes.NewReader(body)

	resp, err := http.Post(sessionURL, "application/json", bodyReader)

	if err != nil {
		return fmt.Errorf("fail to register session: %w", err)
	}
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("fail to register session: %s", resp.Status)
	}

	return nil
}

func (s *Server) WaitForSessionStart(sessionID string) ([]string, error) {
	sessionURL := s.vultisigRelay + "/start/" + sessionID
	for {
		fmt.Println("start waiting for someone to start session...")
		resp, err := http.Get(sessionURL)
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
		if err := json.Unmarshal(buff, &parties); err != nil {
			return nil, fmt.Errorf("fail to unmarshal session body: %w", err)
		}
		if len(parties) > 1 {
			fmt.Println("all parties joined: ", parties)
			return parties, nil
		}

		fmt.Println("waiting for someone to start session...")

		// backoff
		time.Sleep(2 * time.Second)
	}
}

func (s *Server) EndSession(sessionID string) error {
	sessionURL := s.vultisigRelay + "/" + sessionID
	client := http.Client{}
	req, err := http.NewRequest(http.MethodDelete, sessionURL, nil)
	if err != nil {
		return fmt.Errorf("fail to end session: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("fail to end session: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fail to end session: %s", resp.Status)
	}
	return nil
}
