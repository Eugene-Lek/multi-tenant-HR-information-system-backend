package postgres

import (
	"multi-tenant-HR-information-system-backend/storage"
)

func (s *IntegrationTestSuite) TestCreatePolicies() {
	wantPolicies := storage.Policies{
		Role:     "TENANT_ROLE_ADMIN",
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

	err := s.postgres.CreatePolicies(wantPolicies)
	s.Equal(nil, err)

	for _, resource := range wantPolicies.Resources {
		s.expectSelectQueryToReturnOneRow(
			"casbin_rule",
			map[string]string{
				"Ptype": "p",
				"V0":    wantPolicies.Role,
				"V1":    wantPolicies.TenantId,
				"V2":    resource.Path,
				"V3":    resource.Method,
			},
		)
	}
}

func (s *IntegrationTestSuite) TestCreatePoliciesViolatesUniqueConstraint() {
	tests := []struct {
		name            string
		input           storage.Policies
		policyExistance []bool
	}{
		{
			"Should violate unique constraint because first policy already exists",
			storage.Policies{
				Role:     s.defaultPolicies.Role,
				TenantId: s.defaultTenant.Id,
				Resources: []storage.Resource{
					{
						Path:   "/api/tenants/*",
						Method: "POST",
					},
					{
						Path:   "/api/test",
						Method: "POST",
					},
				},
			},
			[]bool{true, false},
		},
		{
			"Should violate unique constraint because second policy already exists",
			storage.Policies{
				Role:     "ROOT_ROLE_ADMIN",
				TenantId: s.defaultTenant.Id,
				Resources: []storage.Resource{
					{
						Path:   "/api/test",
						Method: "POST",
					},
					{
						Path:   "/api/tenants/*/divisions/*",
						Method: "POST",
					},
				},
			},
			[]bool{false, true},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			err := s.postgres.CreatePolicies(test.input)
			s.expectErrorCode(err, "UNIQUE-VIOLATION-ERROR")

			for i, resource := range test.input.Resources {
				if test.policyExistance[i] {
					s.expectSelectQueryToReturnOneRow(
						"casbin_rule",
						map[string]string{
							"Ptype": "p",
							"V0":    test.input.Role,
							"V1":    test.input.TenantId,
							"V2":    resource.Path,
							"V3":    resource.Method,
						},
					)
				} else {
					s.expectSelectQueryToReturnNoRows(
						"casbin_rule",
						map[string]string{
							"Ptype": "p",
							"V0":    test.input.Role,
							"V1":    test.input.TenantId,
							"V2":    resource.Path,
							"V3":    resource.Method,
						},
					)
				}
			}
		})
	}
}

func (s *IntegrationTestSuite) TestCreatePoliciesDoesNotViolateUniqueConstraint() {
	tests := []struct {
		name  string
		input storage.Policies
	}{
		{
			"Should not violate unique constraint because role is different",
			storage.Policies{
				Role:     "TENANT_ROLE_ADMIN",
				TenantId: s.defaultTenant.Id,
				Resources: []storage.Resource{
					{
						Path:   "/api/tenants/*",
						Method: "POST",
					},
				},
			},
		},
		{
			"Should not violate unique constraint because path is different",
			storage.Policies{
				Role:     "ROOT_ROLE_ADMIN",
				TenantId: s.defaultTenant.Id,
				Resources: []storage.Resource{
					{
						Path:   "/api/test",
						Method: "POST",
					},
				},
			},
		},
		{
			"Should not violate unique constraint because method is different",
			storage.Policies{
				Role:     "ROOT_ROLE_ADMIN",
				TenantId: s.defaultTenant.Id,
				Resources: []storage.Resource{
					{
						Path:   "/api/tenants/*/divisions/*",
						Method: "GET",
					},
				},
			},
		},
		{
			"Should not violate unique constraint because tenant_id is different",
			storage.Policies{
				Role:     "ROOT_ROLE_ADMIN",
				TenantId: "afe3c4c8-b0a6-422a-b285-99123aecc6bf",
				Resources: []storage.Resource{
					{
						Path:   "/api/tenants/*/divisions/*",
						Method: "POST",
					},
				},
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			err := s.postgres.CreatePolicies(test.input)
			s.Equal(nil, err)

			for _, resource := range test.input.Resources {
				s.expectSelectQueryToReturnOneRow(
					"casbin_rule",
					map[string]string{
						"Ptype": "p",
						"V0":    test.input.Role,
						"V1":    test.input.TenantId,
						"V2":    resource.Path,
						"V3":    resource.Method,
					},
				)
			}
		})
	}
}
