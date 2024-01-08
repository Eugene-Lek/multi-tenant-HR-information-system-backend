package main

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	pgadapter "github.com/casbin/casbin-pg-adapter"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/util"
	"github.com/go-pg/pg/v10"
	"github.com/gorilla/sessions"
	_ "github.com/lib/pq"
	"github.com/quasoft/memstore"
	"github.com/stretchr/testify/suite"

	"multi-tenant-HR-information-system-backend/postgres"
	"multi-tenant-HR-information-system-backend/routes"
)

// API integration tests
// Purposes:
//  1. Verify that all happy paths work
//  2. Verify that errors from input validation & postgres models are handled correctly
//	3. Verify that all expected logs are generated

type errorResponseBody struct {
	Code    string
	Message string
}

type IntegrationTestSuite struct {
	suite.Suite
	router             *routes.Router
	dbRootConn         *sql.DB
	logOutput          *bytes.Buffer
	sessionStore       sessions.Store
	dbTables           []string
	defaultUser        routes.User
	defaultAppointment routes.Appointment
}

func TestAPIEndpointsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	suite.Run(t, &IntegrationTestSuite{
		defaultUser: routes.User{
			Id:            "e7f31b70-ae26-42b3-b7a6-01ec68d5c33a",
			Email:         "root-role-admin@hrisEnterprises.org",
			Tenant:        "HRIS Enterprises",
			Password:      "$argon2id$v=19$m=65536,t=1,p=8$cFTNg+YXrN4U0lvwnamPkg$0RDBxH+EouVxDbBlQUNctdWZ+CNKrayPpzTJaWNq83U",
			TotpSecretKey: "OLDFXRMH35A3DU557UXITHYDK4SKLTXZ",
		},
		defaultAppointment: routes.Appointment{
			Title:      "System Administrator",
			Tenant:     "HRIS Enterprises",
			Division:   "Operations",
			Department: "Administration",
			UserId:     "e7f31b70-ae26-42b3-b7a6-01ec68d5c33a",
			StartDate:  "2024-02-01",
		},
	})
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

func (s *IntegrationTestSuite) SetupSuite() {
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

	// Instantiate server
	logOutputMedium := bytes.Buffer{}
	rootLogger := routes.NewRootLogger(&logOutputMedium) // Set output to a buffer so it can be read & checked

	dbAppConnString := "host=localhost port=5434 user=hr_information_system password=abcd1234 dbname=hr_information_system sslmode=disable"
	log.Println("Connecting to database on url: ", dbAppConnString)
	storage, err := postgres.NewPostgresStorage(dbAppConnString)
	if err != nil {
		rootLogger.Fatal("DB-CONNECTION-FAILED", "errorMessage", fmt.Sprintf("Could not connect to database: %s", err))
	} else {
		opts, _ := pg.ParseURL("postgres://hr_information_system:abcd1234@localhost:5434/hr_information_system?sslmode=disable")
		rootLogger.Info("DB-CONNECTION-ESTABLISHED", "user", opts.User, "host", opts.Addr, "database", opts.Database)
	}

	// A Translator maps tags to text templates (you must register these tags & templates yourself)
	// In the case of cardinals & ordinals, numerical parameters are also taken into account
	// Validation check parameters are then interpolated into these templates
	// By default, a Translator will only contain guiding rules that are based on the nature of its language
	// E.g. English Cardinals are only categorised into either "One" or "Other"
	universalTranslator := routes.NewUniversalTranslator()

	validate, err := routes.NewValidator(universalTranslator)
	if err != nil {
		rootLogger.Fatal("VALIDATOR-INSTANTIATION-FAILED", "errorMessage", fmt.Sprintf("Could not instantiate validator: %s", err))
	} else {
		rootLogger.Info("VALIDATOR-INSTANTIATED")
	}

	// TODO: create env file to set authentication (hashing/signing) & encryption keys
	sessionStore := memstore.NewMemStore(
		[]byte("authkey123"),
		[]byte("enckey12341234567890123456789012"),
	)
	rootLogger.Info("SESSION-STORE-CONNECTION-ESTABLISHED")

	opts, _ := pg.ParseURL("postgres://hr_information_system:abcd1234@localhost:5433/hr_information_system?sslmode=disable")
	db := pg.Connect(opts)

	a, err := pgadapter.NewAdapterByDB(db, pgadapter.SkipTableCreate())
	if err != nil {
		rootLogger.Fatal("AUTHORIZATION-ADAPTER-INSTANTIATION-FAILED", "errorMessage", fmt.Sprintf("Could not instantiate Authorization Adapter: %s", err))
	} else {
		rootLogger.Info("AUTHORIZATION-ADAPTER-INSTANTIATED", "user", opts.User, "host", opts.Addr, "database", opts.Database)
	}

	authEnforcer, err := casbin.NewEnforcer("auth_model.conf", a)
	if err != nil {
		rootLogger.Fatal("AUTHORIZATION-ENFORCER-INSTANTIATION-FAILED", "errorMessage", fmt.Sprintf("Could not instantiate Authorization Enforcer: %s", err))
	} else {
		rootLogger.Info("AUTHORIZATION-ENFORCER-INSTANTIATED")
	}

	if err := authEnforcer.LoadPolicy(); err != nil {
		rootLogger.Fatal("AUTHORIZATION-POLICY-LOAD-FAILED", "errorMessage", fmt.Sprintf("Could not load policy into Authorization Enforcer: %s", err))
	} else {
		rootLogger.Info("AUTHORIZATION-POLICY-LOADED")
	}
	authEnforcer.AddNamedMatchingFunc("g", "KeyMatch2", util.KeyMatch2)
	authEnforcer.AddNamedDomainMatchingFunc("g", "KeyMatch2", util.KeyMatch2)

	s.router = routes.NewRouter(storage, universalTranslator, validate, rootLogger, sessionStore, authEnforcer)
	s.logOutput = &logOutputMedium
	s.sessionStore = sessionStore

	//Clear any logs from the log output buffer
	s.logOutput.Reset()
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
	// Re-insert the root administrator user & privileges
	insertTenant := "INSERT INTO tenant (name) VALUES ($1)"
	insertDivision := "INSERT INTO division (tenant, name) VALUES ($1, $2)"
	insertDepartment := "INSERT INTO department (tenant, division, name) VALUES ($1, $2, $3)"
	insertUser := `
					INSERT INTO user_account (id, email, tenant, password, totp_secret_key) 
					VALUES ($1, $2, $3, $4, $5)
					`
	insertAppointment := `
	INSERT INTO appointment (title, tenant, division, department, user_account_id, start_date)
	VALUES ($1, $2, $3, $4, $5, $6)
	`

	insertPolicies := `INSERT INTO casbin_rule (Ptype, V0, V1, V2, V3) VALUES 
						('p', 'PUBLIC', '*', '/api/session', 'POST'),
						('p', 'PUBLIC', '*', '/api/session', 'DELETE'),
						('p', 'ROOT_ROLE_ADMIN', 'HRIS Enterprises', '/api/tenants/*', 'POST'),
						('p', 'ROOT_ROLE_ADMIN', 'HRIS Enterprises', '/api/tenants/*/divisions/*', 'POST'),
						('p', 'ROOT_ROLE_ADMIN', 'HRIS Enterprises', '/api/tenants/*/divisions/*/departments/*', 'POST'),
						('p', 'ROOT_ROLE_ADMIN', 'HRIS Enterprises', '/api/tenants/*/users/*', 'POST'),
						('p', 'ROOT_ROLE_ADMIN', 'HRIS Enterprises', '/api/tenants/*/users/*/appointments/*', 'POST');						
						`
	insertRoleAssignments := `INSERT INTO casbin_rule (Ptype, V0, V1, V2) VALUES 
								('g', '*', 'PUBLIC', '*'),
								('g', $1, 'ROOT_ROLE_ADMIN', 'HRIS Enterprises');
								`

	_, err := s.dbRootConn.Query(insertTenant, s.defaultUser.Tenant)
	if err != nil {
		log.Fatalf("DB seeding failed: %s", err)
	}

	_, err = s.dbRootConn.Query(insertDivision, s.defaultAppointment.Tenant, s.defaultAppointment.Division)
	if err != nil {
		log.Fatalf("Division seeding failed: %s", err)
	}

	_, err = s.dbRootConn.Query(insertDepartment, s.defaultAppointment.Tenant, s.defaultAppointment.Division, s.defaultAppointment.Department)
	if err != nil {
		log.Fatalf("Department seeding failed: %s", err)
	}

	_, err = s.dbRootConn.Query(insertUser, s.defaultUser.Id, s.defaultUser.Email, s.defaultUser.Tenant, s.defaultUser.Password, s.defaultUser.TotpSecretKey)
	if err != nil {
		log.Fatalf("DB seeding failed: %s", err)
	}

	_, err = s.dbRootConn.Query(insertAppointment, s.defaultAppointment.Title, s.defaultAppointment.Tenant, s.defaultAppointment.Division, s.defaultAppointment.Department, s.defaultAppointment.UserId, s.defaultAppointment.StartDate)
	if err != nil {
		log.Fatalf("Appointment seeding failed: %s", err)
	}

	_, err = s.dbRootConn.Query(insertPolicies)
	if err != nil {
		log.Fatalf("DB seeding failed: %s", err)
	}

	_, err = s.dbRootConn.Query(insertRoleAssignments, s.defaultUser.Id)
	if err != nil {
		log.Fatalf("DB seeding failed: %s", err)
	}
}

func (s *IntegrationTestSuite) TearDownTest() {
	// Clear all data
	query := fmt.Sprintf("TRUNCATE %s", strings.Join(s.dbTables, ", "))
	_, err := s.dbRootConn.Exec(query)
	if err != nil {
		log.Fatalf("Could not clear data from all tables: %s", err)
	}
	// Clear the log buffer
	s.logOutput.Reset()
}

const authSessionName = "authenticated" // TODO: make this an environment variable

func (s *IntegrationTestSuite) addSessionCookieToRequest(r *http.Request, userId string, tenant string, email string) {
	session, err := s.sessionStore.Get(r, authSessionName)
	if err != nil {
		log.Fatalf("Could not add cookie to the request: %s", err)
	}

	session.Values["id"] = userId
	session.Values["tenant"] = tenant
	session.Values["email"] = email

	w := httptest.NewRecorder()
	s.sessionStore.Save(r, w, session)

	cookieString := w.Header().Get("Set-Cookie")
	first, _, _ := strings.Cut(cookieString, ";")
	name, sessionId, _ := strings.Cut(first, "=")
	cookie := &http.Cookie{
		Name:  name,
		Value: sessionId,
	}

	r.AddCookie(cookie)
}

func (s *IntegrationTestSuite) expectNextLogToContain(reader *bufio.Reader, substrings ...string) {
	log, err := reader.ReadBytes('\n')
	s.Equal(nil, err)
	for _, substring := range substrings {
		s.Contains(string(log), substring)
	}
}

func (s *IntegrationTestSuite) expectSelectQueryToReturnNoRows(table string, conditions map[string]string) {
	// Convert the string slice to an any slice
	attributes := []string{}
	values := []any{}

	for attribute, value := range conditions {
		attributes = append(attributes, attribute)
		values = append(values, value)
	}

	query := postgres.NewDynamicConditionQuery(fmt.Sprintf("SELECT created_at FROM %s", table), attributes)
	rows, err := s.dbRootConn.Query(query, values...)
	if err != nil {
		log.Fatal(err)
	}

	count := 0
	for rows.Next() {
		count++
	}
	s.Equal(0, count, "No rows should be returned")
}

func (s *IntegrationTestSuite) expectSelectQueryToReturnOneRow(table string, conditions map[string]string) {
	// Convert the string slice to an any slice
	attributes := []string{}
	values := []any{}

	for attribute, value := range conditions {
		attributes = append(attributes, attribute)
		values = append(values, value)
	}

	query := postgres.NewDynamicConditionQuery(fmt.Sprintf("SELECT created_at FROM %s", table), attributes)
	rows, err := s.dbRootConn.Query(query, values...)
	if err != nil {
		log.Fatal(err)
	}

	count := 0
	for rows.Next() {
		count++
	}
	s.Equal(1, count, "1 row should be returned")
}

// Verifies the happy path works
func (s *IntegrationTestSuite) TestCreateTenant() {
	wantTenant := routes.Tenant{
		Name: "Macdonalds",
	}
	// Serve the request
	req, err := http.NewRequest("POST", fmt.Sprintf("/api/tenants/%s", wantTenant.Name), nil)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(req, s.defaultUser.Id, s.defaultUser.Tenant, s.defaultUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	// Check response
	res := w.Result()
	s.Equal(201, res.StatusCode, "Status should be 201")

	// Check database
	s.expectSelectQueryToReturnOneRow("tenant", map[string]string{"name": wantTenant.Name})

	// Check logs
	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"TENANT-CREATED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

// Verifies that the validation function is executed & validation errors are handled correctly
// (by triggering a validation error with invalid input)
func (s *IntegrationTestSuite) TestCreateTenantInvalidInput() {
	invalidTenant := routes.Tenant{
		Name: "   ",
	}

	// Create the request and add a session cookie to it
	r, err := http.NewRequest("POST", fmt.Sprintf("/api/tenants/%s", invalidTenant.Name), nil)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultUser.Id, s.defaultUser.Tenant, s.defaultUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	// Check the response status and body
	res := w.Result()
	s.Equal(400, res.StatusCode)

	var body errorResponseBody
	err = json.NewDecoder(res.Body).Decode(&body)
	s.Equal(nil, err, "Request body in wrong format")
	s.Equal("INPUT-VALIDATION-ERROR", body.Code)

	// Check the database
	s.expectSelectQueryToReturnNoRows("tenant", map[string]string{"name": invalidTenant.Name})

	// Check logs
	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"INPUT-VALIDATION-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

// Verifies that postgres errors are handled correctly (by triggering a postgres error)
func (s *IntegrationTestSuite) TestCreateTenantAlreadyExists() {
	existingTenant := routes.Tenant{
		Name: s.defaultUser.Tenant,
	}
	// Create the request and add a session cookie to it
	req, err := http.NewRequest("POST", fmt.Sprintf("/api/tenants/%s", existingTenant.Name), nil)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(req, s.defaultUser.Id, s.defaultUser.Tenant, s.defaultUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	// Check the response status and body
	res := w.Result()
	s.Equal(409, res.StatusCode, "409 error should be returned")

	var body errorResponseBody
	err = json.NewDecoder(res.Body).Decode(&body)
	s.Equal(nil, err, "Response body format should match error response struct")
	s.Equal("UNIQUE-VIOLATION-ERROR", body.Code)

	// Check the database
	s.expectSelectQueryToReturnOneRow("tenant", map[string]string{"name": existingTenant.Name})

	// Check logs
	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"UNIQUE-VIOLATION-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

// Verifies that the happy path works
func (s *IntegrationTestSuite) TestCreateDivision() {
	wantDivision := routes.Division{
		Tenant: s.defaultUser.Tenant,
		Name:   "Marketing",
	}

	// Create the request and add a session cookie to it
	req, err := http.NewRequest("POST", fmt.Sprintf("/api/tenants/%s/divisions/%s", s.defaultUser.Tenant, wantDivision.Name), nil)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(req, s.defaultUser.Id, s.defaultUser.Tenant, s.defaultUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	// Check the response status and body
	res := w.Result()
	s.Equal(201, res.StatusCode)

	// Check the database
	s.expectSelectQueryToReturnOneRow(
		"division",
		map[string]string{
			"tenant": wantDivision.Tenant,
			"name":   wantDivision.Name,
		},
	)

	// Check the logs
	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"DIVISION-CREATED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

// Verifies that the validation function is executed & validation errors are handled correctly
func (s *IntegrationTestSuite) TestCreateDivisionInvalidInput() {
	invalidDivision := routes.Division{
		Tenant: s.defaultUser.Tenant,
		Name:   "  ",
	}

	// Create the request and add a session cookie to it
	r, err := http.NewRequest("POST", fmt.Sprintf("/api/tenants/%s/divisions/%s", invalidDivision.Tenant, invalidDivision.Name), nil)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultUser.Id, s.defaultUser.Tenant, s.defaultUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	// Check the response status and body
	res := w.Result()
	s.Equal(400, res.StatusCode)

	var body errorResponseBody
	err = json.NewDecoder(res.Body).Decode(&body)
	s.Equal(nil, err, "Response body in wrong format")
	s.Equal("INPUT-VALIDATION-ERROR", body.Code)

	// Check the database
	s.expectSelectQueryToReturnNoRows(
		"division",
		map[string]string{
			"tenant": invalidDivision.Tenant,
			"name":   invalidDivision.Name,
		},
	)

	// Check the logs
	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"INPUT-VALIDATION-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

// Verifies that postgres errors are handled correctly
func (s *IntegrationTestSuite) TestCreateDivisionAlreadyExists() {
	existingDivision := routes.Division{
		Tenant: s.defaultUser.Tenant,
		Name:   s.defaultAppointment.Division,
	}

	// Create the request and add a session cookie to it
	req, err := http.NewRequest("POST", fmt.Sprintf("/api/tenants/%s/divisions/%s", s.defaultUser.Tenant, existingDivision.Name), nil)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(req, s.defaultUser.Id, s.defaultUser.Tenant, s.defaultUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	// Check the response status and body
	res := w.Result()
	s.Equal(409, res.StatusCode)

	var body errorResponseBody
	err = json.NewDecoder(res.Body).Decode(&body)
	s.Equal(nil, err, "Response body format did not match the error response body struct")
	s.Equal("UNIQUE-VIOLATION-ERROR", body.Code)

	// Check the database
	s.expectSelectQueryToReturnOneRow(
		"division",
		map[string]string{
			"tenant": existingDivision.Tenant,
			"name":   existingDivision.Name,
		},
	)

	// Check the logs
	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"UNIQUE-VIOLATION-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

// Verifies that the happy path works
func (s *IntegrationTestSuite) TestCreateDepartment() {
	wantDepartment := routes.Department{
		Tenant:   s.defaultUser.Tenant,
		Division: s.defaultAppointment.Division,
		Name:     "Customer Support",
	}

	// Create the request and add a session cookie to it
	r, err := http.NewRequest("POST", fmt.Sprintf("/api/tenants/%s/divisions/%s/departments/%s", wantDepartment.Tenant, wantDepartment.Division, wantDepartment.Name), nil)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultUser.Id, s.defaultUser.Tenant, s.defaultUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	// Check the response status and body
	res := w.Result()
	s.Equal(201, res.StatusCode)

	// Check the database
	s.expectSelectQueryToReturnOneRow(
		"department",
		map[string]string{
			"tenant":   wantDepartment.Tenant,
			"division": wantDepartment.Division,
			"name":     wantDepartment.Name,
		},
	)

	// Check the logs
	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"DEPARTMENT-CREATED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

// Verifies that the validation function is executed & validation errors are handled correctly
func (s *IntegrationTestSuite) TestCreateDepartmentInvalidInput() {
	invalidDepartment := routes.Department{
		Tenant:   s.defaultUser.Tenant,
		Division: s.defaultAppointment.Division,
		Name:     "   ",
	}

	// Create the request and add a session cookie to it
	r, err := http.NewRequest("POST", fmt.Sprintf("/api/tenants/%s/divisions/%s/departments/%s", invalidDepartment.Tenant, invalidDepartment.Division, invalidDepartment.Name), nil)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultUser.Id, s.defaultUser.Tenant, s.defaultUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	// Check the response status and body
	res := w.Result()
	s.Equal(400, res.StatusCode)

	var body errorResponseBody
	err = json.NewDecoder(res.Body).Decode(&body)
	s.Equal(nil, err, "Response in wrong format")
	s.Equal("INPUT-VALIDATION-ERROR", body.Code)

	// Check the database
	s.expectSelectQueryToReturnNoRows(
		"department",
		map[string]string{
			"tenant":   invalidDepartment.Tenant,
			"division": invalidDepartment.Division,
			"name":     invalidDepartment.Name,
		},
	)

	// Check the logs
	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"INPUT-VALIDATION-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

// Verifies that postgres errors are handled correctly
func (s *IntegrationTestSuite) TestCreateDepartmentAlreadyExists() {
	existingDepartment := routes.Department{
		Tenant:   s.defaultUser.Tenant,
		Division: s.defaultAppointment.Division,
		Name:     s.defaultAppointment.Department,
	}

	// Create the request and add a session cookie to it
	r, err := http.NewRequest("POST", fmt.Sprintf("/api/tenants/%s/divisions/%s/departments/%s", existingDepartment.Tenant, existingDepartment.Division, existingDepartment.Name), nil)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultUser.Id, s.defaultUser.Tenant, s.defaultUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	// Check the response status and body
	res := w.Result()
	s.Equal(409, res.StatusCode)

	// Check the database
	s.expectSelectQueryToReturnOneRow(
		"department",
		map[string]string{
			"tenant":   existingDepartment.Tenant,
			"division": existingDepartment.Division,
			"name":     existingDepartment.Name,
		},
	)

	// Check the logs
	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"UNIQUE-VIOLATION-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}
