package policy

import (
	"errors"
	"vultisigner/internal/database"
	"vultisigner/internal/models"
	"vultisigner/internal/validation"

	"github.com/google/uuid"
)

func SavePolicy(tp *models.TransactionPolicy) error {
	if err := validation.Validate.Struct(tp); err != nil {
		return errors.New("validation failed: " + err.Error())
		// var validationErrors string
		// for _, err := range err.(validator.ValidationErrors) {
		// 	validationErrors += err.Namespace() + ": " + err.Tag() + ", "
		// }
		// return errors.New("validation failed: " + validationErrors)
	}

	tp.ID = uuid.New()

	if err := database.DB.Create(tp).Error; err != nil {
		if err.Error() == "record not found" {
			return errors.New("policy not found")
		}
		return errors.New("failed to save policy" + err.Error())
	}
	return nil
}

func GetPolicyByID(id string) (models.TransactionPolicy, error) {
	var tp models.TransactionPolicy
	if err := database.DB.Where("id = ?", id).First(&tp).Error; err != nil {
		if err.Error() == "record not found" {
			return tp, errors.New("policy not found")
		}
		return tp, err
	}
	return tp, nil
}

func CheckPolicy(transaction models.TransactionPolicy) error {
	return errors.New("not implemented")
}
