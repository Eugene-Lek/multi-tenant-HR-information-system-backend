package postgres

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/lib/pq"

	"multi-tenant-HR-information-system-backend/routes"
)

// These errors are defined here instead of in the Routes package because they originate from DB queries

func NewUniqueViolationError(entity string, pgErr *pq.Error) *routes.HttpError {
	detail := pgErr.Detail
	before, _, _ := strings.Cut(detail, ")=(")
	before = strings.ReplaceAll(before, "_", " ") // Replace all underscores in column names with spaces

	columns := strings.Split(before[5:], ", ") // first 5 characters of "before" is excluded because it is a "Key ("

	message := `A %s with the provided %s already exists`
	subMessage := ""

	if len(columns) == 1 {
		subMessage = columns[0]
	} else if len(columns) == 2 {
		subMessage = columns[0] + " and " + columns[1]
	} else {
		for i, column := range columns {
			isLastColumn := i == len(columns)-1
			if isLastColumn {
				subMessage = subMessage + "and " + column
			} else {
				subMessage = subMessage + column + ", "
			}
		}
	}

	message = fmt.Sprintf(message, entity, subMessage)

	return &routes.HttpError{
		Status:  http.StatusConflict,
		Message: message,
		Code:    "UNIQUE-VIOLATION-ERROR",
	}
}

func NewInvalidForeignKeyError(pgErr *pq.Error) *routes.HttpError {
	detail := pgErr.Detail
	before, _, _ := strings.Cut(detail, ")=(")
	before = strings.ReplaceAll(before, "_", " ") // Replace all underscores in column names with spaces

	columns := strings.Split(before[5:], ", ") // first 5 characters of "before" is excluded because it is a "Key ("

	message := `The provided %s is invalid`
	subMessage := ""

	if len(columns) == 1 {
		subMessage = columns[0]
	} else {
		for i, column := range columns {
			isLastColumn := i == len(columns)-1
			if isLastColumn {
				subMessage = subMessage + column + " combination"
			} else {
				subMessage = subMessage + column + "-"
			}
		}
	}

	message = fmt.Sprintf(message, subMessage)

	return &routes.HttpError{
		Status:  http.StatusBadRequest,
		Message: message,
		Code:    "INVALID-FOREIGN-KEY-ERROR",
	}
}
