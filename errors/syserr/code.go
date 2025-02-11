package syserr

import "fmt"

// Type is a custom integer representing [Error] types.
type Type uint16

const (
	// UnknownCode the error code is not known.
	UnknownCode Type = iota
	// OutOfRange the value is out of the specified range.
	OutOfRange
	// InvalidArgument the value is invalid according to internal rules.
	InvalidArgument
	// MissingPrecondition a precondition is required to be executed.
	MissingPrecondition
	// FailedPrecondition a required precondition was executed, but it resulted in failure.
	FailedPrecondition
	// ResourceExists the resource already exist in a persistence store.
	ResourceExists
	// ResourceNotFound the resource was not present in a persistence store.
	ResourceNotFound
	// PermissionDenied access to a certain operation -or resource- was denied.
	PermissionDenied
	// Unauthenticated the given routine requires the caller (aka. Principal) to be authenticated, yet
	// no authentication was found.
	Unauthenticated
	// Aborted the routine execution was aborted.
	Aborted
	// ResourceExhausted the system cannot handle more calls at this time.
	ResourceExhausted
	// DeadlineExceeded the routine execution could not respond in time, reaching the caller wait deadline.
	DeadlineExceeded
	// Unimplemented the operation has not been implemented yet.
	Unimplemented
	// DataLoss the operation resulted in data missing.
	DataLoss
	// Unavailable the system/operation is not available to accept calls.
	Unavailable
	// Internal a non-public error happened during the routine execution.
	Internal
)

var (
	// compile-time assertions
	_ fmt.Stringer = UnknownCode

	_codeStrings = map[Type]string{
		UnknownCode:         "UNKNOWN",
		OutOfRange:          "OUT_OF_RANGE",
		InvalidArgument:     "INVALID_ARGUMENT",
		MissingPrecondition: "MISSING_PRECONDITION",
		FailedPrecondition:  "FAILED_PRECONDITION",
		ResourceExists:      "RESOURCE_EXISTS",
		ResourceNotFound:    "RESOURCE_NOT_FOUND",
		PermissionDenied:    "PERMISSION_DENIED",
		Unauthenticated:     "UNAUTHENTICATED",
		Aborted:             "ABORTED",
		ResourceExhausted:   "RESOURCE_EXHAUSTED",
		DeadlineExceeded:    "DEADLINE_EXCEEDED",
		Unimplemented:       "UNIMPLEMENTED",
		DataLoss:            "DATA_LOSS",
		Unavailable:         "UNAVAILABLE",
		Internal:            "INTERNAL",
	}
)

func (e Type) String() string {
	return _codeStrings[e]
}
