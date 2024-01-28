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

	"multi-tenant-HR-information-system-backend/storage"
)

func (s *IntegrationTestSuite) TestCreateJobApplication() {
	want := storage.JobApplication{
		Id:               "371bcf41-2a2b-4fb4-bc00-f8a1f248d756",
		TenantId:         s.defaultTenant.Id,
		JobRequisitionId: s.defaultJobRequisition.Id,
		FirstName:        "Eugene",
		LastName:         "Lek",
		CountryCode:      "1",
		PhoneNumber:      "987654321",
		Email:            "test123@gmail.com",
		ResumeS3Url:      fmt.Sprintf("%s/hr-information-system/job-applications/5062a285-e82b-475d-8113-daefd05dcd90/Eugene_Lek_resume.pdf", s.s3Server.URL),
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
		JobRequisitionId: want.JobRequisitionId,
		FirstName:        want.FirstName,
		LastName:         want.LastName,
		CountryCode:      want.CountryCode,
		PhoneNumber:      want.PhoneNumber,
		Email:            want.Email,
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

	path := fmt.Sprintf("/api/tenants/%s/job-applications/%s", want.TenantId, want.Id)
	r, err := http.NewRequest("POST", path, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	r.Header.Set("Content-Type", multiWriter.FormDataContentType())

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 201)

	s.expectS3ToContainFile(want.ResumeS3Url)

	log.Println(want.ResumeS3Url)
	s.expectSelectQueryToReturnOneRow(
		"job_application",
		map[string]string{
			"id":                 want.Id,
			"tenant_id":          want.TenantId,
			"job_requisition_id": want.JobRequisitionId,
			"first_name":         want.FirstName,
			"last_name":          want.LastName,
			"country_code":       want.CountryCode,
			"phone_number":       want.PhoneNumber,
			"email":              want.Email,
			"resume_s3_url":      want.ResumeS3Url,
		},
	)

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"JOB-APPLICATION-CREATED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}
