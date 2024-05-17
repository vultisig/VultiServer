package models

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

var Validate *validator.Validate

func init() {
	Validate = validator.New()

	_ = Validate.RegisterValidation("numeric", func(fl validator.FieldLevel) bool {
		match, _ := regexp.MatchString(`^[0-9]+$`, fl.Field().String())
		return match
	})

	_ = Validate.RegisterValidation("hexadecimal", func(fl validator.FieldLevel) bool {
		match, _ := regexp.MatchString(`^[0-9a-fA-F]+$`, fl.Field().String())
		return match
	})
}
