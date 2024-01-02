package main

import (
	"log"
	"net/http"
	"os"

	"github.com/quasoft/memstore"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/util"	
	pgadapter "github.com/casbin/casbin-pg-adapter"	
	"github.com/go-pg/pg/v10"

	"multi-tenant-HR-information-system-backend/postgres"
	"multi-tenant-HR-information-system-backend/routes"
)

func main() {

	listenAddress := "localhost:3000"
	connString := "host=localhost port=5433 user=hr_information_system password=abcd1234 dbname=hr_information_system sslmode=disable"

	storage, err := postgres.NewPostgresStorage(connString)
	if err != nil {
		log.Fatalf("Could not connect to database: %s", err)
	}

	// A Translator maps tags to text templates (you must register these tags & templates yourself)
	// In the case of cardinals & ordinals, numerical parameters are also taken into account
	// Validation check parameters are then interpolated into these templates
	// By default, a Translator will only contain guiding rules that are based on the nature of its language
	// E.g. English Cardinals are only categorised into either "One" or "Other"
	universalTranslator := routes.NewUniversalTranslator()

	validate, err := routes.NewValidator(universalTranslator)
	if err != nil {
		log.Fatalf("Could not instantiate the validator: %s", err)
	}

	logOutputMedium := os.Stdout
	rootLogger := routes.NewRootLogger(logOutputMedium)

	// TODO: create env file to set authentication (hashing/signing) & encryption keys
	sessionStore := memstore.NewMemStore(
		[]byte("authkey123"),
		[]byte("enckey12341234567890123456789012"),
	)

	opts := &pg.Options{
		Addr:     ":5433",
		User:     "hr_information_system",
		Password: "abcd1234",
		Database: "hr_information_system",
	}

	db := pg.Connect(opts)
	defer db.Close()

	a, err := pgadapter.NewAdapterByDB(db, pgadapter.SkipTableCreate())
	if err != nil {
		log.Fatalf("Could not instantiate Authorization Adapter: %s", err)
	}
	authEnforcer, err := casbin.NewEnforcer("auth_model.conf", a)
	if err != nil {
		log.Fatalf("Could not instantiate Authorization Enforcer: %s", err)
	}	
	if err := authEnforcer.LoadPolicy(); err != nil {
		log.Fatalf("Could not load policy into Authorization Enforcer: %s", err)
	}
	authEnforcer.AddNamedMatchingFunc("g", "KeyMatch2", util.KeyMatch2)
	authEnforcer.AddNamedDomainMatchingFunc("g", "KeyMatch2", util.KeyMatch2)

	router := routes.NewRouter(storage, universalTranslator, validate, rootLogger, sessionStore, authEnforcer)

	log.Println("API server running on port: ", listenAddress)
	http.ListenAndServe(listenAddress, router)
}
