package routes


import  (
	"net/http"
	"log"

	"github.com/gorilla/mux"
)

type Tenant struct {
	Name string `validate:"required"`
	CreatedAt string
	UpdatedAt string
}

type Division struct {
	Name string
	Tenant string
	CreatedAt string
	UpdatedAt string
}

type Department struct {
	Name string
	Tenant string
	Division string
	CreatedAt string
	UpdatedAt string
}

func (router *Router) handleCreateTenant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenant := Tenant{
		Name: vars["tenant"],
	}
	//TODO input validation

	err := router.storage.CreateTenant(tenant)
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