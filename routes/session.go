package routes

import (
	"encoding/json"
	"net/http"

	"github.com/alexedwards/argon2id"
	"github.com/gorilla/sessions"
	"github.com/pquerna/otp/totp"
)

const authSessionName = "authenticated"

func (router *Router) handleLogin(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Tenant   string
		Email    string
		Password string
		Totp     string
	}

	var reqBody requestBody
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		err = NewInvalidJSONError()
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	// Note: No validation because the db query & password checks can handle empty inputs

	// Get the user
	filter := User{
		Tenant: reqBody.Tenant,
		Email:  reqBody.Email,
	}

	users, err := router.storage.GetUsers(filter)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	var user User
	if len(users) == 0 {
		// If the user does not exist, default the password & TOTP to empty strings
		// Executing the password check nonetheless prevents timing attacks
		user = User{} //TODO create default with valid argon2id hash
	} else {
		user = users[0]
	}

	//validate the password & TOTP
	passwordMatch, err := argon2id.ComparePasswordAndHash(reqBody.Password, user.Password)
	if err != nil {
		sendToErrorHandlingMiddleware(NewInternalServerError(err), r)
		return
	}

	valid := totp.Validate(reqBody.Totp, user.TotpSecretKey)

	// Return a session cookie if the user's credentials are correct
	if passwordMatch && valid && len(users) != 0 {
		// If the session isn't in the req context, it tries to retrieve the it from the store
		// If it isn't in the store, it returns a session with an empty session id
		session, err := router.sessionStore.Get(r, authSessionName)
		if err != nil {
			sendToErrorHandlingMiddleware(NewInternalServerError(err), r)
			return
		}

		session.Options = &sessions.Options{
			Path:     "/",
			MaxAge:   86400,
			HttpOnly: true,
			//Secure: true, --> only in production where the frontend has an SSL certificate
		}

		session.Values["tenant"] = reqBody.Tenant
		session.Values["email"] = reqBody.Email
		session.Values["id"] = user.Id

		err = router.sessionStore.Save(r, w, session)
		if err != nil {
			sendToErrorHandlingMiddleware(NewInternalServerError(err), r)
			return
		}

		// Check that session was saved & get its ID
		s, err := router.sessionStore.Get(r, authSessionName)
		if err != nil {
			sendToErrorHandlingMiddleware(NewInternalServerError(err), r)
			return
		}

		reqLogger := getRequestLogger(r)
		reqLogger.Info("SESSION-CREATED", "sessionId", s.ID)
		reqLogger.Info("USER-AUTHENTICATED", "userId", user.Id)

	} else {
		err := NewUnauthenticatedError()
		sendToErrorHandlingMiddleware(err, r)
		return
	}
}

func (router *Router) handleLogout(w http.ResponseWriter, r *http.Request) {
	session, err := router.sessionStore.Get(r, authSessionName)
	if err != nil {
		sendToErrorHandlingMiddleware(NewInternalServerError(err), r)
		return
	}

	userId, sessionExists := session.Values["id"].(string) // Used for logging later

	session.Options = &sessions.Options{
		MaxAge: -1,
	}

	// Deletes the session from the storage & sets the cookie's max age to -1
	err = router.sessionStore.Save(r, w, session)
	if err != nil {
		sendToErrorHandlingMiddleware(NewInternalServerError(err), r)
		return
	}

	reqLogger := getRequestLogger(r)
	if sessionExists {
		reqLogger.Info("SESSION-DELETED", "userId", userId)
	} else {
		reqLogger.Warn("SESSION-ALREADY-DELETED", "sessionId", session.ID)
	}
}
