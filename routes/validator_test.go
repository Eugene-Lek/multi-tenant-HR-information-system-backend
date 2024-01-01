package routes

import (
	"reflect"
	"slices"
	"testing"

	"github.com/go-playground/validator/v10"
)

// Common definitions
type fieldTagPair struct {
	field string
	tag   string
}

type validationTestCase struct {
	name  string
	input any
	want  []fieldTagPair
}

func runValidationTest(t *testing.T, tests []validationTestCase) {
	ut := NewUniversalTranslator()
	validate, err := NewValidator(ut)
	if err != nil {
		t.Fatalf("Validator failed to instantiate: %s", err.Error())
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := validate.Struct(test.input)
			validationErrors, _ := result.(validator.ValidationErrors)

			relevantFieldTagPairs := []fieldTagPair{} // i.e. field-tag pairs that the test is looking for
			for _, fieldError := range validationErrors {
				relevant := slices.Contains(test.want, fieldTagPair{fieldError.Field(), fieldError.Tag()})
				if relevant {
					relevantFieldTagPairs = append(relevantFieldTagPairs, fieldTagPair{fieldError.Field(), fieldError.Tag()})
				}
			}

			if !reflect.DeepEqual(test.want, relevantFieldTagPairs) {
				t.Errorf("want: %s, got: %s", test.want, relevantFieldTagPairs)
			}
		})
	}
}

// Input validation tests
// Purposes:
//  1. Verify that all expected validation tags have been included in each struct
//  2. Verify the accuracy of custom validations
//
// The accuracy of backed in validations are not tested because that would amount to testing someone else's library
func TestTenantValidation(t *testing.T) {
	var tests = []validationTestCase{
		{"Tenant should be valid", Tenant{"tenant", "", ""}, []fieldTagPair{}},
		{"Tenant should be invalid because name is missing", Tenant{}, []fieldTagPair{{"tenant name", "required"}}},
		{"Tenant should be invalid because name is blank", Tenant{"  ", "", ""}, []fieldTagPair{{"tenant name", "notBlank"}}},
	}

	runValidationTest(t, tests)
}

func TestDivisionValidation(t *testing.T) {
	tests := []validationTestCase{
		{"Division should be valid", Division{"division", "tenant", "", ""}, []fieldTagPair{}},
		{"Division should be invalid because name is missing", Division{"", "tenant", "", ""}, []fieldTagPair{{"division name", "required"}}},
		{"Division should be invalid because tenant is missing", Division{"division", "", "", ""}, []fieldTagPair{{"tenant name", "required"}}},
		{"Division should be invalid because name is blank", Division{"   ", "tenant", "", ""}, []fieldTagPair{{"division name", "notBlank"}}},
		{"Division should be invalid because tenant is blank", Division{"division", "   ", "", ""}, []fieldTagPair{{"tenant name", "notBlank"}}},
	}

	runValidationTest(t, tests)
}

func TestDepartmentValidation(t *testing.T) {
	tests := []validationTestCase{
		{"Department should be valid", Department{"department", "tenant", "division", "", ""}, []fieldTagPair{}},
		{"Department should be invalid because name is missing", Department{"", "tenant", "division", "", ""}, []fieldTagPair{{"department name", "required"}}},
		{"Department should be invalid because tenant is missing", Department{"department", "", "division", "", ""}, []fieldTagPair{{"tenant name", "required"}}},
		{"Department should be invalid because tenant is missing", Department{"department", "tenant", "", "", ""}, []fieldTagPair{{"division name", "required"}}},
		{"Department should be invalid because name is blank", Department{"   ", "tenant", "division", "", ""}, []fieldTagPair{{"department name", "notBlank"}}},
		{"Department should be invalid because tenant is blank", Department{"department", "   ", "division", "", ""}, []fieldTagPair{{"tenant name", "notBlank"}}},
		{"Department should be invalid because tenant is blank", Department{"department", "tenant", "   ", "", ""}, []fieldTagPair{{"division name", "notBlank"}}},
	}

	runValidationTest(t, tests)
}

func TestUserValidation(t *testing.T) {
	tests := []validationTestCase{
		{"User should be valid", User{"a62c9359-64cd-4e05-bed4-27b361f882b6", "user@gmail.com", "tenant", "", "", "", "", ""}, []fieldTagPair{}},
		{"User should be invalid because ID is missing", User{"", "user@gmail.com", "tenant", "", "", "", "", ""}, []fieldTagPair{{"user id", "required"}}},
		{"User should be invalid because email is missing", User{"a62c9359-64cd-4e05-bed4-27b361f882b6", "", "tenant", "", "", "", "", ""}, []fieldTagPair{{"user email", "required"}}},
		{"User should be invalid because tenant is missing", User{"a62c9359-64cd-4e05-bed4-27b361f882b6", "user@gmail.com", "", "", "", "", "", ""}, []fieldTagPair{{"tenant name", "required"}}},
		{"User should be invalid because ID is blank", User{"   ", "user@gmail.com", "tenant", "", "", "", "", ""}, []fieldTagPair{{"user id", "notBlank"}}},
		{"User should be invalid because email is blank", User{"a62c9359-64cd-4e05-bed4-27b361f882b6", "   ", "tenant", "", "", "", "", ""}, []fieldTagPair{{"user email", "notBlank"}}},
		{"User should be invalid because tenant is blank", User{"a62c9359-64cd-4e05-bed4-27b361f882b6", "user@gmail.com", "   ", "", "", "", "", ""}, []fieldTagPair{{"tenant name", "notBlank"}}},
		{"User should be invalid because ID is an invalid uuid", User{"a62c9359364cd-4e05-bed4-27b361f882b6", "user@gmail.com", "tenant", "", "", "", "", ""}, []fieldTagPair{{"user id", "uuid"}}},
		{"User should be invalid because email is invalid", User{"a62c9359-64cd-4e05-bed4-27b361f882b6", "user@.com", "tenant", "", "", "", "", ""}, []fieldTagPair{{"user email", "email"}}},
	}

	runValidationTest(t, tests)
}

func TestIsIsoDate(t *testing.T) {
	tests := []validationTestCase{
		{"Start date should be valid because 2024-02-29 is a valid date", Appointment{"title", "tenant", "division", "department", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2024-02-29", "2024-06-21", "", ""}, []fieldTagPair{}},
		{"Start date should be invalid because 2023-02-29 is an invalid date", Appointment{"title", "tenant", "division", "department", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023-02-29", "2024-06-21", "", ""}, []fieldTagPair{{"start date", "isIsoDate"}}},
		{"Start date should be invalid because 13 is an invalid month", Appointment{"title", "tenant", "division", "department", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023-13-01", "2024-06-21", "", ""}, []fieldTagPair{{"start date", "isIsoDate"}}},
		{"Start date should be invalid because preceeding 0 is missing", Appointment{"title", "tenant", "division", "department", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023-2-01", "2024-06-21", "", ""}, []fieldTagPair{{"start date", "isIsoDate"}}},
		{"Start date should be invalid because dashes are missing", Appointment{"title", "tenant", "division", "department", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023/02/01", "2024-06-21", "", ""}, []fieldTagPair{{"start date", "isIsoDate"}}},
		{"Start date should be invalid because year is in YY format", Appointment{"title", "tenant", "division", "department", "a62c9359-64cd-4e05-bed4-27b361f882b6", "23-02-01", "2024-06-21", "", ""}, []fieldTagPair{{"start date", "isIsoDate"}}},
	}

	runValidationTest(t, tests)
}

func TestValidAppointmentDuration(t *testing.T) {
	tests := []validationTestCase{
		{"End date should be valid", Appointment{"title", "tenant", "division", "department", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2024-02-29", "2024-06-21", "", ""}, []fieldTagPair{}},
		{"End date should be valid because the duration is exactly 30 days", Appointment{"title", "tenant", "division", "department", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2024-02-29", "2024-06-21", "", ""}, []fieldTagPair{}},
		{"End date should be invalid because the duration is less than 30 days", Appointment{"title", "tenant", "division", "department", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2024-06-01", "2024-06-21", "", ""}, []fieldTagPair{{"end date", "validAppointmentDuration"}}},
		{"End date should be invalid because it comes before the start date", Appointment{"title", "tenant", "division", "department", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2024-07-01", "2024-06-21", "", ""}, []fieldTagPair{{"end date", "validAppointmentDuration"}}},
	}

	runValidationTest(t, tests)
}

func TestApppointmentValidation(t *testing.T) {
	tests := []validationTestCase{
		{"Appointment should be valid", Appointment{"title", "tenant", "division", "department", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023-12-27", "2024-06-21", "", ""}, []fieldTagPair{}},
		{"Appointment should be valid because end date is optional", Appointment{"title", "tenant", "division", "department", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023-12-27", "", "", ""}, []fieldTagPair{}},
		{"Appointment should be invalid because title is missing", Appointment{"", "tenant", "division", "department", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023-12-27", "2024-06-21", "", ""}, []fieldTagPair{{"appointment title", "required"}}},
		{"Appointment should be invalid because tenant is missing", Appointment{"title", "", "division", "department", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023-12-27", "2024-06-21", "", ""}, []fieldTagPair{{"tenant name", "required"}}},
		{"Appointment should be invalid because division is missing", Appointment{"title", "tenant", "", "department", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023-12-27", "2024-06-21", "", ""}, []fieldTagPair{{"division name", "required"}}},
		{"Appointment should be invalid because department is missing", Appointment{"title", "tenant", "division", "", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023-12-27", "2024-06-21", "", ""}, []fieldTagPair{{"department name", "required"}}},
		{"Appointment should be invalid because user id is missing", Appointment{"title", "tenant", "division", "department", "", "2023-12-27", "2024-06-21", "", ""}, []fieldTagPair{{"user id", "required"}}},
		{"Appointment should be invalid because start date is missing", Appointment{"title", "tenant", "division", "department", "a62c9359-64cd-4e05-bed4-27b361f882b6", "", "2024-06-21", "", ""}, []fieldTagPair{{"start date", "required"}}},
		{"Appointment should be invalid because title is blank", Appointment{"  ", "tenant", "division", "department", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023-12-27", "2024-06-21", "", ""}, []fieldTagPair{{"appointment title", "notBlank"}}},
		{"Appointment should be invalid because tenant is blank", Appointment{"title", "  ", "division", "department", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023-12-27", "2024-06-21", "", ""}, []fieldTagPair{{"tenant name", "notBlank"}}},
		{"Appointment should be invalid because division is blank", Appointment{"title", "tenant", "  ", "department", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023-12-27", "2024-06-21", "", ""}, []fieldTagPair{{"division name", "notBlank"}}},
		{"Appointment should be invalid because department is blank", Appointment{"title", "tenant", "division", "  ", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023-12-27", "2024-06-21", "", ""}, []fieldTagPair{{"department name", "notBlank"}}},
		{"Appointment should be invalid because user id is blank", Appointment{"title", "tenant", "division", "department", "  ", "2023-12-27", "2024-06-21", "", ""}, []fieldTagPair{{"user id", "notBlank"}}},
		{"Appointment should be invalid because start date is blank", Appointment{"title", "tenant", "division", "department", "a62c9359-64cd-4e05-bed4-27b361f882b6", "  ", "2024-06-21", "", ""}, []fieldTagPair{{"start date", "notBlank"}}},
		{"Appointment should be invalid because end date is blank", Appointment{"title", "tenant", "division", "department", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2024-06-21", "   ", "", ""}, []fieldTagPair{{"end date", "notBlank"}}},
		{"Appointment should be invalid because user id is an invalid uuid", Appointment{"title", "tenant", "division", "department", "a62c9359d64cd-4e05-bed4-27b361f882b6", "2023-12-27", "2024-06-21", "", ""}, []fieldTagPair{{"user id", "uuid"}}},
		{"Appointment should be invalid because start date is not in ISO format", Appointment{"title", "tenant", "division", "department", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023-13-13", "2024-06-21", "", ""}, []fieldTagPair{{"start date", "isIsoDate"}}},
		{"Appointment should be invalid because end date is not in ISO format", Appointment{"title", "tenant", "division", "department", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023-06-21", "2024-13-13", "", ""}, []fieldTagPair{{"end date", "isIsoDate"}}},
		{"Appointment should be invalid because end date is less than 30 days after start date", Appointment{"title", "tenant", "division", "department", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023-06-21", "2023-07-01", "", ""}, []fieldTagPair{{"end date", "validAppointmentDuration"}}},
	}

	runValidationTest(t, tests)
}
