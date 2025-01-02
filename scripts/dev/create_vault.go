package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/vultisig/vultisigner/config"
	"github.com/vultisig/vultisigner/internal/types"
)

var vaultName string
var stateDir string

func main() {
	flag.StringVar(&vaultName, "vault", "", "vault name")
	flag.StringVar(&stateDir, "state-dir", "", "state directory")
	flag.Parse()

	if vaultName == "" {
		panic("vault name is required")
	}

	if stateDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}

		stateDir = filepath.Join(homeDir, ".vultiserver", "vaults")
	}

	keyPath := filepath.Join(stateDir, vaultName, "public_key")
	if _, err := os.Stat(keyPath); err == nil {
		panic("vault already exists")
	}

	serverConfig, err := config.ReadConfig("config-server")
	if err != nil {
		panic(err)
	}

	pluginConfig, err := config.ReadConfig("config-plugin")
	if err != nil {
		panic(err)
	}

	createVaultRequest := &types.VaultCreateRequest{
		Name:               vaultName,
		SessionID:          uuid.New().String(),
		HexEncryptionKey:   "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		HexChainCode:       "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		LocalPartyId:       "1",
		EncryptionPassword: "your-secure-password",
		Email:              "example@example.com",
		StartSession:       false,
	}

	serverHost := fmt.Sprintf("http://%s:%d", serverConfig.Server.Host, serverConfig.Server.Port)
	pluginHost := fmt.Sprintf("http://%s:%d", pluginConfig.Server.Host, pluginConfig.Server.Port)

	fmt.Printf("Creating vault on verifier server: %s\n", serverHost)
	reqBytes, err := json.Marshal(createVaultRequest)
	if err != nil {
		panic(err)
	}

	resp, err := http.Post(fmt.Sprintf("%s/vault/create", serverHost), "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		panic(err)
	}
	fmt.Printf("Request sent: %d\n", resp.StatusCode)

	fmt.Printf("Creating vault on plugin server: %s\n", pluginHost)
	createVaultRequest.LocalPartyId = "2"
	createVaultRequest.StartSession = true
	createVaultRequest.Parties = []string{"1", "2"}

	reqBytes, err = json.Marshal(createVaultRequest)
	if err != nil {
		panic(err)
	}

	resp, err = http.Post(fmt.Sprintf("%s/vault/create", pluginHost), "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		panic(err)
	}
	fmt.Printf("Request sent: %d\n", resp.StatusCode)

	fmt.Println("Please watch the logs on the worker nodes and retrieve the ECDSA public key")

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter the ECDSA public key: ")
	publicKey, _ := reader.ReadString('\n')
	publicKey = publicKey[:len(publicKey)-1]

	fmt.Printf("Saving vault %s with key %s\n", vaultName, publicKey)
	vaultPath := filepath.Join(stateDir, vaultName)
	if err := os.MkdirAll(vaultPath, 0755); err != nil {
		panic(err)
	}

	vaultFile, err := os.Create(filepath.Join(vaultPath, "public_key"))
	if err != nil {
		panic(err)
	}

	if _, err := vaultFile.WriteString(publicKey); err != nil {
		panic(err)
	}

	fmt.Println("Vault created successfully")
}
