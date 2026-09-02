package users

import (
	"net/http"
	"time"

	"ccms.com/api/internal/constants"
	"ccms.com/api/internal/utils"
)

func (h *Handler) GetDefaultPermissionsByUserType(w http.ResponseWriter, r *http.Request) {
	userType := r.URL.Query().Get("userType")

	if userType == "" {
		utils.ErrorResponse(w, http.StatusBadRequest, "userType query parameter is required")
		return
	}

	// Validate user type
	validTypes := map[string]bool{
		constants.UserTypeTenantOwner: true,
		constants.UserTypeStaff:       true,
		constants.UserTypeGuest:       true,
	}

	if !validTypes[userType] {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid userType")
		return
	}

	// utils.ErrorResponse(w, http.StatusBadRequest, "Invalid userType")
	// return

	// Get permissions based on user type
	var permissions []utils.PermissionNode
	switch userType {
	case constants.UserTypeTenantOwner:
		permissions = constants.DefaultTenantOwnerPermissions()
	case constants.UserTypeStaff:
		permissions = constants.DefaultStaffPermissions()
	case constants.UserTypeGuest:
		permissions = constants.DefaultGuestPermissions()
	default:
		permissions = constants.DefaultGuestPermissions()
	}

	time.Sleep(3 * time.Second)

	utils.JsonResponse(w, http.StatusOK, permissions)
}
