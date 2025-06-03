package validators

import (
	"net/mail"
)

func ValidateEmail(email string) bool {
	_, err := mail.ParseAddress(email)

	return err == nil
}

func ValidateBody(body map[string]string, requiredFields []string) bool {
	for _, field := range requiredFields {
		if _, ok := body[field]; !ok {
			return false
		}
	}

	return true
}
