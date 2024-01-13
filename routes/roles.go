package routes

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"

	"multi-tenant-HR-information-system-backend/storage"
)

func (router *Router) handleCreatePolicies(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Resources []storage.Resource
	}

	var reqBody requestBody
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		sendToErrorHandlingMiddleware(NewInvalidJSONError(), r)
		return
	}

	vars := mux.Vars(r)

	policies := storage.Policies{
		Role:      vars["roleName"],
		TenantId:  vars["tenantId"],
		Resources: reqBody.Resources,
	}

	// Input validation
	translator, err := getAppropriateTranslator(r, router.universalTranslator)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}
	err = storage.ValidateStruct(router.validate, translator, policies)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	err = router.storage.CreatePolicies(policies)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	w.WriteHeader(http.StatusCreated)

	reqLogger := getRequestLogger(r)
	reqLogger.Info("POLICIES-CREATED", "roleName", policies.Role, "tenantId", policies.TenantId)
}

func (router *Router) handleCreateRoleAssignment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	roleAssignment := storage.RoleAssignment{
		UserId:   vars["userId"],
		Role:     vars["roleName"],
		TenantId: vars["tenantId"],
	}

	translator, err := getAppropriateTranslator(r, router.universalTranslator)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}
	err = storage.ValidateStruct(router.validate, translator, roleAssignment)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	err = router.storage.CreateRoleAssignment(roleAssignment)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	w.WriteHeader(http.StatusCreated)

	reqLogger := getRequestLogger(r)
	reqLogger.Info("ROLE-ASSIGNMENT-CREATED", "userId", roleAssignment.UserId, "role", roleAssignment.Role, "tenantId", roleAssignment.TenantId)

	// Re-load the updated policy into the enforcer
	err = router.authEnforcer.LoadPolicy()
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	reqLogger.Info("AUTHORIZATION-ENFORCER-RELOADED")
}
