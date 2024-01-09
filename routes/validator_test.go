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
		{"Tenant should be valid", Tenant{"796d707b-6f0a-4004-be01-f4d63b6866de", "tenant", "", ""}, []fieldTagPair{}},
		{"Tenant should be invalid because id is missing", Tenant{Name: "tenant"}, []fieldTagPair{{"tenant id", "required"}}},
		{"Tenant should be invalid because id is blank", Tenant{"  ", "tenant", "", ""}, []fieldTagPair{{"tenant id", "notBlank"}}},
		{"Tenant should be invalid because name is missing", Tenant{Id: "796d707b-6f0a-4004-be01-f4d63b6866de"}, []fieldTagPair{{"tenant name", "required"}}},
		{"Tenant should be invalid because name is blank", Tenant{"796d707b-6f0a-4004-be01-f4d63b6866de", "  ", "", ""}, []fieldTagPair{{"tenant name", "notBlank"}}},		
		{"Tenant should be invalid because ID is an invalid uuid", Tenant{"a62c9359364cd-4e05-bed4-27b361f882b6", "tenant", "", ""}, []fieldTagPair{{"tenant id", "uuid"}}},		
	}

	runValidationTest(t, tests)
}

func TestDivisionValidation(t *testing.T) {
	tests := []validationTestCase{
		{"Division should be valid", Division{"152e1276-31de-4c08-878f-ac46952ea3c1", "9d8269e6-d0fc-4bc3-a09b-159ef293a363", "division", "", ""}, []fieldTagPair{}},
		{"Division should be invalid because id is missing", Division{"", "9d8269e6-d0fc-4bc3-a09b-159ef293a363", "division", "", ""}, []fieldTagPair{{"division id", "required"}}},
		{"Division should be invalid because name is missing", Division{"152e1276-31de-4c08-878f-ac46952ea3c1", "9d8269e6-d0fc-4bc3-a09b-159ef293a363", "", "", ""}, []fieldTagPair{{"division name", "required"}}},
		{"Division should be invalid because tenant id is missing", Division{"152e1276-31de-4c08-878f-ac46952ea3c1", "", "division", "", ""}, []fieldTagPair{{"tenant id", "required"}}},
		{"Division should be invalid because id is blank", Division{"   ", "9d8269e6-d0fc-4bc3-a09b-159ef293a363", "division", "", ""}, []fieldTagPair{{"division id", "notBlank"}}},				
		{"Division should be invalid because name is blank", Division{"152e1276-31de-4c08-878f-ac46952ea3c1", "9d8269e6-d0fc-4bc3-a09b-159ef293a363", "   ", "", ""}, []fieldTagPair{{"division name", "notBlank"}}},
		{"Division should be invalid because tenant id is blank", Division{"152e1276-31de-4c08-878f-ac46952ea3c1", "   ", "division", "", ""}, []fieldTagPair{{"tenant id", "notBlank"}}},
		{"Division should be invalid because id is an invalid uuid", Division{"a62c9359364cd-4e05-bed4-27b361f882b6", "9d8269e6-d0fc-4bc3-a09b-159ef293a363", "division", "", ""}, []fieldTagPair{{"division id", "uuid"}}},		
		{"Division should be invalid because tenant id is an invalid uuid", Division{"152e1276-31de-4c08-878f-ac46952ea3c1", "9d8269e6-d0fc4bc3-a09b-159ef293a363", "division", "", ""}, []fieldTagPair{{"tenant id", "uuid"}}},
	}

	runValidationTest(t, tests)
}

func TestDepartmentValidation(t *testing.T) {
	tests := []validationTestCase{
		{"Department should be valid", Department{"c40a529e-5624-4e4e-abce-9d112bf4d2b2", "796d707b-6f0a-4004-be01-f4d63b6866de", "e010029f-e146-406f-93dc-7da45cf86078", "department", "", ""}, []fieldTagPair{}},
		{"Department should be invalid because id is missing", Department{"", "796d707b-6f0a-4004-be01-f4d63b6866de", "e010029f-e146-406f-93dc-7da45cf86078", "department", "", ""}, []fieldTagPair{{"department id", "required"}}},		
		{"Division should be invalid because tenant id is missing", Department{"c40a529e-5624-4e4e-abce-9d112bf4d2b2", "", "e010029f-e146-406f-93dc-7da45cf86078", "department", "", ""}, []fieldTagPair{{"tenant id", "required"}}},				
		{"Department should be invalid because division id is missing", Department{"c40a529e-5624-4e4e-abce-9d112bf4d2b2", "796d707b-6f0a-4004-be01-f4d63b6866de", "", "department", "", ""}, []fieldTagPair{{"division id", "required"}}},		
		{"Department should be invalid because name is missing", Department{"c40a529e-5624-4e4e-abce-9d112bf4d2b2", "796d707b-6f0a-4004-be01-f4d63b6866de", "e010029f-e146-406f-93dc-7da45cf86078", "", "", ""}, []fieldTagPair{{"department name", "required"}}},
		{"Department should be invalid because id is blank", Department{"   ", "796d707b-6f0a-4004-be01-f4d63b6866de", "e010029f-e146-406f-93dc-7da45cf86078", "department", "", ""}, []fieldTagPair{{"department id", "notBlank"}}},		
		{"Division should be invalid because tenant id is blank", Department{"c40a529e-5624-4e4e-abce-9d112bf4d2b2", "   ", "e010029f-e146-406f-93dc-7da45cf86078", "department", "", ""}, []fieldTagPair{{"tenant id", "notBlank"}}},				
		{"Department should be invalid because division id is blank", Department{"c40a529e-5624-4e4e-abce-9d112bf4d2b2", "796d707b-6f0a-4004-be01-f4d63b6866de", "   ", "department", "", ""}, []fieldTagPair{{"division id", "notBlank"}}},		
		{"Department should be invalid because name is blank", Department{"c40a529e-5624-4e4e-abce-9d112bf4d2b2", "796d707b-6f0a-4004-be01-f4d63b6866de", "e010029f-e146-406f-93dc-7da45cf86078", "   ", "", ""}, []fieldTagPair{{"department name", "notBlank"}}},
		{"Department should be invalid because id is an invalid uuid", Department{"a62c9359364cd-4e05-bed4-27b361f882b6", "796d707b-6f0a-4004-be01-f4d63b6866de", "9d8269e6-d0fc-4bc3-a09b-159ef293a363", "department", "", ""}, []fieldTagPair{{"department id", "uuid"}}},				
		{"Division should be invalid because tenant id is an invalid uuid", Department{"c40a529e-5624-4e4e-abce-9d112bf4d2b2", "796d707b6f0a-4004-be01-f4d63b6866de", "e010029f-e146-406f-93dc-7da45cf86078", "department", "", ""}, []fieldTagPair{{"tenant id", "uuid"}}},		
		{"Department should be invalid because division id is an invalid uuid", Department{"c40a529e-5624-4e4e-abce-9d112bf4d2b2", "796d707b-6f0a-4004-be01-f4d63b6866de", "9d8269e6-d0fc4bc3-a09b-159ef293a363", "department", "", ""}, []fieldTagPair{{"division id", "uuid"}}},						
	}

	runValidationTest(t, tests)
}

func TestUserValidation(t *testing.T) {
	tests := []validationTestCase{
		{"User should be valid", User{"a62c9359-64cd-4e05-bed4-27b361f882b6", "a4239332-9c7c-429b-877b-1b58e411c29d", "user@gmail.com", "", "", "", "", ""}, []fieldTagPair{}},
		{"User should be invalid because ID is missing", User{"", "a4239332-9c7c-429b-877b-1b58e411c29d", "user@gmail.com", "", "", "", "", ""}, []fieldTagPair{{"user id", "required"}}},
		{"User should be invalid because email is missing", User{"a62c9359-64cd-4e05-bed4-27b361f882b6", "a4239332-9c7c-429b-877b-1b58e411c29d", "", "", "", "", "", ""}, []fieldTagPair{{"user email", "required"}}},
		{"User should be invalid because tenant id is missing", User{"a62c9359-64cd-4e05-bed4-27b361f882b6", "", "user@gmail.com", "", "", "", "", ""}, []fieldTagPair{{"tenant id", "required"}}},
		{"User should be invalid because ID is blank", User{"   ", "a4239332-9c7c-429b-877b-1b58e411c29d", "user@gmail.com", "", "", "", "", ""}, []fieldTagPair{{"user id", "notBlank"}}},
		{"User should be invalid because email is blank", User{"a62c9359-64cd-4e05-bed4-27b361f882b6", "a4239332-9c7c-429b-877b-1b58e411c29d", "   ", "", "", "", "", ""}, []fieldTagPair{{"user email", "notBlank"}}},
		{"User should be invalid because tenant id is blank", User{"a62c9359-64cd-4e05-bed4-27b361f882b6", "   ", "user@gmail.com", "", "", "", "", ""}, []fieldTagPair{{"tenant id", "notBlank"}}},
		{"User should be invalid because ID is an invalid uuid", User{"a62c9359364cd-4e05-bed4-27b361f882b6", "a4239332-9c7c-429b-877b-1b58e411c29d", "user@gmail.com", "", "", "", "", ""}, []fieldTagPair{{"user id", "uuid"}}},
		{"User should be invalid because email is invalid", User{"a62c9359-64cd-4e05-bed4-27b361f882b6", "a4239332-9c7c-429b-877b-1b58e411c29d", "user@.com", "", "", "", "", ""}, []fieldTagPair{{"user email", "email"}}},
		{"User should be invalid because tenant id is an invalid uuid", User{"a62c9359-64cd-4e05-bed4-27b361f882b6", "a4239332-9c7c429b-877b-1b58e411c29d", "user@gmail.com", "", "", "", "", ""}, []fieldTagPair{{"tenant id", "uuid"}}},
	}

	runValidationTest(t, tests)
}

func TestIsIsoDate(t *testing.T) {
	tests := []validationTestCase{
		{"Start date should be valid because 2024-02-29 is a valid date", Appointment{"1cc45a08-58ce-4583-a1b5-f6d2d036ee17", "796d707b-6f0a-4004-be01-f4d63b6866de", "title", "43e52216-bfff-440b-b1d8-6c3c7a29dbbb", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2024-02-29", "2024-06-21", "", ""}, []fieldTagPair{}},
		{"Start date should be invalid because 2023-02-29 is an invalid date", Appointment{"1cc45a08-58ce-4583-a1b5-f6d2d036ee17", "796d707b-6f0a-4004-be01-f4d63b6866de", "title", "43e52216-bfff-440b-b1d8-6c3c7a29dbbb", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023-02-29", "2024-06-21", "", ""}, []fieldTagPair{{"start date", "isIsoDate"}}},
		{"Start date should be invalid because 13 is an invalid month", Appointment{"1cc45a08-58ce-4583-a1b5-f6d2d036ee17", "796d707b-6f0a-4004-be01-f4d63b6866de", "title", "43e52216-bfff-440b-b1d8-6c3c7a29dbbb", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023-13-01", "2024-06-21", "", ""}, []fieldTagPair{{"start date", "isIsoDate"}}},
		{"Start date should be invalid because preceeding 0 is missing", Appointment{"1cc45a08-58ce-4583-a1b5-f6d2d036ee17", "796d707b-6f0a-4004-be01-f4d63b6866de", "title", "43e52216-bfff-440b-b1d8-6c3c7a29dbbb", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023-2-01", "2024-06-21", "", ""}, []fieldTagPair{{"start date", "isIsoDate"}}},
		{"Start date should be invalid because dashes are missing", Appointment{"1cc45a08-58ce-4583-a1b5-f6d2d036ee17", "796d707b-6f0a-4004-be01-f4d63b6866de", "title", "43e52216-bfff-440b-b1d8-6c3c7a29dbbb", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023/02/01", "2024-06-21", "", ""}, []fieldTagPair{{"start date", "isIsoDate"}}},
		{"Start date should be invalid because year is in YY format", Appointment{"1cc45a08-58ce-4583-a1b5-f6d2d036ee17", "796d707b-6f0a-4004-be01-f4d63b6866de", "title", "43e52216-bfff-440b-b1d8-6c3c7a29dbbb", "a62c9359-64cd-4e05-bed4-27b361f882b6", "23-02-01", "2024-06-21", "", ""}, []fieldTagPair{{"start date", "isIsoDate"}}},
	}

	runValidationTest(t, tests)
}

func TestValidAppointmentDuration(t *testing.T) {
	tests := []validationTestCase{
		{"End date should be valid", Appointment{"3cd59970-3942-448b-af1d-8b99310f1701", "796d707b-6f0a-4004-be01-f4d63b6866de", "title", "43e52216-bfff-440b-b1d8-6c3c7a29dbbb", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2024-02-29", "2024-06-21", "", ""}, []fieldTagPair{}},
		{"End date should be valid because the duration is exactly 30 days", Appointment{"3cd59970-3942-448b-af1d-8b99310f1701", "796d707b-6f0a-4004-be01-f4d63b6866de", "title", "43e52216-bfff-440b-b1d8-6c3c7a29dbbb", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2024-02-29", "2024-06-21", "", ""}, []fieldTagPair{}},
		{"End date should be invalid because the duration is less than 30 days", Appointment{"3cd59970-3942-448b-af1d-8b99310f1701", "796d707b-6f0a-4004-be01-f4d63b6866de", "title", "43e52216-bfff-440b-b1d8-6c3c7a29dbbb", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2024-06-01", "2024-06-21", "", ""}, []fieldTagPair{{"end date", "validAppointmentDuration"}}},
		{"End date should be invalid because it comes before the start date", Appointment{"3cd59970-3942-448b-af1d-8b99310f1701", "796d707b-6f0a-4004-be01-f4d63b6866de", "title", "43e52216-bfff-440b-b1d8-6c3c7a29dbbb", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2024-07-01", "2024-06-21", "", ""}, []fieldTagPair{{"end date", "validAppointmentDuration"}}},
	}

	runValidationTest(t, tests)
}

func TestApppointmentValidation(t *testing.T) {
	tests := []validationTestCase{
		{"Appointment should be valid", Appointment{"62306293-a8a2-4530-9511-f1d8585c46e5", "796d707b-6f0a-4004-be01-f4d63b6866de" ,"title", "ff3e248a-a135-4733-8210-c11ee4b46afc", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023-12-27", "2024-06-21", "", ""}, []fieldTagPair{}},
		{"Appointment should be valid because end date is optional", Appointment{"62306293-a8a2-4530-9511-f1d8585c46e5", "796d707b-6f0a-4004-be01-f4d63b6866de" ,"title", "ff3e248a-a135-4733-8210-c11ee4b46afc", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023-12-27", "", "", ""}, []fieldTagPair{}},
		{"Appointment should be invalid because the id is missing", Appointment{"", "796d707b-6f0a-4004-be01-f4d63b6866de" ,"title", "ff3e248a-a135-4733-8210-c11ee4b46afc", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023-12-27", "2024-06-21", "", ""}, []fieldTagPair{{"appointment id", "required"}}},		
		{"Appointment should be invalid because the tenant id is missing", Appointment{"62306293-a8a2-4530-9511-f1d8585c46e5", "" ,"title", "ff3e248a-a135-4733-8210-c11ee4b46afc", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023-12-27", "2024-06-21", "", ""}, []fieldTagPair{{"tenant id", "required"}}},		
		{"Appointment should be invalid because title is missing", Appointment{"62306293-a8a2-4530-9511-f1d8585c46e5", "796d707b-6f0a-4004-be01-f4d63b6866de", "", "ff3e248a-a135-4733-8210-c11ee4b46afc", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023-12-27", "2024-06-21", "", ""}, []fieldTagPair{{"appointment title", "required"}}},
		{"Appointment should be invalid because department id is missing", Appointment{"62306293-a8a2-4530-9511-f1d8585c46e5", "796d707b-6f0a-4004-be01-f4d63b6866de" ,"title", "", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023-12-27", "2024-06-21", "", ""}, []fieldTagPair{{"department id", "required"}}},
		{"Appointment should be invalid because user id is missing", Appointment{"62306293-a8a2-4530-9511-f1d8585c46e5", "796d707b-6f0a-4004-be01-f4d63b6866de" ,"title", "ff3e248a-a135-4733-8210-c11ee4b46afc", "", "2023-12-27", "2024-06-21", "", ""}, []fieldTagPair{{"user id", "required"}}},
		{"Appointment should be invalid because start date is missing", Appointment{"62306293-a8a2-4530-9511-f1d8585c46e5", "796d707b-6f0a-4004-be01-f4d63b6866de" ,"title", "ff3e248a-a135-4733-8210-c11ee4b46afc", "a62c9359-64cd-4e05-bed4-27b361f882b6", "", "2024-06-21", "", ""}, []fieldTagPair{{"start date", "required"}}},
		{"Appointment should be invalid because the id is blank", Appointment{"   ", "796d707b-6f0a-4004-be01-f4d63b6866de" ,"title", "ff3e248a-a135-4733-8210-c11ee4b46afc", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023-12-27", "2024-06-21", "", ""}, []fieldTagPair{{"appointment id", "notBlank"}}},				
		{"Appointment should be invalid because the tenant id is blank", Appointment{"62306293-a8a2-4530-9511-f1d8585c46e5", "  " ,"title", "ff3e248a-a135-4733-8210-c11ee4b46afc", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023-12-27", "2024-06-21", "", ""}, []fieldTagPair{{"tenant id", "notBlank"}}},		
		{"Appointment should be invalid because title is blank", Appointment{"62306293-a8a2-4530-9511-f1d8585c46e5", "796d707b-6f0a-4004-be01-f4d63b6866de", "  ", "ff3e248a-a135-4733-8210-c11ee4b46afc", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023-12-27", "2024-06-21", "", ""}, []fieldTagPair{{"appointment title", "notBlank"}}},
		{"Appointment should be invalid because department id is blank", Appointment{"62306293-a8a2-4530-9511-f1d8585c46e5", "796d707b-6f0a-4004-be01-f4d63b6866de" ,"title", "  ", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023-12-27", "2024-06-21", "", ""}, []fieldTagPair{{"department id", "notBlank"}}},
		{"Appointment should be invalid because user id is blank", Appointment{"62306293-a8a2-4530-9511-f1d8585c46e5", "796d707b-6f0a-4004-be01-f4d63b6866de" ,"title", "ff3e248a-a135-4733-8210-c11ee4b46afc", "  ", "2023-12-27", "2024-06-21", "", ""}, []fieldTagPair{{"user id", "notBlank"}}},
		{"Appointment should be invalid because start date is blank", Appointment{"62306293-a8a2-4530-9511-f1d8585c46e5", "796d707b-6f0a-4004-be01-f4d63b6866de" ,"title", "ff3e248a-a135-4733-8210-c11ee4b46afc", "a62c9359-64cd-4e05-bed4-27b361f882b6", "  ", "2024-06-21", "", ""}, []fieldTagPair{{"start date", "notBlank"}}},
		{"Appointment should be invalid because end date is blank", Appointment{"62306293-a8a2-4530-9511-f1d8585c46e5", "796d707b-6f0a-4004-be01-f4d63b6866de" ,"title", "ff3e248a-a135-4733-8210-c11ee4b46afc", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2024-06-21", "   ", "", ""}, []fieldTagPair{{"end date", "notBlank"}}},
		{"Appointment should be invalid because the id is an invalid uuid", Appointment{"62306293-a8a24530-9511-f1d8585c46e5", "796d707b-6f0a-4004-be01-f4d63b6866de" ,"title", "ff3e248a-a135-4733-8210-c11ee4b46afc", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023-12-27", "2024-06-21", "", ""}, []fieldTagPair{{"appointment id", "uuid"}}},
		{"Appointment should be invalid because the tenant id is an invalid uuid", Appointment{"62306293-a8a2-4530-9511-f1d8585c46e5", "796d707b6f0a-4004-be01-f4d63b6866de" ,"title", "ff3e248a-a135-4733-8210-c11ee4b46afc", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023-12-27", "2024-06-21", "", ""}, []fieldTagPair{{"tenant id", "uuid"}}},		
		{"Appointment should be invalid because the department id is an invalid uuid", Appointment{"62306293-a8a2-4530-9511-f1d8585c46e5", "796d707b-6f0a-4004-be01-f4d63b6866de" ,"title", "ff3e248aa135-4733-8210-c11ee4b46afc", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023-12-27", "2024-06-21", "", ""}, []fieldTagPair{{"department id", "uuid"}}},		
		{"Appointment should be invalid because user id is an invalid uuid", Appointment{"62306293-a8a2-4530-9511-f1d8585c46e5", "796d707b-6f0a-4004-be01-f4d63b6866de" ,"title", "ff3e248a-a135-4733-8210-c11ee4b46afc", "a62c9359d64cd-4e05-bed4-27b361f882b6", "2023-12-27", "2024-06-21", "", ""}, []fieldTagPair{{"user id", "uuid"}}},
		{"Appointment should be invalid because start date is not in ISO format", Appointment{"62306293-a8a2-4530-9511-f1d8585c46e5", "796d707b-6f0a-4004-be01-f4d63b6866de" ,"title", "ff3e248a-a135-4733-8210-c11ee4b46afc", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023-13-13", "2024-06-21", "", ""}, []fieldTagPair{{"start date", "isIsoDate"}}},
		{"Appointment should be invalid because end date is not in ISO format", Appointment{"62306293-a8a2-4530-9511-f1d8585c46e5", "796d707b-6f0a-4004-be01-f4d63b6866de" ,"title", "ff3e248a-a135-4733-8210-c11ee4b46afc", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023-06-21", "2024-13-13", "", ""}, []fieldTagPair{{"end date", "isIsoDate"}}},
		{"Appointment should be invalid because end date is less than 30 days after start date", Appointment{"62306293-a8a2-4530-9511-f1d8585c46e5", "796d707b-6f0a-4004-be01-f4d63b6866de" ,"title", "ff3e248a-a135-4733-8210-c11ee4b46afc", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023-06-21", "2023-07-01", "", ""}, []fieldTagPair{{"end date", "validAppointmentDuration"}}},
	}

	runValidationTest(t, tests)
}
