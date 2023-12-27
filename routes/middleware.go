package routes

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	ut "github.com/go-playground/universal-translator"
)

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
func getAppropriateTranslator(r *http.Request, universalTranslator *ut.UniversalTranslator) (ut.Translator, error) {
	language, ok := r.Context().Value(languageKey).(string)
	if !ok {
		return nil, NewInternalServerError(errors.New("Could not obtain preferred language"))
	}
	translator, _ := universalTranslator.GetTranslator(language)

	return translator, nil
}

type ErrorTransport struct {
	Error error
}

func errorHandlingMiddleware(next http.Handler) http.Handler {
	type errorResponseBody struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create an ErrorTransport struct instance and pass it into the Request Context by reference
		// This way, any modifications to the ErrorTransport (i.e. adding an Error to it can be accessed inside this
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
		if err.Code == "INTERNAL-SERVER-ERROR" {
			// Do not reveal internal server error stack traces to the client!!
			message = "Something went wrong. Trace ID: " // TODO create trace id and add it to the logged error
		} else {
			message = err.Error()
		}

		body := errorResponseBody{
			Code:    err.Code,
			Message: message,
		}

		w.Header().Add("content-type", "application/json")
		w.WriteHeader(err.Status)
		json.NewEncoder(w).Encode(body)

		log.Print(err.Error()) //TODO error logging
	})
}

// Sends the HttpError to the error handling middleware via the Request Context
// The error is assigned to an existing pointer in the Request context
func sendToErrorHandlingMiddleware(err error, r *http.Request) {
	if errTransport, ok := r.Context().Value(errorHandlingKey).(*ErrorTransport); ok {
		errTransport.Error = err
	}
}
