package routes

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Tenant struct {
	Name      string `validate:"required,notBlank" name:"tenant name"`
	CreatedAt string
	UpdatedAt string
}

type Division struct {
	Name      string `validate:"required,notBlank" name:"division name"`
	Tenant    string `validate:"required,notBlank" name:"tenant name"`
	CreatedAt string
	UpdatedAt string
}

type Department struct {
	Name      string `validate:"required,notBlank" name:"department name"`
	Tenant    string `validate:"required,notBlank" name:"tenant name"`
	Division  string `validate:"required,notBlank" name:"division name"`
	CreatedAt string
	UpdatedAt string
}

func (router *Router) handleCreateTenant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenant := Tenant{
		Name: vars["tenant"],
	}

	translator, err := getAppropriateTranslator(r, router.universalTranslator)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	err = validateStruct(router.validate, translator, tenant)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	err = router.storage.CreateTenant(tenant)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	w.WriteHeader(http.StatusCreated)

	requestLogger := getRequestLogger(r)
	requestLogger.Info("CREATED-TENANT", "tenant", tenant.Name)
}

func (router *Router) handleCreateDivision(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	division := Division{
		Name:   vars["division"],
		Tenant: vars["tenant"],
	}

	//TODO parameter validation
	translator, err := getAppropriateTranslator(r, router.universalTranslator)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	err = validateStruct(router.validate, translator, division)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	err = router.storage.CreateDivision(division)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	w.WriteHeader(http.StatusCreated)

	requestLogger := getRequestLogger(r)
	requestLogger.Info("CREATED-DIVISION", "tenant", division.Tenant, "division", division.Name)
}

func (router *Router) handleCreateDepartment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	department := Department{
		Name:     vars["department"],
		Tenant:   vars["tenant"],
		Division: vars["division"],
	}

	//TODO parameter validation
	translator, err := getAppropriateTranslator(r, router.universalTranslator)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	err = validateStruct(router.validate, translator, department)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	err = router.storage.CreateDepartment(department)
	if err != nil {
		sendToErrorHandlingMiddleware(err, r)
		return
	}

	w.WriteHeader(http.StatusCreated)
	requestLogger := getRequestLogger(r)
	requestLogger.Info("CREATED-DIVISION", "tenant", department.Tenant, "division", department.Division, "department", department.Name)
}
