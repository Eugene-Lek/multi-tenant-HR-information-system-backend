package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"multi-tenant-HR-information-system-backend/httperror"
	"multi-tenant-HR-information-system-backend/storage"
)

func (router *Router) handleCreateJobApplication(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 2*1024*1024) // Limit the request body size to 2MB (json + file size)

	type requestData struct {
		JobRequisitionId string
		FirstName        string
		LastName         string
		CountryCode      string
		PhoneNumber      string
		Email            string
	}
	var reqData requestData
	data := r.FormValue("data")
	err := json.Unmarshal([]byte(data), &reqData)
	if err != nil {
		sendToErrorHandlingMiddleware(httperror.NewInternalServerError(err), r)
		return
	}

	vars := mux.Vars(r)

	file, header, err := r.FormFile("resume")
	if _, ok := err.(*http.MaxBytesError); ok {
		sendToErrorHandlingMiddleware(ErrFileTooBig, r)
	}
	if err != nil {
		sendToErrorHandlingMiddleware(httperror.NewInternalServerError(err), r)
		return
	}
	defer file.Close()

	type Input struct {
		Id               string `validate:"required,notBlank,uuid" name:"job application id"`
		TenantId         string `validate:"required,notBlank,uuid" name:"tenant id"`
		JobRequisitionId string `validate:"required,notBlank,uuid" name:"job requisition id"`
		FirstName        string `validate:"required,notBlank,alpha" name:"first name"`
		LastName         string `validate:"required,notBlank,alpha" name:"last name"`
		CountryCode      string `validate:"required,notBlank,number" name:"first name"`
		PhoneNumber      string `validate:"required,notBlank,number" name:"phone number"`
		Email            string `validate:"required,notBlank" name:"email"`
		FileExtension    string `validate:"required,notBlank,oneof=.pdf .docx" name:"file extension"`
	}
	input := Input{
		Id:               vars["jobApplicationId"],
		TenantId:         vars["tenantId"],
		JobRequisitionId: reqData.JobRequisitionId,
		FirstName:        reqData.FirstName,
		LastName:         reqData.LastName,
		CountryCode:      reqData.CountryCode,
		PhoneNumber:      reqData.PhoneNumber,
		Email:            reqData.Email,
		FileExtension:    filepath.Ext(header.Filename),
	}
	translator := getTranslator(r)
	err = validateStruct(router.validate, translator, input)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	resumeS3Url, err := router.fileStorage.UploadResume(file, input.JobRequisitionId, input.FirstName, input.LastName, input.FileExtension)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	jobApplication := storage.JobApplication{
		Id:               input.Id,
		TenantId:         input.TenantId,
		JobRequisitionId: input.JobRequisitionId,
		FirstName:        input.FirstName,
		LastName:         input.LastName,
		CountryCode:      input.CountryCode,
		PhoneNumber:      input.PhoneNumber,
		Email:            input.Email,
		ResumeS3Url:      resumeS3Url,
	}
	err = router.storage.CreateJobApplication(jobApplication)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	reqLogger := getRequestLogger(r)
	reqLogger.Info("JOB-APPLICATION-CREATED", "jobApplicationId", jobApplication.Id, "tenantId", jobApplication.TenantId, "resumeS3Url", jobApplication.ResumeS3Url)

	w.WriteHeader(http.StatusCreated)
}

func (router *Router) handleSetRecruiterDecision(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		RecruiterDecision string
	}
	var reqBody requestBody
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		sendToErrorHandlingMiddleware(ErrInvalidJSON, r)
		return
	}
	vars := mux.Vars(r)

	type Input struct {
		Id                string `validate:"required,notBlank,uuid" name:"job application id"`
		TenantId          string `validate:"required,notBlank,uuid" name:"tenant id"`
		JobRequisitionId  string `validate:"required,notBlank,uuid" name:"job requisition id"`
		Recruiter         string `validate:"required,notBlank,uuid" name:"recruiter id"`
		RecruiterDecision string `validate:"required,notBlank,oneof=SHORTLISTED REJECTED" name:"recruiter decision"`
	}
	input := Input{
		Id:                vars["jobApplicationId"],
		TenantId:          vars["tenantId"],
		JobRequisitionId:  vars["jobRequisitionId"],
		Recruiter:         vars["userId"],
		RecruiterDecision: reqBody.RecruiterDecision,
	}
	translator := getTranslator(r)
	err = validateStruct(router.validate, translator, input)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	// Check that neither the Supervisor nor HR approver have not rescinded their approval of the corresponding job requisition
	// Must be filtered by recruiter to ensure that the user is assigned to the requisition as a recruiter
	jobReqfilter := storage.JobRequisition{
		Id:        input.JobRequisitionId,
		TenantId:  input.TenantId,
		Recruiter: input.Recruiter,
	}
	jobRequisitions, err := router.storage.GetJobRequisitions(jobReqfilter)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}
	if len(jobRequisitions) == 0 {
		sendToErrorHandlingMiddleware(Err404NotFound, r)
		return
	}
	if jobRequisitions[0].SupervisorDecision != "APPROVED" {
		sendToErrorHandlingMiddleware(ErrMissingSupervisorApproval, r)
		return
	}
	if jobRequisitions[0].HrApproverDecision != "APPROVED" {
		sendToErrorHandlingMiddleware(ErrMissingHrApproval, r)
		return
	}

	// Must be filtered by job requisition id to ensure that the altered job application belongs to the given job requisition,
	// which we previously verified "belongs to" the user
	newValues := storage.JobApplication{
		RecruiterDecision: input.RecruiterDecision,
	}
	filter := storage.JobApplication{
		Id:               input.Id,
		TenantId:         input.TenantId,
		JobRequisitionId: input.JobRequisitionId,
	}
	err = router.storage.UpdateJobApplication(newValues, filter)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	reqLogger := getRequestLogger(r)
	if input.RecruiterDecision == "SHORTLISTED" {
		reqLogger.Info("JOB-APPLICATION-RECRUITER-SHORTLISTED", "jobApplicationId", input.Id, "tenantId", input.TenantId, "recruiter", input.Recruiter)
	} else if input.RecruiterDecision == "REJECTED" {
		reqLogger.Info("JOB-APPLICATION-RECRUITER-REJECTED", "jobApplicationId", input.Id, "tenantId", input.TenantId, "recruiter", input.Recruiter)
	}

	w.WriteHeader(http.StatusNoContent)
}

func (router *Router) handleRecruiterSetInterviewDate(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		InterviewDate string
	}
	var reqBody requestBody
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		sendToErrorHandlingMiddleware(ErrInvalidJSON, r)
		return
	}
	vars := mux.Vars(r)

	type Input struct {
		Id               string `validate:"required,notBlank,uuid" name:"job application id"`
		TenantId         string `validate:"required,notBlank,uuid" name:"tenant id"`
		JobRequisitionId string `validate:"required,notBlank,uuid" name:"job requisition id"`
		Recruiter        string `validate:"required,notBlank,uuid" name:"recruiter id"`
		InterviewDate    string `validate:"required,notBlank,isIsoDate" name:"interview date"`
	}
	input := Input{
		Id:               vars["jobApplicationId"],
		TenantId:         vars["tenantId"],
		JobRequisitionId: vars["jobRequisitionId"],
		Recruiter:        vars["userId"],
		InterviewDate:    reqBody.InterviewDate,
	}
	translator := getTranslator(r)
	err = validateStruct(router.validate, translator, input)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	// Check that neither the Supervisor nor HR approver have not rescinded their approval of the corresponding job requisition
	// Must be filtered by recruiter to ensure that the user is assigned to the requisition as a recruiter
	jobReqfilter := storage.JobRequisition{
		Id:        input.JobRequisitionId,
		TenantId:  input.TenantId,
		Recruiter: input.Recruiter,
	}
	jobRequisitions, err := router.storage.GetJobRequisitions(jobReqfilter)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}
	if len(jobRequisitions) == 0 {
		sendToErrorHandlingMiddleware(Err404NotFound, r)
		return
	}
	if jobRequisitions[0].SupervisorDecision != "APPROVED" {
		sendToErrorHandlingMiddleware(ErrMissingSupervisorApproval, r)
		return
	}
	if jobRequisitions[0].HrApproverDecision != "APPROVED" {
		sendToErrorHandlingMiddleware(ErrMissingHrApproval, r)
		return
	}

	// Must be filtered by job requisition id to ensure that the altered job application belongs to the given job requisition,
	// which we previously verified "belongs to" the user
	newValues := storage.JobApplication{
		InterviewDate: input.InterviewDate,
	}
	filter := storage.JobApplication{
		Id:               input.Id,
		TenantId:         input.TenantId,
		JobRequisitionId: input.JobRequisitionId,
	}
	err = router.storage.UpdateJobApplication(newValues, filter)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	reqLogger := getRequestLogger(r)
	reqLogger.Info("JOB-APPLICATION-INTERVIEW-DATE-SET", "jobApplicationId", input.Id, "tenantId", input.TenantId, "recruiter", input.Recruiter)

	w.WriteHeader(http.StatusNoContent)
}

func (router *Router) handleSetHiringManagerDecision(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		HiringManagerDecision string
		OfferStartDate        string
		OfferEndDate          string
	}
	var reqBody requestBody
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		sendToErrorHandlingMiddleware(ErrInvalidJSON, r)
		return
	}
	vars := mux.Vars(r)

	type Input struct {
		Id                    string `validate:"required,notBlank,uuid" name:"job application id"`
		TenantId              string `validate:"required,notBlank,uuid" name:"tenant id"`
		JobRequisitionId      string `validate:"required,notBlank,uuid" name:"job requisition id"`
		Requestor             string `validate:"required,notBlank,uuid" name:"hiring manager id"`
		HiringManagerDecision string `validate:"required,notBlank,oneof=OFFERED REJECTED" name:"hiring manager decision"`
		OfferStartDate        string `validate:"required,notBlank,isIsoDate" name:"offer start date"`
		OfferEndDate          string `validate:"omitempty,notBlank,isIsoDate,validPositionAssignmentDuration" name:"offer end date"`
	}
	input := Input{
		Id:                    vars["jobApplicationId"],
		TenantId:              vars["tenantId"],
		JobRequisitionId:      vars["jobRequisitionId"],
		Requestor:             vars["userId"],
		HiringManagerDecision: reqBody.HiringManagerDecision,
		OfferStartDate:        reqBody.OfferStartDate,
		OfferEndDate:          reqBody.OfferEndDate,
	}
	translator := getTranslator(r)
	err = validateStruct(router.validate, translator, input)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	// Check that neither the Supervisor nor HR approver have not rescinded their approval of the corresponding job requisition
	// Must be filtered by requestor to ensure that the user is assigned to the requisition as the requestor (hiring manager)
	jobReqfilter := storage.JobRequisition{
		Id:        input.JobRequisitionId,
		TenantId:  input.TenantId,
		Requestor: input.Requestor,
	}
	jobRequisitions, err := router.storage.GetJobRequisitions(jobReqfilter)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}
	if len(jobRequisitions) == 0 {
		sendToErrorHandlingMiddleware(Err404NotFound, r)
		return
	}
	if jobRequisitions[0].SupervisorDecision != "APPROVED" {
		sendToErrorHandlingMiddleware(ErrMissingSupervisorApproval, r)
		return
	}
	if jobRequisitions[0].HrApproverDecision != "APPROVED" {
		sendToErrorHandlingMiddleware(ErrMissingHrApproval, r)
		return
	}

	// Must be filtered by job requisition id to ensure that the altered job application belongs to the given job requisition,
	// which we previously verified "belongs to" the user
	newValues := storage.JobApplication{
		HiringManagerDecision: input.HiringManagerDecision,
		OfferStartDate: input.OfferStartDate,
		OfferEndDate: input.OfferEndDate,
	}
	filter := storage.JobApplication{
		Id:               input.Id,
		TenantId:         input.TenantId,
		JobRequisitionId: input.JobRequisitionId,
	}
	err = router.storage.UpdateJobApplication(newValues, filter)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	reqLogger := getRequestLogger(r)
	if input.HiringManagerDecision == "OFFERED" {
		reqLogger.Info("JOB-APPLICATION-HIRING-MANAGER-OFFERED", "jobApplicationId", input.Id, "tenantId", input.TenantId, "hiringManager", input.Requestor)
	} else if input.HiringManagerDecision == "REJECTED" {
		reqLogger.Info("JOB-APPLICATION-HIRING-MANAGER-REJECTED", "jobApplicationId", input.Id, "tenantId", input.TenantId, "hiringManager", input.Requestor)
	}

	w.WriteHeader(http.StatusNoContent)
}

func (router *Router) handleSetApplicantDecision(w http.ResponseWriter, r *http.Request) {
	// TODO: Add the upload of signed documents to the database & s3

	type requestBody struct {
		ApplicantDecision string
	}

	type responseBody struct {
		Password      string `json:"password"`
		TotpSecretKey string `json:"totpSecretKey"`
	}

	var reqBody requestBody
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		sendToErrorHandlingMiddleware(ErrInvalidJSON, r)
		return
	}
	vars := mux.Vars(r)

	type Input struct {
		Id                string `validate:"required,notBlank,uuid" name:"job application id"`
		TenantId          string `validate:"required,notBlank,uuid" name:"tenant id"`
		JobRequisitionId  string `validate:"required,notBlank,uuid" name:"job requisition id"`
		Recruiter         string `validate:"required,notBlank,uuid" name:"recruiter id"`
		ApplicantDecision string `validate:"required,notBlank,oneof=ACCEPTED REJECTED" name:"applicant decision"`
	}
	input := Input{
		Id:                vars["jobApplicationId"],
		TenantId:          vars["tenantId"],
		JobRequisitionId:  vars["jobRequisitionId"],
		Recruiter:         vars["userId"],
		ApplicantDecision: reqBody.ApplicantDecision,
	}
	translator := getTranslator(r)
	err = validateStruct(router.validate, translator, input)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	// Check that neither the Supervisor nor HR approver have not rescinded their approval of the corresponding job requisition
	// Must be filtered by recruiter to ensure that the user is assigned to the requisition as a recruiter
	jobReqfilter := storage.JobRequisition{
		Id:        input.JobRequisitionId,
		TenantId:  input.TenantId,
		Recruiter: input.Recruiter,
	}
	jobRequisitions, err := router.storage.GetJobRequisitions(jobReqfilter)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}
	if len(jobRequisitions) == 0 {
		sendToErrorHandlingMiddleware(Err404NotFound, r)
		return
	}
	if jobRequisitions[0].SupervisorDecision != "APPROVED" {
		sendToErrorHandlingMiddleware(ErrMissingSupervisorApproval, r)
		return
	}
	if jobRequisitions[0].HrApproverDecision != "APPROVED" {
		sendToErrorHandlingMiddleware(ErrMissingHrApproval, r)
		return
	}

	if input.ApplicantDecision == "ACCEPTED" {
		// Retrieve the tenant to get its name, which is used in the email domain
		tenants, err := router.storage.GetTenants(storage.Tenant{Id: input.TenantId})
		if err != nil {
			sendToErrorHandlingMiddleware(err, r)
			return
		}

		// Must be filtered by job requisition id to ensure that the altered job application belongs to the given job requisition,
		// which we previously verified "belongs to" the user
		jobAppFilter := storage.JobApplication{
			Id:               input.Id,
			TenantId:         input.TenantId,
			JobRequisitionId: input.JobRequisitionId,
		}
		jobApplications, err := router.storage.GetJobApplications(jobAppFilter)
		if err != nil {
			sendToErrorHandlingMiddleware(err, r)
			return
		}
		if len(jobApplications) == 0 {
			sendToErrorHandlingMiddleware(Err404NotFound, r)
			return
		}

		firstName := strings.ReplaceAll(jobApplications[0].FirstName, " ", "_")
		lastName := strings.ReplaceAll(jobApplications[0].LastName, " ", "_")
		emailDomain := strings.ReplaceAll(tenants[0].Name, " ", "")
		email := fmt.Sprintf("%s_%s@%s.com", firstName, lastName, emailDomain)
		password, hashedPassword, totp_secret_key, err := generateDefaultCredentials(email)
		if err != nil {
			sendToErrorHandlingMiddleware(err, r)
			return
		}

		newUser := storage.User{
			Id:            uuid.New().String(),
			TenantId:      input.TenantId,
			Email:         email,
			Password:      hashedPassword,
			TotpSecretKey: totp_secret_key,
		}

		err = router.storage.OnboardNewHire(jobApplications[0], newUser)
		if err != nil {
			sendToErrorHandlingMiddleware(err, r)
			return
		}

		reqLogger := getRequestLogger(r)		
		reqLogger.Info("JOB-APPLICATION-APPLICANT-ACCEPTED", "jobApplicationId", input.Id, "tenantId", input.TenantId, "recruiter", input.Recruiter)

		w.WriteHeader(http.StatusCreated)
		w.Header().Add("content-type", "application/json")
		resBody := responseBody{
			Password:      password,
			TotpSecretKey: totp_secret_key,
		}
		json.NewEncoder(w).Encode(resBody)

	} else if input.ApplicantDecision == "REJECTED" {
		// Must be filtered by job requisition id to ensure that the altered job application belongs to the given job requisition,
		// which we previously verified "belongs to" the user
		newValues := storage.JobApplication{
			ApplicantDecision: input.ApplicantDecision,
		}
		filter := storage.JobApplication{
			Id:               input.Id,
			TenantId:         input.TenantId,
			JobRequisitionId: input.JobRequisitionId,
		}
		err = router.storage.UpdateJobApplication(newValues, filter)
		if err != nil {
			sendToErrorHandlingMiddleware(err, r)
			return
		}

		reqLogger := getRequestLogger(r)		
		reqLogger.Info("JOB-APPLICATION-APPLICANT-REJECTED", "jobApplicationId", input.Id, "tenantId", input.TenantId, "recruiter", input.Recruiter)

		w.WriteHeader(http.StatusNoContent)		
	}
}
