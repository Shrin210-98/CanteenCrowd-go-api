package server

import (
	"net/http"

	"ccms.com/api/internal/constants"
	"ccms.com/api/internal/handlers"
	"ccms.com/api/internal/utils"
)

type Route struct {
	Pattern    string
	Handler    http.HandlerFunc
	Permission string
}

func RegisterRoutes(handler *handlers.Handler) http.Handler {
	rootRouter := http.NewServeMux()
	v1Router := http.NewServeMux()

	registerRoutesV1(v1Router, handler)
	rootRouter.Handle("/api/v1/", http.StripPrefix("/api/v1", v1Router))

	return rootRouter
}

func registerRoutesV1(mux *http.ServeMux, h *handlers.Handler) {

	mux.HandleFunc("GET    /health", h.DatabaseHealth)

	mux.HandleFunc("POST   /register", h.Register)
	mux.HandleFunc("POST   /login", h.Login)
	mux.HandleFunc("POST   /logout", h.Logout)
	mux.HandleFunc("GET    /me", h.GetUserProfile)
	mux.HandleFunc("PUT    /me", h.UpdateUserProfile)

	permissionedRoutes := []Route{

		// ---{ Path, Handler, Permission }----

		// ===== Users Module =====
		{"GET    /users", h.Users.ListUsers, "users"},
		{"POST   /users", h.Users.CreateUser, "users.add"},
		{"GET    /users/{id}", h.Users.GetUserById, "users"},
		{"PUT    /users/{id}", h.Users.UpdateUser, "users.edit"},
		{"DELETE /users/{id}", h.Users.DeleteUser, "users.delete"},
		{"GET    /users/permissions", h.Users.GetDefaultPermissionsByUserType, "users"},

		{"GET    /users/roles", h.Users.ListRoles, "users"},
		{"POST   /users/roles", h.Users.CreateRole, "users.roles"},

		{"GET    /users/roles/{id}", h.Users.GetRoleById, "users"},
		{"PUT    /users/roles/{id}", h.Users.UpdateRole, "users.roles"},
		{"DELETE /users/roles/{id}", h.Users.DeleteRole, "users.roles"},

		// ===== Employees Module =====
		{"GET    /employees", h.Employees.ListEmployees, "employees"},
		{"POST   /employees", h.Employees.CreateEmployee, "employees.add"},
		{"PUT    /employees", h.Employees.UpdateEmployee, "employees.edit"},
		{"DELETE /employees/{id}", h.Employees.DeleteEmployee, "employees.delete"},
		{"GET    /employees/{id}", h.Employees.GetEmployeeById, "employees"},

		{"GET    /employees/hierarchy", h.Employees.GetHierarchy, "employees"},
		{"PUT    /employees/hierarchy", h.Employees.UpdateHierarchy, "employees"},
		{"GET    /employees/hierarchy/unassigned", h.Employees.GetUnassignedEmployees, "employees"},
		{"GET    /employees/{id}/direct-reports", h.Employees.GetDirectReports, "employees"},

		{"GET    /employees/departments", h.Employees.ListDepartments, "employees.departments"},
		{"POST   /employees/departments", h.Employees.CreateDepartment, "employees.departments"},
		{"PUT    /employees/departments", h.Employees.UpdateDepartment, "employees.departments"},
		{"DELETE /employees/departments/{id}", h.Employees.DeleteDepartment, "employees.departments"},

		{"GET    /employees/positions", h.Employees.ListPositions, "employees.positions"},
		{"POST   /employees/positions", h.Employees.CreatePosition, "employees.positions"},
		{"PUT    /employees/positions", h.Employees.UpdatePosition, "employees.positions"},
		{"DELETE /employees/positions/{id}", h.Employees.DeletePosition, "employees.positions"},

		// ===== Operations Hub =====
		{"POST   /requests", h.Requests.CreateRequest, "operations-hub.create"},
		{"GET    /requests", h.Requests.GetMyRequests, "operations-hub"},
		{"GET    /requests/assigned", h.Requests.GetAssignedToMe, "operations-hub.view"},
		{"GET    /requests/inbox", h.Requests.GetInbox, "operations-hub.view"},
		{"GET    /requests/{id}", h.Requests.GetRequestByID, "operations-hub.view"},
		{"PUT    /requests/{id}/status", h.Requests.UpdateStatus, "operations-hub.respond"},
		{"POST   /requests/{id}/comments", h.Requests.AddComment, "operations-hub.respond"},

		// ===== Support Center =====
		{"POST   /requests/support", h.Requests.CreateSupportRequest, "support-center.raise"},
		{"GET    /requests/department", h.Requests.GetDepartmentRequests, "support-center"},
		{"PUT    /requests/{id}/accept", h.Requests.AcceptRequest, "support-center.accept"},
		{"POST   /requests/{id}/chat", h.Requests.SendChatMessage, "support-center.chat"},
		{"GET    /requests/{id}/chat", h.Requests.GetChatMessages, "support-center.chat"},
		{"PUT    /requests/{id}/chat/read", h.Requests.MarkChatRead, "support-center.chat"},
		{"GET    /requests/{id}/chat/unread", h.Requests.GetUnreadCount, "support-center.chat"},
		{"GET    /ws/chat/{id}", h.Requests.ChatWebSocket, "support-center.chat"},
		{"GET    /support/departments", h.Requests.GetSupportDepartments, "support-center"},
		{"GET    /support/departments/current/members", h.Requests.GetCurrentDepartmentMembers, "support-center.accept"},
		{"GET    /support/departments/{id}/members", h.Requests.GetSupportTeamMembers, "support-center.accept"},

		// ===== Settings / Bulk Import =====
		{"POST /settings/bulk-import/users", h.Settings.BulkImportUsers, "users.add"},
		{"POST /settings/bulk-import/employees", h.Settings.BulkImportEmployees, "employees.add"},
		{"GET  /settings/bulk-import/users/template", h.Settings.DownloadBulkUsersTemplate, "users.add"},
		{"GET  /settings/bulk-import/employees/template", h.Settings.DownloadBulkEmployeesTemplate, "employees.add"}}

	// Register permission protected routes
	for _, route := range permissionedRoutes {
		if route.Permission != "" {
			mux.HandleFunc(route.Pattern, RequirePermission(route.Permission)(route.Handler))
		} else {
			mux.HandleFunc(route.Pattern, route.Handler)
		}
	}
}

// For route-level permission checks
func RequirePermission(permission string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Get permissions from context
			permissions, ok := r.Context().Value(constants.ContextKeyPermissions).([]string)
			if !ok {
				permissions = []string{}
			}

			// Check permission
			hasPermission := false
			for _, p := range permissions {
				if p == permission {
					hasPermission = true
					break
				}
			}

			// Return error if no permission
			if !hasPermission {
				utils.JsonResponse(w, http.StatusForbidden, map[string]any{
					"message": "Insufficient permissions",
					// "required": permission,
				})
				return
			}

			// Permission granted, continue
			next(w, r)
		}
	}
}
