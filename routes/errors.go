package routes

import (
	"net/http"

	"multi-tenant-HR-information-system-backend/httperror"
)

func New404NotFoundError() *httperror.Error {
	return &httperror.Error{
		Status:  404,
		Message: "Not found",
		Code:    "RESOURCE-NOT-FOUND-ERROR",
	}
}

func NewInvalidJSONError() *httperror.Error {
	return &httperror.Error{
		Status:  400,
		Message: "Invalid JSON provided as request body",
		Code:    "INVALID-JSON-ERROR",
	}
}

func NewUnauthenticatedError() *httperror.Error {
	return &httperror.Error{
		Status:  http.StatusUnauthorized,
		Message: "User unauthenticated",
		Code:    "USER-UNAUTHENTICATED",
	}
}

func NewUnauthorisedError() *httperror.Error {
	return &httperror.Error{
		Status:  http.StatusForbidden,
		Message: "User unauthorised",
		Code:    "USER-UNAUTHORISED",
	}
}
