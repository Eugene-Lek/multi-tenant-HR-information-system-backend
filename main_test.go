package main

import (
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

type errorResponseBody struct {
	Code    string
	Message string
}

type IntegrationTestSuite struct {
	suite.Suite
	router *routes.Router
	sessionStore sessions.Store
	dbRootConn *sql.DB
	dbTables   []string
	rootUser routes.User
}

func TestAPIEndpointsIntegration(t *testing.T) {
	suite.Run(t, &IntegrationTestSuite{
		rootUser: routes.User{
			Id: "e7f31b70-ae26-42b3-b7a6-01ec68d5c33a", 
			Email: "root-role-admin@hrisEnterprises.org", 
			Tenant: "HRIS Enterprises",
			Password: "$argon2id$v=19$m=65536,t=1,p=8$cFTNg+YXrN4U0lvwnamPkg$0RDBxH+EouVxDbBlQUNctdWZ+CNKrayPpzTJaWNq83U", 
			TotpSecretKey: "OLDFXRMH35A3DU557UXITHYDK4SKLTXZ",
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
	logOutputMedium := os.Stdout
	rootLogger := routes.NewRootLogger(logOutputMedium)

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
	s.sessionStore = sessionStore
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
	insertUser := `
					INSERT INTO user_account (id, email, tenant, password, totp_secret_key) 
					VALUES ($1, $2, $3, $4, $5)
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

	_, err := s.dbRootConn.Query(insertTenant, s.rootUser.Tenant)
	if err != nil {
		log.Fatalf("DB seeding failed: %s", err)
	}

	_, err = s.dbRootConn.Query(insertUser, s.rootUser.Id, s.rootUser.Email, s.rootUser.Tenant, s.rootUser.Password, s.rootUser.TotpSecretKey)
	if err != nil {
		log.Fatalf("DB seeding failed: %s", err)
	}

	_, err = s.dbRootConn.Query(insertPolicies)
	if err != nil {
		log.Fatalf("DB seeding failed: %s", err)
	}

	_, err = s.dbRootConn.Query(insertRoleAssignments, s.rootUser.Id)
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
}

const authSessionName = "authenticated" // TODO: make this an environment variable

func (s *IntegrationTestSuite) addCookie(r *http.Request, userId string, tenant string, email string) {
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
		Name: name,
		Value: sessionId,
	}

	r.AddCookie(cookie)
}

func (s *IntegrationTestSuite) TestCreateTenant() {
	wantTenant := routes.Tenant{
		Name: "Macdonalds",
	}
	// Serve the request
	req, err := http.NewRequest("POST", fmt.Sprintf("/api/tenants/%s",wantTenant.Name), nil)
	if err != nil {
		log.Fatal(err)
	}
	s.addCookie(req, s.rootUser.Id, s.rootUser.Tenant, s.rootUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	// Check response
	res := w.Result()	
	s.Equal(201, res.StatusCode, "Status should be 201")

	// Check database
	query := "SELECT * FROM tenant WHERE name = $1"

	var tenant routes.Tenant
	err = s.dbRootConn.QueryRow(query, wantTenant.Name).Scan(&tenant.Name, &tenant.CreatedAt, &tenant.UpdatedAt)
	s.Equal(nil, err, "No error should be thrown")
	s.Equal(wantTenant.Name, tenant.Name, fmt.Sprintf(`Tenant should be "%s"`, s.rootUser.Tenant))
}

// Verifies that the validation function is executed. No need to test various scenarios as it's been covered by the unit tests
func (s *IntegrationTestSuite) TestCreateTenantInvalidInput() {
	invalidTenant := routes.Tenant{
		Name: "   ",
	}

	r, err := http.NewRequest("POST", fmt.Sprintf("/api/tenants/%s", invalidTenant.Name), nil)
	if err != nil {
		log.Fatal(err)
	}
	s.addCookie(r, s.rootUser.Id, s.rootUser.Tenant, s.rootUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	res := w.Result()
	s.Equal(400, res.StatusCode)

	var body errorResponseBody
	err = json.NewDecoder(res.Body).Decode(&body)
	s.Equal(nil, err, "Request body in wrong format")
	s.Equal("INPUT-VALIDATION-ERROR", body.Code)
}

func (s *IntegrationTestSuite) TestCreateTenantAlreadyExists() {
	req, err := http.NewRequest("POST", fmt.Sprintf("/api/tenants/%s", s.rootUser.Tenant), nil)
	if err != nil {
		log.Fatal(err)
	}
	s.addCookie(req, s.rootUser.Id, s.rootUser.Tenant, s.rootUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	res := w.Result()
	s.Equal(409, res.StatusCode, "409 error should be returned")

	var body errorResponseBody
	err = json.NewDecoder(res.Body).Decode(&body)
	s.Equal(nil, err, "Response body format should match error response struct")
	s.Equal("UNIQUE-VIOLATION-ERROR", body.Code)
}

func (s *IntegrationTestSuite) TestCreateDivision() {
	wantDivision := routes.Division{
		Tenant: s.rootUser.Tenant,
		Name: "Marketing",
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("/api/tenants/%s/divisions/%s", s.rootUser.Tenant, wantDivision.Name), nil)
	if err != nil {
		log.Fatal(err)
	}
	s.addCookie(req, s.rootUser.Id, s.rootUser.Tenant, s.rootUser.Email)	

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	res := w.Result()
	s.Equal(201, res.StatusCode)

	query := "SELECT * FROM division WHERE name = $1 AND tenant = $2"

	var division routes.Division
	err = s.dbRootConn.QueryRow(query, wantDivision.Name, wantDivision.Tenant).Scan(&division.Name, &division.Tenant, &division.CreatedAt, &division.UpdatedAt)
	s.Equal(nil, err, "No error should be thrown")
	s.Equal(wantDivision.Name, division.Name)	
	s.Equal(wantDivision.Tenant, division.Tenant)	
}

func (s *IntegrationTestSuite) TestCreateDivisionAlreadyCreated() {
	existingDivision := routes.Division{
		Tenant: s.rootUser.Tenant,
		Name: "Marketing",
	}

	_, err := s.dbRootConn.Exec("INSERT INTO division (name, tenant) VALUES ($1, $2)", existingDivision.Name, existingDivision.Tenant)
	if err != nil {
		log.Fatal(err)
	}	

	req, err := http.NewRequest("POST", fmt.Sprintf("/api/tenants/%s/divisions/%s", s.rootUser.Tenant, existingDivision.Name), nil)
	if err != nil {
		log.Fatal(err)
	}
	s.addCookie(req, s.rootUser.Id, s.rootUser.Tenant, s.rootUser.Email)	

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	res := w.Result()
	s.Equal(409, res.StatusCode)	

	var body errorResponseBody
	err = json.NewDecoder(res.Body).Decode(&body)
	s.Equal(nil, err, "Response body format did not match the error response body struct")
	s.Equal("UNIQUE-VIOLATION-ERROR", body.Code)	
}

func (s *IntegrationTestSuite) TestCreateDivisionInvalidTenant() {
	desiredDivision := routes.Division{
		Tenant: "Non-Existent",
		Name: "Marketing",
	}

	r, err := http.NewRequest("POST", fmt.Sprintf("/api/tenants/%s/divisions/%s", desiredDivision.Tenant, desiredDivision.Name), nil)
	if err != nil {
		log.Fatal(err)
	}
	s.addCookie(r, s.rootUser.Id, s.rootUser.Tenant, s.rootUser.Email)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	res := w.Result()
	s.Equal(400, res.StatusCode)

	var body errorResponseBody
	err = json.NewDecoder(res.Body).Decode(&body)
	s.Equal(nil, err, "Response body format did not match the error response body struct")
	s.Equal("INVALID-FOREIGN-KEY-ERROR", body.Code)
}
