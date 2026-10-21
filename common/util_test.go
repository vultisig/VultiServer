package common

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestDataCompression(t *testing.T) {
	data := "message"
	compressedData, err := CompressData([]byte(data))
	if err != nil {
		t.Fatal(err)
	}

	decompressedData, err := DecompressData(compressedData)
	if err != nil {
		t.Fatal(err)
	}

	if string(decompressedData) != data {
		t.Fatalf("decompressed: %s, expected: %s", decompressedData, data)
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

func TestVaultBackupCompatible(t *testing.T) {
	filePathName := filepath.Join("test_vault_backup_files", "test_ios_vault_backup.bak")
	_, err := os.Stat(filePathName)
	if err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(filePathName)
	if err != nil {
		t.Fatal(err)
	}

	iosVault, err := DecryptVaultFromBackup("ios_test_pwd", content)
	if err != nil {
		t.Fatal(err)
	}

	filePathName = filepath.Join("test_vault_backup_files", "test_android_vault_backup.bak")
	_, err = os.Stat(filePathName)
	if err != nil {
		t.Fatal(err)
	}

	content, err = os.ReadFile(filePathName)
	if err != nil {
		t.Fatal(err)
	}

	androidVault, err := DecryptVaultFromBackup("android_test_pwd", content)
	if err != nil {
		t.Fatal(err)
	}

	if iosVault.PublicKeyEcdsa != androidVault.PublicKeyEcdsa {
		t.Fatalf("ios backup is not compatible with android backup")
	}
}

// Blob from vultisig-sdk vaultBackup.test.ts "golden vector": password
// "golden-cross-platform", plaintext "xy", salt 0x03*16, nonce 0x04*12.
const pbkdf2GoldenVectorHex = "564c5402030303030303030303030303030303030404040404040404040404041b0a21328e42d6527c5ba6f5817300bc2c6f"

func TestDecryptVaultPBKDF2GoldenVector(t *testing.T) {
	blob, err := hex.DecodeString(pbkdf2GoldenVectorHex)
	if err != nil {
		t.Fatal(err)
	}

	plaintext, err := DecryptVault("golden-cross-platform", blob)
	if err != nil {
		t.Fatal(err)
	}
	if string(plaintext) != "xy" {
		t.Fatalf("decrypted: %q, expected: %q", plaintext, "xy")
	}

	if _, err := DecryptVault("wrong-password", blob); err == nil {
		t.Fatal("wrong password decrypted")
	}
	for _, truncated := range [][]byte{blob[:len(vaultBackupMagic)], blob[:vaultBackupHeaderLen+vaultBackupTagLen-1]} {
		if _, err := DecryptVault("golden-cross-platform", truncated); err == nil {
			t.Fatalf("truncated blob of %d bytes decrypted", len(truncated))
		}
	}
}

func TestDecryptVaultLegacyNonceWithMagic(t *testing.T) {
	password := "legacy-password"
	key := sha256.Sum256([]byte(password))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		t.Fatal(err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatal(err)
	}
	nonce := append([]byte(vaultBackupMagic), 1, 2, 3, 4, 5, 6, 7, 8)
	blob := gcm.Seal(nonce, nonce, []byte("legacy-share"), nil)

	plaintext, err := DecryptVault(password, blob)
	if err != nil {
		t.Fatal(err)
	}
	if string(plaintext) != "legacy-share" {
		t.Fatalf("decrypted: %q, expected: %q", plaintext, "legacy-share")
	}
	if _, err := DecryptVault("wrong-password", blob); err == nil {
		t.Fatal("wrong password decrypted")
	}
}

func TestEncryptVaultPBKDF2GoldenVector(t *testing.T) {
	salt := bytes.Repeat([]byte{0x03}, vaultBackupSaltLen)
	nonce := bytes.Repeat([]byte{0x04}, vaultBackupNonceLen)
	blob, err := sealVaultPBKDF2("golden-cross-platform", []byte("xy"), salt, nonce)
	if err != nil {
		t.Fatal(err)
	}
	if got := hex.EncodeToString(blob); got != pbkdf2GoldenVectorHex {
		t.Fatalf("blob: %s, expected: %s", got, pbkdf2GoldenVectorHex)
	}

	first, err := EncryptVault("password", []byte("vault_bytes"))
	if err != nil {
		t.Fatal(err)
	}
	second, err := EncryptVault("password", []byte("vault_bytes"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(first, []byte(vaultBackupMagic)) {
		t.Fatalf("EncryptVault output does not start with the format magic: %x", first[:len(vaultBackupMagic)])
	}
	saltEnd := len(vaultBackupMagic) + vaultBackupSaltLen
	if bytes.Equal(first[len(vaultBackupMagic):saltEnd], second[len(vaultBackupMagic):saltEnd]) {
		t.Fatal("EncryptVault reused the salt")
	}
	if bytes.Equal(first[saltEnd:vaultBackupHeaderLen], second[saltEnd:vaultBackupHeaderLen]) {
		t.Fatal("EncryptVault reused the nonce")
	}
}
