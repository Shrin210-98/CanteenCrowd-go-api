package server

import (
	"net/http"

	"ccms.com/api/internal/handlers"
)

func RegisterRoutes(handler *handlers.Handler) http.Handler {
	rootRouter := http.NewServeMux()
	v1Router := http.NewServeMux()

	registerRoutesV1(v1Router, handler)
	rootRouter.Handle("/api/v1/", http.StripPrefix("/api/v1", v1Router))

	return rootRouter
}

func registerRoutesV1(mux *http.ServeMux, handler *handlers.Handler) {

	mux.HandleFunc("GET    /health", handler.DatabaseHealth)

	mux.HandleFunc("POST   /login", handler.Login)
	mux.HandleFunc("POST   /register", handler.Register)

	// mux.HandleFunc("GET    /users", handler.ListUsers)
	// mux.HandleFunc("POST   /users", handler.CreateUser)
	// mux.HandleFunc("PUT    /users", handler.UpdateUser)
	// mux.HandleFunc("DELETE /users/{id}", handler.DeleteUser)
	mux.HandleFunc("GET    /user-profile", handler.GetUserProfile)

	mux.HandleFunc("GET    /employees", handler.Employees.ListEmployees)
	// mux.HandleFunc("GET    /employees", handler.SearchEmployees)
	mux.HandleFunc("GET    /employees/{id}", handler.Employees.GetEmployeeById)
	mux.HandleFunc("POST   /employees", handler.Employees.CreateEmployee)
	mux.HandleFunc("PUT    /employees", handler.Employees.UpdateEmployee)
	mux.HandleFunc("DELETE /employees/{id}", handler.Employees.DeleteEmployee)

	mux.HandleFunc("GET    /employees/departments", handler.Employees.ListDepartments)
	mux.HandleFunc("POST   /employees/departments", handler.Employees.CreateDepartment)
	mux.HandleFunc("PUT    /employees/departments", handler.Employees.UpdateDepartment)
	mux.HandleFunc("DELETE /employees/departments/{id}", handler.Employees.DeleteDepartment)

	mux.HandleFunc("GET    /employees/positions", handler.Employees.ListPositions)
	mux.HandleFunc("POST   /employees/positions", handler.Employees.CreatePosition)
	mux.HandleFunc("PUT    /employees/positions", handler.Employees.UpdatePosition)
	mux.HandleFunc("DELETE /employees/positions/{id}", handler.Employees.DeletePosition)

}

func OldRegisterRoutes(handler *handlers.Handler) http.Handler {
	mux := http.NewServeMux()

	// mux.HandleFunc("GET    /departments", handler.ListDepartments)
	// mux.HandleFunc("POST   /departments", handler.CreateDepartment)
	// mux.HandleFunc("PUT    /departments", handler.UpdateDepartment)
	// mux.HandleFunc("DELETE /departments/{id}", handler.DeleteDepartment)

	return mux
}
