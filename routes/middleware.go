package routes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	ut "github.com/go-playground/universal-translator"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type contextKey int

const (
	requestLoggerKey contextKey = iota
	errorHandlingKey
	languageKey
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

// Wrapper for http.ResponseWriter so that the logRequestCompletion middleware can access the response status too
type loggingResponseWriter struct {
	http.ResponseWriter
	status int
}

// Re-implements the WriteHeader receiver function in order to store the response status for logging
func (lrw *loggingResponseWriter) WriteHeader(status int) {
	lrw.ResponseWriter.WriteHeader(status)
	lrw.status = status
}

// Logs the result of each request
// Note: "X-Real-Ip" and "X-Forwarded-For" headers are not used for the clientIp because they can be modified by the client == security risk
func logRequestCompletion(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		loggingResponseWriter := loggingResponseWriter{w, 0}
		next.ServeHTTP(&loggingResponseWriter, r)

		duration := time.Since(startTime)

		requestLogger := getRequestLogger(r)
		requestLogger.Info("REQUEST-COMPLETED", "responseTime", duration.String(), "status", loggingResponseWriter.status)
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

		err, ok := errTransport.Error.(*HttpError)
		if !ok {
			// If the error provided is not a HttpError, convert it to an InternalServerError
			err = NewInternalServerError(err)
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

// Sends the HttpError to the error handling middleware via the Request Context
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

// Fetches the locale from the Request Context & uses that to fetch the desired translator
func getAppropriateTranslator(r *http.Request, universalTranslator *ut.UniversalTranslator) (ut.Translator, error) {
	language, ok := r.Context().Value(languageKey).(string)
	if !ok {
		return nil, NewInternalServerError(errors.New("could not obtain preferred language"))
	}
	translator, _ := universalTranslator.GetTranslator(language)

	return translator, nil
}
