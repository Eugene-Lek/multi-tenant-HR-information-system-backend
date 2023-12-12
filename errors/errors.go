package errors

import (
	"fmt"
)

type ClientError struct {
	Code string
	Message string
}

func (c *ClientError) Error() string {
	return fmt.Sprintf("%s: %s", c.Code, c.Message)
}

type InternalError struct {
	ErrorStack string
}

func (i *InternalError) Error() string {
	return i.ErrorStack
}

const (
	SingleAttributeUniqueViolation = `A %s with the %s "%s" already exists.`
	MultiAttributeUniqueViolation = `A %s with the %s already exists.`
	InsertForeignKeyViolation = `%s is an invalid %s.`
)