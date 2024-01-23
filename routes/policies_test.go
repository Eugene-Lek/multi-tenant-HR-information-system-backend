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

func (s *IntegrationTestSuite) TestCreatePolicies() {
	wantPolicies := storage.Policies{
		Subject:  "TENANT_ROLE_ADMIN",
		TenantId: s.defaultTenant.Id,
		Resources: []storage.Resource{
			{
				Path:   "/api/tenants/*",
				Method: "POST",
			},
			{
				Path:   "/api/tenants/*/divisions/*",
				Method: "POST",
			},
		},
	}

	type requestBody struct {
		Subject   string
		Resources []storage.Resource
	}
	reqBody := requestBody{
		Subject:   wantPolicies.Subject,
		Resources: wantPolicies.Resources,
	}
	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)

	r, err := http.NewRequest("POST", fmt.Sprintf("/api/tenants/%s/policies", wantPolicies.TenantId), bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultUser.Id, s.defaultUser.TenantId, s.defaultUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 201)

	for _, resource := range wantPolicies.Resources {
		s.expectSelectQueryToReturnOneRow(
			"casbin_rule",
			map[string]string{
				"Ptype": "p",
				"V0":    wantPolicies.Subject,
				"V1":    wantPolicies.TenantId,
				"V2":    resource.Path,
				"V3":    resource.Method,
			},
		)
	}

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"POLICIES-CREATED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestCreatePoliciesInvalidInput() {
	wantPolicies := storage.Policies{
		Subject:  "TENANT_ROLE_ADMIN",
		TenantId: s.defaultTenant.Id,
		Resources: []storage.Resource{
			{
				Path:   "/api/tenants/*",
				Method: "POST",
			},
			{
				Path:   "/api/tenants/*/divisions/*",
				Method: "",
			},
		},
	}

	type requestBody struct {
		Subject   string
		Resources []storage.Resource
	}
	reqBody := requestBody{
		Subject:   wantPolicies.Subject,
		Resources: wantPolicies.Resources,
	}
	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)

	r, err := http.NewRequest("POST", fmt.Sprintf("/api/tenants/%s/policies", wantPolicies.TenantId), bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultUser.Id, s.defaultUser.TenantId, s.defaultUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 400)
	s.expectErrorCode(w, "INPUT-VALIDATION-ERROR")

	for _, resource := range wantPolicies.Resources {
		s.expectSelectQueryToReturnNoRows(
			"casbin_rule",
			map[string]string{
				"Ptype": "p",
				"V0":    wantPolicies.Subject,
				"V1":    wantPolicies.TenantId,
				"V2":    resource.Path,
				"V3":    resource.Method,
			},
		)
	}

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"INPUT-VALIDATION-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestCreatePoliciesViolatesUniqueConstraint() {
	wantPolicies := s.defaultPolicies

	type requestBody struct {
		Subject   string
		Resources []storage.Resource
	}
	reqBody := requestBody{
		Subject:   wantPolicies.Subject,
		Resources: wantPolicies.Resources,
	}
	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)

	r, err := http.NewRequest("POST", fmt.Sprintf("/api/tenants/%s/policies", wantPolicies.TenantId), bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultUser.Id, s.defaultUser.TenantId, s.defaultUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 409)
	s.expectErrorCode(w, "UNIQUE-VIOLATION-ERROR")

	for _, resource := range wantPolicies.Resources {
		s.expectSelectQueryToReturnOneRow(
			"casbin_rule",
			map[string]string{
				"Ptype": "p",
				"V0":    wantPolicies.Subject,
				"V1":    wantPolicies.TenantId,
				"V2":    resource.Path,
				"V3":    resource.Method,
			},
		)
	}

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"UNIQUE-VIOLATION-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestCreateRoleAssignment() {
	wantRoleAssignment := storage.RoleAssignment{
		UserId:   s.defaultUser.Id,
		Role:     "TENANT_ROLE_ADMIN",
		TenantId: s.defaultTenant.Id,
	}

	seedPolicy := storage.Policies{
		Subject:  wantRoleAssignment.Role,
		TenantId: wantRoleAssignment.TenantId,
		Resources: []storage.Resource{
			{
				Path:   fmt.Sprintf("/api/tenants/%s/roles/*/policies", wantRoleAssignment.TenantId),
				Method: "POST",
			},
		},
	}

	// Seed another policy
	query := "INSERT INTO casbin_rule (Ptype, V0, V1, V2, V3) VALUES ('p', $1, $2, $3, $4)"
	s.dbRootConn.Exec(query, seedPolicy.Subject, seedPolicy.TenantId, seedPolicy.Resources[0].Path, seedPolicy.Resources[0].Method)

	r, err := http.NewRequest("POST", fmt.Sprintf("/api/tenants/%s/users/%s/roles/%s", wantRoleAssignment.TenantId, wantRoleAssignment.UserId, wantRoleAssignment.Role), nil)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultUser.Id, s.defaultUser.TenantId, s.defaultUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 201)

	s.expectSelectQueryToReturnOneRow(
		"casbin_rule",
		map[string]string{
			"Ptype": "g",
			"V0":    wantRoleAssignment.UserId,
			"V1":    wantRoleAssignment.Role,
			"V2":    wantRoleAssignment.TenantId,
		},
	)

	// Check that the authorization enforcer was reloaded
	authorized, err := s.router.authEnforcer.Enforce(wantRoleAssignment.UserId, wantRoleAssignment.TenantId, seedPolicy.Resources[0].Path, seedPolicy.Resources[0].Method)
	s.Equal(nil, err)
	s.Equal(true, authorized, "User should be authorized")

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"ROLE-ASSIGNMENT-CREATED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"AUTHORIZATION-ENFORCER-RELOADED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestCreateRoleAssignmentInvalidInput() {
	wantRoleAssignment := storage.RoleAssignment{
		UserId:   s.defaultUser.Id,
		Role:     "  ",
		TenantId: s.defaultTenant.Id,
	}

	r, err := http.NewRequest("POST", fmt.Sprintf("/api/tenants/%s/users/%s/roles/%s", wantRoleAssignment.TenantId, wantRoleAssignment.UserId, wantRoleAssignment.Role), nil)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultUser.Id, s.defaultUser.TenantId, s.defaultUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 400)
	s.expectErrorCode(w, "INPUT-VALIDATION-ERROR")

	s.expectSelectQueryToReturnNoRows(
		"casbin_rule",
		map[string]string{
			"Ptype": "g",
			"V0":    wantRoleAssignment.UserId,
			"V1":    wantRoleAssignment.Role,
			"V2":    wantRoleAssignment.TenantId,
		},
	)

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"INPUT-VALIDATION-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestCreateRoleAssignmentViolatesUniqueConstraint() {
	wantRoleAssignment := storage.RoleAssignment{
		UserId:   s.defaultUser.Id,
		Role:     "ROOT_ROLE_ADMIN",
		TenantId: s.defaultTenant.Id,
	}

	r, err := http.NewRequest("POST", fmt.Sprintf("/api/tenants/%s/users/%s/roles/%s", wantRoleAssignment.TenantId, wantRoleAssignment.UserId, wantRoleAssignment.Role), nil)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultUser.Id, s.defaultUser.TenantId, s.defaultUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 409)
	s.expectErrorCode(w, "UNIQUE-VIOLATION-ERROR")

	s.expectSelectQueryToReturnOneRow(
		"casbin_rule",
		map[string]string{
			"Ptype": "g",
			"V0":    wantRoleAssignment.UserId,
			"V1":    wantRoleAssignment.Role,
			"V2":    wantRoleAssignment.TenantId,
		},
	)

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"UNIQUE-VIOLATION-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}
