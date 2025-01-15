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
	"time"

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
	rawKey, err := os.ReadFile(keyPath)
	if err != nil {
		panic(err)
	}

	key := string(rawKey)

	fmt.Printf("Public key for vault %s:\n%s\n", vaultName, key)

	serverConfig, err := config.ReadConfig("config-server")
	if err != nil {
		panic(err)
	}

	pluginConfig, err := config.ReadConfig("config-plugin")
	if err != nil {
		panic(err)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter token contract: ")
	tokenContract, _ := reader.ReadString('\n')
	tokenContract = tokenContract[:len(tokenContract)-1]

	fmt.Println("Enter recipients and amounts one by one - enter 'done' when finished")
	var done bool
	var recipientAddresses []string
	var recipientAmounts []string
	for !done {
		fmt.Print("Enter a recipient address: ")
		recipientAddress, _ := reader.ReadString('\n')
		recipientAddress = recipientAddress[:len(recipientAddress)-1]

		if recipientAddress == "done" {
			done = true
			break
		}

		fmt.Print("Enter the amount for this recipient: ")
		recipientAmount, _ := reader.ReadString('\n')
		recipientAmount = recipientAmount[:len(recipientAmount)-1]

		recipientAddresses = append(recipientAddresses, recipientAddress)
		recipientAmounts = append(recipientAmounts, recipientAmount)
	}

	fmt.Printf("Token contract: %s\n", tokenContract)
	fmt.Printf("Recipients: %d\n", len(recipientAddresses))
	for i, recipient := range recipientAddresses {
		fmt.Printf("Recipient %d: %s, Amount: %s\n", i+1, recipient, recipientAmounts[i])
	}

	fmt.Print("Enter schedule frequency: ")
	frequency, _ := reader.ReadString('\n')
	frequency = frequency[:len(frequency)-1]

	policyId := uuid.New().String()
	policy := types.PluginPolicy{
		ID:            policyId,
		PublicKey:     key,
		PluginID:      "payroll",
		PluginVersion: "1.0.0",
		PolicyVersion: "1.0.0",
		PluginType:    "payroll",
		Signature:     "0x0000000000000000000000000000000000000000000000000000000000000000",
	}

	payrollPolicy := types.PayrollPolicy{
		ChainID:    "1",
		TokenID:    tokenContract,
		Recipients: []types.PayrollRecipient{},
		Schedule: types.Schedule{
			Frequency: frequency,
			StartTime: time.Now().UTC().Add(20 * time.Second).Format(time.RFC3339),
		},
	}

	for i, recipient := range recipientAddresses {
		payrollPolicy.Recipients = append(payrollPolicy.Recipients, types.PayrollRecipient{
			Address: recipient,
			Amount:  recipientAmounts[i],
		})
	}

	policyBytes, err := json.Marshal(payrollPolicy)
	if err != nil {
		panic(err)
	}

	fmt.Println("Payroll policy", string(policyBytes))
	policy.Policy = policyBytes

	serverHost := fmt.Sprintf("http://%s:%d", serverConfig.Server.Host, serverConfig.Server.Port)
	pluginHost := fmt.Sprintf("http://%s:%d", pluginConfig.Server.Host, pluginConfig.Server.Port)

	fmt.Printf("Creating policy on verifier server: %s\n", serverHost)
	reqBytes, err := json.Marshal(policy)
	if err != nil {
		panic(err)
	}

	resp, err := http.Post(fmt.Sprintf("%s/plugin/policy", serverHost), "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		panic(err)
	}
	fmt.Printf("Request sent: %d\n", resp.StatusCode)

	fmt.Printf("Creating policy on plugin server: %s\n", pluginHost)

	resp, err = http.Post(fmt.Sprintf("%s/plugin/policy", pluginHost), "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		panic(err)
	}
	fmt.Printf("Request sent: %d\n", resp.StatusCode)
}
