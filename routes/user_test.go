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

func (s *IntegrationTestSuite) TestCreateUser() {
	wantUser := storage.User{
		Id:       "054cf786-7a54-4ebe-9cb7-d9750bbdedac",
		TenantId: s.defaultTenant.Id,
		Email:    "test@gmail.com",
	}

	type requestBody struct {
		Email string
	}
	reqBody := requestBody{
		Email: wantUser.Email,
	}

	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)

	r, err := http.NewRequest("POST", fmt.Sprintf("/api/tenants/%s/users/%s", wantUser.TenantId, wantUser.Id), bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultUser.Id, s.defaultUser.TenantId, s.defaultUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 201)

	s.expectSelectQueryToReturnOneRow(
		"user_account",
		map[string]string{
			"id":        wantUser.Id,
			"tenant_id": wantUser.TenantId,
			"email":     wantUser.Email,
		},
	)

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-CREATED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestCreateUserInvalidUser() {
	wantUser := storage.User{
		Id:       "054cf786-7a54-4ebe-9cb7-d9750bbdedac",
		TenantId: s.defaultTenant.Id,
		Email:    "    ",
	}

	type requestBody struct {
		Email string
	}
	reqBody := requestBody{
		Email: wantUser.Email,
	}

	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)

	r, err := http.NewRequest("POST", fmt.Sprintf("/api/tenants/%s/users/%s", wantUser.TenantId, wantUser.Id), bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultUser.Id, s.defaultUser.TenantId, s.defaultUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 400)
	s.expectErrorCode(w, "INPUT-VALIDATION-ERROR")

	s.expectSelectQueryToReturnNoRows(
		"user_account",
		map[string]string{
			"id":        wantUser.Id,
			"tenant_id": wantUser.TenantId,
			"email":     wantUser.Email,
		},
	)

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"INPUT-VALIDATION-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestCreateUserViolatesUniqueConstraint() {
	type requestBody struct {
		Email string
	}
	reqBody := requestBody{
		Email: s.defaultUser.Email,
	}

	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)

	r, err := http.NewRequest("POST", fmt.Sprintf("/api/tenants/%s/users/%s", s.defaultUser.TenantId, s.defaultUser.Id), bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultUser.Id, s.defaultUser.TenantId, s.defaultUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 409)
	s.expectErrorCode(w, "UNIQUE-VIOLATION-ERROR")

	s.expectSelectQueryToReturnOneRow(
		"user_account",
		map[string]string{
			"id":        s.defaultUser.Id,
			"tenant_id": s.defaultUser.TenantId,
			"email":     s.defaultUser.Email,
		},
	)

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"UNIQUE-VIOLATION-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestCreatePosition() {
	wantPosition := storage.Position{
		Id:           "cfc61cce-3d5a-4014-8490-3302ddd187b8",
		TenantId:     s.defaultTenant.Id,
		Title:        "Manager",
		DepartmentId: s.defaultDepartment.Id,
	}

	type requestBody struct {
		Title        string
		DepartmentId string
	}
	reqBody := requestBody{
		Title:        wantPosition.Title,
		DepartmentId: wantPosition.DepartmentId,
	}
	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)

	resource := fmt.Sprintf("/api/tenants/%s/positions/%s", wantPosition.TenantId, wantPosition.Id)
	r, err := http.NewRequest("POST", resource, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultUser.Id, s.defaultTenant.Id, s.defaultUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 201)

	s.expectSelectQueryToReturnOneRow(
		"position",
		map[string]string{
			"id":            wantPosition.Id,
			"tenant_id":     wantPosition.TenantId,
			"title":         wantPosition.Title,
			"department_id": wantPosition.DepartmentId,
		})

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"POSITION-CREATED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

// Creatapptvalidationerror
func (s *IntegrationTestSuite) TestCreatePositionInvalidInput() {
	wantPosition := storage.Position{
		Id:           "3e4216c5-d85c-4d4d-a48a-9aae1503261a",
		TenantId:     s.defaultTenant.Id,
		Title:        "   ",
		DepartmentId: s.defaultDepartment.Id,
	}

	type requestBody struct {
		Title        string
		DepartmentId string
		StartDate    string
		EndDate      string
	}
	reqBody := requestBody{
		Title:        wantPosition.Title,
		DepartmentId: wantPosition.DepartmentId,
	}
	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)

	path := fmt.Sprintf("/api/tenants/%s/positions/%s", wantPosition.TenantId, wantPosition.Id)
	r, err := http.NewRequest("POST", path, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultUser.Id, s.defaultUser.TenantId, s.defaultUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 400)
	s.expectErrorCode(w, "INPUT-VALIDATION-ERROR")

	s.expectSelectQueryToReturnNoRows(
		"position",
		map[string]string{
			"id":            wantPosition.Id,
			"tenant_id":     wantPosition.TenantId,
			"title":         wantPosition.Title,
			"department_id": wantPosition.DepartmentId,
		})

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"INPUT-VALIDATION-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

// Createapptpgerr
func (s *IntegrationTestSuite) TestCreatePositionAlreadyExists() {
	type requestBody struct {
		Title        string
		DepartmentId string
		StartDate    string
		EndDate      string
	}
	reqBody := requestBody{
		Title:        s.defaultPosition.Title,
		DepartmentId: s.defaultPosition.DepartmentId,
	}
	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)

	path := fmt.Sprintf("/api/tenants/%s/positions/%s", s.defaultPosition.TenantId, s.defaultPosition.Id)
	r, err := http.NewRequest("POST", path, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultUser.Id, s.defaultUser.TenantId, s.defaultUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 409)
	s.expectErrorCode(w, "UNIQUE-VIOLATION-ERROR")

	s.expectSelectQueryToReturnOneRow(
		"position",
		map[string]string{
			"id":            s.defaultPosition.Id,
			"tenant_id":     s.defaultPosition.TenantId,
			"title":         s.defaultPosition.Title,
			"department_id": s.defaultPosition.DepartmentId,
		})

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"UNIQUE-VIOLATION-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestCreatePositionAssignment() {
	wantPositionAssignment := storage.PositionAssignment{
		TenantId:   s.defaultTenant.Id,
		PositionId: "c1ddb117-94e0-40d1-908d-a07f43f319e8",
		UserId:     s.defaultUser.Id,
		StartDate:  "2024-02-01",
	}

	// Seed another position	
	query := "INSERT INTO position (id, tenant_id, title, department_id) VALUES ($1, $2, $3, $4)"
	_, err := s.dbRootConn.Exec(query, wantPositionAssignment.PositionId, s.defaultTenant.Id, "Random", s.defaultDepartment.Id)
	if err != nil {
		log.Fatalf("Could not seed position: %s", err)
	}


	type requestBody struct {
		StartDate string
		EndDate   string
	}
	reqBody := requestBody{
		StartDate: wantPositionAssignment.StartDate,
		EndDate:   wantPositionAssignment.EndDate,
	}
	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)

	resource := fmt.Sprintf("/api/tenants/%s/users/%s/positions/%s", wantPositionAssignment.TenantId, wantPositionAssignment.UserId, wantPositionAssignment.PositionId)
	r, err := http.NewRequest("POST", resource, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultUser.Id, s.defaultTenant.Id, s.defaultUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 201)

	s.expectSelectQueryToReturnOneRow(
		"position_assignment",
		map[string]string{
			"tenant_id":       wantPositionAssignment.TenantId,
			"position_id":   wantPositionAssignment.PositionId,
			"user_account_id": wantPositionAssignment.UserId,
			"start_date":      wantPositionAssignment.StartDate,
		})

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"POSITION-ASSIGNMENT-CREATED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

// Creatapptvalidationerror
func (s *IntegrationTestSuite) TestCreatePositionAssignmentInvalidInput() {
	wantPositionAssignment := storage.PositionAssignment{
		TenantId:   s.defaultTenant.Id,
		PositionId: "53832d4e2e0e-4275-a5cf-24c1a5f37148",
		UserId:     s.defaultUser.Id,
		StartDate:  "2024-06-02",
	}

	type requestBody struct {
		StartDate string
		EndDate   string
	}
	reqBody := requestBody{
		StartDate: wantPositionAssignment.StartDate,
		EndDate:   wantPositionAssignment.EndDate,
	}
	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)

	path := fmt.Sprintf("/api/tenants/%s/users/%s/positions/%s", wantPositionAssignment.TenantId, wantPositionAssignment.UserId, wantPositionAssignment.PositionId)
	r, err := http.NewRequest("POST", path, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultUser.Id, s.defaultUser.TenantId, s.defaultUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 400)
	s.expectErrorCode(w, "INPUT-VALIDATION-ERROR")

	s.expectSelectQueryToReturnNoRows(
		"position_assignment",
		map[string]string{
			"tenant_id":       wantPositionAssignment.TenantId,
			"position_id":     wantPositionAssignment.PositionId,
			"user_account_id": wantPositionAssignment.UserId,
			"start_date":      wantPositionAssignment.StartDate,
		})

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"INPUT-VALIDATION-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

// Createapptpgerr
func (s *IntegrationTestSuite) TestCreatePositionAssignmentAlreadyExists() {
	type requestBody struct {
		Title        string
		DepartmentId string
		StartDate    string
		EndDate      string
	}
	reqBody := requestBody{
		StartDate: s.defaultPositionAssignment.StartDate,
		EndDate:   s.defaultPositionAssignment.EndDate,
	}
	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)

	path := fmt.Sprintf("/api/tenants/%s/users/%s/positions/%s", s.defaultPositionAssignment.TenantId, s.defaultPositionAssignment.UserId, s.defaultPositionAssignment.PositionId)
	r, err := http.NewRequest("POST", path, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultUser.Id, s.defaultUser.TenantId, s.defaultUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 409)
	s.expectErrorCode(w, "UNIQUE-VIOLATION-ERROR")

	s.expectSelectQueryToReturnOneRow(
		"position_assignment",
		map[string]string{
			"tenant_id":       s.defaultPositionAssignment.TenantId,
			"position_id":     s.defaultPositionAssignment.PositionId,
			"user_account_id": s.defaultPositionAssignment.UserId,
			"start_date":      s.defaultPositionAssignment.StartDate,
		})

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"UNIQUE-VIOLATION-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}
