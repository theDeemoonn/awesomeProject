package validator

import (
	"gopkg.in/go-playground/validator.v9"
	"regexp"
)

// RegisterCustomValidators регистрирует все кастомные функции валидации
func RegisterCustomValidators(validate *validator.Validate) {
	validate.RegisterValidation("ogrn", validateOGRN)
	validate.RegisterValidation("inn", validateINN)
}

// validateOGRN проверяет, что значение является корректным ОГРН
func validateOGRN(fl validator.FieldLevel) bool {
	ogrn := fl.Field().String()
	match, _ := regexp.MatchString(`^\d{13}$`, ogrn)
	return match
}

// validateINN проверяет, что значение является корректным ИНН
func validateINN(fl validator.FieldLevel) bool {
	inn := fl.Field().String()
	// ИНН может быть 10 или 12 цифр
	match, _ := regexp.MatchString(`^\d{10}$|^\d{12}$`, inn)
	return match
}
