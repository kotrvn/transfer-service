package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// decodeAndValidate читает JSON и проверяет его на ошибки
func decodeAndValidate(r *http.Request, v interface{}) error {
	// 1. Читаем JSON
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return fmt.Errorf("неверный формат JSON")
	}

	// 2. Валидируем
	if err := validate.Struct(v); err != nil {
		// Превращаем ошибки в понятный текст
		var errors []string
		for _, err := range err.(validator.ValidationErrors) {
			field := strings.ToLower(err.Field())
			switch err.Tag() {
			case "required":
				errors = append(errors, fmt.Sprintf("поле '%s' обязательно", field))
			case "email":
				errors = append(errors, fmt.Sprintf("поле '%s' должно быть валидным email", field))
			case "min":
				errors = append(errors, fmt.Sprintf("поле '%s' должно быть не короче %s", field, err.Param()))
			case "gt":
				errors = append(errors, fmt.Sprintf("поле '%s' должно быть больше %s", field, err.Param()))
			case "gte":
				errors = append(errors, fmt.Sprintf("поле '%s' должно быть больше или равно %s", field, err.Param()))
			case "oneof":
				errors = append(errors, fmt.Sprintf("поле '%s' должно быть одним из: %s", field, err.Param()))
			default:
				errors = append(errors, fmt.Sprintf("поле '%s' неверно", field))
			}
		}
		return fmt.Errorf("%s", strings.Join(errors, ", "))
	}

	return nil
}
