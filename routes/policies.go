package routes

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"multi-tenant-HR-information-system-backend/storage"
)

func (router *Router) handleCreatePolicies(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Subject   string
		Resources []storage.Resource
	}

	var reqBody requestBody
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		sendToErrorHandlingMiddleware(ErrInvalidJSON, r)
		return
	}

	vars := mux.Vars(r)

	// Input validation
	type Resource struct {
		Path   string `validate:"required,notBlank" name:"resource path"`
		Method string `validate:"required,notBlank,oneof=POST GET PUT DELETE" name:"resource method"`
	}
	type Input struct {
		Subject   string     `validate:"required,notBlank" name:"role name"`
		TenantId  string     `validate:"required,notBlank,uuid" name:"tenant id"`
		Resources []Resource `validate:"dive"`
	}
	resources := []Resource{}
	for _, resource := range reqBody.Resources {
		resources = append(resources, Resource{resource.Path, resource.Method})
	}
	input := Input{
		Subject:   reqBody.Subject,
		TenantId:  vars["tenantId"],
		Resources: resources,
	}
	translator := getTranslator(r)
	err = validateStruct(router.validate, translator, input)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	policies := storage.Policies{
		Subject:   input.Subject,
		TenantId:  input.TenantId,
		Resources: reqBody.Resources,
	}
	err = router.storage.CreatePolicies(policies)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	reqLogger := getRequestLogger(r)
	reqLogger.Info("POLICIES-CREATED", "subject", policies.Subject, "tenantId", policies.TenantId)

	w.WriteHeader(http.StatusCreated)
}

func (router *Router) handleCreateRoleAssignment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	type Input struct {
		UserId   string `validate:"required,notBlank,uuid" name:"user id"`
		Role     string `validate:"required,notBlank" name:"role name"`
		TenantId string `validate:"required,notBlank,uuid" name:"tenant id"`
	}
	input := Input{
		UserId:   vars["userId"],
		Role:     vars["roleName"],
		TenantId: vars["tenantId"],
	}
	translator := getTranslator(r)
	err := validateStruct(router.validate, translator, input)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	roleAssignment := storage.RoleAssignment{
		UserId:   input.UserId,
		Role:     input.Role,
		TenantId: input.TenantId,
	}
	err = router.storage.CreateRoleAssignment(roleAssignment)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	reqLogger := getRequestLogger(r)
	reqLogger.Info("ROLE-ASSIGNMENT-CREATED", "userId", roleAssignment.UserId, "role", roleAssignment.Role, "tenantId", roleAssignment.TenantId)

	// Re-load the updated policy into the enforcer
	err = router.authEnforcer.LoadPolicy()
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	reqLogger.Info("AUTHORIZATION-ENFORCER-RELOADED")

	w.WriteHeader(http.StatusCreated)
}
