package routes

import (
	"errors"
	"net/http"

	"github.com/casbin/casbin/v2"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"multi-tenant-HR-information-system-backend/httperror"
	"multi-tenant-HR-information-system-backend/storage"	
)



// A wrapper for the Router that adds its dependencies as properties/fields. This way, they can be accessed by route handlers
type Router struct {
	*mux.Router
	storage             storage.Storage
	universalTranslator *ut.UniversalTranslator
	validate            *validator.Validate
	rootLogger          *tailoredLogger
	sessionStore        sessions.Store
	authEnforcer        casbin.IEnforcer
}

func NewRouter(storage storage.Storage, universalTranslator *ut.UniversalTranslator, validate *validator.Validate, rootLogger *tailoredLogger, sessionStore sessions.Store, authEnforcer casbin.IEnforcer) *Router {
	r := mux.NewRouter()

	router := &Router{
		Router:              r,
		storage:             storage,
		universalTranslator: universalTranslator,
		validate:            validate,
		rootLogger:          rootLogger,
		sessionStore:        sessionStore,
		authEnforcer:        authEnforcer,
	}

	// Logging middleware wraps around error handling middleware because an error in logging has zero impact on the user
	router.Use(newRequestLogger(router.rootLogger))
	router.Use(logRequestCompletion)
	router.Use(errorHandling)
	router.Use(getAcceptedLanguage)
	router.Use(authenticateUser(router.sessionStore))
	router.Use(verifyAuthorization(router.authEnforcer))

	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.HandleFunc("/session", router.handleLogin).Methods("POST")
	apiRouter.HandleFunc("/session", router.handleLogout).Methods("DELETE")

	tenantRouter := apiRouter.PathPrefix("/tenants/{tenantId}").Subrouter()
	tenantRouter.HandleFunc("", router.handleCreateTenant).Methods("POST")
	tenantRouter.HandleFunc("/divisions/{divisionId}", router.handleCreateDivision).Methods("POST")
	tenantRouter.HandleFunc("/divisions/{divisionId}/departments/{departmentId}", router.handleCreateDepartment).Methods("POST")

	userRouter := tenantRouter.PathPrefix("/users").Subrouter()
	userRouter.HandleFunc("/{userId}", router.handleCreateUser).Methods("POST")
	userRouter.HandleFunc("/{userId}/appointments/{appointmentId}", router.handleCreateAppointment).Methods("POST")

	rolesRouter := tenantRouter.PathPrefix("/roles").Subrouter()
	rolesRouter.HandleFunc("/{roleName}/policies", router.handleCreatePolicies).Methods("POST")
	//jobRequisitionRouter := tenantRouter.PathPrefix("/job-requisition").Subrouter()
	//jobRequisitionRouter.HandleFunc("", )

	router.NotFoundHandler = newRequestLogger(router.rootLogger)(errorHandling(http.HandlerFunc(router.handleNotFound))) // Custom 404 handler

	return router
}

// Fetches the locale from the Request Context & uses that to fetch the desired translator
func getAppropriateTranslator(r *http.Request, universalTranslator *ut.UniversalTranslator) (ut.Translator, error) {
	language, ok := r.Context().Value(languageKey).(string)
	if !ok {
		return nil, httperror.NewInternalServerError(errors.New("could not obtain preferred language"))
	}
	translator, _ := universalTranslator.GetTranslator(language)

	return translator, nil
}

func (router *Router) handleNotFound(w http.ResponseWriter, r *http.Request) {
	sendToErrorHandlingMiddleware(New404NotFoundError(), r)
}
