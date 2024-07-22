package common

import (
	"testing"
)

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
