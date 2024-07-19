package common

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}

func pkcs7Unpad(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, fmt.Errorf("input data length must be greater than zero")
	}
	padding := int(data[length-1])
	return data[:length-padding], nil
}
func Encrypt(password, src string) (string, error) {
	salt := make([]byte, 8)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", fmt.Errorf("fail to generate salt: %w", err)
	}
	key := pbkdf2.Key([]byte(password), salt, 4096, 32, sha256.New)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("fail to create cipher: %w", err)
	}
	srcBytes := pkcs7Pad([]byte(src), block.BlockSize())
	ciphertext := make([]byte, aes.BlockSize+len(srcBytes))
	iv := ciphertext[:aes.BlockSize]
	copy(iv, salt)
	if _, err := io.ReadFull(rand.Reader, iv[len(salt):]); err != nil {
		return "", fmt.Errorf("fail to generate random iv: %w", err)
	}
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], []byte(srcBytes))
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}
func Decrypt(password string, src string) (string, error) {
	ciphertextDec, err := base64.StdEncoding.DecodeString(src)
	if err != nil {
		return "", err
	}

	if len(ciphertextDec) < aes.BlockSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	iv := ciphertextDec[:aes.BlockSize]
	salt := iv[:8] // Assuming the salt was stored in the first 8 bytes of the IV
	key := pbkdf2.Key([]byte(password), salt, 4096, 32, sha256.New)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("fail to create cipher: %w", err)
	}
	ciphertextDec = ciphertextDec[aes.BlockSize:]
	if len(ciphertextDec)%aes.BlockSize != 0 {
		return "", fmt.Errorf("ciphertext is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(ciphertextDec, ciphertextDec)

	plaintext, err := pkcs7Unpad(ciphertextDec)
	if err != nil {
		return "", fmt.Errorf("fail to unpad plaintext: %w", err)
	}

	return string(plaintext), nil
}
func EncryptVault(password string, vault []byte) ([]byte, error) {
	// Hash the password to create a key
	hash := sha256.Sum256([]byte(password))
	key := hash[:]

	// Create a new AES cipher using the key
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Use GCM (Galois/Counter Mode)
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Create a nonce. Nonce size is specified by GCM
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Seal encrypts and authenticates plaintext
	ciphertext := gcm.Seal(nonce, nonce, vault, nil)
	return ciphertext, nil
}
func DecryptVault(password string, vault []byte) ([]byte, error) {
	// Hash the password to create a key
	hash := sha256.Sum256([]byte(password))
	key := hash[:]

	// Create a new AES cipher using the key
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Use GCM (Galois/Counter Mode)
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Check that the encrypted data is at least as long as the nonce
	nonceSize := gcm.NonceSize()
	if len(vault) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	// Extract the nonce and ciphertext
	nonce, ciphertext := vault[:nonceSize], vault[nonceSize:]

	// Decrypt the data
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
