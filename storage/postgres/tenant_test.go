package postgres

import (
	"log"

	"multi-tenant-HR-information-system-backend/storage"
)

func (s *IntegrationTestSuite) TestCreateTenant() {
	wantTenant := storage.Tenant{
		Id:   "5338d729-32bd-4ad2-a8d1-22cbf81113de",
		Name: "Macdonalds",
	}

	err := s.postgres.CreateTenant(wantTenant)
	s.Equal(nil, err, "Should not return an error")

	s.expectSelectQueryToReturnOneRow(
		"tenant",
		map[string]any{
			"id":   wantTenant.Id,
			"name": wantTenant.Name,
		},
	)
}

func (s *IntegrationTestSuite) TestCreateTenantViolatesUniqueConstraint() {
	tests := []struct {
		name  string
		input storage.Tenant
	}{
		{
			"Should violate unique constraint because id is the same",
			storage.Tenant{
				Id:   s.defaultTenant.Id,
				Name: "Different Name",
			},
		},
		{
			"Should violate unique constraint because name is the same",
			storage.Tenant{
				Id:   "5338d729-32bd-4ad2-a8d1-22cbf81113de",
				Name: s.defaultTenant.Name,
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			err := s.postgres.CreateTenant(test.input)
			s.expectErrorCode(err, "UNIQUE-VIOLATION-ERROR")

			s.expectSelectQueryToReturnNoRows(
				"tenant",
				map[string]any{
					"id":   test.input.Id,
					"name": test.input.Name,
				},
			)
		})
	}
}

func (s *IntegrationTestSuite) TestCreateDivision() {
	wantDivision := storage.Division{
		Id:       "738f74df-72a3-4389-a4de-c4f7ad75f101",
		TenantId: s.defaultTenant.Id,
		Name:     "Marketing",
	}

	err := s.postgres.CreateDivision(wantDivision)
	s.Equal(nil, err, "Should not return an error")

	s.expectSelectQueryToReturnOneRow(
		"division",
		map[string]any{
			"id":        wantDivision.Id,
			"tenant_id": wantDivision.TenantId,
			"name":      wantDivision.Name,
		},
	)
}

func (s *IntegrationTestSuite) TestCreateDivisionViolatesUniqueConstraint() {
	tests := []struct {
		name  string
		input storage.Division
	}{
		{
			"Should violate unique constraint because id is the same",
			storage.Division{
				Id:       s.defaultDivision.Id,
				TenantId: s.defaultTenant.Id,
				Name:     "Different Name",
			},
		},
		{
			"Should violate unique constraint because tenantId-name combination is the same",
			storage.Division{
				Id:       "738f74df-72a3-4389-a4de-c4f7ad75f101",
				TenantId: s.defaultTenant.Id,
				Name:     s.defaultDivision.Name,
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			err := s.postgres.CreateDivision(test.input)
			s.expectErrorCode(err, "UNIQUE-VIOLATION-ERROR")

			s.expectSelectQueryToReturnNoRows(
				"division",
				map[string]any{
					"id":        test.input.Id,
					"tenant_id": test.input.TenantId,
					"name":      test.input.Name,
				},
			)
		})
	}
}

func (s *IntegrationTestSuite) TestCreateDivisionDoesNotViolateUniqueConstraint() {
	// Seed a tenant and division
	_, err := s.dbRootConn.Exec("INSERT INTO tenant (id, name) VALUES ($1, $2)", "5338d729-32bd-4ad2-a8d1-22cbf81113de", "Macdonalds")
	if err != nil {
		log.Fatalf("Could not seed tenant: %s", err)
	}

	tests := []struct {
		name  string
		input storage.Division
	}{
		{
			"Should not violate unique constraint because tenantId is different",
			storage.Division{
				Id:       "2e3f733c-926a-4754-981d-774832725bc7",
				TenantId: "5338d729-32bd-4ad2-a8d1-22cbf81113de",
				Name:     s.defaultDivision.Name,
			},
		},
		{
			"Should not violate unique constraint because name is different",
			storage.Division{
				Id:       "2e3f733c-926a-4754-981d-774832725bc7",
				TenantId: s.defaultTenant.Id,
				Name:     "Division2",
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			err := s.postgres.CreateDivision(test.input)
			s.Equal(nil, err)

			s.expectSelectQueryToReturnOneRow(
				"division",
				map[string]any{
					"id":        test.input.Id,
					"tenant_id": test.input.TenantId,
					"name":      test.input.Name,
				},
			)

			query := "DELETE FROM division WHERE id = $1"
			s.dbRootConn.Exec(query, test.input.Id)
		})
	}
}

func (s *IntegrationTestSuite) TestCreateDivisionViolatesForeignKeyConstraint() {
	wantDivision := storage.Division{
		Id:       "3d04353f-bbb8-4b98-99e8-1181771316c7",
		TenantId: "116e5c82-6782-418d-8f89-58d893e433e2",
		Name:     "Marketing",
	}

	err := s.postgres.CreateDivision(wantDivision)
	s.expectErrorCode(err, "INVALID-FOREIGN-KEY-ERROR")

	s.expectSelectQueryToReturnNoRows(
		"division",
		map[string]any{
			"id":        wantDivision.Id,
			"tenant_id": wantDivision.TenantId,
			"name":      wantDivision.Name,
		},
	)
}

func (s *IntegrationTestSuite) TestCreateDepartment() {
	wantDepartment := storage.Department{
		Id:         "3d3ef27c-9dc9-4e83-b39c-a42aa003dd2e",
		TenantId:   s.defaultTenant.Id,
		DivisionId: s.defaultDivision.Id,
		Name:       "Outreach",
	}

	err := s.postgres.CreateDepartment(wantDepartment)
	s.Equal(nil, err, "Should not return an error")

	s.expectSelectQueryToReturnOneRow(
		"department",
		map[string]any{
			"id":          wantDepartment.Id,
			"tenant_id":   wantDepartment.TenantId,
			"division_id": wantDepartment.DivisionId,
			"name":        wantDepartment.Name,
		},
	)
}

func (s *IntegrationTestSuite) TestCreateDepartmentViolatesUniqueConstraint() {
	tests := []struct {
		name  string
		input storage.Department
	}{
		{
			"Should violate unique constraint because id is the same",
			storage.Department{
				Id:         s.defaultDepartment.Id,
				TenantId:   s.defaultTenant.Id,
				DivisionId: s.defaultDivision.Id,
				Name:       "Different Name",
			},
		},
		{
			"Should violate unique constraint because divisionId-name combination is the same",
			storage.Department{
				Id:         "738f74df-72a3-4389-a4de-c4f7ad75f101",
				TenantId:   s.defaultTenant.Id,
				DivisionId: s.defaultDivision.Id,
				Name:       s.defaultDepartment.Name,
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			err := s.postgres.CreateDepartment(test.input)
			s.expectErrorCode(err, "UNIQUE-VIOLATION-ERROR")

			s.expectSelectQueryToReturnNoRows(
				"department",
				map[string]any{
					"id":          test.input.Id,
					"tenant_id":   test.input.TenantId,
					"division_id": test.input.DivisionId,
					"name":        test.input.Name,
				},
			)
		})
	}
}

func (s *IntegrationTestSuite) TestCreateDepartmentDoesNotViolateUniqueConstraint() {
	// Seed another division
	_, err := s.dbRootConn.Exec("INSERT INTO division (id, tenant_id, name) VALUES ($1, $2, $3)",
		"edbdc4e9-7bdd-4819-a6aa-1d3a4e208620", s.defaultTenant.Id, "Marketing",
	)
	if err != nil {
		log.Fatalf("Could not seed tenant: %s", err)
	}

	tests := []struct {
		name  string
		input storage.Department
	}{
		{
			"Should not violate unique constraint because divisionId is different",
			storage.Department{
				Id:         "738f74df-72a3-4389-a4de-c4f7ad75f101",
				TenantId:   s.defaultTenant.Id,
				DivisionId: "edbdc4e9-7bdd-4819-a6aa-1d3a4e208620",
				Name:       s.defaultDepartment.Name,
			},
		},
		{
			"Should not violate unique constraint because name is different",
			storage.Department{
				Id:         "738f74df-72a3-4389-a4de-c4f7ad75f101",
				TenantId:   s.defaultTenant.Id,
				DivisionId: s.defaultDivision.Id,
				Name:       "NewDepartment",
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			err := s.postgres.CreateDepartment(test.input)
			s.Equal(nil, err)

			s.expectSelectQueryToReturnOneRow(
				"department",
				map[string]any{
					"id":          test.input.Id,
					"tenant_id":   test.input.TenantId,
					"division_id": test.input.DivisionId,
					"name":        test.input.Name,
				},
			)

			query := "DELETE FROM department WHERE id = $1"
			s.dbRootConn.Exec(query, test.input.Id)
		})
	}
}

func (s *IntegrationTestSuite) TestCreateDepartmentViolatesForeignKeyConstraint() {
	tests := []struct {
		name  string
		input storage.Department
	}{
		{
			"Should violate foreign key constraint as division doesn't exist",
			storage.Department{
				Id:         "bf3d0982-6222-43d6-8c31-9965d6c8cf32",
				TenantId:   s.defaultTenant.Id,
				DivisionId: "1c284191-fc5a-457f-9937-4066c660ca66",
				Name:       "Outreach",
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			err := s.postgres.CreateDepartment(test.input)
			s.expectErrorCode(err, "INVALID-FOREIGN-KEY-ERROR")

			s.expectSelectQueryToReturnNoRows(
				"department",
				map[string]any{
					"id":          test.input.Id,
					"tenant_id":   test.input.TenantId,
					"division_id": test.input.DivisionId,
					"name":        test.input.Name,
				},
			)
		})
	}
}
