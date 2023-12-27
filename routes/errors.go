package routes

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

// Why is HttpError struct defined in the routes package?
// Because route handlers consume HttpErrors (from various sources e.g Models)
// Therefore, the routes package is responsible for defining exactly what it needs in a HttpError

type HttpError struct {
	Status  int
	Message string
	Code    string
}

func (err *HttpError) Error() string {
	return err.Message
}

func NewInternalServerError(err error) *HttpError {
	err = errors.New(err.Error()) // wraps the original error in a pkg/errors error. This way, the stack trace is included

	return &HttpError{
		Status:  http.StatusInternalServerError,
		Message: fmt.Sprintf("%+v", err), // include the stack trace in the error message
		Code:    "INTERNAL-SERVER-ERROR",
	}
}

func NewInvalidJSONError() *HttpError {
	return &HttpError{
		Status:  400,
		Message: "Invalid JSON provided as request body",
		Code:    "INVALID-JSON-ERROR",
	}
}

func NewInputValidationError(validationErrors map[string]string) *HttpError {
	message := "There are one or more errors with your input(s):"
	for _, errorMessage := range validationErrors {
		message = message + "\n" + errorMessage
	}

	return &HttpError{
		Status:  http.StatusBadRequest,
		Message: message,
		Code:    "INPUT-VALIDATION-ERROR",
	}
}
