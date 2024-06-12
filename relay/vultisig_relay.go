package relay

import (
	"bytes"
	"fmt"
	"net/http"
)

type Server struct {
	vultisigRelay string
}

func NewServer(vultisigRelay string) *Server {
	return &Server{
		vultisigRelay: vultisigRelay,
	}
}

// StartSession starts a new session on vultisig relay server
func (s *Server) StartSession(sessionID string) error {
	sessionURL := s.vultisigRelay + "/" + sessionID
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
