package errors

import (
	"fmt"
)

// TODO add Status = default value as an argument in the constructor functions

type InternalError struct {
	Status int
	ErrorStack string
}

func (i *InternalError) Error() string {
	return i.ErrorStack
}

func NewInternalError(ErrorStack string) *InternalError {
	return &InternalError{
		Status: 500, 
		ErrorStack: ErrorStack,
	}
}

type InputValidationError struct {
	Status int
	ValidationErrors map[string]string
}

func (err *InputValidationError) Error() string {
	message := "There are one or more errors with your inputs:\n"
	for _, errorMessage := range err.ValidationErrors {
		message = message + errorMessage + "\n"
	}
	return message
}

func NewInputValidationError(ValidationErrors map[string]string) *InputValidationError {
	return &InputValidationError{
		Status: 400,
		ValidationErrors: ValidationErrors,
	}
}

type UniqueViolationError struct {
	Status int
	Entity string
	DuplicateAttributeValuePairs [][2]string
}

func (err *UniqueViolationError) Error() string {
	message := `A %s with the %s already exists`

	subMessage := ""
	for i, pair := range err.DuplicateAttributeValuePairs {
		attribute := pair[0]
		value := pair[1]
		
		if i != len(err.DuplicateAttributeValuePairs) - 1 {
			subMessage = subMessage + fmt.Sprintf(`%s "%s"`, attribute, value) + ", "
		} else {
			subMessage = subMessage + "and" + fmt.Sprintf(`%s "%s"`, attribute, value)
		}
	}

	return fmt.Sprintf(message, err.Entity, subMessage)
}

func NewUniqueViolationError(Entity string, DuplicateAttributeValuePairs [][2]string) *UniqueViolationError {
	return &UniqueViolationError{
		Status: 409,
		Entity: Entity,
		DuplicateAttributeValuePairs: DuplicateAttributeValuePairs,
	}
}

type InvalidForeignKeyError struct {
	Status int
	ForeignKeyAttributeValuePairs [][2]string
}

func (err *InvalidForeignKeyError) Error() string {
	message := `%s is an invalid %s`

	providedCombination := ""
	foreignKeyCombination := ""
	for i, pair := range err.ForeignKeyAttributeValuePairs {
		attribute := pair[0]
		value := pair[1]
		
		if i != len(err.ForeignKeyAttributeValuePairs) - 1 {
			providedCombination = providedCombination + value + "-"
			foreignKeyCombination = foreignKeyCombination + attribute + "-"
		} else {
			providedCombination = providedCombination + value
			foreignKeyCombination = foreignKeyCombination + attribute
		}
	}

	if len(err.ForeignKeyAttributeValuePairs) > 1 {
		foreignKeyCombination = foreignKeyCombination + " combination"
	}

	return fmt.Sprintf(message, providedCombination, foreignKeyCombination)	
}

func NewInvalidForeignKeyError(ForeignKeyAttributeValuePairs [][2]string) *InvalidForeignKeyError{
	return &InvalidForeignKeyError{
		Status: 400, 
		ForeignKeyAttributeValuePairs: ForeignKeyAttributeValuePairs,
	}
}