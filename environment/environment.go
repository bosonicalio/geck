package environment

import (
	"encoding"
	"errors"
	"fmt"
	"strings"
)

// An Environment represents a software deployment environment used in enterprise systems.
// It defines the context in which an application operates, such as development, testing, or production.
//
// This structure implements [encoding.TextMarshaler], [encoding.TextUnmarshaler] and
// [fmt.Stringer] for easier integration with external components.
type Environment uint8

const (
	Unknown Environment = iota
	Production
	Staging
	Development
	Local
)

var (
	// compile-time assertions
	_ encoding.TextMarshaler   = (*Environment)(nil)
	_ encoding.TextUnmarshaler = (*Environment)(nil)
	_ fmt.Stringer             = Unknown

	// ErrIsInvalid the given environment is not valid (i.e. is unknown).
	ErrIsInvalid = errors.New("invalid environment")

	_stringToInternalMap = map[string]Environment{
		"production":  Production,
		"staging":     Staging,
		"development": Development,
		"prod":        Production,
		"stage":       Staging,
		"stg":         Staging,
		"dev":         Development,
		"local":       Local,
		"sandbox":     Staging,
		"snx":         Staging,
		"pilot":       Staging,
	}
	_internalToStringMap = map[Environment]string{
		Production:  "production",
		Staging:     "staging",
		Development: "development",
		Local:       "local",
	}
)

// Parse allocates a new [Environment] instance based on its string value.
func Parse(value string) (Environment, error) {
	value = strings.ToLower(value)
	environment, ok := _stringToInternalMap[value]
	if !ok {
		return Unknown, ErrIsInvalid
	}
	return environment, nil
}

func (e Environment) MarshalText() (text []byte, err error) {
	return []byte(e.String()), nil
}

func (e *Environment) UnmarshalText(text []byte) error {
	environment, err := Parse(string(text))
	if err != nil {
		return err
	}
	*e = environment
	return nil
}

func (e Environment) String() string {
	return _internalToStringMap[e]
}
