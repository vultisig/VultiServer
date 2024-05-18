package keyGeneration

import (
	"errors"
	"vultisigner/internal/database"
	"vultisigner/internal/validation"
	"vultisigner/pkg/models"

	"github.com/spf13/viper"
	"github.com/vultisig/mobile-tss-lib/coordinator"
)

func JoinKeyGeneration(kg *models.KeyGeneration) error {
	server := viper.GetString("director.server")

	_, err := coordinator.ExecuteKeyGeneration(coordinator.KeygenInput{
		Key:       "vultisigner",
		Parties:   kg.Parties,
		Session:   kg.Session,
		Server:    server,
		ChainCode: kg.ChainCode,
	})
	if err != nil {
		return err
	}

	return nil
}

func SaveKeyGeneration(kg *models.KeyGeneration) error {
	if err := validation.Validate.Struct(kg); err != nil {
		return errors.New("validation failed: " + err.Error())
	}

	if len(kg.Parties) < 2 {
		return errors.New("parties must be 2 or more")
	}

	vultisignerIncluded := false
	for _, party := range kg.Parties {
		if party == "vultisigner" {
			vultisignerIncluded = true
			break
		}
	}
	if !vultisignerIncluded {
		return errors.New("vultisigner must be included in the parties")
	}

	if err := database.DB.Create(kg).Error; err != nil {
		return errors.New("failed to save key generation: " + err.Error())
	}
	return nil
}

func GetKeyGenerationByID(id string) (models.KeyGeneration, error) {
	var kg models.KeyGeneration
	if err := database.DB.Where("id = ?", id).First(&kg).Error; err != nil {
		if err.Error() == "record not found" {
			return kg, errors.New("key generation not found")
		}
		return kg, err
	}
	return kg, nil
}
