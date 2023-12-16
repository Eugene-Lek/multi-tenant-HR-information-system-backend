package routes

import (
	"net/http"
	
	"github.com/gorilla/mux"
	"github.com/go-playground/validator/v10"
)

type Storage interface {
	CreateTenant(tenant Tenant) error
	CreateDivision(division Division) error
	CreateDepartment(department Department) error
	CreateUser(user User) error
	CreateAppointment(appointment Appointment) error
}

type Router struct {
	router *mux.Router
	storage Storage
	validate *validator.Validate	
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.router.ServeHTTP(w, req)
}

func NewRouter(storage Storage, validate *validator.Validate) *Router {
	router := mux.NewRouter()

	r := &Router{
		router: router,
		storage: storage,
		validate: validate,
	}

	tenantRouter := r.router.PathPrefix("/api/{tenant}").Subrouter()
	tenantRouter.HandleFunc("", r.handleCreateTenant).Methods("POST")
	tenantRouter.HandleFunc("/{division}", r.handleCreateDivision).Methods("POST")
	tenantRouter.HandleFunc("/{division}/{department}", r.handleCreateDepartment).Methods("POST")

	userRouter := tenantRouter.PathPrefix("/users").Subrouter()
	userRouter.HandleFunc("/{email}", r.handleCreateUser)
	userRouter.HandleFunc("/{user-id}/appointments/{id}", r.handleCreateAppointment)

	//jobRequisitionRouter := tenantRouter.PathPrefix("/job-requisition").Subrouter()
	//jobRequisitionRouter.HandleFunc("", )	

	return r
}

