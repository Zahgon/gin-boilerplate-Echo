package forms

import (
	"encoding/json"
	"errors"

	"github.com/go-playground/validator/v10"
)

// ValidationMessages maps struct field names to per-tag error messages.
type ValidationMessages map[string]map[string]string

// DefaultMessage is returned when no specific message is configured.
const DefaultMessage = "Something went wrong, please try again later"

// Translate converts a validation error into a user-friendly string
// using the provided message map. Returns the first matched field error.
func Translate(err error, messages ValidationMessages) string {
	var unmarshalTypeError *json.UnmarshalTypeError
	if errors.As(err, &unmarshalTypeError) {
		return DefaultMessage
	}

	switch err.(type) {
	case validator.ValidationErrors:
		for _, e := range err.(validator.ValidationErrors) {
			if fieldMessages, ok := messages[e.Field()]; ok {
				if msg, ok := fieldMessages[e.Tag()]; ok {
					return msg
				}
				if msg, ok := fieldMessages["default"]; ok {
					return msg
				}
			}
		}
	default:
		return "Invalid request"
	}

	return DefaultMessage
}
