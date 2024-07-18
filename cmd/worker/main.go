package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/vultisig/vultisigner/common"
	"github.com/vultisig/vultisigner/config"
	"github.com/vultisig/vultisigner/internal/keygen"
	"github.com/vultisig/vultisigner/internal/logging"
	"github.com/vultisig/vultisigner/internal/tasks"
	"github.com/vultisig/vultisigner/internal/types"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/hibiken/asynq"
	vaultType "github.com/vultisig/commondata/go/vultisig/vault/v1"
)

func main() {
	redisAddr := config.AppConfig.Redis.Host + ":" + config.AppConfig.Redis.Port

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
		},
	)

	logging.Logger.WithFields(logrus.Fields{
		"redis": redisAddr,
	}).Info("Starting server")

	// mux maps a type to a handler
	mux := asynq.NewServeMux()
	mux.HandleFunc(tasks.TypeKeyGeneration, HandleKeyGeneration)
	// mux.Handle(tasks.TypeKeyGeneration, tasks.I())
	// ...register other handlers...

	if err := srv.Run(mux); err != nil {
		log.Fatalf("could not run server: %v", err)
	}
}

type KeyGenerationTaskResult struct {
	EDDSAPublicKey string
	ECDSAPublicKey string
}

func HandleKeyGeneration(ctx context.Context, t *asynq.Task) error {
	var p tasks.KeyGenerationPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	logging.Logger.WithFields(logrus.Fields{
		"name":             p.Name,
		"session":          p.SessionID,
		"local_key":        p.LocalKey,
		"chain_code":       p.ChainCode,
		"HexEncryptionKey": p.HexEncryptionKey,
	}).Info("Joining keygen")

	keyECDSA, keyEDDSA, partiesJoined, ecdsaKeyShare, eddsaKeyShare, err := keygen.JoinKeyGeneration(&types.KeyGeneration{
		Key:              p.LocalKey,
		Session:          p.SessionID,
		ChainCode:        p.ChainCode,
		HexEncryptionKey: p.HexEncryptionKey,
	})
	if err != nil {
		return fmt.Errorf("keygen.JoinKeyGeneration failed: %v: %w", err, asynq.SkipRetry)
	}

	logging.Logger.WithFields(logrus.Fields{
		"keyECDSA": keyECDSA,
		"keyEDDSA": keyEDDSA,
	}).Info("Key generation completed")

	// backup vault
	vault := vaultType.Vault{
		Name:           p.Name,
		PublicKeyEcdsa: keyECDSA,
		PublicKeyEddsa: keyEDDSA,
		Signers:        partiesJoined,
		CreatedAt:      timestamppb.New(time.Now()),
		HexChainCode:   p.ChainCode,
		KeyShares: []*vaultType.Vault_KeyShare{
			{
				PublicKey: keyECDSA,
				Keyshare:  ecdsaKeyShare,
			},
			{
				PublicKey: keyECDSA,
				Keyshare:  eddsaKeyShare,
			},
		},
		LocalPartyId:  p.LocalKey,
		ResharePrefix: "",
	}

	encryptionPassword := "" // retrieve from vaultCacheItem

	isEncrypted := encryptionPassword != ""
	vaultStr := vault.String()
	if isEncrypted {
		vaultStr, err = common.Encrypt(encryptionPassword, vaultStr)
		if err != nil {
			return fmt.Errorf("common.Encrypt failed: %v: %w", err, asynq.SkipRetry)
		}
	}
	vaultBackup := vaultType.VaultContainer{
		Version:     "1",
		Vault:       vaultStr,
		IsEncrypted: isEncrypted,
	}
	filePathName := filepath.Join(config.AppConfig.Server.VaultsFilePath, p.Name+".bak")
	file, err := os.Create(filePathName)

	if err != nil {
		return fmt.Errorf("fail to create file, err: %w", err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			logging.Logger.Errorf("fail to close file, err: %v", err)
		}
	}()

	if _, err := file.Write([]byte(vaultBackup.String())); err != nil {
		return fmt.Errorf("fail to write file, err: %w", err)
	}

	result := KeyGenerationTaskResult{
		EDDSAPublicKey: keyEDDSA,
		ECDSAPublicKey: keyECDSA,
	}

	resultBytes, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("json.Marshal failed: %v: %w", err, asynq.SkipRetry)
	}

	if _, err := t.ResultWriter().Write([]byte(resultBytes)); err != nil {
		return fmt.Errorf("t.ResultWriter.Write failed: %v: %w", err, asynq.SkipRetry)
	}

	return nil
}
