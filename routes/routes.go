package routes

import (
	"errors"
	"net/http"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

type Storage interface {
	CreateTenant(tenant Tenant) error
	CreateDivision(division Division) error
	CreateDepartment(department Department) error
	CreateUser(user User) error
	GetUsers(userFilter User) ([]User, error)
	CreateAppointment(appointment Appointment) error
}

// A wrapper for the Router that adds its dependencies as properties/fields. This way, they can be accessed by route handlers
type Router struct {
	*mux.Router
	storage             Storage
	universalTranslator *ut.UniversalTranslator
	validate            *validator.Validate
	rootLogger          *tailoredLogger
	sessionStore        sessions.Store
}

func NewRouter(storage Storage, universalTranslator *ut.UniversalTranslator, validate *validator.Validate, rootLogger *tailoredLogger, sessionStore sessions.Store) *Router {
	r := mux.NewRouter()

	router := &Router{
		Router:              r,
		storage:             storage,
		universalTranslator: universalTranslator,
		validate:            validate,
		rootLogger:          rootLogger,
		sessionStore:        sessionStore,
	}

	// Logging middleware wraps around error handling middleware because an error in logging has zero impact on the user
	router.Use(newRequestLogger(router.rootLogger))
	router.Use(logRequestCompletion)
	router.Use(errorHandling)
	router.Use(getAcceptedLanguage)
	router.Use(authenticateUser(router.sessionStore))

	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.HandleFunc("/session", router.handleLogin).Methods("POST")
	apiRouter.HandleFunc("/session", router.handleLogout).Methods("DELETE")

	tenantRouter := apiRouter.PathPrefix("/tenants/{tenant}").Subrouter()
	tenantRouter.HandleFunc("", router.handleCreateTenant).Methods("POST")
	tenantRouter.HandleFunc("/divisions/{division}", router.handleCreateDivision).Methods("POST")
	tenantRouter.HandleFunc("/divisions/{division}/departments/{department}", router.handleCreateDepartment).Methods("POST")

	userRouter := tenantRouter.PathPrefix("/users").Subrouter()
	userRouter.HandleFunc("/{user-id}", router.handleCreateUser).Methods("POST")
	userRouter.HandleFunc("/{user-id}/appointments/{id}", router.handleCreateAppointment).Methods("POST")

	//jobRequisitionRouter := tenantRouter.PathPrefix("/job-requisition").Subrouter()
	//jobRequisitionRouter.HandleFunc("", )

	return router
}

// Fetches the locale from the Request Context & uses that to fetch the desired translator
func getAppropriateTranslator(r *http.Request, universalTranslator *ut.UniversalTranslator) (ut.Translator, error) {
	language, ok := r.Context().Value(languageKey).(string)
	if !ok {
		return nil, NewInternalServerError(errors.New("could not obtain preferred language"))
	}
	translator, _ := universalTranslator.GetTranslator(language)

	return translator, nil
}
