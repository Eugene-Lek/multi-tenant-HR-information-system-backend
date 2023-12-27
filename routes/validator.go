package routes

import (
	"fmt"
	"reflect"
	"time"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	validators "github.com/go-playground/validator/v10/non-standard/validators" // convention is for aliases to be 1 word long
	validatortranslations "github.com/go-playground/validator/v10/translations/en"
)

// IMPORTANT NOTE:
// The validator stops at the first validation failure of every field.
// Therefore, you should order your validation tags from broadest (e.g. required) to most specific (e.g. notBlank, then isIsoDate)
// In other words, every subsequent validation tag must test a subset of the "valid" scenarios of the previous tag

func NewUniversalTranslator() *ut.UniversalTranslator {
	enTranslator := en.New()
	return ut.New(enTranslator, enTranslator)
}

func NewValidator(universalTranslator *ut.UniversalTranslator) (*validator.Validate, error) {
	// Fetch all valid translators
	englishTranslator, _ := universalTranslator.GetTranslator("en")

	// Initialise a validator with default validation checks
	validate := validator.New(validator.WithRequiredStructEnabled())

	// Attach the default error message templates & translation functions to the validator
	err := validatortranslations.RegisterDefaultTranslations(validate, englishTranslator)
	if err != nil {
		return nil, err
	}

	// Add custom validation checks & their corresponding error message templates & translation functions
	err = validate.RegisterValidation("notBlank", validators.NotBlank)
	if err != nil {
		return nil, err
	}

	err = validate.RegisterTranslation("notBlank", englishTranslator, registerNotBlankTranslations, executeNotBlankTranslations)
	if err != nil {
		return nil, err
	}

	err = validate.RegisterValidation("isIsoDate", isIsoDate)
	if err != nil {
		return nil, err
	}

	err = validate.RegisterTranslation("isIsoDate", englishTranslator, registerIsIsoDateTranslations, executeIsIsoDateTranslations)
	if err != nil {
		return nil, err
	}

	err = validate.RegisterValidation("validAppointmentDuration", validAppointmentDuration)
	if err != nil {
		return nil, err
	}

	err = validate.RegisterTranslation("validAppointmentDuration", englishTranslator, registerValidAppointmentDurationTranslations, executeValidAppointmentDurationTranslations)
	if err != nil {
		return nil, err
	}

	// Add a tag name function so that way the validator can use the struct tag names in its error messages instead
	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		return field.Tag.Get("name")
	})

	return validate, nil
}

// Validates a struct instance, translates the errors to error messages and returns an error that collates all the error messages
func validateStruct(validate *validator.Validate, translator ut.Translator, s interface{}) error {
	err := validate.Struct(s)
	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		errorMessages := validationErrors.Translate(translator)
		return NewInputValidationError(errorMessages)
	}

	return nil
}

func registerNotBlankTranslations(translator ut.Translator) error {
	if err := translator.Add("notBlank-string", "The {0} cannot be blank", false); err != nil {
		return err
	}
	if err := translator.Add("notBlank-items", "You did not provide any {0}", false); err != nil {
		return err
	}
	if err := translator.Add("notBlank-exist", `You must provide a {0}`, false); err != nil {
		return err
	}
	if err := translator.Add("notBlank-valid", `You provided an invalid {0}`, false); err != nil {
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
		message = "notBlank translation failed"
	}

	return message
}

func isIsoDate(fl validator.FieldLevel) bool {
	field := fl.Field()

	switch field.Kind() {
	case reflect.String:
		layout := "2006-01-02"
		_, err := time.Parse(layout, field.String())
		return err == nil
	default:
		return false
	}
}

func registerIsIsoDateTranslations(translator ut.Translator) error {
	err := translator.Add("isIsoDate", `The {0} must follow the "yyyy-mm-dd" format`, false)
	return err
}

func executeIsIsoDateTranslations(translator ut.Translator, fieldError validator.FieldError) string {
	msg, err := translator.T("isIsoDate", fieldError.Field())
	if err != nil {
		msg = "isIsoDate translation failed"
	}

	return msg
}

const minimumAppointmentDurationDays = 30

// Returns true if the start & end dates are valid ISO dates & are at least 30 days apart. Otherwise, returns false.
func validAppointmentDuration(fl validator.FieldLevel) bool {
	const minimumAppointmentDuration = minimumAppointmentDurationDays * 24 * 3600 // secords

	// Conversion to interface is a necessary intermediate step for conversion to the Parent's concrete type
	entity := fl.Parent().Interface().(Appointment)
	startDate := entity.StartDate
	endDate := entity.EndDate

	// If endDate is not provided, it defaults to 9999-12-31
	if endDate == "" {
		return true
	}

	layout := "2006-01-02"
	startDateTime, err := time.Parse(layout, startDate)
	if err != nil {
		return false
	}
	endDateTime, err := time.Parse(layout, endDate)
	if err != nil {
		return false
	}

	if endDateTime.Unix()-startDateTime.Unix() <= minimumAppointmentDuration {
		return false
	}

	return true
}

func registerValidAppointmentDurationTranslations(translator ut.Translator) error {
	message := fmt.Sprintf("The end date must be at least %v days after the start date", minimumAppointmentDurationDays)
	err := translator.Add("validAppointmentDuration", message, false)
	return err
}

func executeValidAppointmentDurationTranslations(translator ut.Translator, fieldError validator.FieldError) string {
	msg, err := translator.T("validAppointmentDuration")
	if err != nil {
		msg = "validAppointmentDuration translation failed"
	}

	return msg
}
