package configuration

import (
	"context"

	"github.com/caarlos0/env/v11"

	"github.com/hadroncorp/geck/validation"
)

var _globalValidator validation.Validator = validation.NewGoPlaygroundValidator(
	validation.ValidatorConfig{
		StructFieldDriver: validation.JSONDriver,
	},
)

// Parse parses the environment variables into a struct of type T and validates it.
//
// Uses [env] to parse the environment variables and [github.com/go-playground/validator/v10] to validate the struct.
func Parse[T any]() (T, error) {
	config, err := env.ParseAs[T]()
	if err != nil {
		return config, err
	}

	if err = _globalValidator.Validate(context.Background(), config); err != nil {
		return config, err
	}
	return config, nil
}
