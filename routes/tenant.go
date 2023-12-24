package routes

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type Tenant struct {
	Name string `validate:"required,notBlank"`
	CreatedAt string
	UpdatedAt string
}

type Division struct {
	Name string `validate:"required"`
	Tenant string `validate:"required"`
	CreatedAt string
	UpdatedAt string
}

type Department struct {
	Name string `validate:"required"`
	Tenant string `validate:"required"`
	Division string `validate:"required"`
	CreatedAt string
	UpdatedAt string
}

func (router *Router) handleCreateTenant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenant := Tenant{
		Name: vars["tenant"],
	}

	translator, httpErr := getAppropriateTranslator(r, router.universalTranslator)
	if httpErr != nil {
		sendToErrorHandlingMiddleware(httpErr, r)
		log.Print(httpErr.Error())

		return
	}	
	
	httpErr = validateStruct(router.validate, translator, tenant)
	if httpErr != nil {
		sendToErrorHandlingMiddleware(httpErr, r)
		log.Print(httpErr.Error())

		return
	}

	httpErr = router.storage.CreateTenant(tenant)
	if httpErr != nil {
		sendToErrorHandlingMiddleware(httpErr, r)
		log.Print(httpErr.Error())

		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (router *Router) handleCreateDivision(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	division := Division{
		Name: vars["division"],
		Tenant: vars["tenant"],
	}

	//TODO parameter validation
	translator, httpErr := getAppropriateTranslator(r, router.universalTranslator)
	if httpErr != nil {
		sendToErrorHandlingMiddleware(httpErr, r)
		log.Print(httpErr.Error())

		return
	}	
	
	httpErr = validateStruct(router.validate, translator, division)
	if httpErr != nil {
		sendToErrorHandlingMiddleware(httpErr, r)
		log.Print(httpErr.Error())

		return
	}	

	httpErr = router.storage.CreateDivision(division)
	if httpErr != nil {
		sendToErrorHandlingMiddleware(httpErr, r)
		log.Fatalf(httpErr.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (router *Router) handleCreateDepartment (w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	department := Department{
		Name: vars["department"],
		Tenant: vars["tenant"],
		Division: vars["division"],
	}

	//TODO parameter validation
	translator, httpErr := getAppropriateTranslator(r, router.universalTranslator)
	if httpErr != nil {
		sendToErrorHandlingMiddleware(httpErr, r)
		log.Print(httpErr.Error())

		return
	}	
	
	httpErr = validateStruct(router.validate, translator, department)
	if httpErr != nil {
		sendToErrorHandlingMiddleware(httpErr, r)
		log.Print(httpErr.Error())

		return
	}	

	httpErr = router.storage.CreateDepartment(department)
	if httpErr != nil {
		sendToErrorHandlingMiddleware(httpErr, r)
		log.Fatalf(httpErr.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
}