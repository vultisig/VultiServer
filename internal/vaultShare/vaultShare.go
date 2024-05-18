package vaultShare

import (
	"errors"
	"vultisigner/internal/crypto"
	"vultisigner/internal/database"
	"vultisigner/internal/validation"
	"vultisigner/pkg/models"

	"github.com/google/uuid"
)

func SaveVaultShare(tp *models.VaultShare) error {
	if err := validation.Validate.Struct(tp); err != nil {
		return errors.New("validation failed: " + err.Error())
		// var validationErrors string
		// for _, err := range err.(validator.ValidationErrors) {
		// 	validationErrors += err.Namespace() + ": " + err.Tag() + ", "
		// }
		// return errors.New("validation failed: " + validationErrors)
	}

	// encrypt the Xi fields in EcdsaLocalData and EddsaLocalData
	encryptedXi, err := crypto.Encrypt(tp.EcdsaLocalData.Xi)
	if err != nil {
		return errors.New("failed to encrypt EcdsaLocalData.Xi: " + err.Error())
	}
	tp.EcdsaLocalData.Xi = encryptedXi

	if tp.EddsaLocalData.Xi != nil {
		encryptedXi, err := crypto.Encrypt(*tp.EddsaLocalData.Xi)
		if err != nil {
			return errors.New("failed to encrypt EddsaLocalData.Xi: " + err.Error())
		}
		tp.EddsaLocalData.Xi = &encryptedXi
	}

	tp.ID = uuid.New()

	if err := database.DB.Create(tp).Error; err != nil {
		return errors.New("failed to save vault share: " + err.Error())
	}
	return nil
}

func GetVaultShareByID(id string) (models.VaultShare, error) {
	var tp models.VaultShare
	if err := database.DB.Where("id = ?", id).First(&tp).Error; err != nil {
		if err.Error() == "record not found" {
			return tp, errors.New("vault share not found")
		}
		return tp, err
	}
	return tp, nil
}
