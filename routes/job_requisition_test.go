package routes

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"

	"multi-tenant-HR-information-system-backend/storage"
)

func (s *IntegrationTestSuite) TestCreateJobRequisition() {
	want := storage.JobRequisition{
		Id:              "cb180c6e-af87-4a97-9dcf-bcbe503414a7",
		TenantId:        s.defaultTenant.Id,
		Title:           "Database Administrator",
		DepartmentId:    s.defaultDepartment.Id,
		JobDescription:  "Manages databases of HRIS software",
		JobRequirements: "100 years of experience using postgres",
		Requestor:       s.defaultUser.Id,
		Supervisor:      s.defaultSupervisor.Id,
		HrApprover:      s.defaultHrApprover.Id,
	}

	type requestBody struct {
		Title           string
		DepartmentId    string
		JobDescription  string
		JobRequirements string
		Supervisor      string
		HrApprover      string
	}
	reqBody := requestBody{
		Title:           want.Title,
		DepartmentId:    want.DepartmentId,
		JobDescription:  want.JobDescription,
		JobRequirements: want.JobRequirements,
		Supervisor:      want.Supervisor,
		HrApprover:      want.HrApprover,
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
		map[string]string{
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
