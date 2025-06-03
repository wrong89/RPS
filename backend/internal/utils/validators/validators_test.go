package validators_test

import (
	"rps/internal/utils/validators"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateBody(t *testing.T) {
	testCases := []struct {
		name           string
		requiredFields []string
		body           map[string]string
		expectedResult bool
	}{
		{
			name:           "Success: default using",
			requiredFields: []string{"test"},
			body: map[string]string{
				"test": "something",
			},
			expectedResult: true,
		},
		{
			name:           "Success: body has many fields and one requirement",
			requiredFields: []string{"test"},
			body: map[string]string{
				"some":   "field",
				"random": "random",
				"test":   "something",
			},
			expectedResult: true,
		},
		{
			name:           "Fail: required field doesn't exists in body",
			requiredFields: []string{"something"},
			body:           map[string]string{},
			expectedResult: false,
		},
		{
			name:           "Fail: empty requiredFields",
			requiredFields: []string{},
			body: map[string]string{
				"test": "test",
			},
			expectedResult: true,
		},
		{
			name:           "Fail: empty body and requiredFields",
			requiredFields: []string{},
			body:           map[string]string{},
			expectedResult: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := validators.ValidateBody(tc.body, tc.requiredFields...)
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}
