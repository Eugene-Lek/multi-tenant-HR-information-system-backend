package storage

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

func TestPositionValidation(t *testing.T) {
	tests := []validationTestCase{
		{"Position should be valid", Position{"62306293-a8a2-4530-9511-f1d8585c46e5", "796d707b-6f0a-4004-be01-f4d63b6866de", "title", "a62c9359-64cd-4e05-bed4-27b361f882b6", []string{"786f9df6-3de6-42a6-8324-ad0fcd7c9181"}, "", ""}, []fieldTagPair{}},
		{"Position should be invalid because the id is missing", Position{"", "796d707b-6f0a-4004-be01-f4d63b6866de", "title", "a62c9359-64cd-4e05-bed4-27b361f882b6", []string{"786f9df6-3de6-42a6-8324-ad0fcd7c9181"}, "", ""}, []fieldTagPair{{"position id", "required"}}},
		{"Position should be invalid because the tenant id is missing", Position{"62306293-a8a2-4530-9511-f1d8585c46e5", "", "title", "a62c9359-64cd-4e05-bed4-27b361f882b6", []string{"786f9df6-3de6-42a6-8324-ad0fcd7c9181"}, "", ""}, []fieldTagPair{{"tenant id", "required"}}},
		{"Position should be invalid because title is missing", Position{"62306293-a8a2-4530-9511-f1d8585c46e5", "796d707b-6f0a-4004-be01-f4d63b6866de", "", "a62c9359-64cd-4e05-bed4-27b361f882b6", []string{"786f9df6-3de6-42a6-8324-ad0fcd7c9181"}, "", ""}, []fieldTagPair{{"position title", "required"}}},
		{"Position should be invalid because department id is missing", Position{"62306293-a8a2-4530-9511-f1d8585c46e5", "796d707b-6f0a-4004-be01-f4d63b6866de", "title", "", []string{"786f9df6-3de6-42a6-8324-ad0fcd7c9181"}, "", ""}, []fieldTagPair{{"department id", "required"}}},
		{"Position should be invalid because the supervisor id is required", Position{"62306293-a8a2-4530-9511-f1d8585c46e5", "796d707b-6f0a-4004-be01-f4d63b6866de", "title", "a62c9359-64cd-4e05-bed4-27b361f882b6", nil, "", ""}, []fieldTagPair{{"supervisor ids", "required"}}},
		{"Position should be invalid because the id is blank", Position{"   ", "796d707b-6f0a-4004-be01-f4d63b6866de", "title", "a62c9359-64cd-4e05-bed4-27b361f882b6", []string{"786f9df6-3de6-42a6-8324-ad0fcd7c9181"}, "", ""}, []fieldTagPair{{"position id", "notBlank"}}},
		{"Position should be invalid because the tenant id is blank", Position{"62306293-a8a2-4530-9511-f1d8585c46e5", "  ", "title", "a62c9359-64cd-4e05-bed4-27b361f882b6", []string{"786f9df6-3de6-42a6-8324-ad0fcd7c9181"}, "", ""}, []fieldTagPair{{"tenant id", "notBlank"}}},
		{"Position should be invalid because title is blank", Position{"62306293-a8a2-4530-9511-f1d8585c46e5", "796d707b-6f0a-4004-be01-f4d63b6866de", "   ", "a62c9359-64cd-4e05-bed4-27b361f882b6", []string{"786f9df6-3de6-42a6-8324-ad0fcd7c9181"}, "", ""}, []fieldTagPair{{"position title", "notBlank"}}},
		{"Position should be invalid because department id is blank", Position{"62306293-a8a2-4530-9511-f1d8585c46e5", "796d707b-6f0a-4004-be01-f4d63b6866de", "title", "   ", []string{"786f9df6-3de6-42a6-8324-ad0fcd7c9181"}, "", ""}, []fieldTagPair{{"department id", "notBlank"}}},
		{"Position should be invalid because the supervisor id is blank", Position{"62306293-a8a2-4530-9511-f1d8585c46e5", "796d707b-6f0a-4004-be01-f4d63b6866de", "title", "a62c9359-64cd-4e05-bed4-27b361f882b6", []string{"  "}, "", ""}, []fieldTagPair{{"supervisor ids[0]", "notBlank"}}},
		{"Position should be invalid because the id is an invalid uuid", Position{"62306293-a8a24530-9511-f1d8585c46e5", "796d707b-6f0a-4004-be01-f4d63b6866de", "title", "a62c9359-64cd-4e05-bed4-27b361f882b6", []string{"786f9df6-3de6-42a6-8324-ad0fcd7c9181"}, "", ""}, []fieldTagPair{{"position id", "uuid"}}},
		{"Position should be invalid because the tenant id is an invalid uuid", Position{"62306293-a8a2-4530-9511-f1d8585c46e5", "796d707b6f0a-4004-be01-f4d63b6866de", "title", "a62c9359-64cd-4e05-bed4-27b361f882b6", []string{"786f9df6-3de6-42a6-8324-ad0fcd7c9181"}, "", ""}, []fieldTagPair{{"tenant id", "uuid"}}},
		{"Position should be invalid because the department id is an invalid uuid", Position{"62306293-a8a2-4530-9511-f1d8585c46e5", "796d707b-6f0a-4004-be01-f4d63b6866de", "title", "a62c935964cd-4e05-bed4-27b361f882b6", []string{"786f9df6-3de6-42a6-8324-ad0fcd7c9181"}, "", ""}, []fieldTagPair{{"department id", "uuid"}}},
		{"Position should be invalid because the supervisor id is an invalid uuid", Position{"62306293-a8a2-4530-9511-f1d8585c46e5", "796d707b-6f0a-4004-be01-f4d63b6866de", "title", "a62c9359-64cd-4e05-bed4-27b361f882b6", []string{"786f9df63de6-42a6-8324-ad0fcd7c9181"}, "", ""}, []fieldTagPair{{"supervisor ids[0]", "uuid"}}},
	}

	runValidationTest(t, tests)
}

func TestIsIsoDate(t *testing.T) {
	tests := []validationTestCase{
		{"Start date should be valid because 2024-02-29 is a valid date", PositionAssignment{"796d707b-6f0a-4004-be01-f4d63b6866de", "43e52216-bfff-440b-b1d8-6c3c7a29dbbb", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2024-02-29", "2024-06-21", "", ""}, []fieldTagPair{}},
		{"Start date should be invalid because 2023-02-29 is an invalid date", PositionAssignment{"796d707b-6f0a-4004-be01-f4d63b6866de", "43e52216-bfff-440b-b1d8-6c3c7a29dbbb", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023-02-29", "2024-06-21", "", ""}, []fieldTagPair{{"start date", "isIsoDate"}}},
		{"Start date should be invalid because 13 is an invalid month", PositionAssignment{"796d707b-6f0a-4004-be01-f4d63b6866de", "43e52216-bfff-440b-b1d8-6c3c7a29dbbb", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023-13-01", "2024-06-21", "", ""}, []fieldTagPair{{"start date", "isIsoDate"}}},
		{"Start date should be invalid because preceeding 0 is missing", PositionAssignment{"796d707b-6f0a-4004-be01-f4d63b6866de", "43e52216-bfff-440b-b1d8-6c3c7a29dbbb", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023-2-01", "2024-06-21", "", ""}, []fieldTagPair{{"start date", "isIsoDate"}}},
		{"Start date should be invalid because dashes are missing", PositionAssignment{"796d707b-6f0a-4004-be01-f4d63b6866de", "43e52216-bfff-440b-b1d8-6c3c7a29dbbb", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023/02/01", "2024-06-21", "", ""}, []fieldTagPair{{"start date", "isIsoDate"}}},
		{"Start date should be invalid because year is in YY format", PositionAssignment{"796d707b-6f0a-4004-be01-f4d63b6866de", "43e52216-bfff-440b-b1d8-6c3c7a29dbbb", "a62c9359-64cd-4e05-bed4-27b361f882b6", "23-02-01", "2024-06-21", "", ""}, []fieldTagPair{{"start date", "isIsoDate"}}},
	}

	runValidationTest(t, tests)
}

func TestValidPositionAssignmentDuration(t *testing.T) {
	tests := []validationTestCase{
		{"End date should be valid", PositionAssignment{"796d707b-6f0a-4004-be01-f4d63b6866de", "43e52216-bfff-440b-b1d8-6c3c7a29dbbb", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2024-02-29", "2024-06-21", "", ""}, []fieldTagPair{}},
		{"End date should be valid because the duration is exactly 30 days", PositionAssignment{"796d707b-6f0a-4004-be01-f4d63b6866de", "43e52216-bfff-440b-b1d8-6c3c7a29dbbb", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2024-02-29", "2024-06-21", "", ""}, []fieldTagPair{}},
		{"End date should be invalid because the duration is less than 30 days", PositionAssignment{"796d707b-6f0a-4004-be01-f4d63b6866de", "43e52216-bfff-440b-b1d8-6c3c7a29dbbb", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2024-06-01", "2024-06-21", "", ""}, []fieldTagPair{{"end date", "validPositionAssignmentDuration"}}},
		{"End date should be invalid because it comes before the start date", PositionAssignment{"796d707b-6f0a-4004-be01-f4d63b6866de", "43e52216-bfff-440b-b1d8-6c3c7a29dbbb", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2024-07-01", "2024-06-21", "", ""}, []fieldTagPair{{"end date", "validPositionAssignmentDuration"}}},
	}

	runValidationTest(t, tests)
}

func TestPositionAssignmentValidation(t *testing.T) {
	tests := []validationTestCase{
		{"Position Assignment should be valid", PositionAssignment{"796d707b-6f0a-4004-be01-f4d63b6866de", "ff3e248a-a135-4733-8210-c11ee4b46afc", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023-12-27", "2024-06-21", "", ""}, []fieldTagPair{}},
		{"Position Assignment should be valid because end date is optional", PositionAssignment{"796d707b-6f0a-4004-be01-f4d63b6866de", "ff3e248a-a135-4733-8210-c11ee4b46afc", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023-12-27", "", "", ""}, []fieldTagPair{}},
		{"Position Assignment should be invalid because the tenant id is missing", PositionAssignment{"", "ff3e248a-a135-4733-8210-c11ee4b46afc", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023-12-27", "2024-06-21", "", ""}, []fieldTagPair{{"tenant id", "required"}}},
		{"Position Assignment should be invalid because position id is missing", PositionAssignment{"796d707b-6f0a-4004-be01-f4d63b6866de", "", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023-12-27", "2024-06-21", "", ""}, []fieldTagPair{{"position id", "required"}}},
		{"Position Assignment should be invalid because user id is missing", PositionAssignment{"796d707b-6f0a-4004-be01-f4d63b6866de", "ff3e248a-a135-4733-8210-c11ee4b46afc", "", "2023-12-27", "2024-06-21", "", ""}, []fieldTagPair{{"user id", "required"}}},
		{"Position Assignment should be invalid because start date is missing", PositionAssignment{"796d707b-6f0a-4004-be01-f4d63b6866de", "ff3e248a-a135-4733-8210-c11ee4b46afc", "a62c9359-64cd-4e05-bed4-27b361f882b6", "", "2024-06-21", "", ""}, []fieldTagPair{{"start date", "required"}}},
		{"Position Assignment should be invalid because the tenant id is blank", PositionAssignment{"  ", "ff3e248a-a135-4733-8210-c11ee4b46afc", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023-12-27", "2024-06-21", "", ""}, []fieldTagPair{{"tenant id", "notBlank"}}},
		{"Position Assignment should be invalid because position id is blank", PositionAssignment{"796d707b-6f0a-4004-be01-f4d63b6866de", "  ", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023-12-27", "2024-06-21", "", ""}, []fieldTagPair{{"position id", "notBlank"}}},
		{"Position Assignment should be invalid because the user id is blank", PositionAssignment{"796d707b-6f0a-4004-be01-f4d63b6866de", "ff3e248a-a135-4733-8210-c11ee4b46afc", "  ", "2023-12-27", "2024-06-21", "", ""}, []fieldTagPair{{"user id", "notBlank"}}},
		{"Position Assignment should be invalid because start date is blank", PositionAssignment{"796d707b-6f0a-4004-be01-f4d63b6866de", "ff3e248a-a135-4733-8210-c11ee4b46afc", "a62c9359-64cd-4e05-bed4-27b361f882b6", "  ", "2024-06-21", "", ""}, []fieldTagPair{{"start date", "notBlank"}}},
		{"Position Assignment should be invalid because end date is blank", PositionAssignment{"796d707b-6f0a-4004-be01-f4d63b6866de", "ff3e248a-a135-4733-8210-c11ee4b46afc", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2024-06-21", "   ", "", ""}, []fieldTagPair{{"end date", "notBlank"}}},
		{"Position Assignment should be invalid because the tenant id is an invalid uuid", PositionAssignment{"796d707b6f0a-4004-be01-f4d63b6866de", "ff3e248a-a135-4733-8210-c11ee4b46afc", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023-12-27", "2024-06-21", "", ""}, []fieldTagPair{{"tenant id", "uuid"}}},
		{"Position Assignment should be invalid because the position id is an invalid uuid", PositionAssignment{"796d707b-6f0a-4004-be01-f4d63b6866de", "ff3e248aa135-4733-8210-c11ee4b46afc", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023-12-27", "2024-06-21", "", ""}, []fieldTagPair{{"position id", "uuid"}}},
		{"Position Assignment should be invalid because the user id is an invalid uuid", PositionAssignment{"796d707b-6f0a-4004-be01-f4d63b6866de", "ff3e248a-a135-4733-8210-c11ee4b46afc", "a62c9359d64cd-4e05-bed4-27b361f882b6", "2023-12-27", "2024-06-21", "", ""}, []fieldTagPair{{"user id", "uuid"}}},
		{"Position Assignment should be invalid because start date is not in ISO format", PositionAssignment{"796d707b-6f0a-4004-be01-f4d63b6866de", "ff3e248a-a135-4733-8210-c11ee4b46afc", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023-13-13", "2024-06-21", "", ""}, []fieldTagPair{{"start date", "isIsoDate"}}},
		{"Position Assignment should be invalid because end date is not in ISO format", PositionAssignment{"796d707b-6f0a-4004-be01-f4d63b6866de", "ff3e248a-a135-4733-8210-c11ee4b46afc", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023-06-21", "2024-13-13", "", ""}, []fieldTagPair{{"end date", "isIsoDate"}}},
		{"Position Assignment should be invalid because end date is less than 30 days after start date", PositionAssignment{"796d707b-6f0a-4004-be01-f4d63b6866de", "ff3e248a-a135-4733-8210-c11ee4b46afc", "a62c9359-64cd-4e05-bed4-27b361f882b6", "2023-06-21", "2023-07-01", "", ""}, []fieldTagPair{{"end date", "validPositionAssignmentDuration"}}},
	}

	runValidationTest(t, tests)
}

func TestPolicies(t *testing.T) {
	tests := []validationTestCase{
		{"Policies should be valid", Policies{"ROOT-ROLE-ADMIN", "796d707b-6f0a-4004-be01-f4d63b6866de", []Resource{{"/api/tenants/*", "POST"}}, "", ""}, []fieldTagPair{}},
		{"Policies should be invalid because role is missing", Policies{"", "796d707b-6f0a-4004-be01-f4d63b6866de", []Resource{{"/api/tenants/*", "POST"}}, "", ""}, []fieldTagPair{{"role name", "required"}}},
		{"Policies should be invalid because tenant id is missing", Policies{"ROOT-ROLE-ADMIN", "", []Resource{{"/api/tenants/*", "POST"}}, "", ""}, []fieldTagPair{{"tenant id", "required"}}},
		{"Policies should be invalid because resource path is missing", Policies{"ROOT-ROLE-ADMIN", "796d707b-6f0a-4004-be01-f4d63b6866de", []Resource{{"", "POST"}}, "", ""}, []fieldTagPair{{"resource path", "required"}}},
		{"Policies should be invalid because resource method is missing", Policies{"ROOT-ROLE-ADMIN", "796d707b-6f0a-4004-be01-f4d63b6866de", []Resource{{"/api/tenants/*", ""}}, "", ""}, []fieldTagPair{{"resource method", "required"}}},
		{"Policies should be invalid because role is blank", Policies{"   ", "796d707b-6f0a-4004-be01-f4d63b6866de", []Resource{{"/api/tenants/*", "POST"}}, "", ""}, []fieldTagPair{{"role name", "notBlank"}}},
		{"Policies should be invalid because tenant id is blank", Policies{"ROOT-ROLE-ADMIN", "   ", []Resource{{"/api/tenants/*", "POST"}}, "", ""}, []fieldTagPair{{"tenant id", "notBlank"}}},
		{"Policies should be invalid because resource path is blank", Policies{"ROOT-ROLE-ADMIN", "796d707b-6f0a-4004-be01-f4d63b6866de", []Resource{{"   ", "POST"}}, "", ""}, []fieldTagPair{{"resource path", "notBlank"}}},
		{"Policies should be invalid because resource method is blank", Policies{"ROOT-ROLE-ADMIN", "796d707b-6f0a-4004-be01-f4d63b6866de", []Resource{{"/api/tenants/*", "   "}}, "", ""}, []fieldTagPair{{"resource method", "notBlank"}}},
		{"Policies should be invalid because tenant id is an invalid uuid", Policies{"ROOT-ROLE-ADMIN", "796d707b6f0a-4004-be01-f4d63b6866de", []Resource{{"/api/tenants/*", "POST"}}, "", ""}, []fieldTagPair{{"tenant id", "uuid"}}},
	}

	runValidationTest(t, tests)
}

func TestRoleAssignment(t *testing.T) {
	tests := []validationTestCase{
		{"Role Assignment should be valid", RoleAssignment{"a62c9359-64cd-4e05-bed4-27b361f882b6", "ROOT-ROLE-ADMIN", "796d707b-6f0a-4004-be01-f4d63b6866de", "", ""}, []fieldTagPair{}},
		{"Role Assignment should be invalid because user id is missing", RoleAssignment{"", "ROOT-ROLE-ADMIN", "796d707b-6f0a-4004-be01-f4d63b6866de", "", ""}, []fieldTagPair{{"user id", "required"}}},
		{"Role Assignment should be invalid because role name is missing", RoleAssignment{"a62c9359-64cd-4e05-bed4-27b361f882b6", "", "796d707b-6f0a-4004-be01-f4d63b6866de", "", ""}, []fieldTagPair{{"role name", "required"}}},
		{"Role Assignment should be invalid because tenant id is missing", RoleAssignment{"a62c9359-64cd-4e05-bed4-27b361f882b6", "ROOT-ROLE-ADMIN", "", "", ""}, []fieldTagPair{{"tenant id", "required"}}},
		{"Role Assignment should be invalid because user id is blank", RoleAssignment{"   ", "ROOT-ROLE-ADMIN", "796d707b-6f0a-4004-be01-f4d63b6866de", "", ""}, []fieldTagPair{{"user id", "notBlank"}}},
		{"Role Assignment should be invalid because role name is blank", RoleAssignment{"a62c9359-64cd-4e05-bed4-27b361f882b6", "   ", "796d707b-6f0a-4004-be01-f4d63b6866de", "", ""}, []fieldTagPair{{"role name", "notBlank"}}},
		{"Role Assignment should be invalid because tenant id is blank", RoleAssignment{"a62c9359-64cd-4e05-bed4-27b361f882b6", "ROOT-ROLE-ADMIN", "   ", "", ""}, []fieldTagPair{{"tenant id", "notBlank"}}},
		{"Role Assignment should be invalid because user id is an invalid uuid", RoleAssignment{"a62c935964cd-4e05-bed4-27b361f882b6", "ROOT-ROLE-ADMIN", "796d707b-6f0a-4004-be01-f4d63b6866de", "", ""}, []fieldTagPair{{"user id", "uuid"}}},
		{"Role Assignment should be invalid because tenant id is an invalid uuid", RoleAssignment{"a62c9359-64cd-4e05-bed4-27b361f882b6", "ROOT-ROLE-ADMIN", "796d707b6f0a-4004-be01-f4d63b6866de", "", ""}, []fieldTagPair{{"tenant id", "uuid"}}},
	}

	runValidationTest(t, tests)
}
