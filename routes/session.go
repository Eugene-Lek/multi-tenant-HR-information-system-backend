package routes

import (
	"encoding/json"
	"errors"
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
		sendToErrorHandlingMiddleware(ErrInvalidJSON, r)
		return
	}

	type Input struct {
		TenantId string `validate:"required" name:"tenant id"`
		Email    string `validate:"required" name:"user email"`
		Password string
		Totp     string
	}
	input := Input{
		TenantId: reqBody.TenantId,
		Email:    reqBody.Email,
		Password: reqBody.Password,
		Totp:     reqBody.Password,
	}
	translator := getTranslator(r)
	err = validateStruct(router.validate, translator, input)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	valid, err := router.validateCredentials(reqBody.Email, reqBody.TenantId, reqBody.Password, reqBody.Totp)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}
	if !valid {
		sendToErrorHandlingMiddleware(ErrUserUnauthenticated, r)
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
		sendToErrorHandlingMiddleware(err, r)
		return
	}
	if len(users) == 0 {
		err := httperror.NewInternalServerError(
			errors.New("race condition occurred: user was deleted after credentials validation but before session creation"),
		)
		sendToErrorHandlingMiddleware(err, r)
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
	reqLogger.Info("USER-AUTHENTICATED", "userId", users[0].Id, "tenantId", users[0].TenantId)

	w.WriteHeader(http.StatusOK)
}

func (router *Router) handleLogout(w http.ResponseWriter, r *http.Request) {
	session, err := router.sessionStore.Get(r, authSessionName)
	if err != nil {
		sendToErrorHandlingMiddleware(httperror.NewInternalServerError(err), r)
		return
	}

	// Used for logging later
	userId, _ := session.Values["id"].(string)
	tenantId, sessionExists := session.Values["tenantId"].(string)

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
		reqLogger.Info("SESSION-DELETED", "userId", userId, "tenantId", tenantId)
	} else {
		reqLogger.Warn("SESSION-ALREADY-DELETED", "sessionId", session.ID)
	}

	w.WriteHeader(http.StatusOK)
}
