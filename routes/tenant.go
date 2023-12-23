package routes

import (
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"

	"multi-tenant-HR-information-system-backend/errors"	
)

type Tenant struct {
	Name string `validate:"required"`
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

	language, ok := r.Context().Value(languageKey).(string)
	if !ok {
		//TODO error handling
		log.Print("")
	}
	translator, _ := router.universalTranslator.GetTranslator(language)
	
	err := router.validate.Struct(tenant)
	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		errorMessages := validationErrors.Translate(translator)
		err := &errors.InputValidationError{
			Status: 400,
			ValidationErrors: errorMessages,
		}
		log.Print(err.Error())
	}

	err = router.storage.CreateTenant(tenant)
	if err != nil {
		log.Print(err.Error())
	}

	w.WriteHeader(http.StatusNoContent)
}

func (router *Router) handleCreateDivision(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	division := Division{
		Name: vars["division"],
		Tenant: vars["tenant"],
	}

	//TODO parameter validation

	err := router.storage.CreateDivision(division)
	if err != nil {
		log.Fatalf(err.Error())
	}

	w.WriteHeader(http.StatusNoContent)
}

func (router *Router) handleCreateDepartment (w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	department := Department{
		Name: vars["department"],
		Tenant: vars["tenant"],
		Division: vars["division"],
	}

	err := router.storage.CreateDepartment(department)
	if err != nil {
		log.Fatalf(err.Error())
	}

	w.WriteHeader(http.StatusNoContent)
}