package validation

import (
	"time"
)

// ValidateFunc is a function type that defines a custom validation function.
type ValidateFunc func(fieldName string, value any) bool

// Rule represents a validation rule with a name and a validation function.
type Rule struct {
	// Name is the name of the validation rule.
	Name string
	// ValidateFunc is the function that performs the validation.
	ValidateFunc ValidateFunc
}

// - Rules -

// NewDateRule is a custom validation function that checks if the provided value is a valid date string in
// the format "YYYY-MM-DD".
func NewDateRule() Rule {
	return Rule{
		Name: "date",
		ValidateFunc: func(fieldName string, value any) bool {
			_, ok := value.(time.Time)
			if ok {
				// If the value is already a time.Time, it's valid.
				return true
			}
			val, ok := value.(string)
			if !ok {
				return false
			}
			_, err := time.Parse(time.DateOnly, val)
			return err == nil
		},
	}
}
