package postgres

import (
	"fmt"
	"log"

	"multi-tenant-HR-information-system-backend/routes"
)

type userTestCase struct {
	name  string
	input routes.User
}

func (s *IntegrationTestSuite) TestCreateUser() {
	wantUser := routes.User{
		Id:       "d6cfdb0d-b9ae-4f17-ba0a-b141391c42af",
		TenantId: s.defaultUser.TenantId,
		Email:    "test@gmail.com",
	}

	err := s.postgres.CreateUser(wantUser)
	s.Equal(nil, err)

	s.expectSelectQueryToReturnOneRow(
		"user_account",
		map[string]string{
			"id":        wantUser.Id,
			"tenant_id": wantUser.TenantId,
			"email":     wantUser.Email,
		},
	)
}

func (s *IntegrationTestSuite) TestCreateUserViolatesUniqueConstraint() {
	tests := []userTestCase{
		{
			"Should violate unique constraint because ID already exists",
			routes.User{
				Id:       s.defaultUser.Id,
				Email:    "test@gmail.com",
				TenantId: s.defaultUser.TenantId,
			},
		},
		{
			"Should violate unique constraint because Email-TenantId combination already exists",
			routes.User{
				Id:       "d6cfdb0d-b9ae-4f17-ba0a-b141391c42af",
				Email:    s.defaultUser.Email,
				TenantId: s.defaultUser.TenantId,
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			err := s.postgres.CreateUser(test.input)
			s.Equal(true, err != nil, "Error should be returned")

			httpErr, ok := err.(*routes.HttpError)
			s.Equal(true, ok, "Error should be httpError")

			if ok {
				s.Equal("UNIQUE-VIOLATION-ERROR", httpErr.Code)
			}

			s.expectSelectQueryToReturnNoRows(
				"user_account",
				map[string]string{
					"id":        test.input.Id,
					"tenant_id": test.input.TenantId,
					"email":     test.input.Email,
				},
			)
		})
	}
}

func (s *IntegrationTestSuite) TestCreateUserDoesNotViolateUniqueConstraint() {
	// Seed another tenant called "Macdonalds"
	_, err := s.dbRootConn.Exec("INSERT INTO tenant (id, name) VALUES ($1, $2)", "a9f998c6-ba2e-4359-b308-e56404534974", "Macdonalds")
	if err != nil {
		log.Fatalf("Could not seed Macdonalds: %s", err)
	}

	tests := []userTestCase{
		{
			"Should not violate unique constraint because tenant is different",
			routes.User{
				Id:       "d6cfdb0d-b9ae-4f17-ba0a-b141391c42af",
				TenantId: "a9f998c6-ba2e-4359-b308-e56404534974",
				Email:    s.defaultUser.Email,
			},
		},
		{
			"Should not violate unique constraint because email is different",
			routes.User{
				Id:       "d6cfdb0d-b9ae-4f17-ba0a-b141391c42af",
				TenantId: s.defaultUser.TenantId,
				Email:    "test@gmail.com",
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			err := s.postgres.CreateUser(test.input)
			s.Equal(nil, err)

			s.expectSelectQueryToReturnOneRow(
				"user_account",
				map[string]string{
					"id":        test.input.Id,
					"tenant_id": test.input.TenantId,
					"email":     test.input.Email,
				},
			)

			query := "DELETE FROM user_account WHERE id = $1"
			s.dbRootConn.Exec(query, test.input.Id)
		})
	}
}

func (s *IntegrationTestSuite) TestCreateUserViolatesForeignKeyConstraint() {
	tests := []userTestCase{
		{
			"Should violate foreign key constraint because tenant does not exist",
			routes.User{
				Id:       "d6cfdb0d-b9ae-4f17-ba0a-b141391c42af",
				Email:    s.defaultUser.Email,
				TenantId: "a9f998c6-ba2e-4359-b308-e56404534974",
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			err := s.postgres.CreateUser(test.input)
			s.Equal(true, err != nil, "Error should be returned")

			httpError, ok := err.(*routes.HttpError)
			s.Equal(true, ok, "Error should be httpError")

			if ok {
				s.Equal("INVALID-FOREIGN-KEY-ERROR", httpError.Code)
			}

			s.expectSelectQueryToReturnNoRows(
				"user_account",
				map[string]string{
					"id":        test.input.Id,
					"tenant_id": test.input.TenantId,
					"email":     test.input.Email,
				},
			)
		})
	}
}

func (s *IntegrationTestSuite) TestGetUsers() {
	// Seed new users
	_, err := s.dbRootConn.Exec("INSERT INTO tenant (id, name) VALUES ($1, $2)", "a9f998c6-ba2e-4359-b308-e56404534974", "Macdonalds")
	if err != nil {
		log.Fatalf("Could not seed tenant: %s", err)
	}

	newUsers := []routes.User{
		{
			Id:       "1a288b1f-3c53-44e3-9ef9-c902af41cd7e",
			TenantId: s.defaultUser.TenantId,
			Email:    "test1@gmail.com",
		},
		{
			Id:       "20717410-4530-4f83-9cfa-5fd1b32c77a4",
			TenantId: "a9f998c6-ba2e-4359-b308-e56404534974",
			Email:    "test1@gmail.com",
		},
		{
			Id:       "c75d1e25-0e6a-478c-8332-1c48670ba74c",
			TenantId: "a9f998c6-ba2e-4359-b308-e56404534974",
			Email:    "test2@gmail.com",
		},
	}

	for _, user := range newUsers {
		query := `INSERT INTO user_account (id, tenant_id, email, password, totp_secret_key) VALUES ($1, $2, $3, $4, $5)`
		_, err = s.dbRootConn.Exec(query, user.Id, user.TenantId, user.Email, "", "")
		if err != nil {
			log.Fatalf("Could not seed user: %s", err)
		}
	}

	tests := []struct {
		name  string
		input routes.User
		want  int
	}{
		{
			"Should return users by tenant",
			routes.User{
				TenantId: "a9f998c6-ba2e-4359-b308-e56404534974",
			},
			2,
		},
		{
			"Should return users by tenant & email",
			routes.User{
				TenantId: "a9f998c6-ba2e-4359-b308-e56404534974",
				Email:    "test1@gmail.com",
			},
			1,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			users, err := s.postgres.GetUsers(test.input)
			s.Equal(nil, err)

			s.Equal(test.want, len(users), fmt.Sprintf("Should have returned %v row(s)", test.want))
		})
	}
}

func (s *IntegrationTestSuite) TestGetUsersNoTenantId() {
	filter := routes.User{
		Email: s.defaultUser.Email,
	}

	users, err := s.postgres.GetUsers(filter)
	s.expectErrorCode(err, "INTERNAL-SERVER-ERROR")
	s.Equal(0, len(users), "Users should be nil")
}

func (s *IntegrationTestSuite) TestCreateAppointment() {
	wantAppointment := routes.Appointment{
		Id:           "a9f998c6-ba2e-4359-b308-e56404534974",
		TenantId:     s.defaultAppointment.TenantId,
		Title:        "Manager",
		DepartmentId: s.defaultAppointment.DepartmentId,
		UserId:       s.defaultAppointment.UserId,
		StartDate:    s.defaultAppointment.StartDate,
	}

	err := s.postgres.CreateAppointment(wantAppointment)
	s.Equal(nil, err)

	s.expectSelectQueryToReturnOneRow(
		"appointment",
		map[string]string{
			"id":              wantAppointment.Id,
			"tenant_id":       wantAppointment.TenantId,
			"title":           wantAppointment.Title,
			"department_id":   wantAppointment.DepartmentId,
			"user_account_id": wantAppointment.UserId,
			"start_date":      wantAppointment.StartDate,
		},
	)
}

func (s *IntegrationTestSuite) TestCreateAppointmentViolatesUniqueConstraint() {
	err := s.postgres.CreateAppointment(s.defaultAppointment)
	s.expectErrorCode(err, "UNIQUE-VIOLATION-ERROR")

	s.expectSelectQueryToReturnOneRow(
		"appointment",
		map[string]string{
			"id":        s.defaultAppointment.Id,			
			"tenant_id": s.defaultAppointment.TenantId,
			"title": s.defaultAppointment.Title,
			"department_id": s.defaultAppointment.DepartmentId,
			"user_account_id": s.defaultAppointment.UserId,
			"start_date": s.defaultAppointment.StartDate,
		},
	)
}

func (s *IntegrationTestSuite) TestCreateAppointmentDoesNotViolateUniqueConstraint() {
	// Seed another user and department
	query := "INSERT INTO user_account (id, tenant_id, email, password, totp_secret_key) VALUES ($1, $2, $3, $4, $5)"
	_, err := s.dbRootConn.Exec(query, "bc1a0220-fdef-403c-ae0b-5664daefb328", s.defaultUser.TenantId, "test1@gmail.com", "", "")
	if err != nil {
		log.Fatalf("Could not seed user: %s", err)
	}

	query = "INSERT INTO department (id, tenant_id, division_id, name) VALUES ($1, $2, $3, $4)"
	_, err = s.dbRootConn.Exec(query, "583cac89-c402-4655-850f-1635c78d9970", s.defaultUser.TenantId, s.defaultDivision.Id, "Customer Support")
	if err != nil {
		log.Fatalf("Could not seed department: %s", err)
	}

	tests := []struct {
		name  string
		input routes.Appointment
	}{
		{
			"Should not violate unique constraint as the title is different",
			routes.Appointment{
				Id:           "a084e475-2018-4935-81cd-5514c03770db",
				TenantId:     s.defaultAppointment.TenantId,
				Title:        "Manager",
				DepartmentId: s.defaultAppointment.DepartmentId,
				UserId:       s.defaultAppointment.UserId,
				StartDate:    "2024-02-01",
			},
		},
		{
			"Should not violate unique constraint as the user id is different",
			routes.Appointment{
				Id:           "a084e475-2018-4935-81cd-5514c03770db",
				TenantId:     s.defaultAppointment.TenantId,
				Title:        s.defaultAppointment.Title,
				DepartmentId: s.defaultAppointment.DepartmentId,
				UserId:       "bc1a0220-fdef-403c-ae0b-5664daefb328",
				StartDate:    "2024-02-01",
			},
		},
		{
			"Should not violate unique constraint as the department is different",
			routes.Appointment{
				Id:           "a084e475-2018-4935-81cd-5514c03770db",
				TenantId:     s.defaultAppointment.TenantId,
				Title:        s.defaultAppointment.Title,
				DepartmentId: "583cac89-c402-4655-850f-1635c78d9970",
				UserId:       s.defaultUser.Id,
				StartDate:    "2024-02-01",
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			err := s.postgres.CreateAppointment(test.input)
			s.Equal(nil, err)

			s.expectSelectQueryToReturnOneRow(
				"appointment",
				map[string]string{
					"id":              test.input.Id,
					"tenant_id":       test.input.TenantId,
					"title":           test.input.Title,
					"department_id":   test.input.DepartmentId,
					"user_account_id": test.input.UserId,
					"start_date":      test.input.StartDate,
				},
			)

			query := "DELETE FROM appointment WHERE id = $1"
			s.dbRootConn.Exec(query, test.input.Id)
		})
	}
}

func (s *IntegrationTestSuite) TestCreateAppointmentViolatesForeignKeyConstraint() {
	tests := []struct {
		name  string
		input routes.Appointment
	}{
		{
			"Should violate foreign key constraint because department does not exist",
			routes.Appointment{
				Id:           "a084e475-2018-4935-81cd-5514c03770db",
				TenantId:     s.defaultAppointment.TenantId,
				Title:        s.defaultAppointment.Title,
				DepartmentId: "ef6aaa95-921c-4931-bfd3-7635f6be6507",
				UserId:       s.defaultUser.Id,
				StartDate:    "2024-02-01",
			},
		},
		{
			"Should violate foreign key constraint because user id does not exist",
			routes.Appointment{
				Id:           "a084e475-2018-4935-81cd-5514c03770db",
				TenantId:     s.defaultAppointment.TenantId,
				Title:        s.defaultAppointment.Title,
				DepartmentId: s.defaultDivision.Id,
				UserId:       "bc1a0220-fdef-433c-ae0b-5664daefb328",
				StartDate:    "2024-02-01",
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			err := s.postgres.CreateAppointment(test.input)
			s.expectErrorCode(err, "INVALID-FOREIGN-KEY-ERROR")

			s.expectSelectQueryToReturnNoRows(
				"appointment",
				map[string]string{
					"id":              test.input.Id,
					"tenant_id":       test.input.TenantId,
					"title":           test.input.Title,
					"department_id":   test.input.DepartmentId,
					"user_account_id": test.input.UserId,
					"start_date":      test.input.StartDate,
				},
			)
		})
	}
}
