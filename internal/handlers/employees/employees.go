package employees

import (
	"encoding/json"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"ccms.com/api/internal/database"
	"ccms.com/api/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Handler struct {
	db database.Querier
}

func NewHandler(querier database.Querier) *Handler {
	return &Handler{db: querier}
}

func getTenantID(r *http.Request) (uuid.UUID, bool) {
	tenantID, ok := r.Context().Value("tenantID").(uuid.UUID)
	return tenantID, ok
}

// Helper to convert *float64 to pgtype.Numeric
func toNumeric(f *float64) pgtype.Numeric {
	if f == nil {
		return pgtype.Numeric{Valid: false}
	}
	var num pgtype.Numeric
	err := num.Scan(strconv.FormatFloat(*f, 'f', 2, 64))
	if err != nil {
		return pgtype.Numeric{Valid: false}
	}
	return num
}

// Helper to convert *int32 to pgtype.Int4
func toInt4(i *int32) pgtype.Int4 {
	if i == nil {
		return pgtype.Int4{Valid: false}
	}
	return pgtype.Int4{Int32: *i, Valid: true}
}

func (h *Handler) ListEmployees(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := getTenantID(r)
	if !ok {
		utils.JsonResponse(w, http.StatusUnauthorized, map[string]any{"message": "Unauthorized"})
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit == 0 {
		limit = 10
	}

	employees, err := h.db.ListEmployees(r.Context(), database.ListEmployeesParams{
		TenantID: tenantID,
		Limit:    int32(limit),
		Offset:   int32(offset),
	})
	if err != nil {
		log.Printf("ListEmployees error: %v", err)
		utils.JsonResponse(w, http.StatusInternalServerError, map[string]any{"message": "Failed to retrieve employees"})
		return
	}
	if len(employees) == 0 {
		utils.JsonResponse(w, http.StatusOK, map[string]any{"data": []any{}, "message": "No employees found"})
		return
	}
	utils.JsonResponse(w, http.StatusOK, map[string]any{"data": employees, "message": "Employees retrieved successfully"})
}

func (h *Handler) CreateEmployee(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := getTenantID(r)
	if !ok {
		utils.JsonResponse(w, http.StatusUnauthorized, map[string]any{"message": "Unauthorized"})
		return
	}

	var req struct {
		FirstName             string    `json:"firstName"`
		LastName              string    `json:"lastName"`
		Email                 string    `json:"email"`
		Phone                 *string   `json:"phone,omitempty"`
		Address               *string   `json:"address,omitempty"`
		HireDate              string    `json:"hireDate"`
		PositionID            uuid.UUID `json:"positionId"`
		DepartmentID          uuid.UUID `json:"departmentId"`
		Salary                *float64  `json:"salary,omitempty"`
		EmergencyContactName  *string   `json:"emergencyContactName,omitempty"`
		EmergencyContactPhone *string   `json:"emergencyContactPhone,omitempty"`
		ProfileDescription    *string   `json:"profileDescription,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding JSON: %v", err)
		utils.JsonResponse(w, http.StatusBadRequest, map[string]any{"message": "Invalid request body"})
		return
	}

	if req.FirstName == "" || req.LastName == "" || req.Email == "" {
		utils.JsonResponse(w, http.StatusBadRequest, map[string]any{"message": "First name, last name and email are required"})
		return
	}

	// Parse hire date (now time.Time, not pgtype.Date)
	hireDate, err := time.Parse("2006-01-02", req.HireDate)
	if err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, map[string]any{"message": "Invalid hire date format. Use YYYY-MM-DD"})
		return
	}

	// Convert salary to pgtype.Numeric
	salary := toNumeric(req.Salary)

	employee, err := h.db.CreateEmployee(r.Context(), database.CreateEmployeeParams{
		TenantID:              tenantID,
		UserID:                nil, // Optional, can set if linking to user account
		FirstName:             req.FirstName,
		LastName:              req.LastName,
		Email:                 req.Email,
		Phone:                 req.Phone,
		Address:               req.Address,
		HireDate:              hireDate, // time.Time
		TerminationDate:       nil,      // *time.Time, nil for NULL
		PositionID:            req.PositionID,
		DepartmentID:          req.DepartmentID,
		Salary:                salary, // pgtype.Numeric
		EmergencyContactName:  req.EmergencyContactName,
		EmergencyContactPhone: req.EmergencyContactPhone,
		ProfileDescription:    req.ProfileDescription,
		IsActive:              true,
	})

	if err != nil {
		utils.HandleDBError(w, err, "CreateEmployee")
		return
	}

	utils.JsonResponse(w, http.StatusCreated,
		map[string]any{"data": employee, "message": "Employee created successfully"})
}

func (h *Handler) UpdateEmployee(w http.ResponseWriter, r *http.Request) {
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
		FirstName             string    `json:"firstName"`
		LastName              string    `json:"lastName"`
		Email                 string    `json:"email"`
		Phone                 *string   `json:"phone,omitempty"`
		Address               *string   `json:"address,omitempty"`
		HireDate              string    `json:"hireDate"`
		TerminationDate       *string   `json:"terminationDate,omitempty"`
		PositionID            uuid.UUID `json:"positionId"`
		DepartmentID          uuid.UUID `json:"departmentId"`
		Salary                *float64  `json:"salary,omitempty"`
		EmergencyContactName  *string   `json:"emergencyContactName,omitempty"`
		EmergencyContactPhone *string   `json:"emergencyContactPhone,omitempty"`
		ProfileDescription    *string   `json:"profileDescription,omitempty"`
		IsActive              *bool     `json:"isActive,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, map[string]any{"message": "Invalid request body"})
		return
	}

	// Parse hire date (time.Time)
	hireDate, err := time.Parse("2006-01-02", req.HireDate)
	if err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, map[string]any{"message": "Invalid hire date format"})
		return
	}

	// Parse termination date (*time.Time)
	var terminationDate *time.Time
	if req.TerminationDate != nil && *req.TerminationDate != "" {
		t, err := time.Parse("2006-01-02", *req.TerminationDate)
		if err == nil {
			terminationDate = &t
		}
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	// Convert salary to pgtype.Numeric
	salary := toNumeric(req.Salary)

	employee, err := h.db.UpdateEmployee(r.Context(), database.UpdateEmployeeParams{
		ID:                    id,
		FirstName:             req.FirstName,
		LastName:              req.LastName,
		Email:                 req.Email,
		Phone:                 req.Phone,
		Address:               req.Address,
		HireDate:              hireDate,        // time.Time
		TerminationDate:       terminationDate, // *time.Time
		PositionID:            req.PositionID,
		DepartmentID:          req.DepartmentID,
		Salary:                salary, // pgtype.Numeric
		EmergencyContactName:  req.EmergencyContactName,
		EmergencyContactPhone: req.EmergencyContactPhone,
		ProfileDescription:    req.ProfileDescription,
		IsActive:              isActive,
		TenantID:              tenantID,
	})

	if err != nil {
		utils.HandleDBError(w, err, "UpdateEmployee")
		return
	}

	utils.JsonResponse(w, http.StatusOK,
		map[string]any{"data": employee, "message": "Employee updated successfully"})
}

func (h *Handler) DeleteEmployee(w http.ResponseWriter, r *http.Request) {
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

	err = h.db.DeleteEmployee(r.Context(), database.DeleteEmployeeParams{
		ID:       id,
		TenantID: tenantID,
	})
	if err != nil {
		utils.HandleDBError(w, err, "DeleteEmployee")
		return
	}

	utils.JsonResponse(w, http.StatusOK, map[string]any{"message": "Employee deleted successfully"})
}

func (h *Handler) GetEmployeeById(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := getTenantID(r)
	if !ok {
		utils.JsonResponse(w, http.StatusUnauthorized, map[string]any{"message": "Unauthorized"})
		return
	}

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, map[string]any{"message": "Invalid employee ID"})
		return
	}

	employee, err := h.db.GetEmployeeById(r.Context(), database.GetEmployeeByIdParams{
		ID:       id,
		TenantID: tenantID,
	})
	if err != nil {
		utils.HandleDBError(w, err, "GetEmployeeByID", "Employee not found")
		return
	}

	utils.JsonResponse(w, http.StatusOK, map[string]any{"data": employee, "message": "Employee retrieved successfully"})
}

func (h *Handler) SearchEmployees(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := getTenantID(r)
	if !ok {
		utils.JsonResponse(w, http.StatusUnauthorized, map[string]any{"message": "Unauthorized"})
		return
	}

	queryParams := r.URL.Query()
	limit, _ := strconv.Atoi(queryParams.Get("limit"))
	offset, _ := strconv.Atoi(queryParams.Get("offset"))
	if limit == 0 {
		limit = 20
	}

	search := queryParams.Get("q")

	// Get total count
	totalCount, err := h.db.CountSearchEmployees(r.Context(), database.CountSearchEmployeesParams{
		TenantID: tenantID,
		Search:   search,
	})
	if err != nil {
		log.Printf("Error counting employees: %v", err)
		utils.JsonResponse(w, http.StatusInternalServerError,
			map[string]any{"message": "Failed to get employee count"})
		return
	}

	totalPages := int64(math.Ceil(float64(totalCount) / float64(limit)))

	// Get employees
	employees, err := h.db.SearchEmployees(r.Context(), database.SearchEmployeesParams{
		TenantID:    tenantID,
		LimitCount:  int32(limit),
		OffsetCount: int32(offset),
		Search:      search,
	})
	if err != nil {
		log.Printf("Error searching employees: %v", err)
		utils.JsonResponse(w, http.StatusInternalServerError,
			map[string]any{"message": "Failed to retrieve employee data"})
		return
	}

	response := map[string]any{
		"employees":    employees,
		"total_count":  totalCount,
		"total_pages":  totalPages,
		"current_page": (offset / limit) + 1,
		"size":         limit,
	}

	utils.JsonResponse(w, http.StatusOK, response)
}
