package postgres

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/lib/pq"

	"multi-tenant-HR-information-system-backend/httperror"
)

// These errors are defined here instead of in the Routes package because they originate from DB queries
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

func NewUniqueViolationError(entity string, pgErr *pq.Error) *httperror.Error {
	detail := pgErr.Detail
	before, _, _ := strings.Cut(detail, ")=(")
	before = strings.ReplaceAll(before, "_", " ") // Replace all underscores in column names with spaces

	attributes := strings.Split(before[5:], ", ") // first 5 characters of "before" is excluded because it is a "Key ("

	message := `A %s with the provided %s already exists`
	subMessage := ""

	if len(attributes) == 1 {
		subMessage = attributes[0]
	} else if len(attributes) == 2 {
		subMessage = attributes[0] + " and " + attributes[1]
	} else {
		for i, attribute := range attributes {
			isLastColumn := i == len(attributes)-1
			if isLastColumn {
				subMessage = subMessage + "and " + attribute
			} else {
				subMessage = subMessage + attribute + ", "
			}
		}
	}

	message = fmt.Sprintf(message, entity, subMessage)

	return &httperror.Error{
		Status:  http.StatusConflict,
		Message: message,
		Code:    "UNIQUE-VIOLATION-ERROR",
	}
}

func NewInvalidForeignKeyError(pgErr *pq.Error) *httperror.Error {
	detail := pgErr.Detail
	before, _, _ := strings.Cut(detail, ")=(")
	before = strings.ReplaceAll(before, "_", " ") // Replace all underscores in column names with spaces

	attributes := strings.Split(before[5:], ", ") // first 5 characters of "before" is excluded because it is a "Key ("

	message := `The provided %s is invalid`
	subMessage := ""

	if len(attributes) == 1 {
		subMessage = attributes[0]
	} else {
		for i, attribute := range attributes {
			isLastColumn := i == len(attributes)-1
			if isLastColumn {
				subMessage = subMessage + attribute + " combination"
			} else {
				subMessage = subMessage + attribute + "-"
			}
		}
	}

	message = fmt.Sprintf(message, subMessage)

	return &httperror.Error{
		Status:  http.StatusBadRequest,
		Message: message,
		Code:    "INVALID-FOREIGN-KEY-ERROR",
	}
}

func New404NotFoundError(entity string) *httperror.Error {
	return &httperror.Error{
		Status:  404,
		Message: fmt.Sprintf("The %s does not exist", entity),
		Code:    "RESOURCE-NOT-FOUND-ERROR",
	}
}
