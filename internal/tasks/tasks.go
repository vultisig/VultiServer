package tasks

const (
	TypeKeyGeneration = "key:generation"
	// TypeKeySign       = "key:sign"
)

type KeyGenerationPayload struct {
	LocalKey         string
	SessionID        string
	ChainCode        string
	HexEncryptionKey string
}
