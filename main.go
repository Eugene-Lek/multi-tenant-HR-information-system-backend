package main

import (
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"

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

	validate := validator.New(validator.WithRequiredStructEnabled())

	router := routes.NewRouter(postgresStorage, validate)

	log.Println("API server running on port: ", listenAddress)
	http.ListenAndServe(listenAddress, router)
}
