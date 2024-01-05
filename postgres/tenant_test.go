package postgres

import (
	"log"
	"multi-tenant-HR-information-system-backend/routes"
)

func (s *PostgresTestSuite) TestCreateTenant() {
	wantTenant := routes.Tenant{
		Name: "Macdonalds",
	}

	err := s.postgres.CreateTenant(wantTenant)
	s.Equal(nil, err, "Should not return an error")

	s.expectSelectQueryToReturnOneRow(
		"tenant",
		[]string{"name"},
		[]string{wantTenant.Name},
	)
}

func (s *PostgresTestSuite) TestCreateTenantAlreadyExists() {
	wantTenant := routes.Tenant{
		Name: s.defaultUser.Tenant,
	}

	err := s.postgres.CreateTenant(wantTenant)
	s.Equal(false, err == nil, "Error should not be nil")

	httpErr, ok := err.(*routes.HttpError)
	s.Equal(true, ok, "Error should be HttpError")

	if ok {
		s.Equal("UNIQUE-VIOLATION-ERROR", httpErr.Code)
	}

	s.expectSelectQueryToReturnOneRow(
		"tenant",
		[]string{"name"},
		[]string{wantTenant.Name},
	)
}

func (s *PostgresTestSuite) TestCreateDivision() {
	wantDivision := routes.Division{
		Name:   "Marketing",
		Tenant: s.defaultUser.Tenant,
	}

	err := s.postgres.CreateDivision(wantDivision)
	s.Equal(nil, err, "Should not return an error")

	s.expectSelectQueryToReturnOneRow(
		"division",
		[]string{"name", "tenant"},
		[]string{wantDivision.Name, wantDivision.Tenant},
	)
}

func (s *PostgresTestSuite) TestCreateDivisionAlreadyExists() {
	wantDivision := routes.Division{
		Name:   "Marketing",
		Tenant: s.defaultUser.Tenant,
	}

	query := "INSERT INTO division (name, tenant) VALUES ($1, $2)"
	_, err := s.dbRootConn.Exec(query, wantDivision.Name, wantDivision.Tenant)
	if err != nil {
		log.Fatal(err)
	}

	err = s.postgres.CreateDivision(wantDivision)
	s.Equal(false, err == nil, "Error should not be nil")

	httpErr, ok := err.(*routes.HttpError)
	s.Equal(true, ok, "Error should be HttpError")

	if ok {
		s.Equal("UNIQUE-VIOLATION-ERROR", httpErr.Code)
	}

	s.expectSelectQueryToReturnOneRow(
		"division",
		[]string{"name", "tenant"},
		[]string{wantDivision.Name, wantDivision.Tenant},
	)
}

func (s *PostgresTestSuite) TestCreateDivisionInvalidTenant() {
	wantDivision := routes.Division{
		Name:   "Marketing",
		Tenant: "non-existent",
	}

	err := s.postgres.CreateDivision(wantDivision)
	s.Equal(false, err == nil, "Error should not be nil")

	httpErr, ok := err.(*routes.HttpError)
	s.Equal(true, ok, "Error should be HttpError")

	if ok {
		s.Equal("INVALID-FOREIGN-KEY-ERROR", httpErr.Code)
	}

	s.expectSelectQueryToReturnNoRows(
		"division",
		[]string{"name", "tenant"},
		[]string{wantDivision.Name, wantDivision.Tenant},
	)
}

func (s *PostgresTestSuite) TestCreateDepartment() {
	wantDepartment := routes.Department{
		Name:     "Outreach",
		Tenant:   s.defaultUser.Tenant,
		Division: "Marketing",
	}

	query := "INSERT INTO division (name, tenant) VALUES ($1, $2)"
	_, err := s.dbRootConn.Exec(query, wantDepartment.Division, wantDepartment.Tenant)
	if err != nil {
		log.Fatal(err)
	}

	err = s.postgres.CreateDepartment(wantDepartment)
	s.Equal(nil, err, "Should not return an error")

	s.expectSelectQueryToReturnOneRow(
		"department",
		[]string{"name", "tenant", "division"},
		[]string{wantDepartment.Name, wantDepartment.Tenant, wantDepartment.Division},
	)
}

func (s *PostgresTestSuite) TestCreateDepartmentAlreadyExists() {
	wantDepartment := routes.Department{
		Name:     "Outreach",
		Tenant:   s.defaultUser.Tenant,
		Division: "Marketing",
	}

	query := "INSERT INTO division (name, tenant) VALUES ($1, $2)"
	_, err := s.dbRootConn.Exec(query, wantDepartment.Division, wantDepartment.Tenant)
	if err != nil {
		log.Fatal(err)
	}
	query = "INSERT INTO department (name, tenant, division) VALUES ($1, $2, $3)"
	_, err = s.dbRootConn.Exec(query, wantDepartment.Name, wantDepartment.Tenant, wantDepartment.Division)
	if err != nil {
		log.Fatal(err)
	}

	err = s.postgres.CreateDepartment(wantDepartment)
	s.Equal(false, err == nil, "Error should not be nil")

	httpErr, ok := err.(*routes.HttpError)
	s.Equal(true, ok, "Error should be HttpError")

	if ok {
		s.Equal("UNIQUE-VIOLATION-ERROR", httpErr.Code)
	}

	s.expectSelectQueryToReturnOneRow(
		"department",
		[]string{"name", "tenant", "division"},
		[]string{wantDepartment.Name, wantDepartment.Tenant, wantDepartment.Division},
	)
}

func (s *PostgresTestSuite) TestCreateDepartmentInvalidDivision() {
	wantDepartment := routes.Department{
		Name:     "Outreach",
		Tenant:   s.defaultUser.Tenant,
		Division: "Marketing",
	}

	err := s.postgres.CreateDepartment(wantDepartment)
	s.Equal(false, err == nil, "Error should not be nil")

	httpErr, ok := err.(*routes.HttpError)
	s.Equal(true, ok, "Error should be HttpError")

	if ok {
		s.Equal("INVALID-FOREIGN-KEY-ERROR", httpErr.Code)
	}

	s.expectSelectQueryToReturnNoRows(
		"department",
		[]string{"name", "tenant", "division"},
		[]string{wantDepartment.Name, wantDepartment.Tenant, wantDepartment.Division},
	)
}
