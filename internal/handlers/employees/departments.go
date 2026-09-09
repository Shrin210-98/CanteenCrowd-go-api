package employees

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"ccms.com/api/internal/database"
	"ccms.com/api/internal/utils"
	"github.com/google/uuid"
)

func (h *Handler) ListDepartments(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := getTenantID(r)
	if !ok {
		utils.JsonResponse(w, http.StatusUnauthorized, map[string]any{"message": "Unauthorized"})
		return
	}

	departments, err := h.db.ListDepartments(r.Context(), tenantID)
	if err != nil {
		utils.HandleDBError(w, err, "ListDepartments")
		return
	}
	if len(departments) == 0 {
		utils.JsonResponse(w, http.StatusOK, map[string]any{"data": []any{}, "message": "No departments found"})
		return
	}
	utils.JsonResponse(w, http.StatusOK, map[string]any{"data": departments, "message": "Departments retrieved successfully"})
}

func (h *Handler) CreateDepartment(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := getTenantID(r)
	if !ok {
		utils.JsonResponse(w, http.StatusUnauthorized, map[string]any{"message": "Unauthorized"})
		return
	}

	var req struct {
		Name        string  `json:"name"`
		Description *string `json:"description,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("CreateDepartment: Invalid request body: %v", err)
		utils.JsonResponse(w, http.StatusBadRequest, map[string]any{"message": "Invalid Request Body"})
		return
	}

	if req.Name == "" {
		utils.JsonResponse(w, http.StatusBadRequest, map[string]any{"message": "Department name is required"})
		return
	}

	department, err := h.db.CreateDepartment(r.Context(), database.CreateDepartmentParams{
		TenantID:    tenantID,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		utils.HandleDBError(w, err, "CreateDepartment")
		return
	}

	utils.JsonResponse(w, http.StatusCreated, map[string]any{"data": department, "message": "Successfully Created Department"})
}

func (h *Handler) UpdateDepartment(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := getTenantID(r)
	if !ok {
		utils.JsonResponse(w, http.StatusUnauthorized, map[string]any{"message": "Unauthorized"})
		return
	}

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, map[string]any{"message": "Invalid ID"})
		return
	}

	var req struct {
		Name        string  `json:"name"`
		Description *string `json:"description,omitempty"`
		IsActive    *bool   `json:"is_active,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("UpdateDepartment: Invalid request body: %v", err)
		utils.JsonResponse(w, http.StatusBadRequest, map[string]any{"message": "Invalid Request Body"})
		return
	}

	if req.Name == "" {
		utils.JsonResponse(w, http.StatusBadRequest, map[string]any{"message": "Department name is required"})
		return
	}

	department, err := h.db.UpdateDepartment(r.Context(), database.UpdateDepartmentParams{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		TenantID:    tenantID,
	})
	if err != nil {
		utils.HandleDBError(w, err, "UpdateDepartment")
		return
	}

	utils.JsonResponse(w, http.StatusOK, map[string]any{"data": department, "message": "Successfully Updated Department"})
}

func (h *Handler) DeleteDepartment(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := getTenantID(r)
	if !ok {
		utils.JsonResponse(w, http.StatusUnauthorized, map[string]any{"message": "Unauthorized"})
		return
	}

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, map[string]any{"message": "Invalid ID"})
		return
	}

	// Get department first to check if it's a system department
	dept, err := h.db.GetDepartmentByID(r.Context(), database.GetDepartmentByIDParams{
		ID:       id,
		TenantID: tenantID,
	})
	if err != nil {
		utils.HandleDBError(w, err, "GetDepartmentByID", "Department not found")
		return
	}

	// Prevent deletion of system departments
	if dept.IsSystem {
		utils.JsonResponse(w, http.StatusForbidden, map[string]any{
			"message": "Departments created by system cannot be deleted",
		})
		return
	}

	// Check if department has employees
	employeeCount, err := h.db.CountEmployeesInDepartment(r.Context(), database.CountEmployeesInDepartmentParams{
		DepartmentID: id,
		TenantID:     tenantID,
	})
	if err != nil {
		utils.HandleDBError(w, err, "CountEmployeesInDepartment")
		return
	}

	if employeeCount > 0 {
		utils.JsonResponse(w, http.StatusConflict, map[string]any{
			"message": fmt.Sprintf("Cannot delete department with %d employees", employeeCount),
		})
		return
	}

	// Soft delete the department
	deletedDept, err := h.db.DeleteDepartment(r.Context(), database.DeleteDepartmentParams{
		ID:       id,
		TenantID: tenantID,
	})
	if err != nil {
		utils.HandleDBError(w, err, "DeleteDepartment")
		return
	}

	// Optional: Use deletedDept if needed
	_ = deletedDept

	utils.JsonResponse(w, http.StatusOK, map[string]any{
		"message": "Successfully Deleted Department",
	})
}
