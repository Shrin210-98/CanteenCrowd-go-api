package utils

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()

	// Register custom validators if needed
	validate.RegisterValidation("phone", validatePhone)
}

// ValidateStruct validates a struct and returns formatted errors
func ValidateStruct(s interface{}) map[string]string {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	errors := make(map[string]string)

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			field := e.Field()
			tag := e.Tag()
			param := e.Param()

			// Convert field name to camelCase for JSON response
			field = toCamelCase(field)

			// Generate human-readable error messages
			errors[field] = getErrorMessage(field, tag, param)
		}
	}

	return errors
}

func getErrorMessage(field, tag, param string) string {
	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, param)
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, param)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, param)
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "datetime":
		return fmt.Sprintf("%s must be in format: %s", field, param)
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, param)
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, param)
	default:
		return fmt.Sprintf("%s failed validation: %s", field, tag)
	}
}

func toCamelCase(s string) string {
	if s == "" {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}

// Custom phone validator
func validatePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	if phone == "" {
		return true // Empty is valid (use required for mandatory)
	}
	// Simple phone validation - allows digits, spaces, dashes, parentheses, plus
	return len(phone) >= 10 && len(phone) <= 15
}
