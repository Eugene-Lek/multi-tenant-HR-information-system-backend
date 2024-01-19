package routes

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/sessions"

	"multi-tenant-HR-information-system-backend/httperror"
	"multi-tenant-HR-information-system-backend/storage"
)

const authSessionName = "authenticated"

func (router *Router) handleLogin(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		TenantId string
		Email    string
		Password string
		Totp     string
	}

	var reqBody requestBody
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		sendToErrorHandlingMiddleware(NewInvalidJSONError(), r)
		return
	}

	// Note: No validation because the db query & password checks can handle empty inputs

	valid, err := router.validateCredentials(reqBody.Email, reqBody.TenantId, reqBody.Password, reqBody.Totp)
	if err != nil {
		sendToErrorHandlingMiddleware(httperror.NewInternalServerError(err), r)
		return
	}
	if !valid {
		sendToErrorHandlingMiddleware(NewUnauthenticatedError(), r)
		return
	}

	// If the session isn't in the req context, it tries to retrieve the it from the session store
	// If it isn't in the session store, it returns a new session with an empty session id
	session, err := router.sessionStore.Get(r, authSessionName)
	if err != nil {
		sendToErrorHandlingMiddleware(httperror.NewInternalServerError(err), r)
		return
	}

	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: true,
		//Secure: true, --> only in production where the frontend has an SSL certificate
	}

	session.Values["tenantId"] = reqBody.TenantId
	session.Values["email"] = reqBody.Email

	filter := storage.User{
		TenantId: reqBody.TenantId,
		Email:    reqBody.Email,
	}
	users, err := router.storage.GetUsers(filter)
	if err != nil {
		sendToErrorHandlingMiddleware(httperror.NewInternalServerError(err), r)
		return
	}
	session.Values["id"] = users[0].Id

	err = router.sessionStore.Save(r, w, session)
	if err != nil {
		sendToErrorHandlingMiddleware(httperror.NewInternalServerError(err), r)
		return
	}

	// Check that session was saved & get its ID
	s, err := router.sessionStore.Get(r, authSessionName)
	if err != nil {
		sendToErrorHandlingMiddleware(httperror.NewInternalServerError(err), r)
		return
	}

	reqLogger := getRequestLogger(r)
	reqLogger.Info("SESSION-CREATED", "sessionId", s.ID)
	reqLogger.Info("USER-AUTHENTICATED", "userId", users[0].Id)

	w.WriteHeader(http.StatusOK)
}

func (router *Router) handleLogout(w http.ResponseWriter, r *http.Request) {
	session, err := router.sessionStore.Get(r, authSessionName)
	if err != nil {
		sendToErrorHandlingMiddleware(httperror.NewInternalServerError(err), r)
		return
	}

	userId, sessionExists := session.Values["id"].(string) // Used for logging later

	session.Options = &sessions.Options{
		MaxAge: -1,
	}

	// Deletes the session from the storage & sets the cookie's max age to -1
	err = router.sessionStore.Save(r, w, session)
	if err != nil {
		sendToErrorHandlingMiddleware(httperror.NewInternalServerError(err), r)
		return
	}

	reqLogger := getRequestLogger(r)
	if sessionExists {
		reqLogger.Info("SESSION-DELETED", "userId", userId)
	} else {
		reqLogger.Warn("SESSION-ALREADY-DELETED", "sessionId", session.ID)
	}

	w.WriteHeader(http.StatusOK)
}
