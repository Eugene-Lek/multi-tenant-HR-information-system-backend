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
//  1. Verify the accuracy of custom validations
//
// The accuracy of backed in validations are not tested because that would amount to testing someone else's library
func TestIsIsoDate(t *testing.T) {
	type Test struct {
		Date string `validate:"isIsoDate"`
	}

	tests := []validationTestCase{
		{"Date should be valid because 2024-02-29 is a valid date", Test{"2024-02-29"}, []fieldTagPair{}},
		{"Date should be invalid because 2023-02-29 is an invalid date", Test{"2023-02-29"}, []fieldTagPair{{"Date", "isIsoDate"}}},
		{"Date should be invalid because 13 is an invalid month", Test{"2024-13-01"}, []fieldTagPair{{"Date", "isIsoDate"}}},
		{"Date should be invalid because preceeding 0 is missing", Test{"2024-2-29"}, []fieldTagPair{{"Date", "isIsoDate"}}},
		{"Date should be invalid because dashes are missing", Test{"2024/02/01"}, []fieldTagPair{{"Date", "isIsoDate"}}},
		{"Date should be invalid because year is in YY format", Test{"24-02-01"}, []fieldTagPair{{"Date", "isIsoDate"}}},
	}

	runValidationTest(t, tests)
}

func TestValidPositionAssignmentDuration(t *testing.T) {
	type Test struct {	
		StartDate string
		EndDate string 	 `validate:"validPositionAssignmentDuration"`		
	}	

	tests := []validationTestCase{
		{"End date should be valid", Test{"2024-02-01", "2024-04-29"}, []fieldTagPair{}},
		{"End date should be valid because the duration is exactly 30 days", Test{"2024-04-01", "2024-04-30"}, []fieldTagPair{}},
		{"End date should be invalid because the duration is less than 30 days", Test{"2024-02-01", "2024-02-28"}, []fieldTagPair{{"EndDate", "validPositionAssignmentDuration"}}},
		{"End date should be invalid because it comes before the start date", Test{"2024-04-01", "2024-01-01"}, []fieldTagPair{{"EndDate", "validPositionAssignmentDuration"}}},
	}

	runValidationTest(t, tests)
}
