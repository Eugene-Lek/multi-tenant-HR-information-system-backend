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

func (s *IntegrationTestSuite) TestCreateAppointment() {
	wantAppointment := storage.Appointment{
		Id: "cfc61cce-3d5a-4014-8490-3302ddd187b8",
		TenantId: s.defaultTenant.Id,		
		UserId:     s.defaultUser.Id,			
		Title:      "Manager",	
		DepartmentId: s.defaultDepartment.Id,
		StartDate:  "2024-02-01",
	}

	type requestBody struct {
		Title string
		DepartmentId string
		StartDate string
		EndDate string
	}
	reqBody := requestBody{
		Title: wantAppointment.Title,
		DepartmentId: wantAppointment.DepartmentId,
		StartDate: wantAppointment.StartDate,
		EndDate: wantAppointment.EndDate,		
	}
	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)

	resource := fmt.Sprintf("/api/tenants/%s/users/%s/appointments/%s", wantAppointment.TenantId, wantAppointment.UserId, wantAppointment.Id)
	r, err := http.NewRequest("POST", resource, bodyBuf )
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultUser.Id, s.defaultTenant.Id, s.defaultUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 201)

	s.expectSelectQueryToReturnOneRow(
		"appointment", 
		map[string]string{
			"id": wantAppointment.Id,
			"tenant_id": wantAppointment.TenantId,
			"title": wantAppointment.Title,
			"department_id": wantAppointment.DepartmentId,
			"user_account_id": wantAppointment.UserId,
			"start_date": wantAppointment.StartDate,
		})

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"APPOINTMENT-CREATED"`)	
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)		
}


//Creatapptvalidationerror
func (s *IntegrationTestSuite) TestCreateAppointmentInvalidInput() {
	wantAppointment := storage.Appointment{
		Id: "3e4216c5-d85c-4d4d-a48a-9aae1503261a",
		TenantId: s.defaultTenant.Id,
		Title: "   ",
		DepartmentId: s.defaultDepartment.Id,
		UserId: s.defaultUser.Id,
		StartDate: "2024-06-02",
	}

	type requestBody struct {
		Title string
		DepartmentId string
		StartDate string
		EndDate string		
	}
	reqBody := requestBody{
		Title: wantAppointment.Title,
		DepartmentId: wantAppointment.DepartmentId,
		StartDate: wantAppointment.StartDate,
		EndDate: wantAppointment.EndDate,		
	}	
	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)
	
	path := fmt.Sprintf("/api/tenants/%s/users/%s/appointments/%s", wantAppointment.TenantId, wantAppointment.UserId, wantAppointment.Id)
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
		"appointment", 
		map[string]string{
			"id": wantAppointment.Id,
			"tenant_id": wantAppointment.TenantId,
			"title": wantAppointment.Title,
			"department_id": wantAppointment.DepartmentId,
			"user_account_id": wantAppointment.UserId,
			"start_date": wantAppointment.StartDate,
	})

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"INPUT-VALIDATION-ERROR"`)	
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)		
}
//Createapptpgerr
func (s *IntegrationTestSuite) TestCreateAppointmentAlreadyExists() {
	type requestBody struct {
		Title string
		DepartmentId string
		StartDate string
		EndDate string		
	}
	reqBody := requestBody{
		Title: s.defaultAppointment.Title,
		DepartmentId: s.defaultAppointment.DepartmentId,
		StartDate: s.defaultAppointment.StartDate,	
		EndDate: s.defaultAppointment.EndDate,		
	}	
	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)
	
	path := fmt.Sprintf("/api/tenants/%s/users/%s/appointments/%s", s.defaultAppointment.TenantId, s.defaultAppointment.UserId, s.defaultAppointment.Id)
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
		"appointment", 
		map[string]string{
			"id": s.defaultAppointment.Id,
			"tenant_id": s.defaultAppointment.TenantId,
			"title": s.defaultAppointment.Title,
			"department_id": s.defaultAppointment.DepartmentId,
			"user_account_id": s.defaultAppointment.UserId,
			"start_date": s.defaultAppointment.StartDate,
	})

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"UNIQUE-VIOLATION-ERROR"`)	
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)		
}
