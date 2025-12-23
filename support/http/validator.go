package http

import (
	"net/http"

	"go-reasonable-api/support/errors"

	"github.com/go-playground/validator/v10"
)

type Validator struct {
	validator *validator.Validate
}

func NewValidator() *Validator {
	return &Validator{validator: validator.New()}
}

func (v *Validator) Validate(i any) error {
	if err := v.validator.Struct(i); err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			return err
		}

		details := make(map[string]any)
		for _, e := range validationErrors {
			field := e.Field()
			details[field] = []string{messageForTag(e)}
		}

		return errors.NewWithDetails("VALIDATION_ERROR", "validation failed", http.StatusBadRequest, details)
	}
	return nil
}

func messageForTag(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email"
	case "min":
		return "must be at least " + fe.Param() + " characters"
	case "max":
		return "must be at most " + fe.Param() + " characters"
	default:
		return "is invalid"
	}
}
