package utils

import (
	"errors"

	"github.com/go-playground/validator/v10"
)

type BindErrorResponse struct {
	Success bool              `json:"success"`
	Message string            `json:"message"`
	Errors  map[string]string `json:"errors,omitempty"`
}

func ParseValidationError(err error) BindErrorResponse {
	errorMap := make(map[string]string)

	if ve, ok := errors.AsType[validator.ValidationErrors](err); ok {
		for _, fe := range ve {
			errorMap[fe.Field()] = getCustomMessage(fe)
		}
		return BindErrorResponse{
			Success: false,
			Message: "ข้อมูลนำเข้าไม่ถูกต้อง",
			Errors:  errorMap,
		}
	}

	return BindErrorResponse{
		Success: false,
		Message: "รูปแบบ JSON ไม่ถูกต้อง",
	}
}

func getCustomMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "กรุณากรอกข้อมูลในช่องนี้"
	case "email":
		return "รูปแบบอีเมลไม่ถูกต้อง"
	case "min":
		return "ความยาวต้องไม่ต่ำกว่า " + fe.Param() + " ตัวอักษร"
	case "max":
		return "ความยาวต้องไม่เกิน " + fe.Param() + " ตัวอักษร"
	case "strong_password":
		return "รหัสผ่านต้องมีความแข็งแกร่ง"
	case "not_empty_if_present":
		return "ห้ามเป็นค่าว่าง"
	}
	return "ข้อมูลฟิลด์นี้ไม่ถูกต้อง (" + fe.Tag() + ")"
}
