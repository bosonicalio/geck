package validation_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tesserical/geck/syserr"
	"github.com/tesserical/geck/validation"
)

func TestNewGoPlaygroundValidator(t *testing.T) {
	type testStruct struct {
		DateString string    `json:"date_string" validate:"date"`
		Date       time.Time `json:"date" validate:"date"`
		SomeString string    `json:"some_string" validate:"required,lte=3"`
	}
	tests := []struct {
		name    string
		in      testStruct
		expErrs []error
	}{
		{
			name: "Should pass all rules",
			in: testStruct{
				DateString: "2023-10-01",
				Date:       time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC),
				SomeString: "foo",
			},
			expErrs: nil,
		},
		{
			name: "Should fail date string and string validations",
			in: testStruct{
				DateString: "some_invalid_date",
				SomeString: "some_invalid_string",
			},
			expErrs: []error{syserr.ErrInvalidFormat, syserr.ErrInvalidFormat},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			var validator validation.Validator = validation.NewGoPlaygroundValidator(
				validation.WithRules(validation.NewDateRule()),
			)

			// act
			err := validator.Validate(t.Context(), tt.in)

			// assert
			if len(tt.expErrs) == 0 {
				assert.NoError(t, err)
				return
			}
			assert.NotNil(t, err)
			errContainer, ok := err.(syserr.Unwrapper)
			require.True(t, ok)
			errs := errContainer.Unwrap()
			assert.Len(t, errs, len(tt.expErrs))
			for i, expErr := range tt.expErrs {
				assert.ErrorIs(t, errs[i], expErr)
			}
		})
	}
}
