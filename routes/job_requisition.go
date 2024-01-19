package routes

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"multi-tenant-HR-information-system-backend/httperror"
	"multi-tenant-HR-information-system-backend/storage"
)

func (router *Router) handleCreateJobRequisition(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Title           string
		DepartmentId    string
		JobDescription  string
		JobRequirements string
		Supervisor      string
		HrApprover      string
	}
	var reqBody requestBody
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		sendToErrorHandlingMiddleware(NewInvalidJSONError(), r)
		return
	}
	vars := mux.Vars(r)

	type Input struct {
		Id              string `validate:"required,notBlank,uuid" name:"job requisition id"`
		TenantId        string `validate:"required,notBlank,uuid" name:"tenant id"`
		Title           string `validate:"required,notBlank" name:"position title"`
		DepartmentId    string `validate:"required,notBlank,uuid" name:"department id"`
		JobDescription  string `validate:"required,notBlank" name:"job description"`
		JobRequirements string `validate:"required,notBlank" name:"job requirements"`
		Requestor       string `validate:"required,notBlank,uuid" name:"requestor id"`
		Supervisor      string `validate:"required,notBlank,uuid" name:"supervisor id"`
		HrApprover      string `validate:"required,notBlank,uuid" name:"HR approver id"`
	}
	input := Input{
		Id:              vars["jobRequisitionId"],
		TenantId:        vars["tenantId"],
		Title:           reqBody.Title,
		DepartmentId:    reqBody.DepartmentId,
		JobDescription:  reqBody.JobDescription,
		JobRequirements: reqBody.JobRequirements,
		Requestor:       vars["userId"],
		Supervisor:      reqBody.Supervisor,
		HrApprover:      reqBody.HrApprover,
	}
	translator := getTranslator(r)
	err = validateStruct(router.validate, translator, input)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	// TODO: Verify that the supervisor provided is indeed the user's supervisor
	user := getAuthenticatedUser(r)	
	filter := storage.Position{
		TenantId: input.TenantId,		
		SupervisorIds: []string{input.Supervisor},
	}
	userPositions, err := router.storage.GetUserPositions(user.Id, filter)
	if err != nil {
		sendToErrorHandlingMiddleware(httperror.NewInternalServerError(err), r)
		return
	}
	if len(userPositions) == 0 {
		sendToErrorHandlingMiddleware(NewUnauthorisedError(), r)
		return
	}

	jobRequisition := storage.JobRequisition{
		Id:              input.Id,
		TenantId:        input.TenantId,
		Title:           input.Title,
		DepartmentId:    input.DepartmentId,
		JobDescription:  input.JobDescription,
		JobRequirements: input.JobRequirements,
		Requestor:       input.Requestor,
		Supervisor:      input.Supervisor,
		HrApprover:      input.HrApprover,
	}
	err = router.storage.CreateJobRequisition(jobRequisition)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	reqLogger := getRequestLogger(r)
	reqLogger.Info("JOB-REQUISITION-CREATED", "jobRequisitionId", jobRequisition.Id)

	// TODO: Notify approver by email

	w.WriteHeader(http.StatusCreated)
}

func (router *Router) handleSupervisorApproveJobRequisition(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Password           string
		Totp               string
		SupervisorDecision string
	}
	var reqBody requestBody
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		sendToErrorHandlingMiddleware(NewInvalidJSONError(), r)
		return
	}
	vars := mux.Vars(r)

	type Input struct {
		Id                 string `validate:"required,notBlank,uuid" name:"job requisition id"`
		TenantId           string `validate:"required,notBlank,uuid" name:"tenant id"`
		Supervisor         string `validate:"required,notBlank,uuid" name:"supervisor id"`
		SupervisorDecision string `validate:"required,notBlank,oneof=APPROVED REJECTED" name:"supervisor's decision"`
		Password           string `validate:"required,notBlank" name:"password"`
		Totp               string `validate:"required,notBlank" name:"totp"`
	}
	input := Input{
		Id:                 vars["jobRequisitionId"],
		TenantId:           vars["tenantId"],
		Supervisor:         vars["userId"],
		SupervisorDecision: reqBody.SupervisorDecision,
		Password:           reqBody.Password,
		Totp:               reqBody.Totp,
	}
	translator := getTranslator(r)
	err = validateStruct(router.validate, translator, input)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	// Validate credentials. Credentials are revalidated because approval is akin to signing off on something
	// This guards against the abuse of a logged in yet unattended computer
	user := getAuthenticatedUser(r)
	valid, err := router.validateCredentials(user.Email, user.TenantId, input.Password, input.Totp)
	if err != nil {
		sendToErrorHandlingMiddleware(httperror.NewInternalServerError(err), r)
		return
	}
	if !valid {
		sendToErrorHandlingMiddleware(NewUnauthenticatedError(), r)
		return
	}

	// TODO: Verify that the supervisor provided is indeed the user's supervisor


	newValues := storage.JobRequisition{
		SupervisorDecision: input.SupervisorDecision,
	}
	filter := storage.JobRequisition{
		Id:         input.Id,
		TenantId:   input.TenantId,
		Supervisor: input.Supervisor,
	}
	err = router.storage.UpdateJobRequisition(newValues, filter)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	reqLogger := getRequestLogger(r)
	if input.SupervisorDecision == "APPROVED" {
		reqLogger.Info("JOB-REQUISITION-SUPERVISOR-APPROVED", "jobRequisitionId", input.Id, "supervisor", input.Supervisor)
	} else if input.SupervisorDecision == "REJECTED" {
		reqLogger.Info("JOB-REQUISITION-SUPERVISOR-REJECTED", "jobRequisitionId", input.Id, "supervisor", input.Supervisor)
	}

	// TODO: Notify requestor & hr approver by email

	w.WriteHeader(http.StatusNoContent)
}

func (router *Router) handleHrApproveJobRequisition(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Password           string
		Totp               string
		HrApproverDecision string
		Recruiter          string
	}
	var reqBody requestBody
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		sendToErrorHandlingMiddleware(NewInvalidJSONError(), r)
		return
	}
	vars := mux.Vars(r)

	type Input struct {
		Id                 string `validate:"required,notBlank,uuid" name:"job requisition id"`
		TenantId           string `validate:"required,notBlank,uuid" name:"tenant id"`
		HrApprover         string `validate:"required,notBlank,uuid" name:"supervisor id"`
		HrApproverDecision string `validate:"required,notBlank,oneof=APPROVED REJECTED" name:"supervisor's decision"`
		Recruiter          string `validate:"required,notBlank,uuid" name:"recruiter id"`
		Password           string `validate:"required,notBlank" name:"password"`
		Totp               string `validate:"required,notBlank" name:"totp"`
	}
	input := Input{
		Id:                 vars["jobRequisitionId"],
		TenantId:           vars["tenantId"],
		HrApprover:         vars["userId"],
		HrApproverDecision: reqBody.HrApproverDecision,
		Recruiter:          reqBody.Recruiter,
		Password:           reqBody.Password,
		Totp:               reqBody.Totp,
	}
	translator := getTranslator(r)
	err = validateStruct(router.validate, translator, input)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	// Validate credentials. Credentials are revalidated because approval is akin to signing off on something
	// This guards against the abuse of a logged in yet unattended computer
	user := getAuthenticatedUser(r)
	valid, err := router.validateCredentials(user.Email, user.TenantId, input.Password, input.Totp)
	if err != nil {
		sendToErrorHandlingMiddleware(httperror.NewInternalServerError(err), r)
		return
	}
	if !valid {
		sendToErrorHandlingMiddleware(NewUnauthenticatedError(), r)
		return
	}

	newValues := storage.JobRequisition{
		HrApproverDecision: input.HrApproverDecision,
		Recruiter:          input.Recruiter,
	}
	filter := storage.JobRequisition{
		Id:         input.Id,
		TenantId:   input.TenantId,
		HrApprover: input.HrApprover,
	}
	err = router.storage.UpdateJobRequisition(newValues, filter)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	reqLogger := getRequestLogger(r)
	if input.HrApproverDecision == "APPROVED" {
		reqLogger.Info("JOB-REQUISITION-HR-APPROVED", "jobRequisitionId", input.Id, "hrApprover", input.HrApprover)
	} else if input.HrApproverDecision == "REJECTED" {
		reqLogger.Info("JOB-REQUISITION-HR-REJECTED", "jobRequisitionId", input.Id, "hrApprover", input.HrApprover)
	}

	// TODO: Notify requestor, superior, and recruiter

	w.WriteHeader(http.StatusNoContent)
}
