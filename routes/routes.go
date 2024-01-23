package routes

import (
	"net/http"

	"github.com/alexedwards/argon2id"
	"github.com/casbin/casbin/v2"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/pquerna/otp/totp"

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
	router.Use(setRequestLogger(router.rootLogger))
	router.Use(logRequestCompletion)
	router.Use(errorHandling)
	router.Use(setTranslator(router.universalTranslator))
	router.Use(authenticateUser(router.sessionStore))
	router.Use(verifyAuthorization(router.authEnforcer))

	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.HandleFunc("/session", router.handleLogin).Methods("POST")
	apiRouter.HandleFunc("/session", router.handleLogout).Methods("DELETE")

	tenantRouter := apiRouter.PathPrefix("/tenants/{tenantId}").Subrouter()
	tenantRouter.HandleFunc("", router.handleCreateTenant).Methods("POST")
	tenantRouter.HandleFunc("/divisions/{divisionId}", router.handleCreateDivision).Methods("POST")
	tenantRouter.HandleFunc("/divisions/{divisionId}/departments/{departmentId}", router.handleCreateDepartment).Methods("POST")

	policiesRouter := tenantRouter.PathPrefix("/policies").Subrouter()
	policiesRouter.HandleFunc("", router.handleCreatePolicies).Methods("POST")

	positionRouter := tenantRouter.PathPrefix("/positions/{positionId}").Subrouter()
	positionRouter.HandleFunc("", router.handleCreatePosition).Methods("POST")

	userRouter := tenantRouter.PathPrefix("/users/{userId}").Subrouter()
	userRouter.HandleFunc("", router.handleCreateUser).Methods("POST")

	positionAssignmentRouter := userRouter.PathPrefix("/positions/{positionId}").Subrouter()
	positionAssignmentRouter.HandleFunc("", router.handleCreatePositionAssignment).Methods("POST")

	userRolesRouter := userRouter.PathPrefix("/roles").Subrouter()
	userRolesRouter.HandleFunc("/{roleName}", router.handleCreateRoleAssignment).Methods("POST")

	jobRequisitionRouter := userRouter.PathPrefix("/job-requisitions").Subrouter()
	jobRequisitionRouter.HandleFunc("/role-requestor/{jobRequisitionId}", router.handleCreateJobRequisition).Methods("POST")
	jobRequisitionRouter.HandleFunc("/role-supervisor/{jobRequisitionId}/supervisor-approval", router.handleSupervisorApproveJobRequisition).Methods("POST")
	jobRequisitionRouter.HandleFunc("/role-hr-approver/{jobRequisitionId}/hr-approval", router.handleHrApproveJobRequisition).Methods("POST")

	//jobRequisitionRouter := tenantRouter.PathPrefix("/job-requisition").Subrouter()
	//jobRequisitionRouter.HandleFunc("", )

	router.NotFoundHandler = setRequestLogger(router.rootLogger)(errorHandling(http.HandlerFunc(router.handleNotFound))) // Custom 404 handler

	return router
}

func (router *Router) handleNotFound(w http.ResponseWriter, r *http.Request) {
	sendToErrorHandlingMiddleware(Err404NotFound, r)
}

func (router *Router) validateCredentials(email string, tenantId string, password string, otp string) (bool, error) {
	filter := storage.User{
		TenantId: tenantId,
		Email:    email,
	}
	users, err := router.storage.GetUsers(filter)
	if err != nil {
		return false, err
	}

	var user storage.User
	if len(users) == 0 {
		// If the user does not exist, use the default password and totp secret key
		// The password hash is pre-generated using the password "default"
		// Executing the password check nonetheless prevents timing attacks
		user = storage.User{
			Password:      `$argon2id$v=19$m=65536,t=1,p=8$RWNiQ1R3UTVnQ1Fxb3dQdg$y0BaFbMhsPz4YqIuXWe5pUPF/1g66t2fogccTlkYpyQ`,
			TotpSecretKey: `default`,
		}
	} else {
		user = users[0]
	}

	//validate the password & TOTP
	passwordMatch, err := argon2id.ComparePasswordAndHash(password, user.Password)
	if err != nil {
		return false, httperror.NewInternalServerError(err)
	}
	valid := totp.Validate(otp, user.TotpSecretKey)

	return passwordMatch && valid && len(users) != 0, nil
}
