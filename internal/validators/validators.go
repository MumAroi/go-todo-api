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

	// ถ้าเป็น pointer และ nil → ข้าม
	if field.Kind() == reflect.Pointer {
		return field.IsNil() || field.Elem().String() != ""
	}

	// ถ้ามีค่า → ต้องไม่ empty
	return field.String() != ""
}

func strongPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	// ห้ามเป็นค่าว่าง
	if password == "" {
		return false
	}

	// ตรวจสอบความยาว (6-10)
	if len(password) < 6 || len(password) > 10 {
		return false
	}

	// ตรวจสอบว่ามีตัวอักษรแต่ละประเภทไหม
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
