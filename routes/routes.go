package routes

import (
	"context"
	"net/http"
	"strings"
	"encoding/json"
	"reflect"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"

	"multi-tenant-HR-information-system-backend/errors"	
)

type Storage interface {
	CreateTenant(tenant Tenant) errors.HttpError
	CreateDivision(division Division) errors.HttpError
	CreateDepartment(department Department) errors.HttpError
	CreateUser(user User) errors.HttpError
	CreateAppointment(appointment Appointment) errors.HttpError
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

func validateStruct(validate *validator.Validate, translator ut.Translator, s interface{}) errors.HttpError {
	err := validate.Struct(s)
	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		errorMessages := validationErrors.Translate(translator)
		return errors.NewInputValidationError(errorMessages)		
	}

	return nil
}

type contextKey int

const (
	errorHandlingKey contextKey = iota	
	languageKey
)

func getLanguageMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		acceptLanguageHeader := r.Header.Get("Accept-Language")
		languageValue := strings.Split(acceptLanguageHeader, "-")[0]

		r = r.WithContext(context.WithValue(r.Context(), languageKey, languageValue))

		next.ServeHTTP(w, r)
	})
}

// Fetches the locale from the Request Context & uses that to fetch the desired translator
func getAppropriateTranslator(r *http.Request, universalTranslator *ut.UniversalTranslator) (ut.Translator, errors.HttpError) {
	language, ok := r.Context().Value(languageKey).(string)
	if !ok {
		return nil, errors.NewInternalServerError("Could not obtain preferred language")
	}
	translator, _ := universalTranslator.GetTranslator(language)	

	return translator, nil
}

type ErrorWrapper struct {
	Error errors.HttpError
}

func errorHandlingMiddleware(next http.Handler) http.Handler {
	type errorResponseBody struct {
		Code string `json:"code"`
		Message string `json:"message"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a pointer (address) to the ErrorWrapper & store it in the request context
		// In the route handlers, the ErrorWrapper & hence its error object value will be assigned to this pointer
		// This way, the middleware can access the error object via the ErrorWrapper
		// An ErrorWrapper is necessary because the pointer created must point to a struct, not an interface. 
		// Otherwise, you cannot assign the error object to the pointer as it does not implement the pointer to the interface

		errWrapper := new(ErrorWrapper)
		r = r.WithContext(context.WithValue(r.Context(), errorHandlingKey, errWrapper))

		next.ServeHTTP(w, r)

		errWrapper, ok := r.Context().Value(errorHandlingKey).(*ErrorWrapper)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			//w.Write()
			return
		}

		httpErr := errWrapper.Error

		body := errorResponseBody{
			Code: reflect.TypeOf(httpErr).Name(), //TODO derive error code from error type name
			Message: httpErr.Error(),
		}

		w.Header().Add("content-type", "application/json")
		w.WriteHeader(httpErr.Status())
		json.NewEncoder(w).Encode(body)
	})
}

// Sends the HttpError to the error handling middleware via the Request Context
// The error is assigned to an existing pointer in the Request context
func sendToErrorHandlingMiddleware(err errors.HttpError, r *http.Request) {
	if errWrapper, ok := r.Context().Value(errorHandlingKey).(*ErrorWrapper); ok {
		errWrapper.Error = err
	}		
}