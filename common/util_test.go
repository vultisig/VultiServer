package common

import (
	"testing"
)

func TestEncryption(t *testing.T) {
	password := "password"
	src := "helloworld"
	encrypted, err := Encrypt(password, src)
	if err != nil {
		t.Fatal(err)
	}
	decrypted, err := Decrypt(password, encrypted)
	if err != nil {
		t.Fatal(err)
	}
	if decrypted != src {
		t.Fatalf("decrypted: %s, expected: %s", decrypted, src)
	}
}

func TestVaultEncryption(t *testing.T) {
	password := "password"
	src := "vault_bytes"
	encrypted, err := EncryptVault(password, []byte(src))
	if err != nil {
		t.Fatal(err)
	}
	decrypted, err := DecryptVault(password, encrypted)
	if err != nil {
		t.Fatal(err)
	}
	if string(decrypted) != src {
		t.Fatalf("decrypted: %s, expected: %s", decrypted, src)
	}
}
