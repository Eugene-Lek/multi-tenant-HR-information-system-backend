package errors

import (
	"fmt"
)

// TODO add Status = default value as an argument in the constructor functions

type HttpError interface {
	Error() string
	Status() int
}


type InternalServerError struct {
	status int
	errorStack string
}

func (i *InternalServerError) Error() string {
	return i.errorStack
}

func (i *InternalServerError) Status() int {
	return i.status
}

func NewInternalServerError(errorStack string) *InternalServerError {
	return &InternalServerError{
		status: 500, 
		errorStack: errorStack,
	}
}

type InputValidationError struct {
	status int
	validationErrors map[string]string
}

func (err *InputValidationError) Error() string {
	message := "There are one or more errors with your inputs:\n"
	for _, errorMessage := range err.validationErrors {
		message = message + errorMessage + "\n"
	}
	return message
}

func (err *InputValidationError) Status() int {
	return err.status
}

func NewInputValidationError(validationErrors map[string]string) *InputValidationError {
	return &InputValidationError{
		status: 400,
		validationErrors: validationErrors,
	}
}

type UniqueViolationError struct {
	status int
	entity string
	duplicateAttributeValuePairs [][2]string
}

func (err *UniqueViolationError) Error() string {
	message := `A %s with the %s already exists`

	subMessage := ""
	for i, pair := range err.duplicateAttributeValuePairs {
		attribute := pair[0]
		value := pair[1]
		
		if i != len(err.duplicateAttributeValuePairs) - 1 {
			subMessage = subMessage + fmt.Sprintf(`%s "%s"`, attribute, value) + ", "
		} else {
			subMessage = subMessage + "and" + fmt.Sprintf(`%s "%s"`, attribute, value)
		}
	}

	return fmt.Sprintf(message, err.entity, subMessage)
}

func (err *UniqueViolationError) Status() int {
	return err.status
}

func NewUniqueViolationError(entity string, duplicateAttributeValuePairs [][2]string) *UniqueViolationError {
	return &UniqueViolationError{
		status: 409,
		entity: entity,
		duplicateAttributeValuePairs: duplicateAttributeValuePairs,
	}
}

type InvalidForeignKeyError struct {
	status int
	foreignKeyAttributeValuePairs [][2]string
}

func (err *InvalidForeignKeyError) Error() string {
	message := `%s is an invalid %s`

	providedCombination := ""
	foreignKeyCombination := ""
	for i, pair := range err.foreignKeyAttributeValuePairs {
		attribute := pair[0]
		value := pair[1]
		
		if i != len(err.foreignKeyAttributeValuePairs) - 1 {
			providedCombination = providedCombination + value + "-"
			foreignKeyCombination = foreignKeyCombination + attribute + "-"
		} else {
			providedCombination = providedCombination + value
			foreignKeyCombination = foreignKeyCombination + attribute
		}
	}

	if len(err.foreignKeyAttributeValuePairs) > 1 {
		foreignKeyCombination = foreignKeyCombination + " combination"
	}

	return fmt.Sprintf(message, providedCombination, foreignKeyCombination)	
}

func (err *InvalidForeignKeyError) Status() int {
	return err.status
}

func NewInvalidForeignKeyError(foreignKeyAttributeValuePairs [][2]string) *InvalidForeignKeyError{
	return &InvalidForeignKeyError{
		status: 400, 
		foreignKeyAttributeValuePairs: foreignKeyAttributeValuePairs,
	}
}