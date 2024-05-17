package policy

import (
	"errors"
	"vultisigner/pkg/models"
)

func SavePolicy(tp models.TransactionPolicy) error {
	// save policy to database
	return nil
}

func CheckPolicy(transaction models.TransactionPolicy) error {
	// get policy from the db & check it
	return errors.New("not implemented")
}
