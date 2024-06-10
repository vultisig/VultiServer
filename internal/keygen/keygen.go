package keygen

import (
	"github.com/vultisig/vultisigner/config"
	"github.com/vultisig/vultisigner/internal/types"

	"github.com/vultisig/mobile-tss-lib/coordinator"
)

func JoinKeyGeneration(kg *types.KeyGeneration) error {
	relayServer := config.AppConfig.Relay.Server

	_, err := coordinator.ExecuteKeyGeneration(coordinator.KeygenInput{
		Server:    relayServer,
		Key:       kg.Key,
		Parties:   kg.Parties,
		Session:   kg.Session,
		ChainCode: kg.ChainCode,
	})
	if err != nil {
		return err
	}

	return nil
}
