package routes

import (
	"net/http"

	"multi-tenant-HR-information-system-backend/httperror"
)

var Err404NotFound = &httperror.Error{
	Status:  404,
	Message: "Not found",
	Code:    "RESOURCE-NOT-FOUND-ERROR",
}

var ErrInvalidJSON = &httperror.Error{
	Status:  400,
	Message: "Invalid JSON provided as request body",
	Code:    "INVALID-JSON-ERROR",
}

var ErrFileTooBig = &httperror.Error{
	Status: 400,
	Message: "Uploaded file has exceeded the size limit",
	Code: "FILE-TOO-BIG-ERROR",
}

var ErrUserUnauthenticated = &httperror.Error{
	Status:  http.StatusUnauthorized,
	Message: "User unauthenticated",
	Code:    "USER-UNAUTHENTICATED",
}

var ErrUserUnauthorised = &httperror.Error{
	Status:  http.StatusForbidden,
	Message: "User unauthorised",
	Code:    "USER-UNAUTHORISED",
}

var ErrInvalidSupervisor = &httperror.Error{
	Status:  http.StatusBadRequest,
	Message: "You have provided an invalid supervisor",
	Code:    "INVALID-SUPERVISOR-ERROR",
}

var ErrMissingSupervisorApproval = &httperror.Error{
	Status:  403,
	Message: "Supervisor approval is missing",
	Code:    "MISSING-SUPERVISOR-APPROVAL-ERROR",
}

var ErrMissingHrApproval = &httperror.Error{
	Status:  403,
	Message: "HR approval is missing",
	Code:    "MISSING-HR-APPROVAL-ERROR",
}

var ErrJobRequisitionAlreadyFilled = &httperror.Error{
	Status:  409,
	Message: "The job requisition has already been filled",
	Code:    "JOB-REQUISITION-ALREADY-FILLED",	
}