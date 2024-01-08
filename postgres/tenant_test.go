package postgres

import (
	"log"
	"multi-tenant-HR-information-system-backend/routes"
)

func (s *IntegrationTestSuite) TestCreateTenant() {
	wantTenant := routes.Tenant{
		Name: "Macdonalds",
	}

	err := s.postgres.CreateTenant(wantTenant)
	s.Equal(nil, err, "Should not return an error")

	s.expectSelectQueryToReturnOneRow(
		"tenant",
		map[string]string{
			"name": wantTenant.Name,
		},
	)
}

func (s *IntegrationTestSuite) TestCreateTenantViolatesUniqueConstraint() {
	wantTenant := routes.Tenant{
		Name: s.defaultUser.Tenant,
	}

	err := s.postgres.CreateTenant(wantTenant)
	s.expectErrorCode(err, "UNIQUE-VIOLATION-ERROR")

	s.expectSelectQueryToReturnOneRow(
		"tenant",
		map[string]string{
			"name": wantTenant.Name,
		},
	)
}

func (s *IntegrationTestSuite) TestCreateDivision() {
	wantDivision := routes.Division{
		Tenant: s.defaultUser.Tenant,
		Name:   "Marketing",		
	}

	err := s.postgres.CreateDivision(wantDivision)
	s.Equal(nil, err, "Should not return an error")

	s.expectSelectQueryToReturnOneRow(
		"division",
		map[string]string{
			"tenant": wantDivision.Tenant,
			"name":   wantDivision.Name,
		},
	)
}

func (s *IntegrationTestSuite) TestCreateDivisionViolatesUniqueConstraint() {
	wantDivision := routes.Division{
		Tenant: s.defaultUser.Tenant,
		Name:   s.defaultAppointment.Division,		
	}

	err := s.postgres.CreateDivision(wantDivision)
	s.expectErrorCode(err, "UNIQUE-VIOLATION-ERROR")

	s.expectSelectQueryToReturnOneRow(
		"division",
		map[string]string{
			"tenant": wantDivision.Tenant,
			"name":   wantDivision.Name,
		},
	)
}

func (s *IntegrationTestSuite) TestCreateDivisionDoesNotViolateUniqueConstraint() {
	// Seed a tenant and division
	_, err := s.dbRootConn.Exec("INSERT INTO tenant (name) VALUES ($1)", "Macdonalds")
	if err != nil {
		log.Fatalf("Could not seed tenant: %s", err)
	}

	tests := []struct {
		name  string
		input routes.Division
	}{
		{
			"Should not violate unique constraint because tenant is different",
			routes.Division{
				Tenant: "Macdonalds",
				Name:   s.defaultAppointment.Division,
			},
		},
		{
			"Should not violate unique constraint because name is different",
			routes.Division{
				Tenant: "HRIS Enterprises",
				Name:   "Division2",
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			err := s.postgres.CreateDivision(test.input)
			s.Equal(nil, err)

			s.expectSelectQueryToReturnOneRow(
				"division",
				map[string]string{
					"tenant": test.input.Tenant,
					"name":   test.input.Name,
				},
			)

			query := "DELETE FROM division WHERE tenant = $1 AND name = $2"
			s.dbRootConn.Exec(query, test.input.Tenant, test.input.Name)
		})
	}
}

func (s *IntegrationTestSuite) TestCreateDivisionViolatesForeignKeyConstraint() {
	wantDivision := routes.Division{
		Name:   "Marketing",
		Tenant: "non-existent",
	}

	err := s.postgres.CreateDivision(wantDivision)
	s.expectErrorCode(err, "INVALID-FOREIGN-KEY-ERROR")

	s.expectSelectQueryToReturnNoRows(
		"division",
		map[string]string{
			"tenant": wantDivision.Tenant,
			"name":   wantDivision.Name,
		},
	)
}

func (s *IntegrationTestSuite) TestCreateDepartment() {
	wantDepartment := routes.Department{
		Tenant:   s.defaultUser.Tenant,
		Division: s.defaultAppointment.Division,
		Name:     "Outreach",
	}

	err := s.postgres.CreateDepartment(wantDepartment)
	s.Equal(nil, err, "Should not return an error")

	s.expectSelectQueryToReturnOneRow(
		"department",
		map[string]string{
			"tenant":   wantDepartment.Tenant,
			"division": wantDepartment.Division,
			"name":     wantDepartment.Name,
		},
	)
}

func (s *IntegrationTestSuite) TestCreateDepartmentViolatesUniqueConstraint() {
	wantDepartment := routes.Department{
		Tenant:   s.defaultUser.Tenant,
		Division: s.defaultAppointment.Division,
		Name:     s.defaultAppointment.Department,
	}

	err := s.postgres.CreateDepartment(wantDepartment)
	s.expectErrorCode(err, "UNIQUE-VIOLATION-ERROR")

	s.expectSelectQueryToReturnOneRow(
		"department",
		map[string]string{
			"tenant":   wantDepartment.Tenant,
			"division": wantDepartment.Division,
			"name":     wantDepartment.Name,
		},
	)
}

func (s *IntegrationTestSuite) TestCreateDepartmentDoesNotViolateUniqueConstraint() {
	// Seed a tenant and a division
	_, err := s.dbRootConn.Exec("INSERT INTO tenant (name) VALUES ($1)", "Macdonalds")
	if err != nil {
		log.Fatalf("Could not seed tenant: %s", err)
	}

	_, err = s.dbRootConn.Exec("INSERT INTO division (tenant, name) VALUES ($1, $2), ($3, $4)", 
								"Macdonalds", s.defaultAppointment.Division,
								"HRIS Enterprises", "Marketing",								
							)
	if err != nil {
		log.Fatalf("Could not seed tenant: %s", err)
	}

	tests := []struct {
		name  string
		input routes.Department
	}{
		{
			"Should not violate unique constraint because tenant is different",
			routes.Department{
				Tenant:   "Macdonalds",
				Division: s.defaultAppointment.Division,
				Name:     s.defaultAppointment.Department,
			},
		},
		{
			"Should not violate unique constraint because division is different",
			routes.Department{
				Tenant:   s.defaultAppointment.Tenant,
				Division: "Marketing",
				Name:     s.defaultAppointment.Department,
			},
		},
		{
			"Should not violate unique constraint because name is different",
			routes.Department{
				Tenant:   s.defaultAppointment.Tenant,
				Division: s.defaultAppointment.Division,
				Name:     "Customer Support",
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			err := s.postgres.CreateDepartment(test.input)
			s.Equal(nil, err)

			s.expectSelectQueryToReturnOneRow(
				"department",
				map[string]string{
					"tenant":   test.input.Tenant,
					"division": test.input.Division,
					"name":     test.input.Name,
				},
			)

			query := "DELETE FROM department WHERE tenant = $1 AND division = $2 AND name = $3"
			s.dbRootConn.Exec(query, test.input.Tenant, test.input.Division, test.input.Name)
		})
	}
}

func (s *IntegrationTestSuite) TestCreateDepartmentViolatesForeignKeyConstraint() {
	tests := []struct {
		name  string
		input routes.Department
	}{
		{
			"Should violate foreign key constraint as division doesn't exist",
			routes.Department{
				Name:     "Outreach",
				Tenant:   s.defaultUser.Tenant,
				Division: "Marketing",
			},
		},
		{
			"Should violate foreign key constraint as tenant doesn't exist",
			routes.Department{
				Name:     "Outreach",
				Tenant:   "Does not exist",
				Division: "Anything",
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			err := s.postgres.CreateDepartment(test.input)
			s.expectErrorCode(err, "INVALID-FOREIGN-KEY-ERROR")

			s.expectSelectQueryToReturnNoRows(
				"department",
				map[string]string{
					"tenant":   test.input.Tenant,
					"division": test.input.Division,
					"name":     test.input.Name,
				},
			)
		})
	}
}
