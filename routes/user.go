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

	var body requestBody
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		sendToErrorHandlingMiddleware(NewInvalidJSONError(), r)
		return
	}
	vars := mux.Vars(r)

	//Input validation
	type Input struct {
		Id       string `validate:"required,notBlank,uuid" name:"user id"`
		TenantId string `validate:"required,notBlank,uuid" name:"tenant id"`
		Email    string `validate:"required,notBlank,email" name:"user email"`
	}
	input := Input{
		Id:       vars["userId"],
		TenantId: vars["tenantId"],
		Email:    body.Email,
	}
	translator := getTranslator(r)
	err = validateStruct(router.validate, translator, input)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	// Create default password + totpsecretkey
	defaultPassword := generateRandomPassword(12, 2, 2, 2, 2)
	hashedPassword, err := argon2id.CreateHash(defaultPassword, argon2id.DefaultParams)
	if err != nil {
		sendToErrorHandlingMiddleware(httperror.NewInternalServerError(err), r)
		return
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "HRIS.com",
		AccountName: body.Email,
		SecretSize:  20,
		Period:      30,
	})
	if err != nil {
		sendToErrorHandlingMiddleware(httperror.NewInternalServerError(err), r)
		return
	}

	// Make DB query
	user := storage.User{
		Id:            input.Id,
		TenantId:      input.TenantId,
		Email:         input.Email,
		Password:      hashedPassword,
		TotpSecretKey: key.Secret(),
	}
	err = router.storage.CreateUser(user)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	requestLogger := getRequestLogger(r)
	requestLogger.Info("USER-CREATED", "userId", user.Id)

	w.WriteHeader(http.StatusCreated)
	w.Header().Add("content-type", "application/json")

	resBody := responseBody{
		Password:      defaultPassword,
		TotpSecretKey: key.Secret(),
	}
	json.NewEncoder(w).Encode(resBody)
}

func (router *Router) handleCreatePosition(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Id            string
		Title         string
		DepartmentId  string
		SupervisorIds []string
	}

	var body requestBody
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		sendToErrorHandlingMiddleware(NewInvalidJSONError(), r)
		return
	}
	vars := mux.Vars(r)

	//Input validation
	type Input struct {
		Id            string   `validate:"required,notBlank,uuid" name:"position id"`
		TenantId      string   `validate:"required,notBlank,uuid" name:"tenant id"`
		Title         string   `validate:"required,notBlank" name:"position title"`
		DepartmentId  string   `validate:"required,notBlank,uuid" name:"department id"`
		SupervisorIds []string `validate:"required,dive,notBlank,uuid" name:"supervisor ids"`
	}
	input := Input{
		Id:            vars["positionId"],
		TenantId:      vars["tenantId"],
		Title:         body.Title,
		DepartmentId:  body.DepartmentId,
		SupervisorIds: body.SupervisorIds,
	}
	translator := getTranslator(r)
	err = validateStruct(router.validate, translator, input)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	userPosition := storage.Position{
		Id:            input.Id,
		TenantId:      input.TenantId,
		Title:         input.Title,
		DepartmentId:  input.DepartmentId,
		SupervisorIds: input.SupervisorIds,
	}
	err = router.storage.CreatePosition(userPosition)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	requestLogger := getRequestLogger(r)
	requestLogger.Info("POSITION-CREATED", "positionId", userPosition.Id, "title", userPosition.Title, "departmentId", userPosition.DepartmentId)

	w.WriteHeader(http.StatusCreated)
}

func (router *Router) handleCreatePositionAssignment(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		StartDate string
		EndDate   string
	}

	var body requestBody
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		sendToErrorHandlingMiddleware(NewInvalidJSONError(), r)
		return
	}
	vars := mux.Vars(r)

	type Input struct {
		TenantId   string `validate:"required,notBlank,uuid" name:"tenant id"`
		PositionId string `validate:"required,notBlank,uuid" name:"position id"`
		UserId     string `validate:"required,notBlank,uuid" name:"user id"`
		StartDate  string `validate:"required,notBlank,isIsoDate" name:"start date"`
		EndDate    string `validate:"omitempty,notBlank,isIsoDate,validPositionAssignmentDuration" name:"end date"`
	}
	input := Input{
		TenantId:   vars["tenantId"],
		PositionId: vars["positionId"],
		UserId:     vars["userId"],
		StartDate:  body.StartDate,
		EndDate:    body.EndDate,
	}
	translator := getTranslator(r)
	err = validateStruct(router.validate, translator, input)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	userPositionAssignment := storage.PositionAssignment{
		TenantId:   input.TenantId,
		PositionId: input.PositionId,
		UserId:     input.UserId,
		StartDate:  input.StartDate,
		EndDate:    input.EndDate,
	}
	err = router.storage.CreatePositionAssignment(userPositionAssignment)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	requestLogger := getRequestLogger(r)
	requestLogger.Info("POSITION-ASSIGNMENT-CREATED", "tenantId", userPositionAssignment.TenantId, "positionId", userPositionAssignment.PositionId, "userId", userPositionAssignment.UserId)

	w.WriteHeader(http.StatusCreated)
}
