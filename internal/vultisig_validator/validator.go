package vultisigValidator

import "github.com/go-playground/validator/v10"

type VultisigValidator struct {
	Validator *validator.Validate
}

func (v *VultisigValidator) Validate(i interface{}) error {
	if err := v.Validator.Struct(i); err != nil {
		return v.Validator.Struct(i)
	}
	return nil
}
