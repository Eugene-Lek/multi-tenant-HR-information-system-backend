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
		Id:     "d6cfdb0d-b9ae-4f17-ba0a-b141391c42af",
		Email:  "test@gmail.com",
		Tenant: s.defaultUser.Tenant,
	}

	err := s.postgres.CreateUser(wantUser)
	s.Equal(nil, err)

	s.expectSelectQueryToReturnOneRow(
		"user_account",
		map[string]string{"id": wantUser.Id},
	)
}

func (s *IntegrationTestSuite) TestCreateUserViolatesUniqueConstraint() {
	tests := []userTestCase{
		{
			"Should violate unique constraint because ID already exists",
			routes.User{
				Id:     s.defaultUser.Id,
				Email:  "test@gmail.com",
				Tenant: s.defaultUser.Tenant,
			},
		},
		{
			"Should violate unique constraint because Email-Tenant combination already exists",
			routes.User{
				Id:     "d6cfdb0d-b9ae-4f17-ba0a-b141391c42af",
				Email:  s.defaultUser.Email,
				Tenant: s.defaultUser.Tenant,
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
					"id":     test.input.Id,
					"tenant": test.input.Tenant,
					"email":  test.input.Email,
				},
			)
		})
	}
}

func (s *IntegrationTestSuite) TestCreateUserDoesNotViolateUniqueConstraint() {
	// Seed another tenant called "Macdonalds"
	_, err := s.dbRootConn.Exec("INSERT INTO tenant (name) VALUES ($1)", "Macdonalds")
	if err != nil {
		log.Fatalf("Could not seed Macdonalds: %s", err)
	}

	tests := []userTestCase{
		{
			"Should not violate unique constraint because tenant is different",
			routes.User{
				Id:     "d6cfdb0d-b9ae-4f17-ba0a-b141391c42af",
				Email:  s.defaultUser.Email,
				Tenant: "Macdonalds",
			},
		},
		{
			"Should not violate unique constraint because email is different",
			routes.User{
				Id:     "d6cfdb0d-b9ae-4f17-ba0a-b141391c42af",
				Email:  "test@gmail.com",
				Tenant: s.defaultUser.Tenant,
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
					"id":     test.input.Id,
					"tenant": test.input.Tenant,
					"email":  test.input.Email,
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
				Id:     "d6cfdb0d-b9ae-4f17-ba0a-b141391c42af",
				Email:  s.defaultUser.Email,
				Tenant: "Macdonalds",
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
					"id":     test.input.Id,
					"tenant": test.input.Tenant,
					"email":  test.input.Email,
				},
			)
		})
	}
}

func (s *IntegrationTestSuite) TestGetUsers() {
	// Seed new users
	_, err := s.dbRootConn.Exec("INSERT INTO tenant (name) VALUES ($1)", "Macdonalds")
	if err != nil {
		log.Fatalf("Could not seed tenant: %s", err)
	}

	newUsers := []routes.User{
		{
			Tenant: "HRIS Enterprises",
			Email:  "test1@gmail.com",
		},
		{
			Tenant: "Macdonalds",
			Email:  "test1@gmail.com",
		},
		{
			Tenant: "Macdonalds",
			Email:  "test2@gmail.com",
		},
	}

	for _, user := range newUsers {
		query := `INSERT INTO user_account (tenant, email, password, totp_secret_key) VALUES ($1, $2, $3, $4)`
		_, err = s.dbRootConn.Exec(query, user.Tenant, user.Email, "", "")
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
				Tenant: "Macdonalds",
			},
			2,
		},
		{
			"Should return users by email",
			routes.User{
				Email: "test1@gmail.com",
			},
			2,
		},
		{
			"Should return users by tenant & email",
			routes.User{
				Tenant: "Macdonalds",
				Email:  "test1@gmail.com",
			},
			1,
		},
		{
			"Should return all users",
			routes.User{},
			4,
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

func (s *IntegrationTestSuite) TestCreateAppointment() {
	wantAppointment := routes.Appointment{
		Title:      "Manager",
		Tenant:     s.defaultAppointment.Tenant,
		Division:   s.defaultAppointment.Division,
		Department: s.defaultAppointment.Department,
		UserId:     s.defaultAppointment.UserId,
		StartDate:  s.defaultAppointment.StartDate,
	}

	err := s.postgres.CreateAppointment(wantAppointment)
	s.Equal(nil, err)

	s.expectSelectQueryToReturnOneRow(
		"appointment",
		map[string]string{
			"title": "Manager",
			"tenant":          s.defaultAppointment.Tenant,
			"division":        s.defaultAppointment.Division,
			"department":      s.defaultAppointment.Department,
			"user_account_id": s.defaultAppointment.UserId,
			"start_date":      s.defaultAppointment.StartDate,
		},
	)
}

func (s *IntegrationTestSuite) TestCreateAppointmentViolatesUniqueConstraint() {
	err := s.postgres.CreateAppointment(s.defaultAppointment)
	s.expectErrorCode(err, "UNIQUE-VIOLATION-ERROR")

	s.expectSelectQueryToReturnOneRow(
		"appointment",
		map[string]string{
			"title": s.defaultAppointment.Title,
			"tenant":          s.defaultAppointment.Tenant,
			"division":        s.defaultAppointment.Division,
			"department":      s.defaultAppointment.Department,
			"user_account_id": s.defaultAppointment.UserId,
			"start_date":      s.defaultAppointment.StartDate,
		},		
	)
}

func (s *IntegrationTestSuite) TestCreateAppointmentDoesNotViolateUniqueConstraint() {
	// Seed another user and department
	query := "INSERT INTO user_account (id, tenant, email, password, totp_secret_key) VALUES ($1, $2, $3, $4, $5)"
	_, err := s.dbRootConn.Exec(query, "bc1a0220-fdef-403c-ae0b-5664daefb328", s.defaultUser.Tenant, "test1@gmail.com", "", "")
	if err != nil {
		log.Fatalf("Could not seed user: %s", err)
	}

	query = "INSERT INTO department (tenant, division, name) VALUES ($1, $2, $3)"
	_, err = s.dbRootConn.Exec(query, s.defaultAppointment.Tenant, s.defaultAppointment.Division, "Customer Support")
	if err != nil {
		log.Fatalf("Could not seed user: %s", err)
	}

	tests := []struct{
		name string
		input routes.Appointment
	} {
		{
			"Should not violate unique constraint as the title is different",
			routes.Appointment{
				Title: "Manager",
				Tenant: s.defaultAppointment.Tenant,
				Division: s.defaultAppointment.Division,
				Department: s.defaultAppointment.Department,
				UserId:  s.defaultAppointment.UserId,
				StartDate: "2024-02-01",
			},			
		},				
		{
			"Should not violate unique constraint as the user id is different",
			routes.Appointment{
				Title: s.defaultAppointment.Title,
				Tenant: s.defaultAppointment.Tenant,
				Division: s.defaultAppointment.Division,
				Department: s.defaultAppointment.Department,
				UserId: "bc1a0220-fdef-403c-ae0b-5664daefb328",
				StartDate: "2024-02-01",				
			},
		},
		{
			"Should not violate unique constraint as the department is different",
			routes.Appointment{
				Title: s.defaultAppointment.Title,
				Tenant: s.defaultAppointment.Tenant,
				Division: s.defaultAppointment.Division,
				Department: "Customer Support",
				UserId: s.defaultUser.Id,
				StartDate: "2024-02-01",				
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
					"title":           test.input.Title,
					"tenant":          test.input.Tenant,
					"division":        test.input.Division,
					"department":      test.input.Department,
					"user_account_id": test.input.UserId,
					"start_date":      test.input.StartDate,
				},					
			)

			query := "DELETE FROM appointment WHERE title = $1 AND tenant = $2 AND division = $3 AND department = $4 AND user_account_id = $5"
			s.dbRootConn.Exec(query, test.input.Tenant, test.input.Division, test.input.Department, test.input.UserId, test.input.StartDate)
		})
	}
}

func (s *IntegrationTestSuite) TestCreateAppointmentViolatesForeignKeyConstraint() {
	tests := []struct{
		name string
		input routes.Appointment
	} {
		{
			"Should violate foreign key constraint because department does not exist",
			routes.Appointment{
				Title: s.defaultAppointment.Title,
				Tenant: "Does",
				Division: "Not",
				Department: "Exist",
				UserId: s.defaultUser.Id,
				StartDate: "2024-02-01",				
			},
		},
		{
			"Should violate foreign key constraint because user id does not exist",
			routes.Appointment{
				Title: s.defaultAppointment.Title,
				Tenant: s.defaultAppointment.Tenant,
				Division: s.defaultAppointment.Division,
				Department: s.defaultAppointment.Division,
				UserId: "bc1a0220-fdef-433c-ae0b-5664daefb328",
				StartDate: "2024-02-01",				
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
					"title":           test.input.Title,
					"tenant":          test.input.Tenant,
					"division":        test.input.Division,
					"department":      test.input.Department,
					"user_account_id": test.input.UserId,
					"start_date":      test.input.StartDate,
				},						
			)
		})
	}
}