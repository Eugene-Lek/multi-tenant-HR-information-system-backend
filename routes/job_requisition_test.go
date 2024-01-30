package routes

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/pquerna/otp/totp"

	"multi-tenant-HR-information-system-backend/storage"
)

func (s *IntegrationTestSuite) TestCreateJobRequisition() {
	want := storage.JobRequisition{
		Id:                    "cb180c6e-af87-4a97-9dcf-bcbe503414a7",
		TenantId:              s.defaultTenant.Id,
		Title:                 "Database Administrator",
		DepartmentId:          s.defaultDepartment.Id,
		SupervisorPositionIds: []string{s.defaultSupervisorPosition.Id},
		JobDescription:        "Manages databases of HRIS software",
		JobRequirements:       "100 years of experience using postgres",
		Requestor:             s.defaultUser.Id,
		Supervisor:            s.defaultSupervisor.Id,
		HrApprover:            s.defaultHrApprover.Id,
	}

	type requestBody struct {
		PositionId            string
		Title                 string
		DepartmentId          string
		SupervisorPositionIds []string
		JobDescription        string
		JobRequirements       string
		Supervisor            string
		HrApprover            string
	}
	reqBody := requestBody{
		PositionId:            want.PositionId,
		Title:                 want.Title,
		DepartmentId:          want.DepartmentId,
		SupervisorPositionIds: want.SupervisorPositionIds,
		JobDescription:        want.JobDescription,
		JobRequirements:       want.JobRequirements,
		Supervisor:            want.Supervisor,
		HrApprover:            want.HrApprover,
	}
	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)

	path := fmt.Sprintf("/api/tenants/%s/users/%s/job-requisitions/role-requestor/%s", want.TenantId, want.Requestor, want.Id)
	r, err := http.NewRequest("POST", path, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultUser.Id, s.defaultUser.TenantId, s.defaultUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 201)

	s.expectSelectQueryToReturnOneRow(
		"job_requisition",
		map[string]any{
			"id":               want.Id,
			"tenant_id":        want.TenantId,
			"title":            want.Title,
			"department_id":    want.DepartmentId,
			"job_description":  want.JobDescription,
			"job_requirements": want.JobRequirements,
			"requestor":        want.Requestor,
			"supervisor":       want.Supervisor,
			"hr_approver":      want.HrApprover,
		},
	)

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"JOB-REQUISITION-CREATED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestCreateJobRequisitionShouldValidateSupervisor() {
	want := storage.JobRequisition{
		Id:                    "cb180c6e-af87-4a97-9dcf-bcbe503414a7",
		TenantId:              s.defaultTenant.Id,
		PositionId:            "979e87ea-63f8-4cc1-8fa7-3555ffc41a0a",
		Title:                 "Database Administrator",
		DepartmentId:          s.defaultDepartment.Id,
		SupervisorPositionIds: []string{s.defaultSupervisorPosition.Id},
		JobDescription:        "Manages databases of HRIS software",
		JobRequirements:       "100 years of experience using postgres",
		Requestor:             s.defaultUser.Id,
		Supervisor:            s.defaultHrApprover.Id,
		HrApprover:            s.defaultHrApprover.Id,
	}

	type requestBody struct {
		PositionId            string
		Title                 string
		DepartmentId          string
		SupervisorPositionIds []string
		JobDescription        string
		JobRequirements       string
		Supervisor            string
		HrApprover            string
	}
	reqBody := requestBody{
		PositionId:            want.PositionId,
		Title:                 want.Title,
		DepartmentId:          want.DepartmentId,
		SupervisorPositionIds: want.SupervisorPositionIds,
		JobDescription:        want.JobDescription,
		JobRequirements:       want.JobRequirements,
		Supervisor:            want.Supervisor,
		HrApprover:            want.HrApprover,
	}
	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)

	path := fmt.Sprintf("/api/tenants/%s/users/%s/job-requisitions/role-requestor/%s", want.TenantId, want.Requestor, want.Id)
	r, err := http.NewRequest("POST", path, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultUser.Id, s.defaultUser.TenantId, s.defaultUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 400)
	s.expectErrorCode(w, "INVALID-SUPERVISOR-ERROR")

	s.expectSelectQueryToReturnNoRows(
		"job_requisition",
		map[string]any{
			"id":               want.Id,
			"tenant_id":        want.TenantId,
			"title":            want.Title,
			"department_id":    want.DepartmentId,
			"job_description":  want.JobDescription,
			"job_requirements": want.JobRequirements,
			"requestor":        want.Requestor,
			"supervisor":       want.Supervisor,
			"hr_approver":      want.HrApprover,
		},
	)

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"INVALID-SUPERVISOR-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestCreateJobRequisitionShouldValidateInput() {
	want := storage.JobRequisition{
		Id:                    "cb180c6e-af87-4a97-9dcf-bcbe503414a7",
		TenantId:              s.defaultTenant.Id,
		Title:                 "",
		DepartmentId:          s.defaultDepartment.Id,
		SupervisorPositionIds: []string{s.defaultSupervisorPosition.Id},
		JobDescription:        "Manages databases of HRIS software",
		JobRequirements:       "100 years of experience using postgres",
		Requestor:             s.defaultUser.Id,
		Supervisor:            s.defaultSupervisor.Id,
		HrApprover:            s.defaultHrApprover.Id,
	}

	type requestBody struct {
		PositionId            string
		Title                 string
		DepartmentId          string
		SupervisorPositionIds []string
		JobDescription        string
		JobRequirements       string
		Supervisor            string
		HrApprover            string
	}
	reqBody := requestBody{
		PositionId:            want.PositionId,
		Title:                 want.Title,
		DepartmentId:          want.DepartmentId,
		SupervisorPositionIds: want.SupervisorPositionIds,
		JobDescription:        want.JobDescription,
		JobRequirements:       want.JobRequirements,
		Supervisor:            want.Supervisor,
		HrApprover:            want.HrApprover,
	}
	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)

	path := fmt.Sprintf("/api/tenants/%s/users/%s/job-requisitions/role-requestor/%s", want.TenantId, want.Requestor, want.Id)
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
		"job_requisition",
		map[string]any{
			"id":               want.Id,
			"tenant_id":        want.TenantId,
			"title":            want.Title,
			"department_id":    want.DepartmentId,
			"job_description":  want.JobDescription,
			"job_requirements": want.JobRequirements,
			"requestor":        want.Requestor,
			"supervisor":       want.Supervisor,
			"hr_approver":      want.HrApprover,
		},
	)

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"INPUT-VALIDATION-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestCreateJobRequisitionShouldHandleModelsError() {
	want := storage.JobRequisition{
		Id:                    "cb180c6e-af87-4a97-9dcf-bcbe503414a7",
		TenantId:              s.defaultTenant.Id,
		PositionId:            "979e87ea-63f8-4cc1-8fa7-3555ffc41a0a",
		Title:                 "Database Administrator",
		DepartmentId:          "86a054aa-4597-4082-a95b-cfda716e40dd",
		SupervisorPositionIds: []string{s.defaultSupervisorPosition.Id},
		JobDescription:        "Manages databases of HRIS software",
		JobRequirements:       "100 years of experience using postgres",
		Requestor:             s.defaultUser.Id,
		Supervisor:            s.defaultSupervisor.Id,
		HrApprover:            s.defaultHrApprover.Id,
	}

	type requestBody struct {
		PositionId            string
		Title                 string
		DepartmentId          string
		SupervisorPositionIds []string
		JobDescription        string
		JobRequirements       string
		Supervisor            string
		HrApprover            string
	}
	reqBody := requestBody{
		PositionId:            want.PositionId,
		Title:                 want.Title,
		DepartmentId:          want.DepartmentId,
		SupervisorPositionIds: want.SupervisorPositionIds,
		JobDescription:        want.JobDescription,
		JobRequirements:       want.JobRequirements,
		Supervisor:            want.Supervisor,
		HrApprover:            want.HrApprover,
	}
	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)

	path := fmt.Sprintf("/api/tenants/%s/users/%s/job-requisitions/role-requestor/%s", want.TenantId, want.Requestor, want.Id)
	r, err := http.NewRequest("POST", path, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultUser.Id, s.defaultUser.TenantId, s.defaultUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 400)
	s.expectErrorCode(w, "INVALID-FOREIGN-KEY-ERROR")

	s.expectSelectQueryToReturnNoRows(
		"job_requisition",
		map[string]any{
			"id":               want.Id,
			"tenant_id":        want.TenantId,
			"title":            want.Title,
			"department_id":    want.DepartmentId,
			"job_description":  want.JobDescription,
			"job_requirements": want.JobRequirements,
			"requestor":        want.Requestor,
			"supervisor":       want.Supervisor,
			"hr_approver":      want.HrApprover,
		},
	)

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"INVALID-FOREIGN-KEY-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestSupervisorApproveJobRequisition() {
	want := storage.JobRequisition{
		Id:                 s.defaultJobRequisition.Id,
		TenantId:           s.defaultJobRequisition.TenantId,
		Supervisor:         s.defaultJobRequisition.Supervisor,
		SupervisorDecision: "APPROVED",
	}

	type requestBody struct {
		Password           string
		Totp               string
		SupervisorDecision string
	}
	otp, _ := totp.GenerateCode(s.defaultSupervisor.TotpSecretKey, time.Now().UTC())
	reqBody := requestBody{
		SupervisorDecision: want.SupervisorDecision,
		Password:           "jU%q837d!QP7",
		Totp:               otp,
	}
	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)

	path := fmt.Sprintf("/api/tenants/%s/users/%s/job-requisitions/role-supervisor/%s/supervisor-decision", want.TenantId, want.Supervisor, want.Id)
	r, err := http.NewRequest("POST", path, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultSupervisor.Id, s.defaultSupervisor.TenantId, s.defaultSupervisor.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 204)

	s.expectSelectQueryToReturnOneRow(
		"job_requisition",
		map[string]any{
			"id":                  want.Id,
			"tenant_id":           want.TenantId,
			"supervisor":          want.Supervisor,
			"supervisor_decision": want.SupervisorDecision,
		},
	)

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"JOB-REQUISITION-SUPERVISOR-APPROVED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestSupervisorApproveJobRequisitionShouldValidateIdExistence() {
	want := storage.JobRequisition{
		Id:                 "caaa7845-9601-4528-bd60-7cdae6cf298a",
		TenantId:           s.defaultJobRequisition.TenantId,
		Supervisor:         s.defaultJobRequisition.Supervisor,
		SupervisorDecision: "APPROVED",
	}

	type requestBody struct {
		Password           string
		Totp               string
		SupervisorDecision string
	}
	otp, _ := totp.GenerateCode(s.defaultSupervisor.TotpSecretKey, time.Now().UTC())
	reqBody := requestBody{
		SupervisorDecision: want.SupervisorDecision,
		Password:           "jU%q837d!QP7",
		Totp:               otp,
	}
	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)

	path := fmt.Sprintf("/api/tenants/%s/users/%s/job-requisitions/role-supervisor/%s/supervisor-decision", want.TenantId, want.Supervisor, want.Id)
	r, err := http.NewRequest("POST", path, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultSupervisor.Id, s.defaultSupervisor.TenantId, s.defaultSupervisor.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 404)
	s.expectErrorCode(w, "RESOURCE-NOT-FOUND-ERROR")

	s.expectSelectQueryToReturnNoRows(
		"job_requisition",
		map[string]any{
			"id":                  want.Id,
			"tenant_id":           want.TenantId,
			"supervisor":          want.Supervisor,
			"supervisor_decision": want.SupervisorDecision,
		},
	)

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"RESOURCE-NOT-FOUND-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestSupervisorRejectJobRequisition() {
	want := storage.JobRequisition{
		Id:                 s.defaultJobRequisition.Id,
		TenantId:           s.defaultTenant.Id,
		Supervisor:         s.defaultSupervisor.Id,
		SupervisorDecision: "REJECTED",
	}

	type requestBody struct {
		Password           string
		Totp               string
		SupervisorDecision string
	}
	otp, _ := totp.GenerateCode(s.defaultSupervisor.TotpSecretKey, time.Now().UTC())
	reqBody := requestBody{
		SupervisorDecision: want.SupervisorDecision,
		Password:           "jU%q837d!QP7",
		Totp:               otp,
	}
	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)

	path := fmt.Sprintf("/api/tenants/%s/users/%s/job-requisitions/role-supervisor/%s/supervisor-decision", want.TenantId, want.Supervisor, want.Id)
	r, err := http.NewRequest("POST", path, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultSupervisor.Id, s.defaultSupervisor.TenantId, s.defaultSupervisor.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 204)

	s.expectSelectQueryToReturnOneRow(
		"job_requisition",
		map[string]any{
			"id":                  want.Id,
			"tenant_id":           want.TenantId,
			"supervisor":          want.Supervisor,
			"supervisor_decision": want.SupervisorDecision,
		},
	)

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"JOB-REQUISITION-SUPERVISOR-REJECTED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestSupervisorApproveJobRequisitionShouldValidateCredentials() {
	want := storage.JobRequisition{
		Id:                 s.defaultJobRequisition.Id,
		TenantId:           s.defaultTenant.Id,
		Supervisor:         s.defaultSupervisor.Id,
		SupervisorDecision: "APPROVED",
	}

	type requestBody struct {
		Password           string
		Totp               string
		SupervisorDecision string
	}
	otp, _ := totp.GenerateCode(s.defaultSupervisor.TotpSecretKey, time.Now().UTC())
	reqBody := requestBody{
		SupervisorDecision: want.SupervisorDecision,
		Password:           "wrong password",
		Totp:               otp,
	}
	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)

	path := fmt.Sprintf("/api/tenants/%s/users/%s/job-requisitions/role-supervisor/%s/supervisor-decision", want.TenantId, want.Supervisor, want.Id)
	r, err := http.NewRequest("POST", path, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultSupervisor.Id, s.defaultSupervisor.TenantId, s.defaultSupervisor.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 401)
	s.expectErrorCode(w, "USER-UNAUTHENTICATED")

	s.expectSelectQueryToReturnNoRows(
		"job_requisition",
		map[string]any{
			"id":                  want.Id,
			"tenant_id":           want.TenantId,
			"supervisor":          want.Supervisor,
			"supervisor_decision": want.SupervisorDecision,
		},
	)

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"USER-UNAUTHENTICATED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestSupervisorApproveJobRequisitionShouldValidateSupervisor() {
	// Change the supervisor's position assignment's end date to before today.
	// This will simulate the scenario where the supervisor was reassigned to another role after
	// the job requisition had been created & therefore should no longer be authorised to approve this requisition
	query := "UPDATE position_assignment SET end_date = CURRENT_DATE - 1  WHERE user_account_id = $1"
	_, err := s.dbRootConn.Exec(query, s.defaultSupervisor.Id)
	if err != nil {
		log.Fatalf("Could not alter supervisor's position assignment end date: %s", err)
	}

	want := storage.JobRequisition{
		Id:                 s.defaultJobRequisition.Id,
		TenantId:           s.defaultTenant.Id,
		Supervisor:         s.defaultSupervisor.Id,
		SupervisorDecision: "APPROVED",
	}

	type requestBody struct {
		Password           string
		Totp               string
		SupervisorDecision string
	}
	otp, _ := totp.GenerateCode(s.defaultSupervisor.TotpSecretKey, time.Now().UTC())
	reqBody := requestBody{
		SupervisorDecision: want.SupervisorDecision,
		Password:           "jU%q837d!QP7",
		Totp:               otp,
	}
	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)

	path := fmt.Sprintf("/api/tenants/%s/users/%s/job-requisitions/role-supervisor/%s/supervisor-decision", want.TenantId, want.Supervisor, want.Id)
	r, err := http.NewRequest("POST", path, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultSupervisor.Id, s.defaultSupervisor.TenantId, s.defaultSupervisor.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 403)
	s.expectErrorCode(w, "USER-UNAUTHORISED")

	s.expectSelectQueryToReturnNoRows(
		"job_requisition",
		map[string]any{
			"id":                  want.Id,
			"tenant_id":           want.TenantId,
			"supervisor":          want.Supervisor,
			"supervisor_decision": want.SupervisorDecision,
		},
	)

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"USER-UNAUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestSupervisorApproveJobRequisitionShouldValidateInput() {
	want := storage.JobRequisition{
		Id:                 s.defaultJobRequisition.Id,
		TenantId:           s.defaultTenant.Id,
		Supervisor:         s.defaultSupervisor.Id,
		SupervisorDecision: "APPROVED",
	}

	type requestBody struct {
		Password           string
		Totp               string
		SupervisorDecision string
	}
	otp, _ := totp.GenerateCode(s.defaultSupervisor.TotpSecretKey, time.Now().UTC())
	reqBody := requestBody{
		SupervisorDecision: want.SupervisorDecision,
		Password:           "",
		Totp:               otp,
	}
	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)

	path := fmt.Sprintf("/api/tenants/%s/users/%s/job-requisitions/role-supervisor/%s/supervisor-decision", want.TenantId, want.Supervisor, want.Id)
	r, err := http.NewRequest("POST", path, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultSupervisor.Id, s.defaultSupervisor.TenantId, s.defaultSupervisor.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 400)
	s.expectErrorCode(w, "INPUT-VALIDATION-ERROR")

	s.expectSelectQueryToReturnNoRows(
		"job_requisition",
		map[string]any{
			"id":                  want.Id,
			"tenant_id":           want.TenantId,
			"supervisor":          want.Supervisor,
			"supervisor_decision": want.SupervisorDecision,
		},
	)

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"INPUT-VALIDATION-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestSupervisorApproveJobRequisitionShouldPreventIdExploit() {
	// For performance reasons, authorization to approve job requisitions is not given on a job requisition ID basis
	// As such, the handler must compare the user id (from session) against the superior id

	// To test this, we simulate the scenario where the hr approver tries to update the job requisition as the supervisor.
	// The hr approver would bypass the authorization middleware because he has authorization to approve job reqs as a superior too

	want := storage.JobRequisition{
		Id:                 s.defaultJobRequisition.Id,
		TenantId:           s.defaultTenant.Id,
		Supervisor:         s.defaultHrApprover.Id,
		SupervisorDecision: "APPROVED",
	}

	type requestBody struct {
		Password           string
		Totp               string
		SupervisorDecision string
	}
	otp, _ := totp.GenerateCode(s.defaultSupervisor.TotpSecretKey, time.Now().UTC())
	reqBody := requestBody{
		SupervisorDecision: want.SupervisorDecision,
		Password:           "jU%q837d!QP7",
		Totp:               otp,
	}
	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)

	path := fmt.Sprintf("/api/tenants/%s/users/%s/job-requisitions/role-supervisor/%s/supervisor-decision", want.TenantId, want.Supervisor, want.Id)
	r, err := http.NewRequest("POST", path, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultHrApprover.Id, s.defaultHrApprover.TenantId, s.defaultHrApprover.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 403)
	s.expectErrorCode(w, "USER-UNAUTHORISED")

	s.expectSelectQueryToReturnNoRows(
		"job_requisition",
		map[string]any{
			"id":                  want.Id,
			"tenant_id":           want.TenantId,
			"supervisor":          want.Supervisor,
			"supervisor_decision": want.SupervisorDecision,
		},
	)

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"USER-UNAUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestHrApproveJobRequisitionForNewPosition() {
	want := storage.JobRequisition{
		Id:                 s.defaultJobRequisition.Id,
		TenantId:           s.defaultJobRequisition.TenantId,
		HrApprover:         s.defaultJobRequisition.HrApprover,
		HrApproverDecision: "APPROVED",
		Recruiter:          s.defaultRecruiter.Id,
	}

	// Add supervisor approval to the default job requisition
	_, err := s.dbRootConn.Exec("UPDATE job_requisition SET supervisor_decision = 'APPROVED' WHERE id = $1", want.Id)
	if err != nil {
		log.Fatalf("Could not manually seed supervisor approval: %s", err)
	}

	type requestBody struct {
		Password           string
		Totp               string
		HrApproverDecision string
		Recruiter          string
	}
	otp, _ := totp.GenerateCode(s.defaultHrApprover.TotpSecretKey, time.Now().UTC())
	reqBody := requestBody{
		HrApproverDecision: want.HrApproverDecision,
		Recruiter:          want.Recruiter,
		Password:           "jU%q837d!QP7",
		Totp:               otp,
	}
	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)

	path := fmt.Sprintf("/api/tenants/%s/users/%s/job-requisitions/role-hr-approver/%s/hr-approver-decision", want.TenantId, want.HrApprover, want.Id)
	r, err := http.NewRequest("POST", path, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultHrApprover.Id, s.defaultHrApprover.TenantId, s.defaultHrApprover.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 204)

	s.expectSelectQueryToReturnOneRow(
		"job_requisition",
		map[string]any{
			"id":                   want.Id,
			"tenant_id":            want.TenantId,
			"hr_approver":          want.HrApprover,
			"hr_approver_decision": want.HrApproverDecision,
			"recruiter":            want.Recruiter,
		},
	)

	var positionId string
	s.dbRootConn.QueryRow("SELECT position_id FROM job_requisition WHERE id = $1", want.Id).Scan(&positionId)
	if positionId != "" {
		s.expectSelectQueryToReturnOneRow(
			"position",
			map[string]any{
				"id":            positionId,
				"tenant_id":     want.TenantId,
			},
		)
	}

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"JOB-REQUISITION-HR-APPROVED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestHrApproveJobRequisitionForExistingPosition() {
	jobReqExistingPosition := storage.JobRequisition{
		Id:                 "69805a4f-8bfe-4ca7-bfa9-818682be1fa2",
		TenantId:           "2ad1dcfc-8867-49f7-87a3-8bd8d1154924",
		PositionId:         s.defaultPosition.Id,
		JobDescription:     "Manages databases of HRIS software",
		JobRequirements:    "100 years of experience using postgres",
		Requestor:          "e7f31b70-ae26-42b3-b7a6-01ec68d5c33a",
		Supervisor:         "38d3f831-9a9e-4dfc-ba56-ec68bf2462e0",
		SupervisorDecision: "APPROVED",
		HrApprover:         "9f4c9dd0-7c75-4ea9-a106-948885b6bedf",
	}

	input := storage.JobRequisition{
		Id:                 "4e105cc7-46a1-43b7-b9fa-f6c11d5feb74",
		TenantId:           "2ad1dcfc-8867-49f7-87a3-8bd8d1154924",
		HrApprover:         "9f4c9dd0-7c75-4ea9-a106-948885b6bedf",
		HrApproverDecision: "APPROVED",
		Recruiter:          "ccb2da3b-68ac-419e-b95d-dd6b723035f9",
	}

	// Create a job requisition for an existing position
	query := `INSERT INTO job_requisition (id, tenant_id, position_id, job_description, job_requirements, requestor, supervisor, supervisor_decision, hr_approver)
				 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := s.dbRootConn.Exec(query, jobReqExistingPosition.Id, jobReqExistingPosition.TenantId, jobReqExistingPosition.PositionId,
		jobReqExistingPosition.JobDescription, jobReqExistingPosition.JobRequirements, jobReqExistingPosition.Requestor,
		jobReqExistingPosition.Supervisor, jobReqExistingPosition.SupervisorDecision, jobReqExistingPosition.HrApprover)
	if err != nil {
		log.Fatalf("Could not manually seed existing position job requisition: %s", err)
	}

	type requestBody struct {
		Password           string
		Totp               string
		HrApproverDecision string
		Recruiter          string
	}
	otp, _ := totp.GenerateCode(s.defaultHrApprover.TotpSecretKey, time.Now().UTC())
	reqBody := requestBody{
		HrApproverDecision: input.HrApproverDecision,
		Recruiter:          input.Recruiter,
		Password:           "jU%q837d!QP7",
		Totp:               otp,
	}
	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)

	path := fmt.Sprintf("/api/tenants/%s/users/%s/job-requisitions/role-hr-approver/%s/hr-approver-decision", input.TenantId, input.HrApprover, input.Id)
	r, err := http.NewRequest("POST", path, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultHrApprover.Id, s.defaultHrApprover.TenantId, s.defaultHrApprover.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 204)

	s.expectSelectQueryToReturnOneRow(
		"job_requisition",
		map[string]any{
			"id":                   input.Id,
			"tenant_id":            input.TenantId,
			"hr_approver":          input.HrApprover,
			"hr_approver_decision": input.HrApproverDecision,
			"recruiter":            input.Recruiter,
		},
	)

	s.expectSelectQueryToReturnOneRow(
		"position",
		map[string]any{
			"id":            jobReqExistingPosition.PositionId,
			"tenant_id":     jobReqExistingPosition.TenantId,
		},
	)

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"JOB-REQUISITION-HR-APPROVED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestHrApproveJobRequisitionShouldValidateIdExistence() {
	want := storage.JobRequisition{
		Id:                    "781b3b84-9c4e-4319-abbe-df2b34c33cd7",
		TenantId:              s.defaultJobRequisition.TenantId,
		HrApprover:            s.defaultJobRequisition.HrApprover,
		HrApproverDecision:    "APPROVED",
		Recruiter:             s.defaultRecruiter.Id,
	}

	type requestBody struct {
		Password           string
		Totp               string
		HrApproverDecision string
		Recruiter          string
	}
	otp, _ := totp.GenerateCode(s.defaultHrApprover.TotpSecretKey, time.Now().UTC())
	reqBody := requestBody{
		HrApproverDecision: want.HrApproverDecision,
		Recruiter:          want.Recruiter,
		Password:           "jU%q837d!QP7",
		Totp:               otp,
	}
	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)

	path := fmt.Sprintf("/api/tenants/%s/users/%s/job-requisitions/role-hr-approver/%s/hr-approver-decision", want.TenantId, want.HrApprover, want.Id)
	r, err := http.NewRequest("POST", path, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultHrApprover.Id, s.defaultHrApprover.TenantId, s.defaultHrApprover.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 404)
	s.expectErrorCode(w, "RESOURCE-NOT-FOUND-ERROR")

	s.expectSelectQueryToReturnNoRows(
		"job_requisition",
		map[string]any{
			"id":                   want.Id,
			"tenant_id":            want.TenantId,
			"hr_approver":          want.HrApprover,
			"hr_approver_decision": want.HrApproverDecision,
			"recruiter":            want.Recruiter,
		},
	)

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"RESOURCE-NOT-FOUND-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestHrRejectJobRequisition() {
	want := storage.JobRequisition{
		Id:                    s.defaultJobRequisition.Id,
		TenantId:              s.defaultJobRequisition.TenantId,
		HrApprover:            s.defaultJobRequisition.HrApprover,
		HrApproverDecision:    "REJECTED",
	}

	// Add supervisor approval to the default job requisition
	_, err := s.dbRootConn.Exec("UPDATE job_requisition SET supervisor_decision = 'APPROVED' WHERE id = $1", want.Id)
	if err != nil {
		log.Fatalf("Could not manually seed supervisor approval: %s", err)
	}

	type requestBody struct {
		Password           string
		Totp               string
		HrApproverDecision string
		Recruiter          string
	}
	otp, _ := totp.GenerateCode(s.defaultHrApprover.TotpSecretKey, time.Now().UTC())
	reqBody := requestBody{
		HrApproverDecision: want.HrApproverDecision,
		Recruiter:          want.Recruiter,
		Password:           "jU%q837d!QP7",
		Totp:               otp,
	}
	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)

	path := fmt.Sprintf("/api/tenants/%s/users/%s/job-requisitions/role-hr-approver/%s/hr-approver-decision", want.TenantId, want.HrApprover, want.Id)
	r, err := http.NewRequest("POST", path, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultHrApprover.Id, s.defaultHrApprover.TenantId, s.defaultHrApprover.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 204)

	s.expectSelectQueryToReturnOneRow(
		"job_requisition",
		map[string]any{
			"id":                   want.Id,
			"tenant_id":            want.TenantId,
			"position_id":          "", // position_id = null since it should not be created
			"hr_approver":          want.HrApprover,
			"hr_approver_decision": want.HrApproverDecision,
		},
	)

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"JOB-REQUISITION-HR-REJECTED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestHrApproveJobRequisitionShouldFailIfBeforeSupervisorApproval() {
	want := storage.JobRequisition{
		Id:                    s.defaultJobRequisition.Id,
		TenantId:              s.defaultJobRequisition.TenantId,
		HrApprover:            s.defaultJobRequisition.HrApprover,
		HrApproverDecision:    "APPROVED",
		Recruiter:             s.defaultRecruiter.Id,
	}

	type requestBody struct {
		Password           string
		Totp               string
		HrApproverDecision string
		Recruiter          string
	}
	otp, _ := totp.GenerateCode(s.defaultHrApprover.TotpSecretKey, time.Now().UTC())
	reqBody := requestBody{
		HrApproverDecision: want.HrApproverDecision,
		Recruiter:          want.Recruiter,
		Password:           "jU%q837d!QP7",
		Totp:               otp,
	}
	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)

	path := fmt.Sprintf("/api/tenants/%s/users/%s/job-requisitions/role-hr-approver/%s/hr-approver-decision", want.TenantId, want.HrApprover, want.Id)
	r, err := http.NewRequest("POST", path, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultHrApprover.Id, s.defaultHrApprover.TenantId, s.defaultHrApprover.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 403)
	s.expectErrorCode(w, "MISSING-SUPERVISOR-APPROVAL-ERROR")

	s.expectSelectQueryToReturnNoRows(
		"job_requisition",
		map[string]any{
			"id":                   want.Id,
			"tenant_id":            want.TenantId,
			"hr_approver":          want.HrApprover,
			"hr_approver_decision": want.HrApproverDecision,
			"recruiter":            want.Recruiter,
		},
	)

	s.expectSelectQueryToReturnOneRow(
		"job_requisition",
		map[string]any{
			"id":          want.Id,
			"tenant_id":   want.TenantId,
			"position_id": "", // position_id = null since it should not be created
		},
	)

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"MISSING-SUPERVISOR-APPROVAL-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestHrApproveJobRequisitionShouldValidateInput() {
	want := storage.JobRequisition{
		Id:                    s.defaultJobRequisition.Id,
		TenantId:              s.defaultJobRequisition.TenantId,
		HrApprover:            s.defaultJobRequisition.HrApprover,
		HrApproverDecision:    "APPROVED",
		Recruiter:             s.defaultRecruiter.Id,
	}

	// Add supervisor approval to the default job requisition
	_, err := s.dbRootConn.Exec("UPDATE job_requisition SET supervisor_decision = 'APPROVED' WHERE id = $1", want.Id)
	if err != nil {
		log.Fatalf("Could not manually seed supervisor approval: %s", err)
	}

	type requestBody struct {
		Password           string
		Totp               string
		HrApproverDecision string
		Recruiter          string
	}
	otp, _ := totp.GenerateCode(s.defaultHrApprover.TotpSecretKey, time.Now().UTC())
	reqBody := requestBody{
		HrApproverDecision: want.HrApproverDecision,
		Recruiter:          want.Recruiter,
		Password:           "",
		Totp:               otp,
	}
	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)

	path := fmt.Sprintf("/api/tenants/%s/users/%s/job-requisitions/role-hr-approver/%s/hr-approver-decision", want.TenantId, want.HrApprover, want.Id)
	r, err := http.NewRequest("POST", path, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultHrApprover.Id, s.defaultHrApprover.TenantId, s.defaultHrApprover.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 400)
	s.expectErrorCode(w, "INPUT-VALIDATION-ERROR")

	s.expectSelectQueryToReturnNoRows(
		"job_requisition",
		map[string]any{
			"id":                   want.Id,
			"tenant_id":            want.TenantId,
			"hr_approver":          want.HrApprover,
			"hr_approver_decision": want.HrApproverDecision,
			"recruiter":            want.Recruiter,
		},
	)

	s.expectSelectQueryToReturnOneRow(
		"job_requisition",
		map[string]any{
			"id":          want.Id,
			"tenant_id":   want.TenantId,
			"position_id": "", // position_id = null since it should not be created
		},
	)

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"INPUT-VALIDATION-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestHrApproveJobRequisitionShouldValidateCredentials() {
	want := storage.JobRequisition{
		Id:                    s.defaultJobRequisition.Id,
		TenantId:              s.defaultJobRequisition.TenantId,
		HrApprover:            s.defaultJobRequisition.HrApprover,
		HrApproverDecision:    "APPROVED",
		Recruiter:             s.defaultRecruiter.Id,
	}

	// Add supervisor approval to the default job requisition
	_, err := s.dbRootConn.Exec("UPDATE job_requisition SET supervisor_decision = 'APPROVED' WHERE id = $1", want.Id)
	if err != nil {
		log.Fatalf("Could not manually seed supervisor approval: %s", err)
	}

	type requestBody struct {
		Password           string
		Totp               string
		HrApproverDecision string
		Recruiter          string
	}
	otp, _ := totp.GenerateCode(s.defaultHrApprover.TotpSecretKey, time.Now().UTC())
	reqBody := requestBody{
		HrApproverDecision: want.HrApproverDecision,
		Recruiter:          want.Recruiter,
		Password:           "invalid password",
		Totp:               otp,
	}
	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)

	path := fmt.Sprintf("/api/tenants/%s/users/%s/job-requisitions/role-hr-approver/%s/hr-approver-decision", want.TenantId, want.HrApprover, want.Id)
	r, err := http.NewRequest("POST", path, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultHrApprover.Id, s.defaultHrApprover.TenantId, s.defaultHrApprover.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 401)
	s.expectErrorCode(w, "USER-UNAUTHENTICATED")

	s.expectSelectQueryToReturnNoRows(
		"job_requisition",
		map[string]any{
			"id":                   want.Id,
			"tenant_id":            want.TenantId,
			"hr_approver":          want.HrApprover,
			"hr_approver_decision": want.HrApproverDecision,
			"recruiter":            want.Recruiter,
		},
	)

	s.expectSelectQueryToReturnOneRow(
		"job_requisition",
		map[string]any{
			"id":          want.Id,
			"tenant_id":   want.TenantId,
			"position_id": "", // position_id = null since it should not be created
		},
	)

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"USER-UNAUTHENTICATED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestHrApproveJobRequisitionShouldPreventIdExploit() {
	// In this test, we simulate the scenario where the supervisor is trying to approve the requisiton on the HR approver's
	// behalf by using the same Id except with the HR approval url
	// In this scenario, the supervisor can bypass the authorization middleware because he also has the hr approval policy
	// Even so, it should still fail because job requisitions are filtered by hr approver when updated.
	// As such, the job requisition will not be found

	want := storage.JobRequisition{
		Id:                    s.defaultJobRequisition.Id,
		TenantId:              s.defaultJobRequisition.TenantId,
		HrApprover:            s.defaultJobRequisition.Supervisor,
		HrApproverDecision:    "APPROVED",
		Recruiter:             s.defaultRecruiter.Id,
	}

	// Add supervisor approval to the default job requisition
	_, err := s.dbRootConn.Exec("UPDATE job_requisition SET supervisor_decision = 'APPROVED' WHERE id = $1", want.Id)
	if err != nil {
		log.Fatalf("Could not manually seed supervisor approval: %s", err)
	}

	type requestBody struct {
		Password           string
		Totp               string
		HrApproverDecision string
		Recruiter          string
	}
	otp, _ := totp.GenerateCode(s.defaultSupervisor.TotpSecretKey, time.Now().UTC())
	reqBody := requestBody{
		HrApproverDecision: want.HrApproverDecision,
		Recruiter:          want.Recruiter,
		Password:           "jU%q837d!QP7",
		Totp:               otp,
	}
	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)

	path := fmt.Sprintf("/api/tenants/%s/users/%s/job-requisitions/role-hr-approver/%s/hr-approver-decision", want.TenantId, want.HrApprover, want.Id)
	r, err := http.NewRequest("POST", path, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultSupervisor.Id, s.defaultSupervisor.TenantId, s.defaultSupervisor.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 404)
	s.expectErrorCode(w, "RESOURCE-NOT-FOUND-ERROR")

	s.expectSelectQueryToReturnNoRows(
		"job_requisition",
		map[string]any{
			"id":                   want.Id,
			"tenant_id":            want.TenantId,
			"hr_approver":          want.HrApprover,
			"hr_approver_decision": want.HrApproverDecision,
			"recruiter":            want.Recruiter,
		},
	)

	s.expectSelectQueryToReturnOneRow(
		"job_requisition",
		map[string]any{
			"id":          want.Id,
			"tenant_id":   want.TenantId,
			"position_id": "", // position_id = null since it should not be created
		},
	)

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"RESOURCE-NOT-FOUND-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}
