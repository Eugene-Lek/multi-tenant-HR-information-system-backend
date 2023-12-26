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
	router *mux.Router
	storage Storage
	universalTranslator *ut.UniversalTranslator
	validate *validator.Validate	
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.router.ServeHTTP(w, req)
}

func NewRouter(storage Storage, universalTranslator *ut.UniversalTranslator, validate *validator.Validate) *Router {
	router := mux.NewRouter()

	r := &Router{
		router: router,
		storage: storage,
		universalTranslator: universalTranslator,
		validate: validate,
	}

	r.router.Use(errorHandlingMiddleware)
	r.router.Use(getLanguageMiddleware)

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

// Validates a struct instance, translates the errors to error messages and returns an error that collates all the error messages
func validateStruct(validate *validator.Validate, translator ut.Translator, s interface{}) error {
	err := validate.Struct(s)
	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		errorMessages := validationErrors.Translate(translator)
		return NewInputValidationError(errorMessages)		
	}

	return nil
}