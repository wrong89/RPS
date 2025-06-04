package helpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"rps/internal/utils/validators"
)

type BodyData = map[string]string

var (
	ErrInvalidBody = errors.New("invalid body")
)

func GetBodyFromRequest(r *http.Request) (BodyData, error) {
	const op = "helpers.body.GetBodyFromRequest"
	defer r.Body.Close()

	var result BodyData

	err := json.NewDecoder(r.Body).Decode(&result)
	if err != nil {
		return result, fmt.Errorf("%s: %w", op, err)
	}

	return result, nil
}

func GetValidatedBody(r *http.Request, requiredFields ...string) (BodyData, error) {
	body, err := GetBodyFromRequest(r)
	if err != nil {
		return body, err
	}

	valid := validators.ValidateBody(body, requiredFields...)
	if !valid {
		return nil, ErrInvalidBody
	}

	return body, nil
}
