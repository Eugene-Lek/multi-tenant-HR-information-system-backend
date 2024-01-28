package routes

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"multi-tenant-HR-information-system-backend/storage"
)

func (router *Router) handleCreateTenant(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Name string
	}

	var body requestBody
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		sendToErrorHandlingMiddleware(ErrInvalidJSON, r)
		return
	}
	vars := mux.Vars(r)

	type Input struct {
		Id   string `validate:"required,notBlank,uuid" name:"tenant id"`
		Name string `validate:"required,notBlank" name:"tenant name"`
	}
	input := Input{
		Id:   vars["tenantId"],
		Name: body.Name,
	}
	translator := getTranslator(r)
	err = validateStruct(router.validate, translator, input)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	tenant := storage.Tenant{
		Id:   input.Id,
		Name: input.Name,
	}
	err = router.storage.CreateTenant(tenant)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	requestLogger := getRequestLogger(r)
	requestLogger.Info("TENANT-CREATED", "tenantId", tenant.Id, "tenant", tenant.Name)

	w.WriteHeader(http.StatusCreated)
}

func (router *Router) handleCreateDivision(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Name string
	}

	var body requestBody
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		sendToErrorHandlingMiddleware(ErrInvalidJSON, r)
		return
	}
	vars := mux.Vars(r)

	type Input struct {
		Id       string `validate:"required,notBlank,uuid" name:"division id"`
		TenantId string `validate:"required,notBlank,uuid" name:"tenant id"`
		Name     string `validate:"required,notBlank" name:"division name"`
	}
	input := Input{
		Id:       vars["divisionId"],
		TenantId: vars["tenantId"],
		Name:     body.Name,
	}
	translator := getTranslator(r)
	err = validateStruct(router.validate, translator, input)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	division := storage.Division{
		Id:       input.Id,
		TenantId: input.TenantId,
		Name:     input.Name,
	}
	err = router.storage.CreateDivision(division)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	requestLogger := getRequestLogger(r)
	requestLogger.Info("DIVISION-CREATED", "tenantId", division.TenantId, "divisionId", division.Id)

	w.WriteHeader(http.StatusCreated)
}

func (router *Router) handleCreateDepartment(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Name string
	}

	var body requestBody
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		sendToErrorHandlingMiddleware(ErrInvalidJSON, r)
		return
	}
	vars := mux.Vars(r)

	type Input struct {
		Id         string `validate:"required,notBlank,uuid" name:"department id"`
		TenantId   string `validate:"required,notBlank,uuid" name:"tenant id"`
		DivisionId string `validate:"required,notBlank,uuid" name:"division id"`
		Name       string `validate:"required,notBlank" name:"department name"`
	}
	input := Input{
		Id:         vars["departmentId"],
		TenantId:   vars["tenantId"],
		DivisionId: vars["divisionId"],
		Name:       body.Name,
	}
	translator := getTranslator(r)
	err = validateStruct(router.validate, translator, input)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	department := storage.Department{
		Id:         input.Id,
		TenantId:   input.TenantId,
		DivisionId: input.DivisionId,
		Name:       input.Name,
	}
	err = router.storage.CreateDepartment(department)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	requestLogger := getRequestLogger(r)
	requestLogger.Info("DEPARTMENT-CREATED", "tenantId", department.TenantId, "departmentId", department.Id)

	w.WriteHeader(http.StatusCreated)
}
