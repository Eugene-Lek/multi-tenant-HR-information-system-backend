package routes


import  (
	"net/http"
	"log"

	"github.com/gorilla/mux"
)

func (router *Router) handleCreateTenant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenant := vars["tenant"]
	//TODO input validation

	err := router.storage.CreateTenant(tenant)
	if err != nil {
		log.Print(err.Error())
	}

	w.WriteHeader(http.StatusNoContent)
}

func (router *Router) handleCreateDivision(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenant := vars["tenant"]	
	division := vars["division"]	

	//TODO parameter validation

	err := router.storage.CreateDivision(tenant, division)
	if err != nil {
		log.Fatalf(err.Error())
	}

	w.WriteHeader(http.StatusNoContent)
}

func (router *Router) handleCreateDepartment (w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenant := vars["tenant"]
	division := vars["division"]
	department := vars["department"]

	err := router.storage.CreateDepartment(tenant, division, department)
	if err != nil {
		log.Fatalf(err.Error())
	}

	w.WriteHeader(http.StatusNoContent)
}