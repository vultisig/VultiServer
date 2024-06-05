package vault

import (
	"errors"
	"vultisigner/internal/database"
	"vultisigner/internal/models"
	"vultisigner/internal/types"
	"vultisigner/internal/validation"
)

func SaveVault(kg *types.Vault) error {
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

	kgModel := mapper.KeyGenerationMapper{}.ToModel(*kg)

	if err := database.DB.Create(&kgModel).Error; err != nil {
		return errors.New("failed to save key generation: " + err.Error())
	}
	return nil
}

func GetVaultByID(id string) (types.Vault, error) {
	var kg models.KeyGeneration
	if err := database.DB.Where("id = ?", id).First(&kg).Error; err != nil {
		if err.Error() == "record not found" {
			return types.Vault{}, errors.New("vault not found")
		}
		return types.Vault{}, err
	}

	vaultType := mapper.VaultMapper{}.ToAPI(kg)

	return vaultType, nil
}
