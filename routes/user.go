package routes

import (
	"crypto/rand"
	"encoding/json"
	"math/big"
	mathrand "math/rand"

	"net/http"
	"strings"

	"github.com/alexedwards/argon2id"
	"github.com/gorilla/mux"
	"github.com/pquerna/otp/totp"

	"multi-tenant-HR-information-system-backend/httperror"
	"multi-tenant-HR-information-system-backend/storage"
)

func generateRandomPassword(length int, minLower int, minUpper int, minNumber int, minSpecial int) string {
	lowerCharSet := "abcdedfghijklmnopqrst"
	upperCharSet := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	specialCharSet := "!@#$%&*"
	numberSet := "0123456789"
	allCharSet := lowerCharSet + upperCharSet + specialCharSet + numberSet

	var password strings.Builder
	for i := 0; i < minLower; i++ {
		random, err := rand.Int(rand.Reader, big.NewInt(int64(len(lowerCharSet))))
		if err != nil {
			panic(err)
		}
		password.WriteString(string(lowerCharSet[random.Int64()]))
	}

	for i := 0; i < minUpper; i++ {
		random, err := rand.Int(rand.Reader, big.NewInt(int64(len(upperCharSet))))
		if err != nil {
			panic(err)
		}
		password.WriteString(string(upperCharSet[random.Int64()]))
	}

	for i := 0; i < minSpecial; i++ {
		random, err := rand.Int(rand.Reader, big.NewInt(int64(len(specialCharSet))))
		if err != nil {
			panic(err)
		}
		password.WriteString(string(specialCharSet[random.Int64()]))
	}

	for i := 0; i < minNumber; i++ {
		random, err := rand.Int(rand.Reader, big.NewInt(int64(len(numberSet))))
		if err != nil {
			panic(err)
		}
		password.WriteString(string(numberSet[random.Int64()]))
	}

	remainingChars := length - minLower - minUpper - minSpecial - minNumber
	if remainingChars > 0 {
		for i := 0; i < remainingChars; i++ {
			random, err := rand.Int(rand.Reader, big.NewInt(int64(len(allCharSet))))
			if err != nil {
				panic(err)
			}
			password.WriteString(string(allCharSet[random.Int64()]))
		}
	}

	runeString := []rune(password.String()) // Convert to rune so that the string can be manipulated (strings are read-only)
	mathrand.Shuffle(len(runeString), func(i int, j int) {
		runeString[i], runeString[j] = runeString[j], runeString[i]
	})

	return string(runeString)
}

func (router *Router) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Email string
	}

	type responseBody struct {
		Password      string `json:"password"`
		TotpSecretKey string `json:"totpSecretKey"`
	}

	// Parse inputs
	var body requestBody
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		err := NewInvalidJSONError()
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	vars := mux.Vars(r)

	// Create default password + totpsecretkey
	defaultPassword := generateRandomPassword(12, 2, 2, 2, 2)
	hashedPassword, err := argon2id.CreateHash(defaultPassword, argon2id.DefaultParams)
	if err != nil {
		err := httperror.NewInternalServerError(err)
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "HRIS.com",
		AccountName: body.Email,
		SecretSize:  20,
		Period:      30,
	})
	if err != nil {
		err := httperror.NewInternalServerError(err)
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	user := storage.User{
		Id:            vars["userId"],
		TenantId:        vars["tenantId"],		
		Email:         body.Email,
		Password:      hashedPassword,
		TotpSecretKey: key.Secret(),
	}

	//Input validation
	translator, err := getAppropriateTranslator(r, router.universalTranslator)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	err = storage.ValidateStruct(router.validate, translator, user)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	// Make DB query
	err = router.storage.CreateUser(user)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	resBody := responseBody{
		Password:      defaultPassword,
		TotpSecretKey: key.Secret(),
	}

	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resBody)

	requestLogger := getRequestLogger(r)
	requestLogger.Info("USER-CREATED", "userId", user.Id)
}

func (router *Router) handleCreateAppointment(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Id           string
		Title        string
		DepartmentId string
		StartDate    string
		EndDate      string
	}

	// Parse inputs
	var body requestBody
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		err := NewInvalidJSONError()
		sendToErrorHandlingMiddleware(err, r)
		return
	}
	vars := mux.Vars(r)

	userAppointment := storage.Appointment{
		Id:           vars["appointmentId"],
		TenantId: vars["tenantId"],
		Title:        body.Title,
		DepartmentId: body.DepartmentId,
		UserId:       vars["userId"],
		StartDate:    body.StartDate,
		EndDate:      body.EndDate,
	}

	//Input validation
	translator, err := getAppropriateTranslator(r, router.universalTranslator)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}
	err = storage.ValidateStruct(router.validate, translator, userAppointment)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	err = router.storage.CreateAppointment(userAppointment)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	w.WriteHeader(http.StatusCreated)

	requestLogger := getRequestLogger(r)
	requestLogger.Info("APPOINTMENT-CREATED", "appointmentId", userAppointment.Id, "title", userAppointment.Title, "departmentId", userAppointment.DepartmentId, "userId", userAppointment.UserId)
}
