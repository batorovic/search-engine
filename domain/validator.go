package domain

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func ValidateProviderContent(content ProviderContent) error {
	if err := validate.Struct(content); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			return fmt.Errorf("validation failed: %s", formatValidationErrors(validationErrors))
		}
		return fmt.Errorf("validation error: %w", err)
	}
	return nil
}

func ValidateProviderContents(contents []ProviderContent) ([]ProviderContent, []error) {
	validContents := make([]ProviderContent, 0, len(contents))
	errors := make([]error, 0)

	for i, content := range contents {
		if err := ValidateProviderContent(content); err != nil {
			errors = append(errors, fmt.Errorf("content[%d]: %w", i, err))
		} else {
			validContents = append(validContents, content)
		}
	}

	return validContents, errors
}

func formatValidationErrors(errs validator.ValidationErrors) string {
	var errorMsg string
	for i, err := range errs {
		if i > 0 {
			errorMsg += "; "
		}
		errorMsg += fmt.Sprintf("%s failed %s validation", err.Field(), err.Tag())

		switch err.Tag() {
		case "required":
			errorMsg += " (field is required)"
		case "min":
			errorMsg += fmt.Sprintf(" (minimum value: %s)", err.Param())
		case "max":
			errorMsg += fmt.Sprintf(" (maximum value: %s)", err.Param())
		case "gte":
			errorMsg += fmt.Sprintf(" (must be >= %s)", err.Param())
		case "oneof":
			errorMsg += fmt.Sprintf(" (must be one of: %s)", err.Param())
		}
	}
	return errorMsg
}
