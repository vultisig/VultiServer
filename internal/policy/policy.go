package policy

import (
	"errors"
	"vultisigner/internal/database"
	"vultisigner/pkg/models"

	"github.com/google/uuid"
)

func SavePolicy(tp *models.TransactionPolicy) error {
	// Generate a new UUID for the TransactionPolicy
	tp.ID = uuid.New()

	if err := database.DB.Create(tp).Error; err != nil {
		if err.Error() == "record not found" {
			return errors.New("policy not found")
		}
		return err
	}
	return nil
}

func GetPolicyByID(id string) (models.TransactionPolicy, error) {
	var tp models.TransactionPolicy
	if err := database.DB.Where("id = ?", id).First(&tp).Error; err != nil {
		// if err.Error() == "record not found" {
		// 	return tp, errors.New("policy not found")
		// }
		return tp, err
	}
	return tp, nil
}

func CheckPolicy(transaction models.TransactionPolicy) error {
	// Implement the logic to check the policy
	return errors.New("not implemented")
}
