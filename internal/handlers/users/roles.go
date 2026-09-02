package users

import (
	"encoding/json"
	"fmt"
	"net/http"

	"ccms.com/api/internal/constants"
	"ccms.com/api/internal/database"
	"ccms.com/api/internal/utils"
	"github.com/google/uuid"
)

func (h *Handler) CreateRole(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := getTenantID(r)
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Check authorization - only tenant_owner can create roles
	userType, _ := r.Context().Value(constants.ContextKeyUserType).(string)
	if userType != constants.UserTypeTenantOwner {
		utils.ErrorResponse(w, http.StatusForbidden, "Only tenant owner can create roles")
		return
	}

	var req struct {
		Name        string          `json:"name" validate:"required,min=2,max=50"`
		Description *string         `json:"description,omitempty" validate:"omitempty,max=200"`
		Permissions json.RawMessage `json:"permissions" validate:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if validationErrors := utils.ValidateStruct(req); validationErrors != nil {
		utils.ValidationErrorResponse(w, validationErrors)
		return
	}

	// Check if role name already exists in tenant
	existingRole, err := h.db.GetRoleByTenantAndName(r.Context(), database.GetRoleByTenantAndNameParams{
		TenantID: tenantID,
		Name:     req.Name,
	})
	if err == nil && existingRole.ID != uuid.Nil {
		utils.ErrorResponse(w, http.StatusConflict, "Role with this name already exists")
		return
	}

	// Create role
	role, err := h.db.CreateRole(r.Context(), database.CreateRoleParams{
		TenantID:    tenantID,
		Name:        req.Name,
		Description: req.Description,
		Permissions: req.Permissions,
	})
	if err != nil {
		utils.HandleDBError(w, err, "CreateRole")
		return
	}

	utils.SuccessResponse(w, http.StatusCreated, role, "Role created successfully")
}

func (h *Handler) ListRoles(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := getTenantID(r)
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	roles, err := h.db.ListRolesByTenant(r.Context(), tenantID)
	if err != nil {
		utils.HandleDBError(w, err, "ListRolesByTenant")
		return
	}

	utils.SuccessResponse(w, http.StatusOK, roles, "Roles retrieved successfully")
}

func (h *Handler) GetRoleById(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := getTenantID(r)
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	roleID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid role ID")
		return
	}

	role, err := h.db.GetRoleByID(r.Context(), roleID)
	if err != nil {
		utils.HandleDBError(w, err, "GetRoleByID", "Role not found")
		return
	}

	// Verify role belongs to tenant
	if role.TenantID != tenantID {
		utils.ErrorResponse(w, http.StatusNotFound, "Role not found")
		return
	}

	utils.SuccessResponse(w, http.StatusOK, role, "Role retrieved successfully")
}

func (h *Handler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := getTenantID(r)
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Check authorization - only tenant_owner can delete roles
	userType, _ := r.Context().Value(constants.ContextKeyUserType).(string)
	if userType != constants.UserTypeTenantOwner {
		utils.ErrorResponse(w, http.StatusForbidden, "Only tenant owner can delete roles")
		return
	}

	roleID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid role ID")
		return
	}

	// Check if role exists and belongs to tenant
	role, err := h.db.GetRoleByID(r.Context(), roleID)
	if err != nil {
		utils.HandleDBError(w, err, "GetRoleByID", "Role not found")
		return
	}

	if role.TenantID != tenantID {
		utils.ErrorResponse(w, http.StatusNotFound, "Role not found")
		return
	}

	// Check if users are assigned to this role
	usersCount, err := h.db.CountUsersByRole(r.Context(), database.CountUsersByRoleParams{
		RoleID:   &roleID,
		TenantID: tenantID,
	})
	if err != nil {
		utils.HandleDBError(w, err, "CountUsersByRole")
		return
	}

	if usersCount > 0 {
		utils.ErrorResponse(w, http.StatusConflict, fmt.Sprintf("Cannot delete role. %d users are assigned to this role", usersCount))
		return
	}

	// Delete role
	err = h.db.DeleteRole(r.Context(), database.DeleteRoleParams{
		ID:       roleID,
		TenantID: tenantID,
	})
	if err != nil {
		utils.HandleDBError(w, err, "DeleteRole")
		return
	}

	utils.SuccessResponse(w, http.StatusOK, nil, "Role deleted successfully")
}

// UpdateRole handles PUT /roles/{id}
func (h *Handler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := getTenantID(r)
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	roleID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid role ID")
		return
	}

	var req struct {
		Name        *string         `json:"name,omitempty"`
		Description *string         `json:"description,omitempty"`
		Permissions json.RawMessage `json:"permissions,omitempty"`
		ApplyToAll  *bool           `json:"applyToAll,omitempty"` // Force apply to all users
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get users with overrides for this role
	usersWithOverrides, err := h.db.GetUsersWithOverridesForRole(r.Context(),
		database.GetUsersWithOverridesForRoleParams{
			RoleID:   &roleID,
			TenantID: tenantID,
		})
	if err != nil {
		utils.HandleDBError(w, err, "GetUsersWithOverridesForRole")
		return
	}

	// Update role
	role, err := h.db.UpdateRole(r.Context(), database.UpdateRoleParams{
		ID:          roleID,
		TenantID:    tenantID,
		Name:        *req.Name,
		Description: req.Description,
		Permissions: req.Permissions,
	})
	if err != nil {
		utils.HandleDBError(w, err, "UpdateRole")
		return
	}

	// If applyToAll is true, clear all overrides
	if req.ApplyToAll != nil && *req.ApplyToAll {
		err = h.db.ClearOverridesForRole(r.Context(), database.ClearOverridesForRoleParams{
			RoleID:   &roleID,
			TenantID: tenantID,
		})
		if err != nil {
			utils.HandleDBError(w, err, "ClearOverridesForRole")
			return
		}
	}

	// Response with warning about affected users
	response := map[string]any{
		"role":               role,
		"usersWithOverrides": len(usersWithOverrides),
		"overridesCleared":   req.ApplyToAll != nil && *req.ApplyToAll,
	}

	if len(usersWithOverrides) > 0 && (req.ApplyToAll == nil || !*req.ApplyToAll) {
		response["warning"] = fmt.Sprintf("%d users have custom overrides that will be preserved", len(usersWithOverrides))
	}

	utils.SuccessResponse(w, http.StatusOK, response, "Role updated successfully")
}

// {
//     "data": {
//         "role": {
//             "id": "role-uuid",
//             "name": "Staff",
//             "permissions": [...]
//         },
//         "usersWithOverrides": 5,
//         "overridesCleared": false,
//         "warning": "5 users have custom overrides that will be preserved"
//     },
//     "message": "Role updated successfully"
// }
