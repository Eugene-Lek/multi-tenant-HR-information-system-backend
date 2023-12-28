package routes

import (
	"net/http"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type Storage interface {
	CreateTenant(tenant Tenant) error
	CreateDivision(division Division) error
	CreateDepartment(department Department) error
	CreateUser(user User) error
	CreateAppointment(appointment Appointment) error
}

type Router struct {
	router              *mux.Router
	storage             Storage
	universalTranslator *ut.UniversalTranslator
	validate            *validator.Validate
	rootLogger          *tailoredLogger
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.router.ServeHTTP(w, req)
}

func NewRouter(storage Storage, universalTranslator *ut.UniversalTranslator, validate *validator.Validate, rootLogger *tailoredLogger) *Router {
	router := mux.NewRouter()

	r := &Router{
		router:              router,
		storage:             storage,
		universalTranslator: universalTranslator,
		validate:            validate,
		rootLogger:          rootLogger,
	}

	// Logging middleware wraps around error handling middleware because an error in logging has zero impact on the user
	r.router.Use(newRequestLogger(r.rootLogger))
	r.router.Use(logRequestCompletion)
	r.router.Use(errorHandling)
	r.router.Use(getAcceptedLanguage)

	apiRouter := r.router.PathPrefix("/api").Subrouter()

	tenantRouter := apiRouter.PathPrefix("/tenants/{tenant}").Subrouter()
	tenantRouter.HandleFunc("", r.handleCreateTenant).Methods("POST")
	tenantRouter.HandleFunc("/divisions/{division}", r.handleCreateDivision).Methods("POST")
	tenantRouter.HandleFunc("/divisions/{division}/departments/{department}", r.handleCreateDepartment).Methods("POST")

	userRouter := tenantRouter.PathPrefix("/users").Subrouter()
	userRouter.HandleFunc("/{user-id}", r.handleCreateUser).Methods("POST")
	userRouter.HandleFunc("/{user-id}/appointments/{id}", r.handleCreateAppointment).Methods("POST")

	//jobRequisitionRouter := tenantRouter.PathPrefix("/job-requisition").Subrouter()
	//jobRequisitionRouter.HandleFunc("", )

	return r
}
