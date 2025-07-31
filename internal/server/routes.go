package server

import (
	"net/http"

	"ccms.com/api/internal/handlers"
)

func RegisterRoutes(handler *handlers.Handler) http.Handler {
	// Create a root router
	rootRouter := http.NewServeMux()

	// Create versioned API routers
	v1Router := http.NewServeMux()

	// Register v1 routes
	registerRoutesV1(v1Router, handler)

	// Mount the versioned routers under their paths
	rootRouter.Handle("/api/v1/", http.StripPrefix("/api/v1", v1Router))

	return rootRouter
}

func registerRoutesV1(mux *http.ServeMux, handler *handlers.Handler) {

	// mux.HandleFunc("POST   /register", handler.Register)
	// mux.HandleFunc("POST   /login", handler.Login)

	mux.HandleFunc("GET    /departments", handler.ListDepartments)
	mux.HandleFunc("POST   /departments", handler.CreateDepartment)
	mux.HandleFunc("PUT    /departments", handler.UpdateDepartment)
	mux.HandleFunc("DELETE /departments/{id}", handler.DeleteDepartment)

	mux.HandleFunc("GET    /positions", handler.ListPositions)
	mux.HandleFunc("POST   /positions", handler.CreatePosition)
	mux.HandleFunc("PUT    /positions", handler.UpdatePosition)
	mux.HandleFunc("DELETE /positions/{id}", handler.DeletePosition)

	mux.HandleFunc("GET    /employees", handler.ListEmployees)
	// mux.HandleFunc("GET    /employees", handler.SearchEmployees)
	mux.HandleFunc("GET    /employees/{id}", handler.GetEmployeeById)
	mux.HandleFunc("POST   /employees", handler.CreateEmployee)
	mux.HandleFunc("PUT    /employees", handler.UpdateEmployee)
	mux.HandleFunc("DELETE /employees/{id}", handler.DeleteEmployee)

}

// -- Old Implementation --
func OldRegisterRoutes(handler *handlers.Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET    /departments", handler.ListDepartments)
	mux.HandleFunc("POST   /departments", handler.CreateDepartment)
	mux.HandleFunc("PUT    /departments", handler.UpdateDepartment)
	mux.HandleFunc("DELETE /departments/{id}", handler.DeleteDepartment)

	return mux
}
