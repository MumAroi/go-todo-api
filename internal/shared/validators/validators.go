package validators

import (
	"reflect"
	"unicode"

	"github.com/go-playground/validator/v10"
)

func RegisterCustomValidators(validate *validator.Validate) {
	validate.RegisterValidation("not_empty_if_present", notEmptyIfPresent)
	validate.RegisterValidation("strong_password", strongPassword)
}

func notEmptyIfPresent(fl validator.FieldLevel) bool {
	field := fl.Field()

	if field.Kind() == reflect.Pointer {
		return field.IsNil() || field.Elem().String() != ""
	}

	return field.String() != ""
}

func strongPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	if password == "" {
		return false
	}

	if len(password) < 6 || len(password) > 10 {
		return false
	}

	hasLower := false
	hasUpper := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	return hasLower && hasUpper && hasDigit && hasSpecial
}
