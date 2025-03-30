package syserr

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

var (
	// ErrResourceNotFound is returned when the resource is not found.
	ErrResourceNotFound = errors.New("resource not found")
	// ErrResourceAlreadyExists is returned when the resource already exists.
	ErrResourceAlreadyExists = errors.New("resource already exists")
	// ErrInvalidFormat is returned when the format is invalid.
	ErrInvalidFormat = errors.New("invalid format")
	// ErrMissingValue is returned when the value is missing.
	ErrMissingValue = errors.New("missing value")
)

// NewResourceNotFound allocates a new [Error].
//
// Specifies [Error] properties for `resource not found` error cases.
func NewResourceNotFound[T any]() Error {
	typeof := reflect.TypeFor[T]()
	msg := fmt.Sprintf("resource '%s' not found", typeof.String())
	return New(ResourceNotFound, msg,
		WithInternalCode("RESOURCE_NOT_FOUND"),
		WithStaticError(ErrResourceNotFound),
	)
}

// NewResourceAlreadyExists allocates a new [Error].
//
// Specifies [Error] properties for `resource already exists` error cases.
func NewResourceAlreadyExists[T any]() Error {
	typeof := reflect.TypeFor[T]()
	msg := fmt.Sprintf("resource '%s' already exists", typeof.String())
	return New(ResourceExists, msg,
		WithInternalCode("RESOURCE_ALREADY_EXISTS"),
		WithStaticError(ErrResourceAlreadyExists),
	)
}

// NewInvalidFormat allocates a new [Error].
//
// Specifies [Error] properties for `invalid format` error cases.
func NewInvalidFormat(name, format string) Error {
	msg := fmt.Sprintf("'%s' is invalid", name)
	return New(InvalidArgument, msg,
		WithInternalCode("INVALID_FORMAT"),
		WithInfo("expected_format", format),
		WithStaticError(ErrInvalidFormat),
	)
}

// NewMissingValue allocates a new [Error].
//
// Specifies [Error] properties for `missing value` error cases.
func NewMissingValue(name string) Error {
	msg := fmt.Sprintf("'%s' is missing", name)
	return New(InvalidArgument, msg,
		WithInternalCode("MISSING_VALUE"),
		WithStaticError(ErrMissingValue),
	)
}

// NewNotOneOf allocates a new [Error].
//
// Specifies [Error] properties for `value is not one of` error cases.
func NewNotOneOf(name string, values ...string) Error {
	acceptedValues := strings.Join(values, ",")
	msg := fmt.Sprintf("'%s' is not equals to one of the accepted values (%s)", name, acceptedValues)
	return New(InvalidArgument, msg,
		WithInternalCode("VALUE_NOT_ONE_OF"),
		WithInfo("accepted_values", acceptedValues),
		WithStaticError(ErrInvalidFormat),
	)
}

// NewNotEquals allocates a new [Error].
//
// Specifies [Error] properties for `value is not equals to` error cases.
func NewNotEquals(name, exp string) Error {
	msg := fmt.Sprintf("'%s' is not equals to (%s)", name, exp)
	return New(InvalidArgument, msg,
		WithInternalCode("VALUE_NOT_EQUALS"),
		WithInfo("expected_value", exp),
	)
}

// NewEquals allocates a new [Error].
//
// Specifies [Error] properties for `value is equals to` error cases.
func NewEquals(name string, invalidVals ...string) Error {
	valStr := strings.Join(invalidVals, ",")
	msg := fmt.Sprintf("'%s' is equals to (%s)", name, valStr)
	return New(InvalidArgument, msg,
		WithInternalCode("VALUE_EQUALS"),
		WithInfo("invalid_values", valStr),
		WithStaticError(ErrInvalidFormat),
	)
}

// NewInvalidLength allocates a new [Error].
//
// Specifies [Error] properties for `value does not have the expected length` error cases.
func NewInvalidLength(name string, expLen int) Error {
	msg := fmt.Sprintf("'%s' has an invalid length, expected (%d)", name, expLen)
	return New(InvalidArgument, msg,
		WithInternalCode("VALUE_INVALID_LENGTH"),
		WithInfo("expected_length", strconv.Itoa(expLen)),
		WithStaticError(ErrInvalidFormat),
	)
}

// NewAboveLimit allocates a new [Error].
//
// Specifies [Error] properties for `value is above the expected limit` error cases.
func NewAboveLimit(name string, max int) Error {
	msg := fmt.Sprintf("'%s' has an invalid size, expected maximum value (%d)", name, max)
	return New(InvalidArgument, msg,
		WithInternalCode("VALUE_INVALID_SIZE"),
		WithInfo("max_size", strconv.Itoa(max)),
		WithStaticError(ErrInvalidFormat),
	)
}

// NewBelowLimit allocates a new [Error].
//
// Specifies [Error] properties for `value is below the expected limit` error cases.
func NewBelowLimit(name string, min int) Error {
	msg := fmt.Sprintf("'%s' has an invalid size, expected minimum value (%d)", name, min)
	return New(InvalidArgument, msg,
		WithInternalCode("VALUE_INVALID_SIZE"),
		WithInfo("min_size", strconv.Itoa(min)),
		WithStaticError(ErrInvalidFormat),
	)
}
