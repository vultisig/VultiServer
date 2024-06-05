package vault

import (
	"context"
	"errors"
	"vultisigner/ent"
	"vultisigner/internal/database"
	"vultisigner/internal/types"
)

func SaveVault(vault *types.VaultCreateRequest) (*ent.Vault, error) {
	v, err := database.Client.Vault.Create().SetName(vault.Name).SetLocalPartyKey("vultisigner").SetChainCodeHex("80871c0f885f953e5206e461630a9222148797e66276a83224c7b9b0f75b3ec0").Save(context.Background())
	if err != nil {
		return nil, errors.New("failed to save vault: " + err.Error())
	}

	return v, nil

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
}

func GetVaultByID(id string) (types.Vault, error) {
	// var kg models.KeyGeneration
	// if err := database.DB.Where("id = ?", id).First(&kg).Error; err != nil {
	// 	if err.Error() == "record not found" {
	// 		return types.Vault{}, errors.New("vault not found")
	// 	}
	// 	return types.Vault{}, err
	// }

	// vaultType := mapper.VaultMapper{}.ToAPI(kg)

	// return vaultType, nil
	return types.Vault{}, nil
}
