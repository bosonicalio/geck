package http

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/tesserical/geck/syserr"
)

// Errors is a sentinel structure containing a slice of [Error].
//
// Holds a Code property acting as top level status code.
type Errors struct {
	// Code Top level status code.
	Code int `json:"code" xml:"code"`
	// Errors slice of [Error].
	Errors []Error `json:"errors" xml:"errors"`
}

// Error is an informational structure specifying a system failure with a standard format
// for HTTP APIs.
//
// NOTE: If using XML, Metadata cannot be parsed and therefore, it is ignored.
type Error struct {
	Kind         string            `json:"kind" xml:"kind"`
	Code         int               `json:"code" xml:"code"`
	InternalCode string            `json:"internal_code" xml:"internal_code"`
	Message      string            `json:"message" xml:"message"`
	Metadata     map[string]string `json:"metadata" xml:"-"`
}

// NewErrorHandler allocates a new [echo.HTTPErrorHandler] instance.
//
// This routine generates [Error] structures to comply with a homogeneous error format.
func NewErrorHandler(config ServerConfig) echo.HTTPErrorHandler {
	return func(errSrc error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		topCode, errMap := newErrorMap(errSrc)
		switch config.ResponseFormat {
		case "xml":
			errSrc = c.XML(topCode, errMap["error"])
		case "string":
			errSrc = c.String(topCode, fmt.Sprintf("%+v", errMap))
		default:
			errSrc = c.JSON(topCode, errMap)
		}
	}
}

// Returns top status code and the error map.
func newErrorMap(errSrc error) (int, map[string]interface{}) {
	srcErrContainer, ok := errSrc.(syserr.Unwrapper)
	if !ok {
		code, err := newError(errSrc)
		return code, echo.Map{
			"error": Errors{
				Code: code,
				Errors: []Error{
					err,
				},
			},
		}
	}

	srcErrs := srcErrContainer.Unwrap()
	errs := make([]Error, 0, len(srcErrs))
	topCode := http.StatusBadRequest
	for _, item := range srcErrs {
		code, err := newError(item)
		errs = append(errs, err)
		if code > topCode {
			topCode = code
		}
	}
	return topCode, echo.Map{
		"error": Errors{
			Code:   topCode,
			Errors: errs,
		},
	}
}

func newError(errSrc error) (int, Error) {
	var errHTTP *echo.HTTPError
	ok := errors.As(errSrc, &errHTTP)
	if ok {
		return errHTTP.Code, Error{
			Code:         errHTTP.Code,
			Kind:         syserr.Internal.String(),
			InternalCode: "INTERNAL_SERVER_ERROR",
			Message:      fmt.Sprintf("%+v", errHTTP.Message),
		}
	}

	var errSys syserr.Error
	ok = errors.As(errSrc, &errSys)
	if !ok {
		return http.StatusInternalServerError, Error{
			Code:         http.StatusInternalServerError,
			Kind:         syserr.Internal.String(),
			InternalCode: "INTERNAL_SERVER_ERROR",
			Message:      http.StatusText(http.StatusInternalServerError),
		}
	}

	metadata := errSys.Metadata
	if len(metadata) == 0 {
		metadata = nil
	}
	code := translateSysErrCodes(errSys.Type)
	return code, Error{
		Code:         code,
		Kind:         errSys.Type.String(),
		InternalCode: errSys.InternalCode,
		Message:      errSys.Message,
		Metadata:     metadata,
	}
}

// -- CODE TRANSLATIONS --

var _sysErrCodes = map[syserr.Type]int{
	syserr.UnknownCode:         http.StatusInternalServerError,
	syserr.OutOfRange:          http.StatusBadRequest,
	syserr.InvalidArgument:     http.StatusBadRequest,
	syserr.MissingPrecondition: http.StatusPreconditionRequired,
	syserr.FailedPrecondition:  http.StatusPreconditionFailed,
	syserr.ResourceExists:      http.StatusConflict,
	syserr.ResourceNotFound:    http.StatusNotFound,
	syserr.PermissionDenied:    http.StatusForbidden,
	syserr.Unauthenticated:     http.StatusUnauthorized,
	syserr.Aborted:             http.StatusRequestTimeout,
	syserr.ResourceExhausted:   http.StatusTooManyRequests,
	syserr.DeadlineExceeded:    http.StatusRequestTimeout,
	syserr.Unimplemented:       http.StatusNotImplemented,
	syserr.DataLoss:            http.StatusUnprocessableEntity,
	syserr.Unavailable:         http.StatusServiceUnavailable,
	syserr.Internal:            http.StatusInternalServerError,
}

func translateSysErrCodes(t syserr.Type) int {
	return _sysErrCodes[t]
}
