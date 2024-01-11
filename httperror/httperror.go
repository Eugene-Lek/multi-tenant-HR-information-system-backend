package httperror

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"	
)

// Why is the Error struct defined in a separate package?
// Because it is a wrapper for the standard error interface & is used across multiple packages (i.e. used globally)
// In the same way that pkg/errors is a global wrapper of the standard error interface & is thus defined in its own package, 
// httperror should too

type Error struct {
	Status  int
	Message string
	Code    string
}

func (err *Error) Error() string {
	return err.Message
}

func NewInternalServerError(err error) *Error {
	err = errors.New(err.Error()) // wraps the original error in a pkg/errors error. This way, the stack trace is included

	return &Error{
		Status:  http.StatusInternalServerError,
		Message: fmt.Sprintf("%+v", err), // include the stack trace in the error message
		Code:    "INTERNAL-SERVER-ERROR",
	}
}