package routes

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	"multi-tenant-HR-information-system-backend/storage"
	"multi-tenant-HR-information-system-backend/storage/s3mock"
)

func (s *IntegrationTestSuite) TestCreateJobApplication() {
	input := storage.JobApplication{
		Id:               "371bcf41-2a2b-4fb4-bc00-f8a1f248d756",
		TenantId:         s.defaultTenant.Id,
		JobRequisitionId: s.defaultApprovedJobRequisition.Id,
		FirstName:        "Eugene",
		LastName:         "Lek",
		CountryCode:      "1",
		PhoneNumber:      "987654321",
		Email:            "test123@gmail.com",
		ResumeS3Url:      fmt.Sprintf("%s/hr-information-system/job-applications/371bcf41-2a2b-4fb4-bc00-f8a1f248d756/Eugene_Lek_resume.pdf", s.s3Server.URL),
	}

	bodyBuf := new(bytes.Buffer)
	multiWriter := multipart.NewWriter(bodyBuf)

	type requestBody struct {
		JobRequisitionId string
		FirstName        string
		LastName         string
		CountryCode      string
		PhoneNumber      string
		Email            string
	}
	reqBody := requestBody{
		JobRequisitionId: input.JobRequisitionId,
		FirstName:        input.FirstName,
		LastName:         input.LastName,
		CountryCode:      input.CountryCode,
		PhoneNumber:      input.PhoneNumber,
		Email:            input.Email,
	}
	dw, err := multiWriter.CreateFormField("data")
	if err != nil {
		log.Fatalf("Could not create data body: %s", err)
	}
	err = json.NewEncoder(dw).Encode(reqBody)
	if err != nil {
		log.Fatalf("Could not encode json body: %s", err)
	}

	file, err := os.Open("../test_resume.pdf")
	if err != nil {
		log.Fatalf("Could not open test resume: %s", err)
	}
	defer file.Close()
	fw, err := multiWriter.CreateFormFile("resume", "test_resume.pdf")
	if err != nil {
		log.Fatalf("Could not create file body: %s", err)
	}
	// Copy the contents of the file to the form field
	if _, err := io.Copy(fw, file); err != nil {
		log.Fatal(err)
	}

	multiWriter.Close()

	path := fmt.Sprintf("/api/tenants/%s/job-applications/%s", input.TenantId, input.Id)
	r, err := http.NewRequest("POST", path, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	r.Header.Set("Content-Type", multiWriter.FormDataContentType())

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 201)

	s.expectS3ToContainFile(input.ResumeS3Url)

	s.expectSelectQueryToReturnOneRow(
		"job_application",
		map[string]any{
			"id":                 input.Id,
			"tenant_id":          input.TenantId,
			"job_requisition_id": input.JobRequisitionId,
			"first_name":         input.FirstName,
			"last_name":          input.LastName,
			"country_code":       input.CountryCode,
			"phone_number":       input.PhoneNumber,
			"email":              input.Email,
			"resume_s3_url":      input.ResumeS3Url,
		},
	)

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"RESUME-UPLOADED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"JOB-APPLICATION-CREATED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestCreateJobApplicationValidatesInput() {
	input := storage.JobApplication{
		Id:               "371bcf41-2a2b-4fb4-bc00-f8a1f248d756",
		TenantId:         s.defaultTenant.Id,
		JobRequisitionId: s.defaultApprovedJobRequisition.Id,
		FirstName:        "Eugene",
		LastName:         "Lek",
		CountryCode:      "1",
		PhoneNumber:      "987654321",
		Email:            "test123gmail.com",
		ResumeS3Url:      fmt.Sprintf("%s/hr-information-system/job-applications/371bcf41-2a2b-4fb4-bc00-f8a1f248d756/Eugene_Lek_resume.pdf", s.s3Server.URL),
	}

	bodyBuf := new(bytes.Buffer)
	multiWriter := multipart.NewWriter(bodyBuf)

	type requestBody struct {
		JobRequisitionId string
		FirstName        string
		LastName         string
		CountryCode      string
		PhoneNumber      string
		Email            string
	}
	reqBody := requestBody{
		JobRequisitionId: input.JobRequisitionId,
		FirstName:        input.FirstName,
		LastName:         input.LastName,
		CountryCode:      input.CountryCode,
		PhoneNumber:      input.PhoneNumber,
		Email:            input.Email,
	}
	dw, err := multiWriter.CreateFormField("data")
	if err != nil {
		log.Fatalf("Could not create data body: %s", err)
	}
	err = json.NewEncoder(dw).Encode(reqBody)
	if err != nil {
		log.Fatalf("Could not encode json body: %s", err)
	}

	file, err := os.Open("../test_resume.pdf")
	if err != nil {
		log.Fatalf("Could not open test resume: %s", err)
	}
	defer file.Close()
	fw, err := multiWriter.CreateFormFile("resume", "test_resume.pdf")
	if err != nil {
		log.Fatalf("Could not create file body: %s", err)
	}
	// Copy the contents of the file to the form field
	if _, err := io.Copy(fw, file); err != nil {
		log.Fatal(err)
	}

	multiWriter.Close()

	path := fmt.Sprintf("/api/tenants/%s/job-applications/%s", input.TenantId, input.Id)
	r, err := http.NewRequest("POST", path, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	r.Header.Set("Content-Type", multiWriter.FormDataContentType())

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 400)
	s.expectErrorCode(w, "INPUT-VALIDATION-ERROR")

	s.expectS3ToNotContainFile(input.ResumeS3Url)

	s.expectSelectQueryToReturnNoRows(
		"job_application",
		map[string]any{
			"id":                 input.Id,
			"tenant_id":          input.TenantId,
			"job_requisition_id": input.JobRequisitionId,
			"first_name":         input.FirstName,
			"last_name":          input.LastName,
			"country_code":       input.CountryCode,
			"phone_number":       input.PhoneNumber,
			"email":              input.Email,
			"resume_s3_url":      input.ResumeS3Url,
		},
	)

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"INPUT-VALIDATION-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestCreateJobApplicationValidatesFileSize() {
	input := storage.JobApplication{
		Id:               "371bcf41-2a2b-4fb4-bc00-f8a1f248d756",
		TenantId:         s.defaultTenant.Id,
		JobRequisitionId: s.defaultApprovedJobRequisition.Id,
		FirstName:        "Eugene",
		LastName:         "Lek",
		CountryCode:      "1",
		PhoneNumber:      "987654321",
		Email:            "test123@gmail.com",
		ResumeS3Url:      fmt.Sprintf("%s/hr-information-system/job-applications/371bcf41-2a2b-4fb4-bc00-f8a1f248d756/Eugene_Lek_resume.pdf", s.s3Server.URL),
	}

	bodyBuf := new(bytes.Buffer)
	multiWriter := multipart.NewWriter(bodyBuf)

	type requestBody struct {
		JobRequisitionId string
		FirstName        string
		LastName         string
		CountryCode      string
		PhoneNumber      string
		Email            string
	}
	reqBody := requestBody{
		JobRequisitionId: input.JobRequisitionId,
		FirstName:        input.FirstName,
		LastName:         input.LastName,
		CountryCode:      input.CountryCode,
		PhoneNumber:      input.PhoneNumber,
		Email:            input.Email,
	}
	dw, err := multiWriter.CreateFormField("data")
	if err != nil {
		log.Fatalf("Could not create data body: %s", err)
	}
	err = json.NewEncoder(dw).Encode(reqBody)
	if err != nil {
		log.Fatalf("Could not encode json body: %s", err)
	}

	file, err := os.Open("../test_resume_too_big.pdf")
	if err != nil {
		log.Fatalf("Could not open test resume: %s", err)
	}
	defer file.Close()
	fw, err := multiWriter.CreateFormFile("resume", "test_resume_too_big.pdf")
	if err != nil {
		log.Fatalf("Could not create file body: %s", err)
	}
	// Copy the contents of the file to the form field
	if _, err := io.Copy(fw, file); err != nil {
		log.Fatal(err)
	}

	multiWriter.Close()

	path := fmt.Sprintf("/api/tenants/%s/job-applications/%s", input.TenantId, input.Id)
	r, err := http.NewRequest("POST", path, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	r.Header.Set("Content-Type", multiWriter.FormDataContentType())

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 400)
	s.expectErrorCode(w, "FILE-TOO-BIG-ERROR")

	s.expectS3ToNotContainFile(input.ResumeS3Url)

	s.expectSelectQueryToReturnNoRows(
		"job_application",
		map[string]any{
			"id":                 input.Id,
			"tenant_id":          input.TenantId,
			"job_requisition_id": input.JobRequisitionId,
			"first_name":         input.FirstName,
			"last_name":          input.LastName,
			"country_code":       input.CountryCode,
			"phone_number":       input.PhoneNumber,
			"email":              input.Email,
			"resume_s3_url":      input.ResumeS3Url,
		},
	)

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"FILE-TOO-BIG-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestCreateJobApplicationValidatesJobRequisition() {
	input := storage.JobApplication{
		Id:               "371bcf41-2a2b-4fb4-bc00-f8a1f248d756",
		TenantId:         s.defaultTenant.Id,
		JobRequisitionId: s.defaultApprovedJobRequisition.Id,
		FirstName:        "Eugene",
		LastName:         "Lek",
		CountryCode:      "1",
		PhoneNumber:      "987654321",
		Email:            "test123@gmail.com",
		ResumeS3Url:      fmt.Sprintf("%s/hr-information-system/job-applications/371bcf41-2a2b-4fb4-bc00-f8a1f248d756/Eugene_Lek_resume.pdf", s.s3Server.URL),
	}

	// Remove HR approval from the job requisition
	query := "UPDATE job_requisition SET hr_approver_decision = 'REJECTED' WHERE id = $1"
	_, err := s.dbRootConn.Exec(query, input.JobRequisitionId)
	if err != nil {
		log.Fatalf("Could not remove hr approval: %s", err)
	}

	bodyBuf := new(bytes.Buffer)
	multiWriter := multipart.NewWriter(bodyBuf)

	type requestBody struct {
		JobRequisitionId string
		FirstName        string
		LastName         string
		CountryCode      string
		PhoneNumber      string
		Email            string
	}
	reqBody := requestBody{
		JobRequisitionId: input.JobRequisitionId,
		FirstName:        input.FirstName,
		LastName:         input.LastName,
		CountryCode:      input.CountryCode,
		PhoneNumber:      input.PhoneNumber,
		Email:            input.Email,
	}
	dw, err := multiWriter.CreateFormField("data")
	if err != nil {
		log.Fatalf("Could not create data body: %s", err)
	}
	err = json.NewEncoder(dw).Encode(reqBody)
	if err != nil {
		log.Fatalf("Could not encode json body: %s", err)
	}

	file, err := os.Open("../test_resume.pdf")
	if err != nil {
		log.Fatalf("Could not open test resume: %s", err)
	}
	defer file.Close()
	fw, err := multiWriter.CreateFormFile("resume", "test_resume.pdf")
	if err != nil {
		log.Fatalf("Could not create file body: %s", err)
	}
	// Copy the contents of the file to the form field
	if _, err := io.Copy(fw, file); err != nil {
		log.Fatal(err)
	}

	multiWriter.Close()

	path := fmt.Sprintf("/api/tenants/%s/job-applications/%s", input.TenantId, input.Id)
	r, err := http.NewRequest("POST", path, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	r.Header.Set("Content-Type", multiWriter.FormDataContentType())

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 403)
	s.expectErrorCode(w, "MISSING-HR-APPROVAL-ERROR")

	s.expectS3ToNotContainFile(input.ResumeS3Url)

	s.expectSelectQueryToReturnNoRows(
		"job_application",
		map[string]any{
			"id":                 input.Id,
			"tenant_id":          input.TenantId,
			"job_requisition_id": input.JobRequisitionId,
			"first_name":         input.FirstName,
			"last_name":          input.LastName,
			"country_code":       input.CountryCode,
			"phone_number":       input.PhoneNumber,
			"email":              input.Email,
			"resume_s3_url":      input.ResumeS3Url,
		},
	)

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"MISSING-HR-APPROVAL-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestCreateJobApplicationHandlesS3Error() {
	input := storage.JobApplication{
		Id:               "371bcf41-2a2b-4fb4-bc00-f8a1f248d756",
		TenantId:         s.defaultTenant.Id,
		JobRequisitionId: s.defaultApprovedJobRequisition.Id,
		FirstName:        "Eugene",
		LastName:         "Lek",
		CountryCode:      "1",
		PhoneNumber:      "987654321",
		Email:            "test123@gmail.com",
		ResumeS3Url:      fmt.Sprintf("%s/hr-information-system/job-applications/371bcf41-2a2b-4fb4-bc00-f8a1f248d756/Eugene_Lek_resume.pdf", s.s3Server.URL),
	}

	bodyBuf := new(bytes.Buffer)
	multiWriter := multipart.NewWriter(bodyBuf)

	type requestBody struct {
		JobRequisitionId string
		FirstName        string
		LastName         string
		CountryCode      string
		PhoneNumber      string
		Email            string
	}
	reqBody := requestBody{
		JobRequisitionId: input.JobRequisitionId,
		FirstName:        input.FirstName,
		LastName:         input.LastName,
		CountryCode:      input.CountryCode,
		PhoneNumber:      input.PhoneNumber,
		Email:            input.Email,
	}
	dw, err := multiWriter.CreateFormField("data")
	if err != nil {
		log.Fatalf("Could not create data body: %s", err)
	}
	err = json.NewEncoder(dw).Encode(reqBody)
	if err != nil {
		log.Fatalf("Could not encode json body: %s", err)
	}

	file, err := os.Open("../test_resume.pdf")
	if err != nil {
		log.Fatalf("Could not open test resume: %s", err)
	}
	defer file.Close()
	fw, err := multiWriter.CreateFormFile("resume", "test_resume.pdf")
	if err != nil {
		log.Fatalf("Could not create file body: %s", err)
	}
	// Copy the contents of the file to the form field
	if _, err := io.Copy(fw, file); err != nil {
		log.Fatal(err)
	}

	multiWriter.Close()

	path := fmt.Sprintf("/api/tenants/%s/job-applications/%s", input.TenantId, input.Id)
	r, err := http.NewRequest("POST", path, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	r.Header.Set("Content-Type", multiWriter.FormDataContentType())

	w := httptest.NewRecorder()

	// Temporarily modifying the global router will not interfere with other tests because these tests are run sequentially.
	realFileStorage := s.router.fileStorage
	s.router.fileStorage = s3mock.NewS3Mock()
	s.router.ServeHTTP(w, r)
	s.router.fileStorage = realFileStorage

	s.expectHttpStatus(w, 500)
	s.expectErrorCode(w, "INTERNAL-SERVER-ERROR")

	s.expectS3ToNotContainFile(input.ResumeS3Url)

	s.expectSelectQueryToReturnNoRows(
		"job_application",
		map[string]any{
			"id":                 input.Id,
			"tenant_id":          input.TenantId,
			"job_requisition_id": input.JobRequisitionId,
			"first_name":         input.FirstName,
			"last_name":          input.LastName,
			"country_code":       input.CountryCode,
			"phone_number":       input.PhoneNumber,
			"email":              input.Email,
			"resume_s3_url":      input.ResumeS3Url,
		},
	)

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"ERROR"`, `"msg":"INTERNAL-SERVER-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestCreateJobApplicationHandlesModelsError() {
	input := storage.JobApplication{
		Id:               s.defaultJobApplication.Id, // Should trigger unique constraint error
		TenantId:         s.defaultTenant.Id,
		JobRequisitionId: s.defaultApprovedJobRequisition.Id,
		FirstName:        "Eugene",
		LastName:         "Lek",
		CountryCode:      "1",
		PhoneNumber:      "987654321",
		Email:            "test123@gmail.com",
		ResumeS3Url:      fmt.Sprintf("%s/hr-information-system/job-applications/%s/Eugene_Lek_resume.pdf", s.s3Server.URL, s.defaultJobApplication.Id),
	}

	bodyBuf := new(bytes.Buffer)
	multiWriter := multipart.NewWriter(bodyBuf)

	type requestBody struct {
		JobRequisitionId string
		FirstName        string
		LastName         string
		CountryCode      string
		PhoneNumber      string
		Email            string
	}
	reqBody := requestBody{
		JobRequisitionId: input.JobRequisitionId,
		FirstName:        input.FirstName,
		LastName:         input.LastName,
		CountryCode:      input.CountryCode,
		PhoneNumber:      input.PhoneNumber,
		Email:            input.Email,
	}
	dw, err := multiWriter.CreateFormField("data")
	if err != nil {
		log.Fatalf("Could not create data body: %s", err)
	}
	err = json.NewEncoder(dw).Encode(reqBody)
	if err != nil {
		log.Fatalf("Could not encode json body: %s", err)
	}

	file, err := os.Open("../test_resume.pdf")
	if err != nil {
		log.Fatalf("Could not open test resume: %s", err)
	}
	defer file.Close()
	fw, err := multiWriter.CreateFormFile("resume", "test_resume.pdf")
	if err != nil {
		log.Fatalf("Could not create file body: %s", err)
	}
	// Copy the contents of the file to the form field
	if _, err := io.Copy(fw, file); err != nil {
		log.Fatal(err)
	}

	multiWriter.Close()

	path := fmt.Sprintf("/api/tenants/%s/job-applications/%s", input.TenantId, input.Id)
	r, err := http.NewRequest("POST", path, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	r.Header.Set("Content-Type", multiWriter.FormDataContentType())

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 409)
	s.expectErrorCode(w, "UNIQUE-VIOLATION-ERROR")

	s.expectS3ToContainFile(input.ResumeS3Url)

	s.expectSelectQueryToReturnNoRows(
		"job_application",
		map[string]any{
			"id":                 input.Id,
			"tenant_id":          input.TenantId,
			"job_requisition_id": input.JobRequisitionId,
			"first_name":         input.FirstName,
			"last_name":          input.LastName,
			"country_code":       input.CountryCode,
			"phone_number":       input.PhoneNumber,
			"email":              input.Email,
			"resume_s3_url":      input.ResumeS3Url,
		},
	)

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"RESUME-UPLOADED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"UNIQUE-VIOLATION-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestRecruiterShortlistJobApplication() {
	input := storage.JobApplication{
		Id:                s.defaultJobApplication.Id,
		TenantId:          s.defaultJobApplication.TenantId,
		JobRequisitionId:  s.defaultJobApplication.JobRequisitionId,
		RecruiterDecision: "SHORTLISTED",
	}
	inputRecruiter := s.defaultRecruiter.Id

	type requestBody struct {
		RecruiterDecision string
	}
	reqBody := requestBody{
		RecruiterDecision: input.RecruiterDecision,
	}
	bodyBuf := new(bytes.Buffer)
	err := json.NewEncoder(bodyBuf).Encode(reqBody)
	if err != nil {
		log.Fatalf("Could not encode json body: %s", err)
	}

	path := fmt.Sprintf("/api/tenants/%s/users/%s/job-requisitions/role-recruiter/%s/job-applications/%s/recruiter-decision", input.TenantId, inputRecruiter, input.JobRequisitionId, input.Id)
	r, err := http.NewRequest("POST", path, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultRecruiter.Id, s.defaultRecruiter.TenantId, s.defaultRecruiter.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 204)

	s.expectSelectQueryToReturnOneRow(
		"job_application",
		map[string]any{
			"id":                 input.Id,
			"tenant_id":          input.TenantId,
			"job_requisition_id": input.JobRequisitionId,
			"recruiter_decision": input.RecruiterDecision,
		},
	)

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"JOB-APPLICATION-RECRUITER-SHORTLISTED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestRecruiterRejectJobApplication() {
	input := storage.JobApplication{
		Id:                s.defaultJobApplication.Id,
		TenantId:          s.defaultJobApplication.TenantId,
		JobRequisitionId:  s.defaultJobApplication.JobRequisitionId,
		RecruiterDecision: "REJECTED",
	}
	inputRecruiter := s.defaultRecruiter.Id

	type requestBody struct {
		RecruiterDecision string
	}
	reqBody := requestBody{
		RecruiterDecision: input.RecruiterDecision,
	}
	bodyBuf := new(bytes.Buffer)
	err := json.NewEncoder(bodyBuf).Encode(reqBody)
	if err != nil {
		log.Fatalf("Could not encode json body: %s", err)
	}

	path := fmt.Sprintf("/api/tenants/%s/users/%s/job-requisitions/role-recruiter/%s/job-applications/%s/recruiter-decision", input.TenantId, inputRecruiter, input.JobRequisitionId, input.Id)
	r, err := http.NewRequest("POST", path, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultRecruiter.Id, s.defaultRecruiter.TenantId, s.defaultRecruiter.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 204)

	s.expectSelectQueryToReturnOneRow(
		"job_application",
		map[string]any{
			"id":                 input.Id,
			"tenant_id":          input.TenantId,
			"job_requisition_id": input.JobRequisitionId,
			"recruiter_decision": input.RecruiterDecision,
		},
	)

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"JOB-APPLICATION-RECRUITER-REJECTED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

// Tests that recruiters cannot shortlist job applications that are tied to job requisitions that they are not assigned to as a recruiter
func (s *IntegrationTestSuite) TestRecruiterShortlistJobApplicationShouldPreventIdExploit() {
	input := storage.JobApplication{
		Id:                s.defaultJobApplication.Id,
		TenantId:          s.defaultJobApplication.TenantId,
		JobRequisitionId:  s.defaultJobApplication.JobRequisitionId,
		RecruiterDecision: "SHORTLISTED",
	}
	inputRecruiter := s.defaultUser.Id // Not the recruiter, should trigger a 404 error

	// Temporarily give the defaultUser recruiter shortlisting rights for the purpose of this test
	query := "INSERT INTO casbin_rule (Ptype, V0, V1, V2, V3) VALUES ('p', $1, $2, $3, $4)"
	resourcePath := fmt.Sprintf("/api/tenants/%s/users/%s/job-requisitions/role-recruiter/{jobReqId}/job-applications/{jobAppId}/recruiter-decision", s.defaultTenant.Id, inputRecruiter)
	_, err := s.dbRootConn.Exec(query, inputRecruiter, s.defaultJobApplication.TenantId, resourcePath, "POST")
	if err != nil {
		log.Fatalf("Could not give user temporary recruiter rights: %s", err)
	}
	err = s.router.authEnforcer.LoadPolicy()
	if err != nil {
		log.Fatalf("Could not reload enforcer to give user temporary recruiter rights: %s", err)
	}

	type requestBody struct {
		RecruiterDecision string
	}
	reqBody := requestBody{
		RecruiterDecision: input.RecruiterDecision,
	}
	bodyBuf := new(bytes.Buffer)
	err = json.NewEncoder(bodyBuf).Encode(reqBody)
	if err != nil {
		log.Fatalf("Could not encode json body: %s", err)
	}

	path := fmt.Sprintf("/api/tenants/%s/users/%s/job-requisitions/role-recruiter/%s/job-applications/%s/recruiter-decision", input.TenantId, inputRecruiter, input.JobRequisitionId, input.Id)
	r, err := http.NewRequest("POST", path, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultUser.Id, s.defaultUser.TenantId, s.defaultUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 404)
	s.expectErrorCode(w, "RESOURCE-NOT-FOUND-ERROR")

	s.expectSelectQueryToReturnNoRows(
		"job_application",
		map[string]any{
			"id":                 input.Id,
			"tenant_id":          input.TenantId,
			"job_requisition_id": input.JobRequisitionId,
			"recruiter_decision": input.RecruiterDecision,
		},
	)

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"RESOURCE-NOT-FOUND-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestRecruiterShortlistValidatesInput() {
	input := storage.JobApplication{
		Id:                s.defaultJobApplication.Id,
		TenantId:          s.defaultJobApplication.TenantId,
		JobRequisitionId:  s.defaultJobApplication.JobRequisitionId,
		RecruiterDecision: "invalid decision",
	}
	inputRecruiter := s.defaultRecruiter.Id

	type requestBody struct {
		RecruiterDecision string
	}
	reqBody := requestBody{
		RecruiterDecision: input.RecruiterDecision,
	}
	bodyBuf := new(bytes.Buffer)
	err := json.NewEncoder(bodyBuf).Encode(reqBody)
	if err != nil {
		log.Fatalf("Could not encode json body: %s", err)
	}

	path := fmt.Sprintf("/api/tenants/%s/users/%s/job-requisitions/role-recruiter/%s/job-applications/%s/recruiter-decision", input.TenantId, inputRecruiter, input.JobRequisitionId, input.Id)
	r, err := http.NewRequest("POST", path, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultRecruiter.Id, s.defaultRecruiter.TenantId, s.defaultRecruiter.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 400)
	s.expectErrorCode(w, "INPUT-VALIDATION-ERROR")

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"INPUT-VALIDATION-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestRecruiterShortlistValidatesJobRequisition() {
	input := storage.JobApplication{
		Id:                s.defaultJobApplication.Id,
		TenantId:          s.defaultJobApplication.TenantId,
		JobRequisitionId:  s.defaultJobApplication.JobRequisitionId,
		RecruiterDecision: "SHORTLISTED",
	}
	inputRecruiter := s.defaultRecruiter.Id

	// Remove HR approval from the job requisition
	query := "UPDATE job_requisition SET hr_approver_decision = 'REJECTED' WHERE id = $1"
	_, err := s.dbRootConn.Exec(query, input.JobRequisitionId)
	if err != nil {
		log.Fatalf("Could not remove hr approval: %s", err)
	}

	type requestBody struct {
		RecruiterDecision string
	}
	reqBody := requestBody{
		RecruiterDecision: input.RecruiterDecision,
	}
	bodyBuf := new(bytes.Buffer)
	err = json.NewEncoder(bodyBuf).Encode(reqBody)
	if err != nil {
		log.Fatalf("Could not encode json body: %s", err)
	}

	path := fmt.Sprintf("/api/tenants/%s/users/%s/job-requisitions/role-recruiter/%s/job-applications/%s/recruiter-decision", input.TenantId, inputRecruiter, input.JobRequisitionId, input.Id)
	r, err := http.NewRequest("POST", path, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultRecruiter.Id, s.defaultRecruiter.TenantId, s.defaultRecruiter.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 403)
	s.expectErrorCode(w, "MISSING-HR-APPROVAL-ERROR")

	s.expectSelectQueryToReturnNoRows(
		"job_application",
		map[string]any{
			"id":                 input.Id,
			"tenant_id":          input.TenantId,
			"job_requisition_id": input.JobRequisitionId,
			"recruiter_decision": input.RecruiterDecision,
		},
	)

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"MISSING-HR-APPROVAL-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestRecruiterSetInterviewDate() {
	input := storage.JobApplication{
		Id:               s.defaultJobApplication.Id,
		TenantId:         s.defaultJobApplication.TenantId,
		JobRequisitionId: s.defaultJobApplication.JobRequisitionId,
		InterviewDate:    "2024-01-29",
	}
	inputRecruiter := s.defaultRecruiter.Id

	// Sets recruiter shortlist
	query := "UPDATE job_application SET recruiter_decision = 'SHORTLISTED' WHERE id = $1"
	_, err := s.dbRootConn.Exec(query, input.Id)
	if err != nil {
		log.Fatalf("Could not set recruiter shortlist: %s", err)
	}

	type requestBody struct {
		InterviewDate string
	}
	reqBody := requestBody{
		InterviewDate: input.InterviewDate,
	}
	bodyBuf := new(bytes.Buffer)
	err = json.NewEncoder(bodyBuf).Encode(reqBody)
	if err != nil {
		log.Fatalf("Could not encode json body: %s", err)
	}

	path := fmt.Sprintf("/api/tenants/%s/users/%s/job-requisitions/role-recruiter/%s/job-applications/%s/interview-date", input.TenantId, inputRecruiter, input.JobRequisitionId, input.Id)
	r, err := http.NewRequest("POST", path, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultRecruiter.Id, s.defaultRecruiter.TenantId, s.defaultRecruiter.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 204)

	s.expectSelectQueryToReturnOneRow(
		"job_application",
		map[string]any{
			"id":                 input.Id,
			"tenant_id":          input.TenantId,
			"job_requisition_id": input.JobRequisitionId,
			"interview_date":     input.InterviewDate,
		},
	)

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"JOB-APPLICATION-INTERVIEW-DATE-SET"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestRecruiterSetInterviewDateShouldPreventIdExploit() {
	input := storage.JobApplication{
		Id:               s.defaultJobApplication.Id,
		TenantId:         s.defaultJobApplication.TenantId,
		JobRequisitionId: s.defaultJobApplication.JobRequisitionId,
		InterviewDate:    "2024-01-29",
	}
	inputRecruiter := s.defaultUser.Id // Not the recruiter, should trigger a 403 error

	// Sets recruiter shortlist
	query := "UPDATE job_application SET recruiter_decision = 'SHORTLISTED' WHERE id = $1"
	_, err := s.dbRootConn.Exec(query, input.Id)
	if err != nil {
		log.Fatalf("Could not set recruiter shortlist: %s", err)
	}

	// Temporarily give the defaultUser recruiter interview date rights for the purpose of this test
	query = "INSERT INTO casbin_rule (Ptype, V0, V1, V2, V3) VALUES ('p', $1, $2, $3, $4)"
	resourcePath := fmt.Sprintf("/api/tenants/%s/users/%s/job-requisitions/role-recruiter/{jobReqId}/job-applications/{jobAppId}/interview-date", s.defaultTenant.Id, inputRecruiter)
	_, err = s.dbRootConn.Exec(query, inputRecruiter, s.defaultJobApplication.TenantId, resourcePath, "POST")
	if err != nil {
		log.Fatalf("Could not give user temporary recruiter rights: %s", err)
	}
	err = s.router.authEnforcer.LoadPolicy()
	if err != nil {
		log.Fatalf("Could not reload enforcer to give user temporary recruiter rights: %s", err)
	}

	type requestBody struct {
		InterviewDate string
	}
	reqBody := requestBody{
		InterviewDate: input.InterviewDate,
	}
	bodyBuf := new(bytes.Buffer)
	err = json.NewEncoder(bodyBuf).Encode(reqBody)
	if err != nil {
		log.Fatalf("Could not encode json body: %s", err)
	}

	path := fmt.Sprintf("/api/tenants/%s/users/%s/job-requisitions/role-recruiter/%s/job-applications/%s/interview-date", input.TenantId, inputRecruiter, input.JobRequisitionId, input.Id)
	r, err := http.NewRequest("POST", path, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultUser.Id, s.defaultUser.TenantId, s.defaultUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 404)
	s.expectErrorCode(w, "RESOURCE-NOT-FOUND-ERROR")

	s.expectSelectQueryToReturnNoRows(
		"job_application",
		map[string]any{
			"id":                 input.Id,
			"tenant_id":          input.TenantId,
			"job_requisition_id": input.JobRequisitionId,
			"interview_date":     input.InterviewDate,
		},
	)

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"RESOURCE-NOT-FOUND-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestRecruiterSetInterviewDateValidatesInput() {
	input := storage.JobApplication{
		Id:               s.defaultJobApplication.Id,
		TenantId:         s.defaultJobApplication.TenantId,
		JobRequisitionId: s.defaultJobApplication.JobRequisitionId,
		InterviewDate:    "2024-01-32",
	}
	inputRecruiter := s.defaultRecruiter.Id

	// Sets recruiter shortlist
	query := "UPDATE job_application SET recruiter_decision = 'SHORTLISTED' WHERE id = $1"
	_, err := s.dbRootConn.Exec(query, input.Id)
	if err != nil {
		log.Fatalf("Could not set recruiter shortlist: %s", err)
	}

	type requestBody struct {
		InterviewDate string
	}
	reqBody := requestBody{
		InterviewDate: input.InterviewDate,
	}
	bodyBuf := new(bytes.Buffer)
	err = json.NewEncoder(bodyBuf).Encode(reqBody)
	if err != nil {
		log.Fatalf("Could not encode json body: %s", err)
	}

	path := fmt.Sprintf("/api/tenants/%s/users/%s/job-requisitions/role-recruiter/%s/job-applications/%s/interview-date", input.TenantId, inputRecruiter, input.JobRequisitionId, input.Id)
	r, err := http.NewRequest("POST", path, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultRecruiter.Id, s.defaultRecruiter.TenantId, s.defaultRecruiter.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 400)
	s.expectErrorCode(w, "INPUT-VALIDATION-ERROR")

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"INPUT-VALIDATION-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestRecruiterSetInterviewDateValidatesJobRequisition() {
	input := storage.JobApplication{
		Id:               s.defaultJobApplication.Id,
		TenantId:         s.defaultJobApplication.TenantId,
		JobRequisitionId: s.defaultJobApplication.JobRequisitionId,
		InterviewDate:    "2024-01-29",
	}
	inputRecruiter := s.defaultRecruiter.Id

	// Sets recruiter shortlist
	query := "UPDATE job_application SET recruiter_decision = 'SHORTLISTED' WHERE id = $1"
	_, err := s.dbRootConn.Exec(query, input.Id)
	if err != nil {
		log.Fatalf("Could not set recruiter shortlist: %s", err)
	}

	// Remove HR approval from the job requisition
	query = "UPDATE job_requisition SET hr_approver_decision = 'REJECTED' WHERE id = $1"
	_, err = s.dbRootConn.Exec(query, input.JobRequisitionId)
	if err != nil {
		log.Fatalf("Could not remove hr approval: %s", err)
	}

	type requestBody struct {
		InterviewDate string
	}
	reqBody := requestBody{
		InterviewDate: input.InterviewDate,
	}
	bodyBuf := new(bytes.Buffer)
	err = json.NewEncoder(bodyBuf).Encode(reqBody)
	if err != nil {
		log.Fatalf("Could not encode json body: %s", err)
	}

	path := fmt.Sprintf("/api/tenants/%s/users/%s/job-requisitions/role-recruiter/%s/job-applications/%s/interview-date", input.TenantId, inputRecruiter, input.JobRequisitionId, input.Id)
	r, err := http.NewRequest("POST", path, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultRecruiter.Id, s.defaultRecruiter.TenantId, s.defaultRecruiter.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 403)
	s.expectErrorCode(w, "MISSING-HR-APPROVAL-ERROR")

	s.expectSelectQueryToReturnNoRows(
		"job_application",
		map[string]any{
			"id":                 input.Id,
			"tenant_id":          input.TenantId,
			"job_requisition_id": input.JobRequisitionId,
			"interview_date":     input.InterviewDate,
		},
	)

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"MISSING-HR-APPROVAL-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestRecruiterSetInterviewDateHandlesModelError() {
	// Should trigger check constraint error because the recruiter's decision has not been made
	input := storage.JobApplication{
		Id:               s.defaultJobApplication.Id,
		TenantId:         s.defaultJobApplication.TenantId,
		JobRequisitionId: s.defaultJobApplication.JobRequisitionId,
		InterviewDate:    "2024-01-29",
	}
	inputRecruiter := s.defaultRecruiter.Id

	type requestBody struct {
		InterviewDate string
	}
	reqBody := requestBody{
		InterviewDate: input.InterviewDate,
	}
	bodyBuf := new(bytes.Buffer)
	err := json.NewEncoder(bodyBuf).Encode(reqBody)
	if err != nil {
		log.Fatalf("Could not encode json body: %s", err)
	}

	path := fmt.Sprintf("/api/tenants/%s/users/%s/job-requisitions/role-recruiter/%s/job-applications/%s/interview-date", input.TenantId, inputRecruiter, input.JobRequisitionId, input.Id)
	r, err := http.NewRequest("POST", path, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultRecruiter.Id, s.defaultRecruiter.TenantId, s.defaultRecruiter.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 409)
	s.expectErrorCode(w, "MISSING-RECRUITER-SHORTLIST-ERROR")

	s.expectSelectQueryToReturnNoRows(
		"job_application",
		map[string]any{
			"id":                 input.Id,
			"tenant_id":          input.TenantId,
			"job_requisition_id": input.JobRequisitionId,
			"interview_date":     input.InterviewDate,
		},
	)

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"MISSING-RECRUITER-SHORTLIST-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestHiringManagerMakeOffer() {
	input := storage.JobApplication{
		Id:                    s.defaultJobApplication.Id,
		TenantId:              s.defaultJobApplication.TenantId,
		JobRequisitionId:      s.defaultJobApplication.JobRequisitionId,
		HiringManagerDecision: "OFFERED",
		OfferStartDate:        "2024-02-01",
	}
	inputHiringManager := s.defaultUser.Id

	// Sets recruiter shortlist & interview date
	query := "UPDATE job_application SET recruiter_decision = 'SHORTLISTED', interview_date = '2024-01-29' WHERE id = $1"
	_, err := s.dbRootConn.Exec(query, input.Id)
	if err != nil {
		log.Fatalf("Could not set recruiter shortlist & interview date: %s", err)
	}

	type requestBody struct {
		HiringManagerDecision string
		OfferStartDate        string
	}
	reqBody := requestBody{
		HiringManagerDecision: input.HiringManagerDecision,
		OfferStartDate:        input.OfferStartDate,
	}
	bodyBuf := new(bytes.Buffer)
	err = json.NewEncoder(bodyBuf).Encode(reqBody)
	if err != nil {
		log.Fatalf("Could not encode json body: %s", err)
	}

	path := fmt.Sprintf("/api/tenants/%s/users/%s/job-requisitions/role-requestor/%s/job-applications/%s/hiring-manager-decision", input.TenantId, inputHiringManager, input.JobRequisitionId, input.Id)
	r, err := http.NewRequest("POST", path, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultUser.Id, s.defaultUser.TenantId, s.defaultUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 204)

	s.expectSelectQueryToReturnOneRow(
		"job_application",
		map[string]any{
			"id":                      input.Id,
			"tenant_id":               input.TenantId,
			"job_requisition_id":      input.JobRequisitionId,
			"hiring_manager_decision": input.HiringManagerDecision,
		},
	)

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"JOB-APPLICATION-HIRING-MANAGER-OFFERED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestHiringManagerRejectJobApplication() {
	input := storage.JobApplication{
		Id:                    s.defaultJobApplication.Id,
		TenantId:              s.defaultJobApplication.TenantId,
		JobRequisitionId:      s.defaultJobApplication.JobRequisitionId,
		HiringManagerDecision: "REJECTED",
		OfferStartDate:        "2024-02-01",
	}
	inputHiringManager := s.defaultUser.Id

	// Sets recruiter shortlist & interview date
	query := "UPDATE job_application SET recruiter_decision = 'SHORTLISTED', interview_date = '2024-01-29' WHERE id = $1"
	_, err := s.dbRootConn.Exec(query, input.Id)
	if err != nil {
		log.Fatalf("Could not set recruiter shortlist & interview date: %s", err)
	}

	type requestBody struct {
		HiringManagerDecision string
		OfferStartDate        string
	}
	reqBody := requestBody{
		HiringManagerDecision: input.HiringManagerDecision,
		OfferStartDate:        input.OfferStartDate,
	}
	bodyBuf := new(bytes.Buffer)
	err = json.NewEncoder(bodyBuf).Encode(reqBody)
	if err != nil {
		log.Fatalf("Could not encode json body: %s", err)
	}

	path := fmt.Sprintf("/api/tenants/%s/users/%s/job-requisitions/role-requestor/%s/job-applications/%s/hiring-manager-decision", input.TenantId, inputHiringManager, input.JobRequisitionId, input.Id)
	r, err := http.NewRequest("POST", path, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultUser.Id, s.defaultUser.TenantId, s.defaultUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 204)

	s.expectSelectQueryToReturnOneRow(
		"job_application",
		map[string]any{
			"id":                      input.Id,
			"tenant_id":               input.TenantId,
			"job_requisition_id":      input.JobRequisitionId,
			"hiring_manager_decision": input.HiringManagerDecision,
		},
	)

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"JOB-APPLICATION-HIRING-MANAGER-REJECTED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestSetHiringManagerDecisionShouldPreventIdExploit() {
	input := storage.JobApplication{
		Id:                    s.defaultJobApplication.Id,
		TenantId:              s.defaultJobApplication.TenantId,
		JobRequisitionId:      s.defaultJobApplication.JobRequisitionId,
		HiringManagerDecision: "OFFERED",
		OfferStartDate:        "2024-02-01",
	}
	inputHiringManager := s.defaultRecruiter.Id // Not the hiring manager, should trigger a 403 error

	// Sets recruiter shortlist & interview date
	query := "UPDATE job_application SET recruiter_decision = 'SHORTLISTED', interview_date = '2024-01-29' WHERE id = $1"
	_, err := s.dbRootConn.Exec(query, input.Id)
	if err != nil {
		log.Fatalf("Could not set recruiter shortlist & interview date: %s", err)
	}

	// Temporarily give the recruiter offer rights for the purpose of this test
	query = "INSERT INTO casbin_rule (Ptype, V0, V1, V2, V3) VALUES ('p', $1, $2, $3, $4)"
	resourcePath := fmt.Sprintf("/api/tenants/%s/users/%s/job-requisitions/role-requestor/{jobReqId}/job-applications/{jobAppId}/hiring-manager-decision", s.defaultTenant.Id, inputHiringManager)
	_, err = s.dbRootConn.Exec(query, inputHiringManager, s.defaultJobApplication.TenantId, resourcePath, "POST")
	if err != nil {
		log.Fatalf("Could not give user temporary hiring manager rights: %s", err)
	}
	err = s.router.authEnforcer.LoadPolicy()
	if err != nil {
		log.Fatalf("Could not reload enforcer to give user temporary recruiter rights: %s", err)
	}

	type requestBody struct {
		HiringManagerDecision string
		OfferStartDate        string
	}
	reqBody := requestBody{
		HiringManagerDecision: input.HiringManagerDecision,
		OfferStartDate:        input.OfferStartDate,
	}
	bodyBuf := new(bytes.Buffer)
	err = json.NewEncoder(bodyBuf).Encode(reqBody)
	if err != nil {
		log.Fatalf("Could not encode json body: %s", err)
	}

	path := fmt.Sprintf("/api/tenants/%s/users/%s/job-requisitions/role-requestor/%s/job-applications/%s/hiring-manager-decision", input.TenantId, inputHiringManager, input.JobRequisitionId, input.Id)
	r, err := http.NewRequest("POST", path, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultRecruiter.Id, s.defaultRecruiter.TenantId, s.defaultRecruiter.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 404)
	s.expectErrorCode(w, "RESOURCE-NOT-FOUND-ERROR")

	s.expectSelectQueryToReturnNoRows(
		"job_application",
		map[string]any{
			"id":                      input.Id,
			"tenant_id":               input.TenantId,
			"job_requisition_id":      input.JobRequisitionId,
			"hiring_manager_decision": input.HiringManagerDecision,
		},
	)

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"RESOURCE-NOT-FOUND-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestSetHiringManagerDecisionValidatesInput() {
	input := storage.JobApplication{
		Id:                    s.defaultJobApplication.Id,
		TenantId:              s.defaultJobApplication.TenantId,
		JobRequisitionId:      s.defaultJobApplication.JobRequisitionId,
		HiringManagerDecision: "invalid decision",
		OfferStartDate:        "2024-02-01",
	}
	inputHiringManager := s.defaultUser.Id

	// Sets recruiter shortlist & interview date
	query := "UPDATE job_application SET recruiter_decision = 'SHORTLISTED', interview_date = '2024-01-29' WHERE id = $1"
	_, err := s.dbRootConn.Exec(query, input.Id)
	if err != nil {
		log.Fatalf("Could not set recruiter shortlist & interview date: %s", err)
	}

	type requestBody struct {
		HiringManagerDecision string
		OfferStartDate        string
	}
	reqBody := requestBody{
		HiringManagerDecision: input.HiringManagerDecision,
		OfferStartDate:        input.OfferStartDate,
	}
	bodyBuf := new(bytes.Buffer)
	err = json.NewEncoder(bodyBuf).Encode(reqBody)
	if err != nil {
		log.Fatalf("Could not encode json body: %s", err)
	}

	path := fmt.Sprintf("/api/tenants/%s/users/%s/job-requisitions/role-requestor/%s/job-applications/%s/hiring-manager-decision", input.TenantId, inputHiringManager, input.JobRequisitionId, input.Id)
	r, err := http.NewRequest("POST", path, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultUser.Id, s.defaultUser.TenantId, s.defaultUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 400)
	s.expectErrorCode(w, "INPUT-VALIDATION-ERROR")

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"INPUT-VALIDATION-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestSetHiringManagerDecisionValidatesJobRequisition() {
	input := storage.JobApplication{
		Id:                    s.defaultJobApplication.Id,
		TenantId:              s.defaultJobApplication.TenantId,
		JobRequisitionId:      s.defaultJobApplication.JobRequisitionId,
		HiringManagerDecision: "OFFERED",
		OfferStartDate:        "2024-02-01",
	}
	inputHiringManager := s.defaultUser.Id

	// Sets recruiter shortlist & interview date
	query := "UPDATE job_application SET recruiter_decision = 'SHORTLISTED', interview_date = '2024-01-29' WHERE id = $1"
	_, err := s.dbRootConn.Exec(query, input.Id)
	if err != nil {
		log.Fatalf("Could not set recruiter shortlist & interview date: %s", err)
	}

	// Remove HR approval from the job requisition
	query = "UPDATE job_requisition SET hr_approver_decision = 'REJECTED' WHERE id = $1"
	_, err = s.dbRootConn.Exec(query, input.JobRequisitionId)
	if err != nil {
		log.Fatalf("Could not remove hr approval: %s", err)
	}

	type requestBody struct {
		HiringManagerDecision string
		OfferStartDate        string
	}
	reqBody := requestBody{
		HiringManagerDecision: input.HiringManagerDecision,
		OfferStartDate:        input.OfferStartDate,
	}
	bodyBuf := new(bytes.Buffer)
	err = json.NewEncoder(bodyBuf).Encode(reqBody)
	if err != nil {
		log.Fatalf("Could not encode json body: %s", err)
	}

	path := fmt.Sprintf("/api/tenants/%s/users/%s/job-requisitions/role-requestor/%s/job-applications/%s/hiring-manager-decision", input.TenantId, inputHiringManager, input.JobRequisitionId, input.Id)
	r, err := http.NewRequest("POST", path, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultUser.Id, s.defaultUser.TenantId, s.defaultUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 403)
	s.expectErrorCode(w, "MISSING-HR-APPROVAL-ERROR")

	s.expectSelectQueryToReturnNoRows(
		"job_application",
		map[string]any{
			"id":                      input.Id,
			"tenant_id":               input.TenantId,
			"job_requisition_id":      input.JobRequisitionId,
			"hiring_manager_decision": input.HiringManagerDecision,
		},
	)

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"MISSING-HR-APPROVAL-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestSetHiringManagerDecisionHandlesModelError() {
	// Should trigger check constraint error because the recruiter's decision has not been made
	input := storage.JobApplication{
		Id:                    s.defaultJobApplication.Id,
		TenantId:              s.defaultJobApplication.TenantId,
		JobRequisitionId:      s.defaultJobApplication.JobRequisitionId,
		HiringManagerDecision: "OFFERED",
		OfferStartDate:        "2024-02-01",
	}
	inputHiringManager := s.defaultUser.Id

	type requestBody struct {
		HiringManagerDecision string
		OfferStartDate        string
	}
	reqBody := requestBody{
		HiringManagerDecision: input.HiringManagerDecision,
		OfferStartDate:        input.OfferStartDate,
	}
	bodyBuf := new(bytes.Buffer)
	err := json.NewEncoder(bodyBuf).Encode(reqBody)
	if err != nil {
		log.Fatalf("Could not encode json body: %s", err)
	}

	path := fmt.Sprintf("/api/tenants/%s/users/%s/job-requisitions/role-requestor/%s/job-applications/%s/hiring-manager-decision", input.TenantId, inputHiringManager, input.JobRequisitionId, input.Id)
	r, err := http.NewRequest("POST", path, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultUser.Id, s.defaultUser.TenantId, s.defaultUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 409)
	s.expectErrorCode(w, "MISSING-INTERVIEW-DATE-ERROR")

	s.expectSelectQueryToReturnNoRows(
		"job_application",
		map[string]any{
			"id":                      input.Id,
			"tenant_id":               input.TenantId,
			"job_requisition_id":      input.JobRequisitionId,
			"hiring_manager_decision": input.HiringManagerDecision,
		},
	)

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"MISSING-INTERVIEW-DATE-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestRecruiterSetApplicantAcceptance() {
	input := storage.JobApplication{
		Id:                s.defaultJobApplication.Id,
		TenantId:          s.defaultJobApplication.TenantId,
		JobRequisitionId:  s.defaultJobApplication.JobRequisitionId,
		ApplicantDecision: "ACCEPTED",
	}
	inputRecruiter := s.defaultRecruiter.Id
	newUserEmail := fmt.Sprintf("%s_%s@%s.com", 
			strings.ToLower(s.defaultJobApplication.FirstName), 
			strings.ToLower(s.defaultJobApplication.LastName), 
			strings.ToLower(strings.ReplaceAll(strings.ToLower(s.defaultTenant.Name), " ", "")),
	)

	// Sets recruiter shortlist, interview date, hiring manager offer, and offer start date
	query := "UPDATE job_application SET recruiter_decision = 'SHORTLISTED', interview_date = '2024-01-29', hiring_manager_decision = 'OFFERED', offer_start_date = '2024-02-01' WHERE id = $1"
	_, err := s.dbRootConn.Exec(query, input.Id)
	if err != nil {
		log.Fatalf("Could not set recruiter shortlist & interview date: %s", err)
	}

	type requestBody struct {
		ApplicantDecision string
	}
	reqBody := requestBody{
		ApplicantDecision: input.ApplicantDecision,
	}
	bodyBuf := new(bytes.Buffer)
	err = json.NewEncoder(bodyBuf).Encode(reqBody)
	if err != nil {
		log.Fatalf("Could not encode json body: %s", err)
	}

	path := fmt.Sprintf("/api/tenants/%s/users/%s/job-requisitions/role-recruiter/%s/job-applications/%s/applicant-decision", input.TenantId, inputRecruiter, input.JobRequisitionId, input.Id)
	r, err := http.NewRequest("POST", path, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultRecruiter.Id, s.defaultRecruiter.TenantId, s.defaultRecruiter.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 204)

	// Fetch the newly created user account's id
	var newUserId string
	s.dbRootConn.QueryRow("SELECT id FROM user_account WHERE tenant_id = $1 AND email = $2", input.TenantId, newUserEmail).Scan(&newUserId)

	if newUserId != "" {
		// Run these only if the user id was found. Otherwise, they will error
		s.expectSelectQueryToReturnOneRow(
			"user_account",
			map[string]any{
				"id": newUserId,
				"tenant_id": input.TenantId,
				"email":     newUserEmail,
			},
		)
	
		s.expectSelectQueryToReturnOneRow(
			"position_assignment",
			map[string]any{
				"tenant_id":   input.TenantId,
				"user_account_id": newUserId,			
				"position_id": s.defaultApprovedJobRequisition.PositionId,
			},
		)
	
		s.expectSelectQueryToReturnOneRow(
			"job_requisition",
			map[string]any{
				"id":        input.JobRequisitionId,
				"tenant_id": input.TenantId,
				"filled_by": newUserId,
			},
		)
	
		s.expectSelectQueryToReturnOneRow(
			"job_application",
			map[string]any{
				"id":                 input.Id,
				"tenant_id":          input.TenantId,
				"job_requisition_id": input.JobRequisitionId,
				"applicant_decision": input.ApplicantDecision,
			},
		)
	}		

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"JOB-APPLICATION-APPLICANT-ACCEPTED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestRecruiterSetApplicantRejection() {
	input := storage.JobApplication{
		Id:                s.defaultJobApplication.Id,
		TenantId:          s.defaultJobApplication.TenantId,
		JobRequisitionId:  s.defaultJobApplication.JobRequisitionId,
		ApplicantDecision: "REJECTED",
	}
	inputRecruiter := s.defaultRecruiter.Id

	// Sets recruiter shortlist, interview date, hiring manager offer, and offer start date
	query := "UPDATE job_application SET recruiter_decision = 'SHORTLISTED', interview_date = '2024-01-29', hiring_manager_decision = 'OFFERED', offer_start_date = '2024-02-01' WHERE id = $1"
	_, err := s.dbRootConn.Exec(query, input.Id)
	if err != nil {
		log.Fatalf("Could not set recruiter shortlist & interview date: %s", err)
	}	

	type requestBody struct {
		ApplicantDecision string
	}
	reqBody := requestBody{
		ApplicantDecision: input.ApplicantDecision,
	}
	bodyBuf := new(bytes.Buffer)
	err = json.NewEncoder(bodyBuf).Encode(reqBody)
	if err != nil {
		log.Fatalf("Could not encode json body: %s", err)
	}

	path := fmt.Sprintf("/api/tenants/%s/users/%s/job-requisitions/role-recruiter/%s/job-applications/%s/applicant-decision", input.TenantId, inputRecruiter, input.JobRequisitionId, input.Id)
	r, err := http.NewRequest("POST", path, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultRecruiter.Id, s.defaultRecruiter.TenantId, s.defaultRecruiter.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 204)

	s.expectSelectQueryToReturnOneRow(
		"job_application",
		map[string]any{
			"id":                 input.Id,
			"tenant_id":          input.TenantId,
			"job_requisition_id": input.JobRequisitionId,
			"applicant_decision": input.ApplicantDecision,
		},
	)

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"JOB-APPLICATION-APPLICANT-REJECTED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestRecruiterSetApplicantAcceptanceShouldPreventIdExploit() {
	input := storage.JobApplication{
		Id:                s.defaultJobApplication.Id,
		TenantId:          s.defaultJobApplication.TenantId,
		JobRequisitionId:  s.defaultJobApplication.JobRequisitionId,
		ApplicantDecision: "ACCEPTED",
	}
	inputRecruiter := s.defaultUser.Id // Not the recruiter, should trigger a 403 error

	// Sets recruiter shortlist, interview date, hiring manager offer, and offer start date
	query := "UPDATE job_application SET recruiter_decision = 'SHORTLISTED', interview_date = '2024-01-29', hiring_manager_decision = 'OFFERED', offer_start_date = '2024-02-01' WHERE id = $1"
	_, err := s.dbRootConn.Exec(query, input.Id)
	if err != nil {
		log.Fatalf("Could not set recruiter shortlist & interview date: %s", err)
	}	

	// Temporarily give the defaultUser recruiter acceptance rights for the purpose of this test
	query = "INSERT INTO casbin_rule (Ptype, V0, V1, V2, V3) VALUES ('p', $1, $2, $3, $4)"
	resourcePath := fmt.Sprintf("/api/tenants/%s/users/%s/job-requisitions/role-recruiter/{jobReqId}/job-applications/{jobAppId}/applicant-decision", s.defaultTenant.Id, inputRecruiter)
	_, err = s.dbRootConn.Exec(query, inputRecruiter, s.defaultJobApplication.TenantId, resourcePath, "POST")
	if err != nil {
		log.Fatalf("Could not give user temporary recruiter rights: %s", err)
	}
	err = s.router.authEnforcer.LoadPolicy()
	if err != nil {
		log.Fatalf("Could not reload enforcer to give user temporary recruiter rights: %s", err)
	}

	type requestBody struct {
		ApplicantDecision string
	}
	reqBody := requestBody{
		ApplicantDecision: input.ApplicantDecision,
	}
	bodyBuf := new(bytes.Buffer)
	err = json.NewEncoder(bodyBuf).Encode(reqBody)
	if err != nil {
		log.Fatalf("Could not encode json body: %s", err)
	}

	path := fmt.Sprintf("/api/tenants/%s/users/%s/job-requisitions/role-recruiter/%s/job-applications/%s/applicant-decision", input.TenantId, inputRecruiter, input.JobRequisitionId, input.Id)
	r, err := http.NewRequest("POST", path, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultUser.Id, s.defaultUser.TenantId, s.defaultUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 404)
	s.expectErrorCode(w, "RESOURCE-NOT-FOUND-ERROR")

	s.expectSelectQueryToReturnNoRows(
		"job_application",
		map[string]any{
			"id":                 input.Id,
			"tenant_id":          input.TenantId,
			"job_requisition_id": input.JobRequisitionId,
			"applicant_decision": input.ApplicantDecision,
		},
	)

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"RESOURCE-NOT-FOUND-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestRecruiterSetApplicantAcceptanceValidatesInput() {
	input := storage.JobApplication{
		Id:                s.defaultJobApplication.Id,
		TenantId:          s.defaultJobApplication.TenantId,
		JobRequisitionId:  s.defaultJobApplication.JobRequisitionId,
		ApplicantDecision: "invalid decision",
	}
	inputRecruiter := s.defaultRecruiter.Id

	// Sets recruiter shortlist, interview date, hiring manager offer, and offer start date
	query := "UPDATE job_application SET recruiter_decision = 'SHORTLISTED', interview_date = '2024-01-29', hiring_manager_decision = 'OFFERED', offer_start_date = '2024-02-01' WHERE id = $1"
	_, err := s.dbRootConn.Exec(query, input.Id)
	if err != nil {
		log.Fatalf("Could not set recruiter shortlist & interview date: %s", err)
	}	

	type requestBody struct {
		ApplicantDecision string
	}
	reqBody := requestBody{
		ApplicantDecision: input.ApplicantDecision,
	}
	bodyBuf := new(bytes.Buffer)
	err = json.NewEncoder(bodyBuf).Encode(reqBody)
	if err != nil {
		log.Fatalf("Could not encode json body: %s", err)
	}

	path := fmt.Sprintf("/api/tenants/%s/users/%s/job-requisitions/role-recruiter/%s/job-applications/%s/applicant-decision", input.TenantId, inputRecruiter, input.JobRequisitionId, input.Id)
	r, err := http.NewRequest("POST", path, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultRecruiter.Id, s.defaultRecruiter.TenantId, s.defaultRecruiter.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 400)
	s.expectErrorCode(w, "INPUT-VALIDATION-ERROR")

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"INPUT-VALIDATION-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestRecruiterSetApplicantAcceptanceValidatesJobRequisition() {
	input := storage.JobApplication{
		Id:                s.defaultJobApplication.Id,
		TenantId:          s.defaultJobApplication.TenantId,
		JobRequisitionId:  s.defaultJobApplication.JobRequisitionId,
		ApplicantDecision: "ACCEPTED",
	}
	inputRecruiter := s.defaultRecruiter.Id

	// Sets recruiter shortlist, interview date, hiring manager offer, and offer start date
	query := "UPDATE job_application SET recruiter_decision = 'SHORTLISTED', interview_date = '2024-01-29', hiring_manager_decision = 'OFFERED', offer_start_date = '2024-02-01' WHERE id = $1"
	_, err := s.dbRootConn.Exec(query, input.Id)
	if err != nil {
		log.Fatalf("Could not set recruiter shortlist & interview date: %s", err)
	}	

	// Remove HR approval from the job requisition
	query = "UPDATE job_requisition SET hr_approver_decision = 'REJECTED' WHERE id = $1"
	_, err = s.dbRootConn.Exec(query, input.JobRequisitionId)
	if err != nil {
		log.Fatalf("Could not remove hr approval: %s", err)
	}

	type requestBody struct {
		ApplicantDecision string
	}
	reqBody := requestBody{
		ApplicantDecision: input.ApplicantDecision,
	}
	bodyBuf := new(bytes.Buffer)
	err = json.NewEncoder(bodyBuf).Encode(reqBody)
	if err != nil {
		log.Fatalf("Could not encode json body: %s", err)
	}

	path := fmt.Sprintf("/api/tenants/%s/users/%s/job-requisitions/role-recruiter/%s/job-applications/%s/applicant-decision", input.TenantId, inputRecruiter, input.JobRequisitionId, input.Id)
	r, err := http.NewRequest("POST", path, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultRecruiter.Id, s.defaultRecruiter.TenantId, s.defaultRecruiter.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 403)
	s.expectErrorCode(w, "MISSING-HR-APPROVAL-ERROR")

	s.expectSelectQueryToReturnNoRows(
		"job_application",
		map[string]any{
			"id":                 input.Id,
			"tenant_id":          input.TenantId,
			"job_requisition_id": input.JobRequisitionId,
			"applicant_decision": input.ApplicantDecision,
		},
	)

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"MISSING-HR-APPROVAL-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}
