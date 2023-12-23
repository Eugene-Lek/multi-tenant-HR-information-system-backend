package main

import (
	"reflect"

	"github.com/go-playground/validator/v10"
	validators "github.com/go-playground/validator/v10/non-standard/validators" // convention is for aliases to be 1 word long
	validatortranslations "github.com/go-playground/validator/v10/translations/en"
	ut "github.com/go-playground/universal-translator"	
)

func newValidator(universalTranslator *ut.UniversalTranslator) (*validator.Validate, error) {
	// Fetch all valid translators
	englishTranslator, _ := universalTranslator.GetTranslator("en")

	// Initialises a validator with default validation checks
	validate := validator.New(validator.WithRequiredStructEnabled())

	// Attaches the default validation error message templates & translation functions to the validator
	err := validatortranslations.RegisterDefaultTranslations(validate, englishTranslator) 
	if err != nil {
		return nil, err
	}	

	// Adds custom validation checks & their corresponding validation error messages
	err = validate.RegisterValidation("notBlank", validators.NotBlank)
	if err != nil {
		return nil, err
	}

	err = validate.RegisterTranslation("notBlank", englishTranslator, registerNotBlankTranslations, executeNotBlankTranslations)
	if err != nil {
		return nil, err
	}

	return validate, nil
}

func registerNotBlankTranslations(translator ut.Translator) error {
	if err := translator.Add("notBlank-string", "{0} cannot be blank", false); err != nil {
		return err
	}
	if err := translator.Add("notBlank-items", "{0} must contain at least 1 item", false); err != nil {
		return err
	}
	if err := translator.Add("notBlank-exist", "{0} is a required field", false); err != nil {
		return err
	}
	if err := translator.Add("notBlank-valid", "You provided an invalid value for the field {0}", false); err != nil {
		return err
	}

	return nil
}

func executeNotBlankTranslations(translator ut.Translator, fieldError validator.FieldError) string {
	var message string
	var err error

	fieldName := fieldError.Field()
	kind := fieldError.Kind()
	switch kind {
	case reflect.String:
		message, err = translator.T("notBlank-string", fieldName)
	case reflect.Slice, reflect.Array, reflect.Chan, reflect.Map:
		message, err = translator.T("notBlank-items", fieldName)
	case reflect.Interface, reflect.Func, reflect.Ptr:
		message, err = translator.T("notBlank-exist", fieldName)
	default:
		message, err = translator.T("notBlank-valid", fieldName)
	}

	if err != nil {
		// TODO return error message
	}

	return message
}