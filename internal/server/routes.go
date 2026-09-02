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

	mux.HandleFunc("POST   /register", handler.Register)
	mux.HandleFunc("POST   /login", handler.Login)
	mux.HandleFunc("POST   /logout", handler.Logout)
	mux.HandleFunc("GET    /me", handler.GetUserProfile)
	mux.HandleFunc("PUT    /me", handler.UpdateUserProfile)

	// ========= Users Module =============

	mux.HandleFunc("GET    /users", handler.Users.ListUsers)
	mux.HandleFunc("GET    /users/{id}", handler.Users.GetUserById)
	mux.HandleFunc("POST   /users", handler.Users.CreateUser)
	mux.HandleFunc("PUT    /users", handler.Users.UpdateUser)
	mux.HandleFunc("DELETE /users/{id}", handler.Users.DeleteUser)

	mux.HandleFunc("GET    /users/roles", handler.Users.ListRoles)
	mux.HandleFunc("POST   /users/roles", handler.Users.CreateRole)
	mux.HandleFunc("PUT    /users/roles", handler.Users.UpdateRole)
	mux.HandleFunc("DELETE /users/roles/{id}", handler.Users.DeleteRole)
	mux.HandleFunc("GET    /users/roles/{id}", handler.Users.GetRoleById)
	mux.HandleFunc("GET    /users/permissions", handler.Users.GetDefaultPermissionsByUserType)

	// ========== Employee Module ===========

	mux.HandleFunc("GET    /employees", handler.Employees.ListEmployees)
	mux.HandleFunc("POST   /employees", handler.Employees.CreateEmployee)
	mux.HandleFunc("PUT    /employees", handler.Employees.UpdateEmployee)
	mux.HandleFunc("DELETE /employees/{id}", handler.Employees.DeleteEmployee)
	mux.HandleFunc("GET    /employees/{id}", handler.Employees.GetEmployeeById)
	// mux.HandleFunc("GET    /employees", handler.SearchEmployees)

	mux.HandleFunc("GET    /employees/departments", handler.Employees.ListDepartments)
	mux.HandleFunc("POST   /employees/departments", handler.Employees.CreateDepartment)
	mux.HandleFunc("PUT    /employees/departments", handler.Employees.UpdateDepartment)
	mux.HandleFunc("DELETE /employees/departments/{id}", handler.Employees.DeleteDepartment)

	mux.HandleFunc("GET    /employees/positions", handler.Employees.ListPositions)
	mux.HandleFunc("POST   /employees/positions", handler.Employees.CreatePosition)
	mux.HandleFunc("PUT    /employees/positions", handler.Employees.UpdatePosition)
	mux.HandleFunc("DELETE /employees/positions/{id}", handler.Employees.DeletePosition)

}
