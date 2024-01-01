package main

import (
	"log"
	"net/http"
	"os"

	"github.com/quasoft/memstore"

	"multi-tenant-HR-information-system-backend/postgres"
	"multi-tenant-HR-information-system-backend/routes"
)

func main() {

	listenAddress := "localhost:3000"
	connString := "host=localhost port=5433 user=hr_information_system password=abcd1234 dbname=hr_information_system sslmode=disable"

	postgresStorage, err := postgres.NewPostgresStorage(connString)
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

	router := routes.NewRouter(postgresStorage, universalTranslator, validate, rootLogger, sessionStore)

	log.Println("API server running on port: ", listenAddress)
	http.ListenAndServe(listenAddress, router)
}
