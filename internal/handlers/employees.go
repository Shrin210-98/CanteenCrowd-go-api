package handlers

import (
	"encoding/json"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"ccms.com/api/internal/database"
	"ccms.com/api/internal/utils"
	"github.com/jackc/pgx/v5/pgtype"
)

func (h *Handler) ListEmployees(w http.ResponseWriter, r *http.Request) {
	// Optional pagination parameters
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit == 0 {
		limit = 10 // Default limit
	}

	employees, err := h.db.ListEmployees(r.Context(), database.ListEmployeesParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, map[string]string{"message": "Failed to retrieve employees"})
		return
	}
	if len(employees) == 0 {
		utils.JsonResponse(w, http.StatusOK, map[string]string{"message": "No employees found"})
		return
	}
	utils.JsonResponse(w, http.StatusOK, employees)
}

func (h *Handler) CreateEmployee(w http.ResponseWriter, r *http.Request) {
	var employeeData database.CreateEmployeeParams
	if err := json.NewDecoder(r.Body).Decode(&employeeData); err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, map[string]string{"message": "Invalid request body"})
		return
	}

	// Validate required fields
	if employeeData.FirstName == "" || employeeData.LastName == "" || employeeData.Email == "" {
		utils.JsonResponse(w, http.StatusBadRequest, map[string]string{"message": "First name, last name and email are required"})
		return
	}

	employee, err := h.db.CreateEmployee(r.Context(), employeeData)
	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, map[string]string{"message": "Failed to create employee"})
		return
	}

	utils.JsonResponse(w, http.StatusCreated,
		map[string]any{"data": employee, "message": "Employee created successfully"})
}

func (h *Handler) UpdateEmployee(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, map[string]string{"message": "Invalid employee ID"})
		return
	}

	var employeeData database.UpdateEmployeeParams
	if err := json.NewDecoder(r.Body).Decode(&employeeData); err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, map[string]string{"message": "Invalid request body"})
		return
	}

	// Ensure the ID from path matches the body
	employeeData.ID = int32(id)

	employee, err := h.db.UpdateEmployee(r.Context(), employeeData)
	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, map[string]string{"message": "Failed to update employee"})
		return
	}

	utils.JsonResponse(w, http.StatusOK,
		map[string]any{"data": employee, "message": "Employee updated successfully"})
}

func (h *Handler) DeleteEmployee(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, map[string]string{"message": "Invalid employee ID"})
		return
	}

	err = h.db.DeleteEmployee(r.Context(), int32(id))
	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, map[string]string{"message": "Failed to delete employee"})
		return
	}

	utils.JsonResponse(w, http.StatusOK, map[string]string{"message": "Employee deleted successfully"})
}

func (h *Handler) GetEmployeeById(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, map[string]string{"message": "Invalid employee ID"})
		return
	}

	employee, err := h.db.GetEmployeeById(r.Context(), int32(id))
	if err != nil {
		utils.JsonResponse(w, http.StatusNotFound, map[string]string{"message": "Employee not found"})
		return
	}

	utils.JsonResponse(w, http.StatusOK, employee)
}

func (h *Handler) SearchEmployees(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	queryParams := r.URL.Query()

	// Set up pagination defaults
	limit, _ := strconv.Atoi(queryParams.Get("limit"))
	offset, _ := strconv.Atoi(queryParams.Get("offset"))
	if limit == 0 {
		limit = 20 // Default page size
	}

	// Prepare search filters (same as before for consistency)
	filterParams := database.CountSearchEmployeesParams{
		Column1: pgtype.Text{String: queryParams.Get("q"), Valid: queryParams.Get("q") != ""}, // Name search
		// Populate other filter fields here, similar to your original code
		DepartmentID: func() int32 {
			if id := queryParams.Get("department_id"); id != "" {
				val, _ := strconv.Atoi(id)
				return int32(val)
			}
			return 0 // or appropriate default
		}(),
		PositionID: func() int32 {
			if id := queryParams.Get("position_id"); id != "" {
				val, _ := strconv.Atoi(id)
				return int32(val)
			}
			return 0 // or appropriate default
		}(),
		IsActive: func() pgtype.Bool {
			if val := queryParams.Get("is_active"); val != "" {
				b, _ := strconv.ParseBool(val)
				return pgtype.Bool{Bool: b, Valid: true}
			}
			return pgtype.Bool{Valid: false} // NULL
		}(),
		HireDate: func() pgtype.Date {
			if date := queryParams.Get("hire_date_from"); date != "" {
				t, err := time.Parse("2006-01-02", date)
				if err == nil {
					return pgtype.Date{Time: t, Valid: true}
				}
			}
			return pgtype.Date{Valid: false} // NULL
		}(),
		HireDate_2: func() pgtype.Date {
			if date := queryParams.Get("hire_date_to"); date != "" {
				t, err := time.Parse("2006-01-02", date)
				if err == nil {
					return pgtype.Date{Time: t, Valid: true}
				}
			}
			return pgtype.Date{Valid: false} // NULL
		}(),
	}

	// --- 1. Get total count of employees matching the filters ---
	totalCount, err := h.db.CountSearchEmployees(r.Context(), filterParams)
	if err != nil {
		log.Printf("Error counting employees: %v", err)
		utils.JsonResponse(w, http.StatusInternalServerError,
			map[string]string{"message": "Failed to get employee count"})
		return
	}

	// Calculate total pages
	totalPages := int64(math.Ceil(float64(totalCount) / float64(limit)))

	// --- 2. Get the paginated list of employees ---
	params := database.SearchEmployeesParams{
		Limit:        int32(limit),
		Offset:       int32(offset),
		Column1:      filterParams.Column1,
		DepartmentID: filterParams.DepartmentID,
		PositionID:   filterParams.PositionID,
		IsActive:     filterParams.IsActive,
		HireDate:     filterParams.HireDate,
		HireDate_2:   filterParams.HireDate_2,
	}
	employees, err := h.db.SearchEmployees(r.Context(), params)
	if err != nil {
		log.Printf("Error searching employees: %v", err)
		utils.JsonResponse(w, http.StatusInternalServerError,
			map[string]string{"message": "Failed to retrieve employee data"})
		return
	}

	// Prepare the response
	response := map[string]any{
		"employees":    employees,
		"total_count":  totalCount,
		"total_pages":  totalPages,
		"current_page": (offset / limit) + 1,
		"size":         limit,
	}

	utils.JsonResponse(w, http.StatusOK, response)
}
