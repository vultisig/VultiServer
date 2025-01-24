package relay

type Message struct {
	SessionID  string   `json:"session_id,omitempty"`
	From       string   `json:"from,omitempty"`
	To         []string `json:"to,omitempty"`
	Body       string   `json:"body,omitempty"`
	Hash       string   `json:"hash,omitempty"`
	SequenceNo int64    `json:"sequence_no,omitempty"`
}
