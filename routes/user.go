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

	// DB query

	err = router.storage.CreateUser(user)
	if err != nil {
		// Add error address to request context	
	}

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

	err = router.storage.CreateAppointment(userAppointment)
	if err != nil {
		// Add error address to request context	
	}
}