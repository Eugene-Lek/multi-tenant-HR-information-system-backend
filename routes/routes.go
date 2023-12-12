package routes

import (
	"net/http"
	
	"github.com/gorilla/mux"
)

type Storage interface {
	CreateTenant(name string) error
	CreateDivision(tenant string, name string) error
	CreateDepartment(tenant string, division string, name string) error
}

type Router struct {
	router *mux.Router
	storage Storage
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.router.ServeHTTP(w, req)
}

func NewRouter(storage Storage) *Router {
	router := mux.NewRouter()

	r := &Router{
		router: router,
		storage: storage,
	}

	tenantRouter := r.router.PathPrefix("/api/{tenant}").Subrouter()
	tenantRouter.HandleFunc("", r.handleCreateTenant).Methods("POST")
	tenantRouter.HandleFunc("/{division}", r.handleCreateDivision).Methods("POST")
	tenantRouter.HandleFunc("/{division}/{department}", r.handleCreateDepartment).Methods("POST")

	//userRouter := tenantRouter.PathPrefix("/users").Subrouter()
	//userRouter.HandleFunc("/{email}", r.handleUser)
	//userRouter.HandleFunc("", )

	//jobRequisitionRouter := tenantRouter.PathPrefix("/job-requisition").Subrouter()
	//jobRequisitionRouter.HandleFunc("", )	

	return r
}

