package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"

	"github.com/spf13/viper"
)

// meant to encrypt the Xi fields in EcdsaLocalData and EddsaLocalData
// but can be used for anything really
func Encrypt(plaintext string) (string, error) {
	password := viper.GetString("encryption.password")
	if password == "" {
		return "", errors.New("encryption password not set in server configuration")
	}

	block, err := aes.NewCipher([]byte(password))
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypts the encrypted ciphertext
func Decrypt(ciphertext string) (string, error) {
	password := viper.GetString("encryption.password")
	if password == "" {
		return "", errors.New("encryption password not set in server configuration")
	}

	block, err := aes.NewCipher([]byte(password))
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	ciphertextBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertextBytes) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertextBytes := ciphertextBytes[:nonceSize], ciphertextBytes[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
