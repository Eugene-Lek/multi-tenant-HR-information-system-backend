package routes

import (
	"net/http"
	"context"
	"strings"

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

	r.router.Use(getLanguageMiddleware)

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

type contextKey int

const (
	languageKey contextKey = iota
)

func getLanguageMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		acceptLanguageHeader := r.Header.Get("Accept-Language")
		languageValue := strings.Split(acceptLanguageHeader, "-")[0]

		r = r.WithContext(context.WithValue(r.Context(), languageKey, languageValue))

		next.ServeHTTP(w, r)
	})
}