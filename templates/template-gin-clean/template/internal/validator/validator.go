package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func Validate(i interface{}) []ValidationError {
	var errs []ValidationError

	err := validate.Struct(i)
	if err == nil {
		return nil
	}

	for _, verr := range err.(validator.ValidationErrors) {
		errs = append(errs, ValidationError{
			Field:   toSnakeCase(verr.Field()),
			Message: messageForTag(verr.Tag(), verr.Param()),
		})
	}

	return errs
}

func messageForTag(tag, param string) string {
	switch tag {
	case "required":
		return "this field is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return fmt.Sprintf("must be at least %s characters", param)
	case "max":
		return fmt.Sprintf("must be at most %s characters", param)
	case "gte":
		return fmt.Sprintf("must be greater than or equal to %s", param)
	case "lte":
		return fmt.Sprintf("must be less than or equal to %s", param)
	default:
		return fmt.Sprintf("validation failed on '%s'", tag)
	}
}

func toSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				result = append(result, '_')
			}
			result = append(result, r+32)
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

func FormatErrors(errs []ValidationError) map[string]string {
	formatted := make(map[string]string)
	for _, e := range errs {
		formatted[e.Field] = e.Message
	}
	return formatted
}

func FormatErrorsSlice(errs []ValidationError) []string {
	var msgs []string
	for _, e := range errs {
		msgs = append(msgs, fmt.Sprintf("%s: %s", e.Field, e.Message))
	}
	return msgs
}

func Valid(i interface{}) bool {
	err := validate.Struct(i)
	return err == nil
}

func HasErrors(errs []ValidationError) bool {
	return len(errs) > 0
}

func ErrorSlice(errs []ValidationError) string {
	var sb strings.Builder
	for i, e := range errs {
		if i > 0 {
			sb.WriteString("; ")
		}
		sb.WriteString(fmt.Sprintf("%s: %s", e.Field, e.Message))
	}
	return sb.String()
}
