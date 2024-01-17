package postgres

import (
	"fmt"
	"log"

	"multi-tenant-HR-information-system-backend/storage"
)

type userTestCase struct {
	name  string
	input storage.User
}

func (s *IntegrationTestSuite) TestCreateUser() {
	wantUser := storage.User{
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
			storage.User{
				Id:       s.defaultUser.Id,
				Email:    "test@gmail.com",
				TenantId: s.defaultUser.TenantId,
			},
		},
		{
			"Should violate unique constraint because Email-TenantId combination already exists",
			storage.User{
				Id:       "d6cfdb0d-b9ae-4f17-ba0a-b141391c42af",
				Email:    s.defaultUser.Email,
				TenantId: s.defaultUser.TenantId,
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			err := s.postgres.CreateUser(test.input)
			s.expectErrorCode(err, "UNIQUE-VIOLATION-ERROR")

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
			storage.User{
				Id:       "d6cfdb0d-b9ae-4f17-ba0a-b141391c42af",
				TenantId: "a9f998c6-ba2e-4359-b308-e56404534974",
				Email:    s.defaultUser.Email,
			},
		},
		{
			"Should not violate unique constraint because email is different",
			storage.User{
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
			storage.User{
				Id:       "d6cfdb0d-b9ae-4f17-ba0a-b141391c42af",
				Email:    s.defaultUser.Email,
				TenantId: "a9f998c6-ba2e-4359-b308-e56404534974",
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			err := s.postgres.CreateUser(test.input)
			s.expectErrorCode(err, "INVALID-FOREIGN-KEY-ERROR")

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

	newUsers := []storage.User{
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
		input storage.User
		want  int
	}{
		{
			"Should return users by tenant",
			storage.User{
				TenantId: "a9f998c6-ba2e-4359-b308-e56404534974",
			},
			2,
		},
		{
			"Should return users by tenant & email",
			storage.User{
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
	filter := storage.User{
		Email: s.defaultUser.Email,
	}

	users, err := s.postgres.GetUsers(filter)
	s.expectErrorCode(err, "INTERNAL-SERVER-ERROR")
	s.Equal(0, len(users), "Users should be nil")
}

func (s *IntegrationTestSuite) TestCreatePosition() {
	wantPosition := storage.Position{
		Id:            "a9f998c6-ba2e-4359-b308-e56404534974",
		TenantId:      s.defaultPosition.TenantId,
		Title:         "Manager",
		DepartmentId:  s.defaultPosition.DepartmentId,
		SupervisorIds: []string{},
	}

	err := s.postgres.CreatePosition(wantPosition)
	s.Equal(nil, err)

	s.expectSelectQueryToReturnOneRow(
		"position",
		map[string]string{
			"id":            wantPosition.Id,
			"tenant_id":     wantPosition.TenantId,
			"title":         wantPosition.Title,
			"department_id": wantPosition.DepartmentId,
		},
	)
}

func (s *IntegrationTestSuite) TestCreatePositionViolatesUniqueConstraint() {
	tests := []struct {
		name  string
		input storage.Position
	}{
		{
			"Should violate the unique constraint because id already exists",
			storage.Position{
				Id:            s.defaultPosition.Id,
				TenantId:      s.defaultPosition.TenantId,
				Title:         "New",
				DepartmentId:  s.defaultPosition.DepartmentId,
				SupervisorIds: []string{},
			},
		},
		{
			"Should violate the unique constraint because title & department already exists",
			storage.Position{
				Id:            "a9f998c6-ba2e-4359-b308-e56404534974",
				TenantId:      s.defaultPosition.TenantId,
				Title:         s.defaultPosition.Title,
				DepartmentId:  s.defaultPosition.DepartmentId,
				SupervisorIds: []string{},
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			err := s.postgres.CreatePosition(test.input)
			s.expectErrorCode(err, "UNIQUE-VIOLATION-ERROR")

			s.expectSelectQueryToReturnNoRows(
				"position",
				map[string]string{
					"id":            test.input.Id,
					"tenant_id":     test.input.TenantId,
					"title":         test.input.Title,
					"department_id": test.input.DepartmentId,
				},
			)
		})
	}
}

func (s *IntegrationTestSuite) TestCreatePositionDoesNotViolateUniqueConstraint() {
	// Seed another department
	query := "INSERT INTO department (id, tenant_id, division_id, name) VALUES ($1, $2, $3, $4)"
	_, err := s.dbRootConn.Exec(query, "583cac89-c402-4655-850f-1635c78d9970", s.defaultUser.TenantId, s.defaultDivision.Id, "Customer Support")
	if err != nil {
		log.Fatalf("Could not seed department: %s", err)
	}

	tests := []struct {
		name  string
		input storage.Position
	}{
		{
			"Should not violate unique constraint as the title is different",
			storage.Position{
				Id:            "a084e475-2018-4935-81cd-5514c03770db",
				TenantId:      s.defaultPosition.TenantId,
				Title:         "Manager",
				DepartmentId:  s.defaultPosition.DepartmentId,
				SupervisorIds: []string{},
			},
		},
		{
			"Should not violate unique constraint as the department is different",
			storage.Position{
				Id:            "a084e475-2018-4935-81cd-5514c03770db",
				TenantId:      s.defaultPosition.TenantId,
				Title:         s.defaultPosition.Title,
				DepartmentId:  "583cac89-c402-4655-850f-1635c78d9970",
				SupervisorIds: []string{},
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			err := s.postgres.CreatePosition(test.input)
			s.Equal(nil, err)

			s.expectSelectQueryToReturnOneRow(
				"position",
				map[string]string{
					"id":            test.input.Id,
					"tenant_id":     test.input.TenantId,
					"title":         test.input.Title,
					"department_id": test.input.DepartmentId,
				},
			)

			query := "DELETE FROM position WHERE id = $1"
			s.dbRootConn.Exec(query, test.input.Id)
		})
	}
}

func (s *IntegrationTestSuite) TestCreatePositionViolatesForeignKeyConstraint() {
	tests := []struct {
		name  string
		input storage.Position
	}{
		{
			"Should violate foreign key constraint because tenant id does not exist",
			storage.Position{
				Id:            "a084e475-2018-4935-81cd-5514c03770db",
				TenantId:      "68df1358-76bc-49ca-bea5-dc4f79afdce3",
				Title:         "Random",
				DepartmentId:  s.defaultDepartment.Id,
				SupervisorIds: []string{},
			},
		},
		{
			"Should violate foreign key constraint because department does not exist",
			storage.Position{
				Id:            "a084e475-2018-4935-81cd-5514c03770db",
				TenantId:      s.defaultPosition.TenantId,
				Title:         s.defaultPosition.Title,
				DepartmentId:  "ef6aaa95-921c-4931-bfd3-7635f6be6507",
				SupervisorIds: []string{},
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			err := s.postgres.CreatePosition(test.input)
			s.expectErrorCode(err, "INVALID-FOREIGN-KEY-ERROR")

			s.expectSelectQueryToReturnNoRows(
				"position",
				map[string]string{
					"id":            test.input.Id,
					"tenant_id":     test.input.TenantId,
					"title":         test.input.Title,
					"department_id": test.input.DepartmentId,
				},
			)
		})
	}
}

func (s *IntegrationTestSuite) TestCreatePositionWithSupervisor() {
	wantPosition := storage.Position{
		Id:            "a9f998c6-ba2e-4359-b308-e56404534974",
		TenantId:      s.defaultPosition.TenantId,
		Title:         "Manager",
		DepartmentId:  s.defaultPosition.DepartmentId,
		SupervisorIds: []string{s.defaultPosition.Id, "975132d2-7b2a-49af-9e73-d090b11ef3b1"},
	}

	// Seed another position
	query := "INSERT INTO position (id, tenant_id, title, department_id) VALUES ($1, $2, $3, $4)"
	_, err := s.dbRootConn.Exec(query, "975132d2-7b2a-49af-9e73-d090b11ef3b1", s.defaultTenant.Id, "Test", s.defaultDepartment.Id)
	if err != nil {
		log.Fatalf("Could not seed position: %s", err)
	}

	err = s.postgres.CreatePosition(wantPosition)
	s.Equal(nil, err)

	s.expectSelectQueryToReturnOneRow(
		"position",
		map[string]string{
			"id":            wantPosition.Id,
			"tenant_id":     wantPosition.TenantId,
			"title":         wantPosition.Title,
			"department_id": wantPosition.DepartmentId,
		},
	)

	for _, supervisorId := range wantPosition.SupervisorIds {
		s.expectSelectQueryToReturnOneRow(
			"subordinate_supervisor_relationship",
			map[string]string{
				"subordinate_position_id": wantPosition.Id,
				"supervisor_position_id":  supervisorId,
			},
		)
	}
}

func (s *IntegrationTestSuite) TestCreatePositionWithSupervisorViolatesUniqueConstraint() {
	wantPosition := storage.Position{
		Id:            "a9f998c6-ba2e-4359-b308-e56404534974",
		TenantId:      s.defaultPosition.TenantId,
		Title:         "Manager",
		DepartmentId:  s.defaultPosition.DepartmentId,
		SupervisorIds: []string{s.defaultPosition.Id, s.defaultPosition.Id},
	}

	err := s.postgres.CreatePosition(wantPosition)
	s.expectErrorCode(err, "UNIQUE-VIOLATION-ERROR")

	s.expectSelectQueryToReturnNoRows(
		"position",
		map[string]string{
			"id":            wantPosition.Id,
			"tenant_id":     wantPosition.TenantId,
			"title":         wantPosition.Title,
			"department_id": wantPosition.DepartmentId,
		},
	)

	for _, supervisorId := range wantPosition.SupervisorIds {
		s.expectSelectQueryToReturnNoRows(
			"subordinate_supervisor_relationship",
			map[string]string{
				"subordinate_position_id": wantPosition.Id,
				"supervisor_position_id":  supervisorId,
			},
		)
	}
}

func (s *IntegrationTestSuite) TestCreatePositionWithSupervisorViolatesCheckConstraint() {
	wantPosition := storage.Position{
		Id:            "a9f998c6-ba2e-4359-b308-e56404534974",
		TenantId:      s.defaultPosition.TenantId,
		Title:         "Manager",
		DepartmentId:  s.defaultPosition.DepartmentId,
		SupervisorIds: []string{"a9f998c6-ba2e-4359-b308-e56404534974"},
	}

	err := s.postgres.CreatePosition(wantPosition)
	s.expectErrorCode(err, "INVALID-SUBORDINATE-SUPERVISOR-PAIR-ERROR")

	s.expectSelectQueryToReturnNoRows(
		"position",
		map[string]string{
			"id":            wantPosition.Id,
			"tenant_id":     wantPosition.TenantId,
			"title":         wantPosition.Title,
			"department_id": wantPosition.DepartmentId,
		},
	)

	s.expectSelectQueryToReturnNoRows(
		"subordinate_supervisor_relationship",
		map[string]string{
			"subordinate_position_id": wantPosition.Id,
			"supervisor_position_id":  wantPosition.SupervisorIds[0],
		},
	)
}

func (s *IntegrationTestSuite) TestCreatePositionWithSupervisorViolatesForeignKeyConstraint() {
	wantPosition := storage.Position{
		Id:            "a9f998c6-ba2e-4359-b308-e56404534974",
		TenantId:      s.defaultPosition.TenantId,
		Title:         "Manager",
		DepartmentId:  s.defaultPosition.DepartmentId,
		SupervisorIds: []string{"091876aa-ee30-49a3-b0b5-f6e8f320c687"},
	}

	err := s.postgres.CreatePosition(wantPosition)
	s.expectErrorCode(err, "INVALID-FOREIGN-KEY-ERROR")

	s.expectSelectQueryToReturnNoRows(
		"position",
		map[string]string{
			"id":            wantPosition.Id,
			"tenant_id":     wantPosition.TenantId,
			"title":         wantPosition.Title,
			"department_id": wantPosition.DepartmentId,
		},
	)

	s.expectSelectQueryToReturnNoRows(
		"subordinate_supervisor_relationship",
		map[string]string{
			"subordinate_position_id": wantPosition.Id,
			"supervisor_position_id":  wantPosition.SupervisorIds[0],
		},
	)
}

func (s *IntegrationTestSuite) TestCreatePositionAssignment() {
	tests := []struct {
		name  string
		input storage.PositionAssignment
	}{
		{
			"Should be valid without end date",
			storage.PositionAssignment{
				TenantId:   s.defaultPositionAssignment.TenantId,
				PositionId: "c1ddb117-94e0-40d1-908d-a07f43f319e8",
				UserId:     s.defaultPositionAssignment.UserId,
				StartDate:  s.defaultPositionAssignment.StartDate,
			},
		},
		{
			"Should be valid with end date",
			storage.PositionAssignment{
				TenantId:   s.defaultPositionAssignment.TenantId,
				PositionId: "c1ddb117-94e0-40d1-908d-a07f43f319e8",
				UserId:     s.defaultPositionAssignment.UserId,
				StartDate:  s.defaultPositionAssignment.StartDate,
				EndDate:    "2024-10-20",
			},
		},
	}

	// Seed another position
	query := "INSERT INTO position (id, tenant_id, title, department_id) VALUES ($1, $2, $3, $4)"
	_, err := s.dbRootConn.Exec(query, "c1ddb117-94e0-40d1-908d-a07f43f319e8", s.defaultTenant.Id, "Random", s.defaultDepartment.Id)
	if err != nil {
		log.Fatalf("Could not seed position: %s", err)
	}

	for _, test := range tests {
		err := s.postgres.CreatePositionAssignment(test.input)
		s.Equal(nil, err)

		s.expectSelectQueryToReturnOneRow(
			"position_assignment",
			map[string]string{
				"tenant_id":       test.input.TenantId,
				"position_id":     test.input.PositionId,
				"user_account_id": test.input.UserId,
				"start_date":      test.input.StartDate,
			},
		)

		query = "DELETE FROM position_assignment WHERE position_id = $1 AND user_account_id = $2"
		s.dbRootConn.Exec(query, test.input.PositionId, test.input.UserId)
	}
}

func (s *IntegrationTestSuite) TestCreatePositionAssignmentViolatesUniqueConstraint() {
	wantPositionAssignment := storage.PositionAssignment{
		TenantId:   s.defaultPositionAssignment.TenantId,
		PositionId: s.defaultPositionAssignment.PositionId,
		UserId:     s.defaultPositionAssignment.UserId,
		StartDate:  s.defaultPositionAssignment.StartDate,
	}

	err := s.postgres.CreatePositionAssignment(wantPositionAssignment)
	s.expectErrorCode(err, "UNIQUE-VIOLATION-ERROR")

	s.expectSelectQueryToReturnOneRow(
		"position_assignment",
		map[string]string{
			"tenant_id":       wantPositionAssignment.TenantId,
			"position_id":     wantPositionAssignment.PositionId,
			"user_account_id": wantPositionAssignment.UserId,
			"start_date":      wantPositionAssignment.StartDate,
		},
	)
}

func (s *IntegrationTestSuite) TestCreatePositionAssignmentDoesNotViolateUniqueConstraint() {
	// Seed another user and position
	query := "INSERT INTO user_account (id, tenant_id, email, password, totp_secret_key) VALUES ($1, $2, $3, $4, $5)"
	_, err := s.dbRootConn.Exec(query, "bc1a0220-fdef-403c-ae0b-5664daefb328", s.defaultUser.TenantId, "test1@gmail.com", "", "")
	if err != nil {
		log.Fatalf("Could not seed user: %s", err)
	}

	query = "INSERT INTO position (id, tenant_id, title, department_id) VALUES ($1, $2, $3, $4)"
	_, err = s.dbRootConn.Exec(query, "50a10748-28a3-466b-b807-947fbc049eda", s.defaultTenant.Id, "Random", s.defaultDepartment.Id)
	if err != nil {
		log.Fatalf("Could not seed position: %s", err)
	}

	tests := []struct {
		name  string
		input storage.PositionAssignment
	}{
		{
			"Should not violate unique constraint as the user id is different",
			storage.PositionAssignment{
				TenantId:   s.defaultPositionAssignment.TenantId,
				PositionId: s.defaultPositionAssignment.PositionId,
				UserId:     "bc1a0220-fdef-403c-ae0b-5664daefb328",
				StartDate:  "2024-02-01",
			},
		},
		{
			"Should not violate unique constraint as the position id is different",
			storage.PositionAssignment{
				TenantId:   s.defaultPositionAssignment.TenantId,
				PositionId: "50a10748-28a3-466b-b807-947fbc049eda",
				UserId:     s.defaultUser.Id,
				StartDate:  "2024-02-01",
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			err := s.postgres.CreatePositionAssignment(test.input)
			s.Equal(nil, err)

			s.expectSelectQueryToReturnOneRow(
				"position_assignment",
				map[string]string{
					"tenant_id":       test.input.TenantId,
					"position_id":     test.input.PositionId,
					"user_account_id": test.input.UserId,
					"start_date":      test.input.StartDate,
				},
			)

			query := "DELETE FROM position_assignment WHERE position_id = $1 AND user_account_id = $2"
			s.dbRootConn.Exec(query, test.input.PositionId, test.input.UserId)
		})
	}
}

func (s *IntegrationTestSuite) TestCreatePositionAssignmentViolatesForeignKeyConstraint() {
	tests := []struct {
		name  string
		input storage.PositionAssignment
	}{
		{
			"Should violate foreign key constraint because tenant id does not exist",
			storage.PositionAssignment{
				TenantId:   "398fe284-f20f-456d-ae0a-270adc4e737a",
				PositionId: s.defaultPositionAssignment.PositionId,
				UserId:     "bc1a0220-fdef-403c-ae0b-5664daefb328",
				StartDate:  "2024-02-01",
			},
		},
		{
			"Should violate foreign key constraint because user id does not exist",
			storage.PositionAssignment{
				TenantId:   s.defaultPositionAssignment.TenantId,
				PositionId: s.defaultPositionAssignment.PositionId,
				UserId:     "bc1a0220-fdef-433c-ae0b-5664daefb328",
				StartDate:  "2024-02-01",
			},
		},
		{
			"Should violate foreign key constraint because position id does not exist",
			storage.PositionAssignment{
				TenantId:   s.defaultPositionAssignment.TenantId,
				PositionId: "b1420372-a814-4f4d-ad77-65c6feb9561a",
				UserId:     s.defaultPositionAssignment.UserId,
				StartDate:  "2024-02-01",
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			err := s.postgres.CreatePositionAssignment(test.input)
			s.expectErrorCode(err, "INVALID-FOREIGN-KEY-ERROR")

			s.expectSelectQueryToReturnNoRows(
				"position_assignment",
				map[string]string{
					"tenant_id":       test.input.TenantId,
					"position_id":     test.input.PositionId,
					"user_account_id": test.input.UserId,
					"start_date":      test.input.StartDate,
				},
			)
		})
	}
}
