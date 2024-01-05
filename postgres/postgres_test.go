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

	"multi-tenant-HR-information-system-backend/routes"
)

type PostgresTestSuite struct {
	suite.Suite
	postgres    *postgresStorage
	dbRootConn  *sql.DB
	dbTables    []string
	defaultUser routes.User
}

func TestPostgresIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	suite.Run(t, &PostgresTestSuite{
		defaultUser: routes.User{
			Id:            "e7f31b70-ae26-42b3-b7a6-01ec68d5c33a",
			Email:         "root-role-admin@hrisEnterprises.org",
			Tenant:        "HRIS Enterprises",
			Password:      "$argon2id$v=19$m=65536,t=1,p=8$cFTNg+YXrN4U0lvwnamPkg$0RDBxH+EouVxDbBlQUNctdWZ+CNKrayPpzTJaWNq83U",
			TotpSecretKey: "OLDFXRMH35A3DU557UXITHYDK4SKLTXZ",
		},
	})
}

func (s *PostgresTestSuite) SetupSuite() {
	// Create the postgres container
	cmd := exec.Command("docker", "run", "--name", "integration_test", "-e", "POSTGRES_PASSWORD=abcd1234", "-e", "POSTGRES_DB=hr_information_system", "-p", "5434:5432", "-v", `C:\Users\perio\Documents\Coding\Projects\multi-tenant-HR-information-system\multi-tenant-HR-information-system-backend\init.sql:/docker-entrypoint-initdb.d/init.sql`, "-d", "postgres")
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
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

func (s *PostgresTestSuite) TearDownSuite() {
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

func (s *PostgresTestSuite) SetupTest() {
	// Re-insert the root administrator user & privileges
	insertTenant := "INSERT INTO tenant (name) VALUES ($1)"
	insertUser := `
					INSERT INTO user_account (id, email, tenant, password, totp_secret_key) 
					VALUES ($1, $2, $3, $4, $5)					
					`
	_, err := s.dbRootConn.Query(insertTenant, s.defaultUser.Tenant)
	if err != nil {
		log.Fatalf("DB seeding failed: %s", err)
	}

	_, err = s.dbRootConn.Query(insertUser, s.defaultUser.Id, s.defaultUser.Email, s.defaultUser.Tenant, s.defaultUser.Password, s.defaultUser.TotpSecretKey)
	if err != nil {
		log.Fatalf("DB seeding failed: %s", err)
	}

}

func (s *PostgresTestSuite) TearDownTest() {
	// Clear all data
	query := fmt.Sprintf("TRUNCATE %s", strings.Join(s.dbTables, ", "))
	_, err := s.dbRootConn.Exec(query)
	if err != nil {
		log.Fatalf("Could not clear data from all tables: %s", err)
	}
}

func (s *PostgresTestSuite) expectSelectQueryToReturnNoRows(table string, attributes []string, values []string) {
	// Convert the string slice to an any slice
	valuesAny := make([]interface{}, len(values))
	for i, v := range values {
		valuesAny[i] = v
	}

	query := NewDynamicConditionQuery(fmt.Sprintf("SELECT created_at FROM %s", table), attributes)
	rows, err := s.dbRootConn.Query(query, valuesAny...)
	if err != nil {
		log.Fatal(err)
	}

	count := 0
	for rows.Next() {
		count++
	}
	s.Equal(0, count, "No rows should be returned")
}

func (s *PostgresTestSuite) expectSelectQueryToReturnOneRow(table string, attributes []string, values []string) {
	// Convert the string slice to an any slice
	valuesAny := make([]interface{}, len(values))
	for i, v := range values {
		valuesAny[i] = v
	}

	query := NewDynamicConditionQuery(fmt.Sprintf("SELECT created_at FROM %s", table), attributes)
	rows, err := s.dbRootConn.Query(query, valuesAny...)
	if err != nil {
		log.Fatal(err)
	}

	count := 0
	for rows.Next() {
		count++
	}
	s.Equal(1, count, "1 row should be returned")
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
			} else if err != nil && err.Error() != "EOF" {
				return nil, err
			}
		}
	}
}
