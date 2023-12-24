package routes

import (
	"encoding/json"
	"net/http"
	"log"

	"github.com/gorilla/mux"
)


type User struct {
	Id string
	Email string
	Tenant string
	Division string
	Department string
	Password string
	TotpSecretKey string
	CreatedAt string
	UpdatedAt string
	LastLogin string
}

type Appointment struct {
	Title string
	Tenant string
	Division string
	Department string
	UserId string
	StartDate string
	EndDate string
	CreatedAt string
	UpdatedAt string
}

func (router *Router) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Division string 
		Department string
	}

	var body requestBody
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		log.Fatal(err.Error())
	}

	vars := mux.Vars(r)

	// Create default password + totpsecretkey

	user := User{
		Email: vars["email"],
		Tenant: vars["tenant"],
		Division: body.Division,	
		Department: body.Department,
	}

	//TODO: Input validation
	translator, httpErr := getAppropriateTranslator(r, router.universalTranslator)
	if httpErr != nil {
		sendToErrorHandlingMiddleware(httpErr, r)
		log.Print(httpErr.Error())

		return
	}	
	
	httpErr = validateStruct(router.validate, translator, user)
	if httpErr != nil {
		sendToErrorHandlingMiddleware(httpErr, r)
		log.Print(httpErr.Error())

		return
	}	

	httpErr = router.storage.CreateUser(user)
	if httpErr != nil {
		sendToErrorHandlingMiddleware(httpErr, r)
		// Logging

		return	
	}

	w.WriteHeader(http.StatusCreated)
}


func (router *Router) handleCreateAppointment(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Title string
		Division string
		Department string
		StartDate string		
		EndDate string
	}

	var body requestBody

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		log.Fatal(err.Error())
	}

	vars := mux.Vars(r)

	userAppointment := Appointment{
		Title: body.Title,
		Tenant: vars["tenant"],
		Division: body.Division,	
		Department: body.Department,
		UserId: vars["user-id"],
		StartDate: body.StartDate,		
		EndDate: body.EndDate,
	}	

	//TODO Input validation
	translator, httpErr := getAppropriateTranslator(r, router.universalTranslator)
	if httpErr != nil {
		sendToErrorHandlingMiddleware(httpErr, r)
		log.Print(httpErr.Error())

		return
	}	
	
	httpErr = validateStruct(router.validate, translator, userAppointment)
	if httpErr != nil {
		sendToErrorHandlingMiddleware(httpErr, r)
		log.Print(httpErr.Error())

		return
	}		

	httpErr = router.storage.CreateAppointment(userAppointment)
	if httpErr != nil {
		sendToErrorHandlingMiddleware(httpErr, r)
		// Logging

		return
	}

	w.WriteHeader(http.StatusCreated)
}