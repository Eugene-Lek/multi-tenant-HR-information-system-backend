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
		sendToErrorHandlingMiddleware(NewInvalidJSONError(), r)
		return
	}

	vars := mux.Vars(r)

	tenant := storage.Tenant{
		Id:   vars["tenantId"],
		Name: body.Name,
	}

	// Input validation
	translator, err := getAppropriateTranslator(r, router.universalTranslator)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}
	err = storage.ValidateStruct(router.validate, translator, tenant)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	err = router.storage.CreateTenant(tenant)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	w.WriteHeader(http.StatusCreated)

	requestLogger := getRequestLogger(r)
	requestLogger.Info("TENANT-CREATED", "tenantId", tenant.Id, "tenant", tenant.Name)
}

func (router *Router) handleCreateDivision(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Name     string
	}

	var body requestBody
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		sendToErrorHandlingMiddleware(NewInvalidJSONError(), r)
		return
	}

	vars := mux.Vars(r)
	division := storage.Division{
		Id:       vars["divisionId"],
		TenantId: vars["tenantId"],
		Name:     body.Name,
	}

	// Input validation
	translator, err := getAppropriateTranslator(r, router.universalTranslator)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}
	err = storage.ValidateStruct(router.validate, translator, division)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	err = router.storage.CreateDivision(division)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	w.WriteHeader(http.StatusCreated)

	requestLogger := getRequestLogger(r)
	requestLogger.Info("DIVISION-CREATED", "divisionId", division.Id, "tenantId", division.TenantId, "name", division.Name)
}

func (router *Router) handleCreateDepartment(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Name       string
	}

	var body requestBody
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		sendToErrorHandlingMiddleware(NewInvalidJSONError(), r)
		return
	}

	vars := mux.Vars(r)
	department := storage.Department{
		Id:         vars["departmentId"],
		TenantId: vars["tenantId"],
		DivisionId: vars["divisionId"],
		Name:       body.Name,
	}

	// Input validation
	translator, err := getAppropriateTranslator(r, router.universalTranslator)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}
	err = storage.ValidateStruct(router.validate, translator, department)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	err = router.storage.CreateDepartment(department)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	w.WriteHeader(http.StatusCreated)
	requestLogger := getRequestLogger(r)
	requestLogger.Info("DEPARTMENT-CREATED", "tenantId", department.TenantId, "departmentId", department.Id, "divisionId", department.DivisionId, "name", department.Name)
}
