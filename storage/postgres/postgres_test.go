package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"

	"multi-tenant-HR-information-system-backend/httperror"
	"multi-tenant-HR-information-system-backend/storage"
)

// Postgres integration tests
// Purposes:
//  1. Verify that all happy paths work
//	2. Verify that all the expected unique, foreign key, and check constraints are present in the database schema
//	3. Verify that unique constraints do not incorrectly block valid inputs
//  4. Verify that all constraint violation errors are handled correctly

type IntegrationTestSuite struct {
	suite.Suite
	postgres                  *postgresStorage
	dbRootConn                *sql.DB
	dbTables                  []string
	defaultTenant             storage.Tenant
	defaultDivision           storage.Division
	defaultDepartment         storage.Department
	defaultUser                         storage.User
	defaultPosition                     storage.Position
	defaultPositionAssignment           storage.PositionAssignment
	defaultPolicies                     storage.Policies
	defaultRoleAssignment               storage.RoleAssignment
	defaultSupervisor                   storage.User
	defaultSupervisorPosition           storage.Position
	defaultSupervisorPositionAssignment storage.PositionAssignment
	defaultHrApprover                   storage.User
	defaultRecruiter                    storage.User
	defaultJobRequisition               storage.JobRequisition
}

func TestPostgresIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	suite.Run(t, &IntegrationTestSuite{
		defaultTenant: storage.Tenant{
			Id:   "2ad1dcfc-8867-49f7-87a3-8bd8d1154924",
			Name: "HRIS Enterprises",
		},
		defaultDivision: storage.Division{
			Id:       "f8b1551a-71bb-48c4-924a-8a25a6bff71d",
			TenantId: "2ad1dcfc-8867-49f7-87a3-8bd8d1154924",
			Name:     "Operations",
		},
		defaultDepartment: storage.Department{
			Id:         "9147b727-1955-437b-be7d-785e9a31f20c",
			TenantId:   "2ad1dcfc-8867-49f7-87a3-8bd8d1154924",
			DivisionId: "f8b1551a-71bb-48c4-924a-8a25a6bff71d",
			Name:       "Operations",
		},
		defaultUser: storage.User{
			Id:            "e7f31b70-ae26-42b3-b7a6-01ec68d5c33a",
			TenantId:      "2ad1dcfc-8867-49f7-87a3-8bd8d1154924",
			Email:         "root-role-admin@hrisEnterprises.org",
			Password:      "$argon2id$v=19$m=65536,t=1,p=8$cFTNg+YXrN4U0lvwnamPkg$0RDBxH+EouVxDbBlQUNctdWZ+CNKrayPpzTJaWNq83U",
			TotpSecretKey: "OLDFXRMH35A3DU557UXITHYDK4SKLTXZ",
		},
		defaultPosition: storage.Position{
			Id:           "e4edbd37-164d-478d-9625-5b1397ef6e45",
			TenantId:     "2ad1dcfc-8867-49f7-87a3-8bd8d1154924",
			Title:        "System Administrator",
			DepartmentId: "9147b727-1955-437b-be7d-785e9a31f20c",
		},
		defaultPositionAssignment: storage.PositionAssignment{
			TenantId:   "2ad1dcfc-8867-49f7-87a3-8bd8d1154924",
			PositionId: "e4edbd37-164d-478d-9625-5b1397ef6e45",
			UserId:     "e7f31b70-ae26-42b3-b7a6-01ec68d5c33a",
			StartDate:  "2024-02-01",
		},
		defaultPolicies: storage.Policies{
			Role:     "ROOT_ROLE_ADMIN",
			TenantId: "2ad1dcfc-8867-49f7-87a3-8bd8d1154924",
			Resources: []storage.Resource{
				{
					Path:   "/api/tenants/*",
					Method: "POST",
				},
				{
					Path:   "/api/tenants/*/divisions/*",
					Method: "POST",
				},
				{
					Path:   "/api/tenants/*/divisions/*/departments",
					Method: "POST",
				},
				{
					Path:   "/api/tenants/*/users/*",
					Method: "POST",
				},
				{
					Path:   "/api/tenants/*/positions/*",
					Method: "POST",
				},
				{
					Path:   "/api/tenants/*/users/*/position-assignments/*",
					Method: "POST",
				},
				{
					Path:   "/api/tenants/*/roles/*",
					Method: "POST",
				},
				{
					Path:   "/api/tenants/*/users/*/role-assignments/*",
					Method: "POST",
				},
			},
		},
		defaultRoleAssignment: storage.RoleAssignment{
			UserId:   "e7f31b70-ae26-42b3-b7a6-01ec68d5c33a",
			Role:     "ROOT_ROLE_ADMIN",
			TenantId: "2ad1dcfc-8867-49f7-87a3-8bd8d1154924",
		},
		defaultSupervisor: storage.User{
			Id:            "38d3f831-9a9e-4dfc-ba56-ec68bf2462e0",
			TenantId:      "2ad1dcfc-8867-49f7-87a3-8bd8d1154924",
			Email:         "administration-manager@hrisEnterprises.org",
			Password:      "$argon2id$v=19$m=65536,t=1,p=8$cFTNg+YXrN4U0lvwnamPkg$0RDBxH+EouVxDbBlQUNctdWZ+CNKrayPpzTJaWNq83U",
			TotpSecretKey: "OLDFXRMH35A3DU557UXITHYDK4SKLTXZ",
		},
		defaultSupervisorPosition: storage.Position{
			Id:           "0c55ff72-a23d-440b-b77f-db6b8002f734",
			TenantId:     "2ad1dcfc-8867-49f7-87a3-8bd8d1154924",
			Title:        "Manager",
			DepartmentId: "9147b727-1955-437b-be7d-785e9a31f20c",
		},
		defaultSupervisorPositionAssignment: storage.PositionAssignment{
			TenantId:   "2ad1dcfc-8867-49f7-87a3-8bd8d1154924",
			PositionId: "0c55ff72-a23d-440b-b77f-db6b8002f734",
			UserId:     "38d3f831-9a9e-4dfc-ba56-ec68bf2462e0",
			StartDate:  "2024-02-01",
		},
		defaultHrApprover: storage.User{
			Id:            "9f4c9dd0-7c75-4ea9-a106-948885b6bedf",
			TenantId:      "2ad1dcfc-8867-49f7-87a3-8bd8d1154924",
			Email:         "hr-director@hrisEnterprises.org",
			Password:      "$argon2id$v=19$m=65536,t=1,p=8$cFTNg+YXrN4U0lvwnamPkg$0RDBxH+EouVxDbBlQUNctdWZ+CNKrayPpzTJaWNq83U",
			TotpSecretKey: "OLDFXRMH35A3DU557UXITHYDK4SKLTXZ",
		},
		defaultRecruiter: storage.User{
			Id:            "ccb2da3b-68ac-419e-b95d-dd6b723035f9",
			TenantId:      "2ad1dcfc-8867-49f7-87a3-8bd8d1154924",
			Email:         "hr-recruiter@hrisEnterprises.org",
			Password:      "$argon2id$v=19$m=65536,t=1,p=8$cFTNg+YXrN4U0lvwnamPkg$0RDBxH+EouVxDbBlQUNctdWZ+CNKrayPpzTJaWNq83U",
			TotpSecretKey: "OLDFXRMH35A3DU557UXITHYDK4SKLTXZ",
		},
		defaultJobRequisition: storage.JobRequisition{
			Id:              "5062a285-e82b-475d-8113-daefd05dcd90",
			TenantId:        "2ad1dcfc-8867-49f7-87a3-8bd8d1154924",
			Title:           "Database Administrator",
			DepartmentId:    "9147b727-1955-437b-be7d-785e9a31f20c",
			JobDescription:  "Manages databases of HRIS software",
			JobRequirements: "100 years of experience using postgres",
			Requestor:       "e7f31b70-ae26-42b3-b7a6-01ec68d5c33a",
			Supervisor:      "38d3f831-9a9e-4dfc-ba56-ec68bf2462e0",
			HrApprover:      "9f4c9dd0-7c75-4ea9-a106-948885b6bedf",
		},		
	})
}

func (s *IntegrationTestSuite) SetupSuite() {
	// Create the postgres container
	cmd := exec.Command("docker", "run", "--name", "integration_test", "-e", "POSTGRES_PASSWORD=abcd1234", "-e", "POSTGRES_DB=hr_information_system", "-p", "5434:5432", "-v", `C:\Users\perio\Documents\Coding\Projects\multi-tenant-HR-information-system\multi-tenant-HR-information-system-backend\init.sql:/docker-entrypoint-initdb.d/init.sql`, "-d", "postgres")
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout

	if err := cmd.Start(); err != nil {
		log.Fatalf("Could not create postgres docker instance for postgres_test: %s", err)
	}

	if err := cmd.Wait(); err != nil {
		log.Fatalf("Could not create postgres docker instance for postgres_test: %s", err)
	}

	// Fetch the database's tables & clear any data the container might have been seeded with
	var err error
	dbRootConnString := "host=localhost port=5434 user=postgres password=abcd1234 dbname=hr_information_system sslmode=disable"
	s.dbRootConn, err = attemptDBconnectionUntilTimeout(dbRootConnString)
	if err != nil {
		log.Fatalf("Could not connect to the docker postgres instance: %s", err)
	}

	query := "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'"
	rows, err := s.dbRootConn.Query(query)
	if err != nil {
		log.Fatalf("Could not fetch database tables: %s", err)
	}

	tables := []string{}
	for rows.Next() {
		var table string
		err := rows.Scan(&table)
		if err != nil {
			log.Fatalf("Could not fetch database tables: %s", err)
		}

		tables = append(tables, table)
	}
	s.dbTables = tables

	query = fmt.Sprintf("TRUNCATE %s", strings.Join(tables, ", "))
	_, err = s.dbRootConn.Exec(query)
	if err != nil {
		log.Fatalf("Could not clear data from all tables")
	}

	// Instantiate a postgresStorage instance
	dbAppConnString := "host=localhost port=5434 user=hr_information_system password=abcd1234 dbname=hr_information_system sslmode=disable"
	s.postgres, err = NewPostgresStorage(dbAppConnString)
	if err != nil {
		log.Fatalf("Could not instantiate postgres storage instance: %s", err)
	}
}

func (s *IntegrationTestSuite) TearDownSuite() {
	// Stop & remove the postgres container
	cmd := exec.Command("docker", "stop", "integration_test")
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}

	cmd2 := exec.Command("docker", "rm", "integration_test")
	cmd2.Env = os.Environ()
	cmd2.Stdout = os.Stdout

	if err := cmd2.Start(); err != nil {
		log.Fatal(err)
	}

	if err := cmd2.Wait(); err != nil {
		log.Fatal(err)
	}
}

func (s *IntegrationTestSuite) SetupTest() {
	// Re-insert the root administrator user
	insertTenant := "INSERT INTO tenant (id, name) VALUES ($1, $2)"
	_, err := s.dbRootConn.Exec(insertTenant, s.defaultTenant.Id, s.defaultTenant.Name)
	if err != nil {
		log.Fatalf("Tenant seeding failed: %s", err)
	}

	insertDivision := "INSERT INTO division (id, tenant_id, name) VALUES ($1, $2, $3)"
	_, err = s.dbRootConn.Exec(insertDivision, s.defaultDivision.Id, s.defaultDivision.TenantId, s.defaultDivision.Name)
	if err != nil {
		log.Fatalf("Division seeding failed: %s", err)
	}

	insertDepartment := "INSERT INTO department (id, tenant_id, division_id, name) VALUES ($1, $2, $3, $4)"
	_, err = s.dbRootConn.Exec(insertDepartment, s.defaultDepartment.Id, s.defaultDepartment.TenantId, s.defaultDepartment.DivisionId, s.defaultDepartment.Name)
	if err != nil {
		log.Fatalf("Department seeding failed: %s", err)
	}

	insertUser := `
					INSERT INTO user_account (id, email, tenant_id, password, totp_secret_key) 
					VALUES ($1, $2, $3, $4, $5)					
					`
	_, err = s.dbRootConn.Exec(insertUser, s.defaultUser.Id, s.defaultUser.Email, s.defaultUser.TenantId, s.defaultUser.Password, s.defaultUser.TotpSecretKey)
	if err != nil {
		log.Fatalf("User seeding failed: %s", err)
	}

	insertPosition := `
					INSERT INTO position (id, tenant_id, title, department_id) 
					VALUES ($1, $2, $3, $4)
					`
	_, err = s.dbRootConn.Exec(insertPosition, s.defaultPosition.Id, s.defaultPosition.TenantId, s.defaultPosition.Title, s.defaultPosition.DepartmentId)
	if err != nil {
		log.Fatalf("Position seeding failed: %s", err)
	}

	insertPositionAssignment := `
								INSERT INTO position_assignment (tenant_id, position_id, user_account_id, start_date) 
								VALUES ($1, $2, $3, $4)
								`
	_, err = s.dbRootConn.Exec(insertPositionAssignment, s.defaultPositionAssignment.TenantId, s.defaultPositionAssignment.PositionId, s.defaultPositionAssignment.UserId, s.defaultPositionAssignment.StartDate)
	if err != nil {
		log.Fatalf("Position seeding failed: %s", err)
	}

	insertPolicy := `INSERT INTO casbin_rule (Ptype, V0, V1, V2, V3) VALUES ('p', $1, $2, $3, $4)`
	for _, resource := range s.defaultPolicies.Resources {
		_, err := s.dbRootConn.Exec(insertPolicy, s.defaultPolicies.Role, s.defaultPolicies.TenantId, resource.Path, resource.Method)
		if err != nil {
			log.Fatalf("Policy seeding failed: %s", err)
		}
	}

	insertPublicPolicies := `INSERT INTO casbin_rule (Ptype, V0, V1, V2, V3) VALUES
							 ('p', 'PUBLIC', '*', '/api/session', 'POST'),
							 ('p', 'PUBLIC', '*', '/api/session', 'DELETE')
							`
	_, err = s.dbRootConn.Exec(insertPublicPolicies)
	if err != nil {
		log.Fatalf("Policy seeding failed: %s", err)
	}

	insertRoleAssignments := `INSERT INTO casbin_rule (Ptype, V0, V1, V2) VALUES ('g', $1, $2, $3)`
	_, err = s.dbRootConn.Exec(insertRoleAssignments, s.defaultRoleAssignment.UserId, s.defaultRoleAssignment.Role, s.defaultRoleAssignment.TenantId)
	if err != nil {
		log.Fatalf("DB seeding failed: %s", err)
	}

	insertPublicRoleAssignments := `INSERT INTO casbin_rule (Ptype, V0, V1, V2) VALUES ('g', '*', 'PUBLIC', '*')`
	_, err = s.dbRootConn.Exec(insertPublicRoleAssignments)
	if err != nil {
		log.Fatalf("DB seeding failed: %s", err)
	}

	insertSupervisor := "INSERT INTO user_account (id, email, tenant_id, password, totp_secret_key) VALUES ($1, $2, $3, $4, $5)"
	_, err = s.dbRootConn.Exec(insertSupervisor, s.defaultSupervisor.Id, s.defaultSupervisor.Email, s.defaultSupervisor.TenantId, s.defaultSupervisor.Password, s.defaultSupervisor.TotpSecretKey)
	if err != nil {
		log.Fatalf("User seeding failed: %s", err)
	}

	insertHrApprover := "INSERT INTO user_account (id, email, tenant_id, password, totp_secret_key) VALUES ($1, $2, $3, $4, $5)"
	_, err = s.dbRootConn.Exec(insertHrApprover, s.defaultHrApprover.Id, s.defaultHrApprover.Email, s.defaultHrApprover.TenantId, s.defaultHrApprover.Password, s.defaultHrApprover.TotpSecretKey)
	if err != nil {
		log.Fatalf("User seeding failed: %s", err)
	}

	insertRecruiter := "INSERT INTO user_account (id, email, tenant_id, password, totp_secret_key) VALUES ($1, $2, $3, $4, $5)"
	_, err = s.dbRootConn.Exec(insertRecruiter, s.defaultRecruiter.Id, s.defaultRecruiter.Email, s.defaultRecruiter.TenantId, s.defaultRecruiter.Password, s.defaultRecruiter.TotpSecretKey)
	if err != nil {
		log.Fatalf("User seeding failed: %s", err)
	}

	insertOtherPolicies := `
		INSERT INTO casbin_rule (Ptype, V0, V1, V2, V3) VALUES 
		('p', 'JOB_REQUISITION_REQUESTOR', '2ad1dcfc-8867-49f7-87a3-8bd8d1154924', '/api/tenants/2ad1dcfc-8867-49f7-87a3-8bd8d1154924/users/e7f31b70-ae26-42b3-b7a6-01ec68d5c33a/job-requisitions/*/supervisor-approval', 'POST'),
		('p', 'JOB_REQUISITION_SUPERVISOR', '2ad1dcfc-8867-49f7-87a3-8bd8d1154924', '/api/tenants/2ad1dcfc-8867-49f7-87a3-8bd8d1154924/users/38d3f831-9a9e-4dfc-ba56-ec68bf2462e0/job-requisitions/*/supervisor-approval', 'POST'),
		('p', 'JOB_REQUISITION_HR_APPROVER', '2ad1dcfc-8867-49f7-87a3-8bd8d1154924', '/api/tenants/2ad1dcfc-8867-49f7-87a3-8bd8d1154924/users/9f4c9dd0-7c75-4ea9-a106-948885b6bedf/job-requisitions/*/hr-approval', 'POST')
	`
	_, err = s.dbRootConn.Exec(insertOtherPolicies)
	if err != nil {
		log.Fatalf("Other policy seeding failed: %s", err)
	}

	insertOtherRoleAssignments := `
	INSERT INTO casbin_rule (Ptype, V0, V1, V2) VALUES 		
		('g', 'e7f31b70-ae26-42b3-b7a6-01ec68d5c33a', 'JOB_REQUISITION_REQUESTOR', '2ad1dcfc-8867-49f7-87a3-8bd8d1154924'),
		('g', '38d3f831-9a9e-4dfc-ba56-ec68bf2462e0', 'JOB_REQUISITION_SUPERVISOR', '2ad1dcfc-8867-49f7-87a3-8bd8d1154924'),				
		('g', '9f4c9dd0-7c75-4ea9-a106-948885b6bedf', 'JOB_REQUISITION_HR_APPROVER', '2ad1dcfc-8867-49f7-87a3-8bd8d1154924')			
	`
	_, err = s.dbRootConn.Exec(insertOtherRoleAssignments)
	if err != nil {
		log.Fatalf("Other role assignment failed: %s", err)
	}
}

func (s *IntegrationTestSuite) TearDownTest() {
	// Clear all data
	query := fmt.Sprintf("TRUNCATE %s", strings.Join(s.dbTables, ", "))
	_, err := s.dbRootConn.Exec(query)
	if err != nil {
		log.Fatalf("Could not clear data from all tables: %s", err)
	}
}

func (s *IntegrationTestSuite) expectSelectQueryToReturnNoRows(table string, filter map[string]string) {
	// Convert the string slice to an any slice
	conditions := []string{}
	values := []any{}

	for column, value := range filter {
		conditions = append(conditions, fmt.Sprintf("%s = $%v", column, len(conditions) + 1))
		values = append(values, value)
	}

	query := NewQueryWithFilter(fmt.Sprintf("SELECT created_at FROM %s", table), conditions)
	rows, err := s.dbRootConn.Query(query, values...)
	if err != nil {
		log.Fatalf("Could not execute select return 0 rows query: %s", err)
	}

	count := 0
	for rows.Next() {
		count++
	}
	s.Equal(0, count, "No rows should be returned")
}

func (s *IntegrationTestSuite) expectSelectQueryToReturnOneRow(table string, filter map[string]string) {
	// Convert the string slice to an any slice
	conditions := []string{}
	values := []any{}

	for column, value := range filter {
		conditions = append(conditions, fmt.Sprintf("%s = $%v", column, len(conditions) + 1))
		values = append(values, value)
	}

	query := NewQueryWithFilter(fmt.Sprintf("SELECT created_at FROM %s", table), conditions)
	rows, err := s.dbRootConn.Query(query, values...)
	if err != nil {
		log.Fatalf("Could not execute select return 1 row query: %s", err)
	}

	count := 0
	for rows.Next() {
		count++
	}
	s.Equal(1, count, "1 row should be returned")
}

func (s *IntegrationTestSuite) expectErrorCode(err error, code string) {
	s.Equal(true, err != nil, "Error should not be nil")

	httpErr, ok := err.(*httperror.Error)
	s.Equal(true, ok, "Error should be HttpError")

	if ok {
		s.Equal(code, httpErr.Code)
	}
}

func attemptDBconnectionUntilTimeout(dbRootConnString string) (*sql.DB, error) {
	tick := time.Tick(500 * time.Millisecond)
	timeout := time.After(10 * time.Second)
	for {
		select {
		case <-timeout:
			return nil, errors.New("Attempt to connect to the Database timed out")
		case <-tick:
			conn, err := sql.Open("postgres", dbRootConnString)
			if err != nil {
				return nil, err
			}

			err = conn.Ping()
			if err == nil {
				return conn, nil
			} else if err != nil && err.Error() != "EOF" && err.Error() != "pq: the database system is starting up" {
				return nil, err
			}
		}
	}
}
