package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"multi-tenant-HR-information-system-backend/httperror"
	"multi-tenant-HR-information-system-backend/storage"
)

type contextKey int

const (
	requestLoggerKey contextKey = iota
	errorHandlingKey
	languageKey
	authenticatedUserKey
)

// Creates a request-specific logger & adds it to the request context
func newRequestLogger(rootLogger *tailoredLogger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestId := uuid.New().String()
			requestLogger := rootLogger.With("requestId", requestId, "clientIp", r.RemoteAddr, "url", r.URL.Path, "method", r.Method)

			r = r.WithContext(context.WithValue(r.Context(), requestLoggerKey, requestLogger))

			next.ServeHTTP(w, r)
		})
	}
}

func getRequestLogger(r *http.Request) *tailoredLogger {
	requestLogger := r.Context().Value(requestLoggerKey).(*tailoredLogger)
	return requestLogger
}

// Implementation of http.ResponseWriter so the response status is recorded for logging purposes too!
// Inherits all methods from http.ResponseWriter except WriteHeader
type ResponseWriterRecorder struct {
	http.ResponseWriter
	status int
}

func (wr *ResponseWriterRecorder) WriteHeader(status int) {
	wr.ResponseWriter.WriteHeader(status)
	wr.status = status
}

// Note: "X-Real-Ip" and "X-Forwarded-For" headers are not used for the clientIp because they can be modified by the client == security risk
func logRequestCompletion(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		wr := ResponseWriterRecorder{w, 0}
		next.ServeHTTP(&wr, r)

		duration := time.Since(startTime)

		requestLogger := getRequestLogger(r)
		requestLogger.Info("REQUEST-COMPLETED", "responseTime", duration.String(), "status", wr.status)
	})
}

type ErrorTransport struct {
	Error error
}

func errorHandling(next http.Handler) http.Handler {
	type errorResponseBody struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create an ErrorTransport struct instance and pass it into the Request Context by reference
		// This way, any modifications to the ErrorTransport (i.e. adding an Error to it) can be accessed inside this
		// middleware scope

		errTransport := new(ErrorTransport)
		r = r.WithContext(context.WithValue(r.Context(), errorHandlingKey, errTransport))

		// Call the remaining middleware(s) & router handler. If an error occurs, it will be added to the same errTransport
		// defined above
		next.ServeHTTP(w, r)

		if errTransport.Error == nil {
			// If no error was attached to the ErrorTransport, end the middleware call
			return
		}

		err, ok := errTransport.Error.(*httperror.Error)
		if !ok {
			// If the error provided is not a httperror.Error, convert it to an InternalServerError
			err = httperror.NewInternalServerError(err)
		}

		var message string
		requestLogger := getRequestLogger(r)
		if err.Code == "INTERNAL-SERVER-ERROR" {
			// Do not reveal internal server error stack traces to the client!!
			traceId := uuid.New().String()
			message = fmt.Sprintf("Something went wrong. Trace ID: %s", traceId)

			errorMessage, stackTrace, _ := strings.Cut(err.Error(), "\n")
			requestLogger.Error(err.Code, "errorMessage", errorMessage, "stackTrace", stackTrace, "traceId", traceId)
		} else {
			message = err.Error()
			requestLogger.Warn(err.Code, "errorMessage", err.Error())
		}

		body := errorResponseBody{
			Code:    err.Code,
			Message: message,
		}
		w.Header().Add("content-type", "application/json")
		w.WriteHeader(err.Status)
		json.NewEncoder(w).Encode(body)
	})
}

// Sends the httperror.Error to the error handling middleware via the Request Context
// The error is assigned to an existing pointer in the Request context
func sendToErrorHandlingMiddleware(err error, r *http.Request) {
	if errTransport, ok := r.Context().Value(errorHandlingKey).(*ErrorTransport); ok {
		errTransport.Error = err
	}
}

func getAcceptedLanguage(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		acceptLanguageHeader := r.Header.Get("Accept-Language")
		languageValue := strings.Split(acceptLanguageHeader, "-")[0]

		r = r.WithContext(context.WithValue(r.Context(), languageKey, languageValue))

		next.ServeHTTP(w, r)
	})
}

func authenticateUser(sessionStore sessions.Store) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := sessionStore.Get(r, authSessionName)
			if err != nil {
				sendToErrorHandlingMiddleware(httperror.NewInternalServerError(err), r)
				return
			}

			var user storage.User
			if _, ok := session.Values["id"].(string); !ok || session.ID == "" {
				// If the session ID is empty, the user does not have an existing session
				// If the sessionID was found but its values have been deleted, the session is invalid & user is not authenticated
				user = storage.User{
					Id:       "public",
					TenantId: "public",
					Email:    "public",
				}

				// If the user attempts to use a deleted session, log a warning (security reasons)
				if session.ID != "" && !ok {
					reqLogger := getRequestLogger(r)
					reqLogger.Warn("DELETED-SESSION-USED", "sessionId", session.ID)
				}
			} else {
				user = storage.User{
					Id:       session.Values["id"].(string),
					TenantId: session.Values["tenantId"].(string),
					Email:    session.Values["email"].(string),
				}
			}

			r = r.WithContext(context.WithValue(r.Context(), authenticatedUserKey, user))

			// Add the userId to the request logger
			reqLogger := getRequestLogger(r)
			reqLoggerWithUserID := reqLogger.With("userId", user.Id)
			r = r.WithContext(context.WithValue(r.Context(), requestLoggerKey, reqLoggerWithUserID))

			next.ServeHTTP(w, r)
		})
	}
}

func verifyAuthorization(authEnforcer casbin.IEnforcer) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := r.Context().Value(authenticatedUserKey).(storage.User)

			authorized, err := authEnforcer.Enforce(user.Id, user.TenantId, r.URL.Path, r.Method)
			if err != nil {
				sendToErrorHandlingMiddleware(httperror.NewInternalServerError(err), r)
				return
			}

			if !authorized {
				sendToErrorHandlingMiddleware(NewUnauthorisedError(), r)
				return
			}

			reqLogger := getRequestLogger(r)
			reqLogger.Info("USER-AUTHORISED", "userId", user.Id, "tenantId", user.TenantId, "resource", r.URL.Path, "method", r.Method)

			next.ServeHTTP(w, r)
		})
	}
}
