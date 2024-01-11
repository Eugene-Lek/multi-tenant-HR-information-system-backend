package routes

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"

	_ "github.com/lib/pq"

	"multi-tenant-HR-information-system-backend/storage"
)

// Verifies the happy path works
func (s *IntegrationTestSuite) TestCreateTenant() {
	rows, _ := s.dbRootConn.Query("SELECT * FROM casbin_rule")
	for rows.Next() {
		var row struct{
			Id string
			Ptype string
			V0 string
			V1 string
			V2 string
			V3 string
			V4 string												
			V5 string															
		} 
		rows.Scan(&row.Id, &row.Ptype, &row.V0, &row.V1, &row.V2, &row.V3, &row.V4, &row.V5)
		log.Print(row)
	}

	wantTenant := storage.Tenant{
		Id:   "5338d729-32bd-4ad2-a8d1-22cbf81113de",
		Name: "Macdonalds",
	}

	// Create & serve the request
	type requestBody struct {
		Name string
	}	
	body := requestBody{
		Name: wantTenant.Name,
	}

	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(body)

	req, err := http.NewRequest("POST", fmt.Sprintf("/api/tenants/%s", wantTenant.Id), bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(req, s.defaultUser.Id, s.defaultUser.TenantId, s.defaultUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	// Check response
	s.expectHttpStatus(w, 201)

	// Check database
	s.expectSelectQueryToReturnOneRow("tenant", map[string]string{"id": wantTenant.Id})

	// Check logs
	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"TENANT-CREATED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

// Verifies that the validation function is executed & validation errors are handled correctly
// (by triggering a validation error with invalid input)
func (s *IntegrationTestSuite) TestCreateTenantInvalidInput() {
	invalidTenant := storage.Tenant{
		Id:   "5338d729-32bd-4ad2-a8d1-22cbf81113de",		
		Name: "   ",
	}

	// Create the request and add a session cookie to it
	type requestBody struct {
		Name string
	}	
	reqBody := requestBody{
		Name: invalidTenant.Name,
	}

	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)

	r, err := http.NewRequest("POST", fmt.Sprintf("/api/tenants/%s", invalidTenant.Id), bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultUser.Id, s.defaultUser.TenantId, s.defaultUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	// Check the response status and body
	s.expectHttpStatus(w, 400)
	s.expectErrorCode(w, "INPUT-VALIDATION-ERROR")

	// Check the database
	s.expectSelectQueryToReturnNoRows("tenant", map[string]string{"id": invalidTenant.Id})

	// Check logs
	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"INPUT-VALIDATION-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

// Verifies that postgres errors are handled correctly (by triggering a postgres error)
func (s *IntegrationTestSuite) TestCreateTenantAlreadyExists() {
	existingTenant := storage.Tenant{
		Id: s.defaultTenant.Id,		
		Name: s.defaultTenant.Name,
	}
	// Create the request and add a session cookie to it
	type requestBody struct {
		Name string
	}	
	reqBody := requestBody{
		Name: existingTenant.Name,
	}
	
	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)	
	req, err := http.NewRequest("POST", fmt.Sprintf("/api/tenants/%s", existingTenant.Id), bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(req, s.defaultUser.Id, s.defaultUser.TenantId, s.defaultUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	// Check the response status and body
	s.expectHttpStatus(w, 409)
	s.expectErrorCode(w, "UNIQUE-VIOLATION-ERROR")	

	// Check the database
	s.expectSelectQueryToReturnOneRow("tenant", map[string]string{"id": existingTenant.Id})

	// Check logs
	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"UNIQUE-VIOLATION-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

// Verifies that the happy path works
func (s *IntegrationTestSuite) TestCreateDivision() {
	wantDivision := storage.Division{
		Id: "f0935407-d43a-47f6-8fdd-bc45ab9c43d9",
		TenantId: s.defaultDivision.TenantId,
		Name:   "Marketing",
	}

	// Create the request and add a session cookie to it
	type requestBody struct {
		Name string
	}	
	reqBody := requestBody{
		Name: wantDivision.Name,
	}
	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)

	req, err := http.NewRequest("POST", fmt.Sprintf("/api/tenants/%s/divisions/%s", wantDivision.TenantId, wantDivision.Id), bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(req, s.defaultUser.Id, s.defaultUser.TenantId, s.defaultUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	// Check the response status and body
	s.expectHttpStatus(w, 201)

	// Check the database
	s.expectSelectQueryToReturnOneRow(
		"division",
		map[string]string{"id": wantDivision.Id},
	)

	// Check the logs
	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"DIVISION-CREATED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

// Verifies that the validation function is executed & validation errors are handled correctly
func (s *IntegrationTestSuite) TestCreateDivisionInvalidInput() {
	invalidDivision := storage.Division{
		Id: "f0935407-d43a-47f6-8fdd-bc45ab9c43d9",		
		TenantId: s.defaultDivision.TenantId,
		Name:   "  ",
	}

	// Create the request and add a session cookie to it
	type requestBody struct {
		Name string
	}	
	reqBody := requestBody{
		Name: invalidDivision.Name,
	}
	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)

	r, err := http.NewRequest("POST", fmt.Sprintf("/api/tenants/%s/divisions/%s", invalidDivision.TenantId, invalidDivision.Id), bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultUser.Id, s.defaultUser.TenantId, s.defaultUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	// Check the response status and body
	s.expectHttpStatus(w, 400)
	s.expectErrorCode(w, "INPUT-VALIDATION-ERROR")

	// Check the database
	s.expectSelectQueryToReturnNoRows(
		"division",
		map[string]string{"id": invalidDivision.Id},
	)

	// Check the logs
	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"INPUT-VALIDATION-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

// Verifies that postgres errors are handled correctly
func (s *IntegrationTestSuite) TestCreateDivisionAlreadyExists() {
	// Create the request and add a session cookie to it
	type requestBody struct {
		Name string
	}	
	reqBody := requestBody{
		Name: s.defaultDivision.Name,
	}
	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)

	req, err := http.NewRequest("POST", fmt.Sprintf("/api/tenants/%s/divisions/%s", s.defaultDivision.TenantId, s.defaultDivision.Id), bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(req, s.defaultUser.Id, s.defaultUser.TenantId, s.defaultUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	// Check the response status and body
	s.expectHttpStatus(w, 409)
	s.expectErrorCode(w, "UNIQUE-VIOLATION-ERROR")

	// Check the database
	s.expectSelectQueryToReturnOneRow(
		"division",
		map[string]string{"id": s.defaultDivision.Id},
	)

	// Check the logs
	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"UNIQUE-VIOLATION-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

// Verifies that the happy path works
func (s *IntegrationTestSuite) TestCreateDepartment() {
	wantDepartment := storage.Department{
		Id: "444aa127-b21b-45cf-b779-eb1c1ef82478",
		TenantId: s.defaultTenant.Id,
		DivisionId: s.defaultDivision.Id,
		Name:     "Customer Support",
	}

	// Create the request and add a session cookie to it
	type requestBody struct {
		Name string
	}	
	reqBody := requestBody{
		Name: wantDepartment.Name,
	}
	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)

	r, err := http.NewRequest("POST", fmt.Sprintf("/api/tenants/%s/divisions/%s/departments/%s", s.defaultUser.TenantId, wantDepartment.DivisionId, wantDepartment.Id), bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultUser.Id, s.defaultUser.TenantId, s.defaultUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	// Check the response status and body
	s.expectHttpStatus(w, 201)

	// Check the database
	s.expectSelectQueryToReturnOneRow(
		"department",
		map[string]string{"id": wantDepartment.Id},
	)

	// Check the logs
	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"DEPARTMENT-CREATED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

// Verifies that the validation function is executed & validation errors are handled correctly
func (s *IntegrationTestSuite) TestCreateDepartmentInvalidInput() {
	invalidDepartment := storage.Department{
		Id: "444aa127-b21b-45cf-b779-eb1c1ef82478",
		DivisionId: s.defaultDivision.Id,
		Name:     "   ",
	}

	// Create the request and add a session cookie to it
	type requestBody struct {
		Name string
	}	
	reqBody := requestBody{
		Name: invalidDepartment.Name,
	}
	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)

	r, err := http.NewRequest("POST", fmt.Sprintf("/api/tenants/%s/divisions/%s/departments/%s", s.defaultUser.TenantId, invalidDepartment.DivisionId, invalidDepartment.Id), bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultUser.Id, s.defaultUser.TenantId, s.defaultUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	// Check the response status and body
	s.expectHttpStatus(w, 400)
	s.expectErrorCode(w, "INPUT-VALIDATION-ERROR")

	// Check the database
	s.expectSelectQueryToReturnNoRows(
		"department",
		map[string]string{"id": invalidDepartment.Id},
	)

	// Check the logs
	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"INPUT-VALIDATION-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

// Verifies that postgres errors are handled correctly
func (s *IntegrationTestSuite) TestCreateDepartmentAlreadyExists() {
	// Create the request and add a session cookie to it
	type requestBody struct {
		Name string
	}	
	reqBody := requestBody{
		Name: s.defaultDepartment.Name,
	}
	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)
	
	r, err := http.NewRequest("POST", fmt.Sprintf("/api/tenants/%s/divisions/%s/departments/%s", s.defaultUser.TenantId, s.defaultDepartment.DivisionId, s.defaultDepartment.Id), bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultUser.Id, s.defaultUser.TenantId, s.defaultUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	// Check the response status and body
	s.expectHttpStatus(w, 409)
	s.expectErrorCode(w, "UNIQUE-VIOLATION-ERROR")

	// Check the database
	s.expectSelectQueryToReturnOneRow(
		"department",
		map[string]string{"id": s.defaultDepartment.Id},
	)

	// Check the logs
	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"UNIQUE-VIOLATION-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}