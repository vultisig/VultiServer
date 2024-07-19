package tasks

const (
	TypeKeyGeneration = "key:generation"
	// TypeKeySign       = "key:sign"
)

type KeyGenerationPayload struct {
	LocalKey           string
	Name               string
	SessionID          string
	ChainCode          string
	HexEncryptionKey   string
	EncryptionPassword string
}
