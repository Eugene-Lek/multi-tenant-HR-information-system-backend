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
	defaultTenant      routes.Tenant
	defaultDivision    routes.Division
	defaultDepartment  routes.Department	
	defaultUser        routes.User
	defaultAppointment routes.Appointment
}

func TestAPIEndpointsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	suite.Run(t, &IntegrationTestSuite{
		defaultTenant: routes.Tenant{
			Id:   "2ad1dcfc-8867-49f7-87a3-8bd8d1154924",
			Name: "HRIS Enterprises",
		},
		defaultDivision: routes.Division{
			Id:       "f8b1551a-71bb-48c4-924a-8a25a6bff71d",
			TenantId: "2ad1dcfc-8867-49f7-87a3-8bd8d1154924",
			Name:     "Operations",
		},
		defaultDepartment: routes.Department{
			Id:         "9147b727-1955-437b-be7d-785e9a31f20c",
			TenantId: "2ad1dcfc-8867-49f7-87a3-8bd8d1154924",			
			DivisionId: "f8b1551a-71bb-48c4-924a-8a25a6bff71d",
			Name:       "Operations",
		},
		defaultUser: routes.User{
			Id:            "e7f31b70-ae26-42b3-b7a6-01ec68d5c33a",
			TenantId:      "2ad1dcfc-8867-49f7-87a3-8bd8d1154924",			
			Email:         "root-role-admin@hrisEnterprises.org",
			Password:      "$argon2id$v=19$m=65536,t=1,p=8$cFTNg+YXrN4U0lvwnamPkg$0RDBxH+EouVxDbBlQUNctdWZ+CNKrayPpzTJaWNq83U",
			TotpSecretKey: "OLDFXRMH35A3DU557UXITHYDK4SKLTXZ",
		},
		defaultAppointment: routes.Appointment{
			Id:           "e4edbd37-164d-478d-9625-5b1397ef6e45",
			TenantId:      "2ad1dcfc-8867-49f7-87a3-8bd8d1154924",			
			Title:        "System Administrator",
			DepartmentId: "9147b727-1955-437b-be7d-785e9a31f20c",
			UserId:       "e7f31b70-ae26-42b3-b7a6-01ec68d5c33a",
			StartDate:    "2024-02-01",
		},
	})
}

func attemptDBconnectionUntilTimeout(dbRootConnString string) (*sql.DB, error) {
	tick := time.Tick(500 * time.Millisecond)
	timeout := time.After(15 * time.Second)
	for {
		select {
		case <-timeout:
			return nil, errors.New("Attempt to connect to the Database timed out")
		case <-tick:
			conn, err := sql.Open("postgres", dbRootConnString)
			if err != nil && err.Error() != "pq: the database system is starting up" {
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
	insertTenant := "INSERT INTO tenant (id, name) VALUES ($1, $2)"
	insertDivision := "INSERT INTO division (id, tenant_id, name) VALUES ($1, $2, $3)"
	insertDepartment := "INSERT INTO department (id, tenant_id, division_id, name) VALUES ($1, $2, $3, $4)"
	insertUser := `
					INSERT INTO user_account (id, email, tenant_id, password, totp_secret_key) 
					VALUES ($1, $2, $3, $4, $5)					
					`
	insertAppointment := `
					INSERT INTO appointment (id, tenant_id, title, department_id, user_account_id, start_date)
					VALUES ($1, $2, $3, $4, $5, $6)
					`

	insertPolicies := `INSERT INTO casbin_rule (Ptype, V0, V1, V2, V3) VALUES 
						('p', 'PUBLIC', '*', '/api/session', 'POST'),
						('p', 'PUBLIC', '*', '/api/session', 'DELETE'),
						('p', 'ROOT_ROLE_ADMIN', $1, '/api/tenants/*', 'POST'),
						('p', 'ROOT_ROLE_ADMIN', $1, '/api/tenants/*/divisions/*', 'POST'),
						('p', 'ROOT_ROLE_ADMIN', $1, '/api/tenants/*/divisions/*/departments/*', 'POST'),
						('p', 'ROOT_ROLE_ADMIN', $1, '/api/tenants/*/users/*', 'POST'),
						('p', 'ROOT_ROLE_ADMIN', $1, '/api/tenants/*/users/*/appointments/*', 'POST');						
						`
	insertRoleAssignments := `INSERT INTO casbin_rule (Ptype, V0, V1, V2) VALUES 
								('g', '*', 'PUBLIC', '*'),
								('g', $1, 'ROOT_ROLE_ADMIN', $2);
								`

	_, err := s.dbRootConn.Exec(insertTenant, s.defaultTenant.Id, s.defaultTenant.Name)
	if err != nil {
		log.Fatalf("Tenant seeding failed: %s", err)
	}

	_, err = s.dbRootConn.Exec(insertDivision, s.defaultDivision.Id, s.defaultDivision.TenantId, s.defaultDivision.Name)
	if err != nil {
		log.Fatalf("Division seeding failed: %s", err)
	}

	_, err = s.dbRootConn.Exec(insertDepartment, s.defaultDepartment.Id, s.defaultDepartment.TenantId, s.defaultDepartment.DivisionId, s.defaultDepartment.Name)
	if err != nil {
		log.Fatalf("Department seeding failed: %s", err)
	}

	_, err = s.dbRootConn.Exec(insertUser, s.defaultUser.Id, s.defaultUser.Email, s.defaultUser.TenantId, s.defaultUser.Password, s.defaultUser.TotpSecretKey)
	if err != nil {
		log.Fatalf("User seeding failed: %s", err)
	}

	_, err = s.dbRootConn.Exec(insertAppointment, s.defaultAppointment.Id, s.defaultAppointment.Id, s.defaultAppointment.Title, s.defaultAppointment.DepartmentId, s.defaultAppointment.UserId, s.defaultAppointment.StartDate)
	if err != nil {
		log.Fatalf("Appointment seeding failed: %s", err)
	}

	_, err = s.dbRootConn.Exec(insertPolicies, s.defaultUser.TenantId)
	if err != nil {
		log.Fatalf("DB seeding failed: %s", err)
	}

	_, err = s.dbRootConn.Exec(insertRoleAssignments, s.defaultUser.Id, s.defaultUser.TenantId)
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

func (s *IntegrationTestSuite) addSessionCookieToRequest(r *http.Request, userId string, tenantId string, email string) {
	session, err := s.sessionStore.Get(r, authSessionName)
	if err != nil {
		log.Fatalf("Could not add cookie to the request: %s", err)
	}

	session.Values["id"] = userId
	session.Values["tenantId"] = tenantId
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
		Id:   "5338d729-32bd-4ad2-a8d1-22cbf81113de",
		Name: "Macdonalds",
	}

	// Create & serve the request
	type requestBody struct {
		Name string
	}	
	body := requestBody{
		Name: wantTenant.Name,
	}

	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(body)

	req, err := http.NewRequest("POST", fmt.Sprintf("/api/tenants/%s", wantTenant.Id), bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(req, s.defaultUser.Id, s.defaultUser.TenantId, s.defaultUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	// Check response
	res := w.Result()
	s.Equal(201, res.StatusCode, "Status should be 201")

	// Check database
	s.expectSelectQueryToReturnOneRow("tenant", map[string]string{"id": wantTenant.Id})

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
		Id:   "5338d729-32bd-4ad2-a8d1-22cbf81113de",		
		Name: "   ",
	}

	// Create the request and add a session cookie to it
	type requestBody struct {
		Name string
	}	
	reqBody := requestBody{
		Name: invalidTenant.Name,
	}

	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)

	r, err := http.NewRequest("POST", fmt.Sprintf("/api/tenants/%s", invalidTenant.Id), bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultUser.Id, s.defaultUser.TenantId, s.defaultUser.Email)

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
	s.expectSelectQueryToReturnNoRows("tenant", map[string]string{"id": invalidTenant.Id})

	// Check logs
	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"INPUT-VALIDATION-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

// Verifies that postgres errors are handled correctly (by triggering a postgres error)
func (s *IntegrationTestSuite) TestCreateTenantAlreadyExists() {
	existingTenant := routes.Tenant{
		Id: s.defaultTenant.Id,		
		Name: s.defaultTenant.Name,
	}
	// Create the request and add a session cookie to it
	type requestBody struct {
		Name string
	}	
	reqBody := requestBody{
		Name: existingTenant.Name,
	}
	
	bodyBuf := new(bytes.Buffer)
	err := json.NewEncoder(bodyBuf).Encode(reqBody)	
	req, err := http.NewRequest("POST", fmt.Sprintf("/api/tenants/%s", existingTenant.Id), bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(req, s.defaultUser.Id, s.defaultUser.TenantId, s.defaultUser.Email)

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
	s.expectSelectQueryToReturnOneRow("tenant", map[string]string{"id": existingTenant.Id})

	// Check logs
	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"UNIQUE-VIOLATION-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

// Verifies that the happy path works
func (s *IntegrationTestSuite) TestCreateDivision() {
	wantDivision := routes.Division{
		Id: "f0935407-d43a-47f6-8fdd-bc45ab9c43d9",
		TenantId: s.defaultDivision.TenantId,
		Name:   "Marketing",
	}

	// Create the request and add a session cookie to it
	type requestBody struct {
		Name string
	}	
	reqBody := requestBody{
		Name: wantDivision.Name,
	}
	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)

	req, err := http.NewRequest("POST", fmt.Sprintf("/api/tenants/%s/divisions/%s", wantDivision.TenantId, wantDivision.Id), bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(req, s.defaultUser.Id, s.defaultUser.TenantId, s.defaultUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	// Check the response status and body
	res := w.Result()
	s.Equal(201, res.StatusCode)

	// Check the database
	s.expectSelectQueryToReturnOneRow(
		"division",
		map[string]string{"id": wantDivision.Id},
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
		Id: "f0935407-d43a-47f6-8fdd-bc45ab9c43d9",		
		TenantId: s.defaultDivision.TenantId,
		Name:   "  ",
	}

	// Create the request and add a session cookie to it
	type requestBody struct {
		Name string
	}	
	reqBody := requestBody{
		Name: invalidDivision.Name,
	}
	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)

	r, err := http.NewRequest("POST", fmt.Sprintf("/api/tenants/%s/divisions/%s", invalidDivision.TenantId, invalidDivision.Id), bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultUser.Id, s.defaultUser.TenantId, s.defaultUser.Email)

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
		map[string]string{"id": invalidDivision.Id},
	)

	// Check the logs
	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"INPUT-VALIDATION-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

// Verifies that postgres errors are handled correctly
func (s *IntegrationTestSuite) TestCreateDivisionAlreadyExists() {
	// Create the request and add a session cookie to it
	type requestBody struct {
		Name string
	}	
	reqBody := requestBody{
		Name: s.defaultDivision.Name,
	}
	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)

	req, err := http.NewRequest("POST", fmt.Sprintf("/api/tenants/%s/divisions/%s", s.defaultDivision.TenantId, s.defaultDivision.Id), bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(req, s.defaultUser.Id, s.defaultUser.TenantId, s.defaultUser.Email)

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
		map[string]string{"id": s.defaultDivision.Id},
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
		Id: "444aa127-b21b-45cf-b779-eb1c1ef82478",
		TenantId: s.defaultTenant.Id,
		DivisionId: s.defaultDivision.Id,
		Name:     "Customer Support",
	}

	// Create the request and add a session cookie to it
	type requestBody struct {
		Name string
	}	
	reqBody := requestBody{
		Name: wantDepartment.Name,
	}
	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)

	r, err := http.NewRequest("POST", fmt.Sprintf("/api/tenants/%s/divisions/%s/departments/%s", s.defaultUser.TenantId, wantDepartment.DivisionId, wantDepartment.Id), bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultUser.Id, s.defaultUser.TenantId, s.defaultUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	// Check the response status and body
	res := w.Result()
	s.Equal(201, res.StatusCode)

	// Check the database
	s.expectSelectQueryToReturnOneRow(
		"department",
		map[string]string{"id": wantDepartment.Id},
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
		Id: "444aa127-b21b-45cf-b779-eb1c1ef82478",
		DivisionId: s.defaultDivision.Id,
		Name:     "   ",
	}

	// Create the request and add a session cookie to it
	type requestBody struct {
		Name string
	}	
	reqBody := requestBody{
		Name: invalidDepartment.Name,
	}
	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)

	r, err := http.NewRequest("POST", fmt.Sprintf("/api/tenants/%s/divisions/%s/departments/%s", s.defaultUser.TenantId, invalidDepartment.DivisionId, invalidDepartment.Id), bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultUser.Id, s.defaultUser.TenantId, s.defaultUser.Email)

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
		map[string]string{"id": invalidDepartment.Id},
	)

	// Check the logs
	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"INPUT-VALIDATION-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

// Verifies that postgres errors are handled correctly
func (s *IntegrationTestSuite) TestCreateDepartmentAlreadyExists() {
	// Create the request and add a session cookie to it
	type requestBody struct {
		Name string
	}	
	reqBody := requestBody{
		Name: s.defaultDepartment.Name,
	}
	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)
	
	r, err := http.NewRequest("POST", fmt.Sprintf("/api/tenants/%s/divisions/%s/departments/%s", s.defaultUser.TenantId, s.defaultDepartment.DivisionId, s.defaultDepartment.Id), bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultUser.Id, s.defaultUser.TenantId, s.defaultUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	// Check the response status and body
	res := w.Result()
	s.Equal(409, res.StatusCode)

	// Check the database
	s.expectSelectQueryToReturnOneRow(
		"department",
		map[string]string{"id": s.defaultDepartment.Id},
	)

	// Check the logs
	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"UNIQUE-VIOLATION-ERROR"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestCreateAppointment() {
	wantAppointment := routes.Appointment{
		Id: "cfc61cce-3d5a-4014-8490-3302ddd187b8",
		TenantId: s.defaultTenant.Id,		
		UserId:     s.defaultUser.Id,			
		Title:      "Manager",	
		DepartmentId: s.defaultDepartment.Id,
		StartDate:  "2024-02-01",
	}

	type requestBody struct {
		Title string
		DepartmentId string
		StartDate string
		EndDate string
	}
	reqBody := requestBody{
		Title: wantAppointment.Id,
		DepartmentId: wantAppointment.DepartmentId,
		StartDate: wantAppointment.StartDate,
		EndDate: wantAppointment.EndDate,		
	}
	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)

	resource := fmt.Sprintf("/api/tenants/%s/users/%s/appointments/%s", wantAppointment.TenantId, wantAppointment.UserId, wantAppointment.Id)
	r, err := http.NewRequest("POST", resource, bodyBuf )
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultUser.Id, s.defaultTenant.Id, s.defaultUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	res := w.Result()
	s.Equal(201, res.StatusCode)

	s.expectSelectQueryToReturnOneRow("appointment", map[string]string{"id": wantAppointment.Id})

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"APPOINTMENT-CREATED"`)	
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)		
}
