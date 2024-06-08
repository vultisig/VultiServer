package keygen

import (
	"vultisigner/config"
	"vultisigner/internal/types"

	"github.com/vultisig/mobile-tss-lib/coordinator"
)

func JoinKeyGeneration(kg *types.KeyGeneration) error {
	relayServer := config.AppConfig.Relay.Server

	// if kg.Key != "vultisigner" {
	// 	return errors.New("key must be vultisigner")
	// }

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

func SaveKeyGeneration(kg *types.KeyGeneration) error {
	// if err := validation.Validate.Struct(kg); err != nil {
	// 	return errors.New("validation failed: " + err.Error())
	// }

	// if len(kg.Parties) < 2 {
	// 	return errors.New("parties must be 2 or more")
	// }

	// vultisignerIncluded := false
	// for _, party := range kg.Parties {
	// 	if party == "vultisigner" {
	// 		vultisignerIncluded = true
	// 		break
	// 	}
	// }
	// if !vultisignerIncluded {
	// 	return errors.New("vultisigner must be included in the parties")
	// }

	// kgModel := mapper.KeyGenerationMapper{}.ToModel(*kg)

	// if err := database.DB.Create(&kgModel).Error; err != nil {
	// 	return errors.New("failed to save key generation: " + err.Error())
	// }
	return nil
}

func GetKeyGenerationByID(id string) (types.KeyGeneration, error) {
	// var kg models.KeyGeneration
	// if err := database.DB.Where("id = ?", id).First(&kg).Error; err != nil {
	// 	if err.Error() == "record not found" {
	// 		return types.KeyGeneration{}, errors.New("key generation not found")
	// 	}
	// 	return types.KeyGeneration{}, err
	// }

	// kgType := mapper.KeyGenerationMapper{}.ToAPI(kg)

	// return kgType, nil
	return types.KeyGeneration{}, nil
}
