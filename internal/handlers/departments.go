package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"ccms.com/api/internal/database"
	"ccms.com/api/internal/utils"
)

func (h *Handler) ListDepartments(w http.ResponseWriter, r *http.Request) {
	departments, err := h.db.ListDepartments(r.Context())
	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, map[string]string{"message": "Failed to retrieve Departments"})
		return
	}
	if len(departments) == 0 {
		utils.JsonResponse(w, http.StatusOK, map[string]string{"message": "No departments found"})
		return
	}
	utils.JsonResponse(w, http.StatusOK, departments)
}

func (h *Handler) CreateDepartment(w http.ResponseWriter, r *http.Request) {
	var departmentData database.CreateDepartmentParams
	if err := json.NewDecoder(r.Body).Decode(&departmentData); err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, map[string]string{"message": "Invalid Request Body"})
		return
	}

	author, err := h.db.CreateDepartment(r.Context(), departmentData)
	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, map[string]string{"message": "Failed to create Department"})
		return
	}

	utils.JsonResponse(w, http.StatusCreated,
		map[string]any{"data": author, "message": "Successfully Created Department"})
}

func (h *Handler) UpdateDepartment(w http.ResponseWriter, r *http.Request) {
	var departmentData database.UpdateDepartmentParams
	if err := json.NewDecoder(r.Body).Decode(&departmentData); err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, map[string]string{"message": "Invalid Request Body"})
		return
	}

	err := h.db.UpdateDepartment(r.Context(), departmentData)
	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, map[string]string{"message": "Department doesn't exist"})
		return
	}

	utils.JsonResponse(w, http.StatusOK, map[string]string{"message": "Successfully Updated Department"})
}

func (h *Handler) DeleteDepartment(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, map[string]string{"message": "Invalid ID"})
		return
	}

	err = h.db.DeleteDepartment(r.Context(), int32(id))
	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, map[string]string{"message": "Department doesn't exist"})
		return
	}
	utils.JsonResponse(w, http.StatusOK, map[string]string{"message": "Successfully Deleted Departmnent"})
}
