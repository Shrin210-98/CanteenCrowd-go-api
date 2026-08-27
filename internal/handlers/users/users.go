package users

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"ccms.com/api/internal/constants"
	"ccms.com/api/internal/database"
	"ccms.com/api/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	db   database.Querier
	pool *pgxpool.Pool
}

func NewHandler(db database.Querier, pool *pgxpool.Pool) *Handler {
	return &Handler{
		db:   db,
		pool: pool,
	}
}

func getTenantID(r *http.Request) (uuid.UUID, bool) {
	tenantID, ok := r.Context().Value(constants.ContextKeyTenantID).(uuid.UUID)
	return tenantID, ok
}

type CreateUserRequest struct {
	EmployeeID *uuid.UUID `json:"employeeId,omitempty" validate:"required_if=UserType staff,uuid"`
	Username   string     `json:"username" validate:"required,min=3,max=50"`
	Email      string     `json:"email" validate:"required,email"`
	Password   string     `json:"password" validate:"required,min=8,max=100"`
	UserType   string     `json:"userType" validate:"required,oneof=staff guest"`
}

type UpdateUserRequest struct {
	Username *string `json:"username,omitempty" validate:"omitempty,min=3,max=50"`
	Email    *string `json:"email,omitempty" validate:"omitempty,email"`
	IsLocked *bool   `json:"isLocked,omitempty"`
}

type UserResponse struct {
	ID            uuid.UUID  `json:"id"`
	Username      string     `json:"username"`
	Email         string     `json:"email"`
	UserType      string     `json:"userType"`
	TenantID      uuid.UUID  `json:"tenantId"`
	IsLocked      bool       `json:"isLocked"`
	EmailVerified bool       `json:"emailVerified"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
	EmployeeID    *uuid.UUID `json:"employeeId,omitempty"`
	EmployeeName  *string    `json:"employeeName,omitempty"`
}

type EmployeeEligibleResponse struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Email     string    `json:"email"`
	FullName  string    `json:"fullName"`
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := getTenantID(r)
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Check authorization
	userType, _ := r.Context().Value("userType").(string)
	if userType != "tenant_owner" {
		utils.ErrorResponse(w, http.StatusForbidden, "Only tenant owner can create users")
		return
	}

	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("CreateUser: Invalid request body: %v", err)
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if validationErrors := utils.ValidateStruct(req); validationErrors != nil {
		utils.ValidationErrorResponse(w, validationErrors)
		return
	}

	// Handle staff user requirements
	if req.UserType == "staff" {
		if req.EmployeeID == nil {
			utils.ErrorResponse(w, http.StatusBadRequest, "employeeId is required for staff users")
			return
		}

		// Verify employee exists and belongs to tenant
		employee, err := h.db.GetEmployeeById(r.Context(), database.GetEmployeeByIdParams{
			ID:       *req.EmployeeID,
			TenantID: tenantID,
		})
		if err != nil {
			utils.HandleDBError(w, err, "GetEmployeeByID", "Employee not found")
			return
		}

		// Check if employee already has user account
		if employee.UserID != nil {
			utils.ErrorResponse(w, http.StatusConflict, "Employee already has a user account")
			return
		}

		// Optionally auto-fill email from employee if not provided
		// if req.Email == "" {
		//     req.Email = employee.Email
		// }
	}

	// Handle guest user (should not have employeeId)
	if req.UserType == "guest" && req.EmployeeID != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "employeeId should not be provided for guest users")
		return
	}

	// Check username availability
	usernameExists, err := h.db.CheckUsernameExists(r.Context(), req.Username)
	if err != nil {
		utils.HandleDBError(w, err, "CheckUsernameExists")
		return
	}
	if usernameExists {
		utils.ErrorResponse(w, http.StatusConflict, "Username already exists")
		return
	}

	// Check email availability
	emailExists, err := h.db.CheckEmailExists(r.Context(), req.Email)
	if err != nil {
		utils.HandleDBError(w, err, "CheckEmailExists")
		return
	}
	if emailExists {
		utils.ErrorResponse(w, http.StatusConflict, "Email already exists")
		return
	}

	// Hash password
	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		log.Printf("CreateUser: Failed to hash password: %v", err)
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to process password")
		return
	}

	// Start transaction
	tx, err := h.pool.Begin(r.Context())
	if err != nil {
		utils.HandleDBError(w, err, "BeginTransaction")
		return
	}
	defer tx.Rollback(r.Context())

	qtx := database.New(h.pool).WithTx(tx)

	// Create user
	user, err := qtx.CreateUser(r.Context(), database.CreateUserParams{
		TenantID:     tenantID,
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: passwordHash,
		UserType:     req.UserType,
	})
	if err != nil {
		utils.HandleDBError(w, err, "CreateUser")
		return
	}

	// Link to employee for staff users
	if req.UserType == "staff" && req.EmployeeID != nil {
		_, err = qtx.LinkUserToEmployee(r.Context(), database.LinkUserToEmployeeParams{
			ID:       *req.EmployeeID,
			UserID:   &user.ID,
			TenantID: tenantID,
		})
		if err != nil {
			utils.HandleDBError(w, err, "LinkUserToEmployee")
			return
		}
	}

	// Commit transaction
	if err := tx.Commit(r.Context()); err != nil {
		utils.HandleDBError(w, err, "CommitTransaction")
		return
	}

	// Prepare response
	response := UserResponse{
		ID:            user.ID,
		Username:      user.Username,
		Email:         user.Email,
		UserType:      user.UserType,
		TenantID:      user.TenantID,
		IsLocked:      user.IsLocked,
		EmailVerified: user.EmailVerified,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}

	if req.EmployeeID != nil {
		response.EmployeeID = req.EmployeeID
	}

	utils.SuccessResponse(w, http.StatusCreated, response, "User created successfully")
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := getTenantID(r)
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	queryParams := r.URL.Query()

	// Parse pagination
	page := 1
	pageSize := 20
	if pageStr := queryParams.Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if sizeStr := queryParams.Get("pageSize"); sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 && s <= 100 {
			pageSize = s
		}
	}

	offset := (page - 1) * pageSize

	// Parse filters
	search := queryParams.Get("search")
	userType := queryParams.Get("userType")

	// Get total count
	totalCount, err := h.db.CountUsersWithFilters(r.Context(), database.CountUsersWithFiltersParams{
		TenantID: tenantID,
		Search:   search,
		UserType: userType,
	})
	if err != nil {
		log.Printf("Error counting users: %v", err)
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to get user count")
		return
	}

	// Get users with filters
	users, err := h.db.ListUsersWithFilters(r.Context(), database.ListUsersWithFiltersParams{
		TenantID:    tenantID,
		Search:      search,
		UserType:    userType,
		LimitCount:  int32(pageSize),
		OffsetCount: int32(offset),
	})
	if err != nil {
		log.Printf("Error listing users: %v", err)
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve user data")
		return
	}

	utils.PaginatedResponse(w, http.StatusOK, users, "Users retrieved successfully", page, pageSize, int(totalCount))
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := getTenantID(r)
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	user, err := h.db.GetUserByTenantAndID(r.Context(), database.GetUserByTenantAndIDParams{
		ID:       userID,
		TenantID: tenantID,
	})
	if err != nil {
		utils.HandleDBError(w, err, "GetUserByID", "User not found")
		return
	}

	// Build response with employee info for staff
	response := UserResponse{
		ID:            user.ID,
		Username:      user.Username,
		Email:         user.Email,
		UserType:      user.UserType,
		TenantID:      user.TenantID,
		IsLocked:      user.IsLocked,
		EmailVerified: user.EmailVerified,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}

	// Add employee info for staff users
	if user.UserType == "staff" {
		employee, err := h.db.GetEmployeeByUserID(r.Context(), database.GetEmployeeByUserIDParams{
			UserID:   &user.ID,
			TenantID: tenantID,
		})
		if err == nil && employee.ID != uuid.Nil {
			response.EmployeeID = &employee.ID
			fullName := employee.FirstName + " " + employee.LastName
			response.EmployeeName = &fullName
		}
	}

	utils.SuccessResponse(w, http.StatusOK, response, "User retrieved successfully")
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := getTenantID(r)
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if validationErrors := utils.ValidateStruct(req); validationErrors != nil {
		utils.ValidationErrorResponse(w, validationErrors)
		return
	}

	// Get current user
	currentUser, err := h.db.GetUserByTenantAndID(r.Context(), database.GetUserByTenantAndIDParams{
		ID:       userID,
		TenantID: tenantID,
	})
	if err != nil {
		utils.HandleDBError(w, err, "GetUserByID", "User not found")
		return
	}

	// Check username if provided
	if req.Username != nil && *req.Username != currentUser.Username {
		exists, err := h.db.CheckUsernameExists(r.Context(), *req.Username)
		if err != nil {
			utils.HandleDBError(w, err, "CheckUsernameExists")
			return
		}
		if exists {
			utils.ErrorResponse(w, http.StatusConflict, "Username already exists")
			return
		}
	}

	// Check email if provided
	if req.Email != nil && *req.Email != currentUser.Email {
		exists, err := h.db.CheckEmailExists(r.Context(), *req.Email)
		if err != nil {
			utils.HandleDBError(w, err, "CheckEmailExists")
			return
		}
		if exists {
			utils.ErrorResponse(w, http.StatusConflict, "Email already exists")
			return
		}
	}

	// Build update params
	updateParams := database.UpdateUserParams{
		ID:                 userID,
		Username:           currentUser.Username,
		Email:              currentUser.Email,
		LastLogin:          currentUser.LastLogin,
		LoginAttempts:      currentUser.LoginAttempts,
		IsLocked:           currentUser.IsLocked,
		LockedUntil:        currentUser.LockedUntil,
		PasswordResetToken: currentUser.PasswordResetToken,
		TokenExpiry:        currentUser.TokenExpiry,
		MustChangePassword: currentUser.MustChangePassword,
		EmailVerified:      currentUser.EmailVerified,
		MfaEnabled:         currentUser.MfaEnabled,
		LastPasswordChange: currentUser.LastPasswordChange,
		UserType:           currentUser.UserType,
		TenantID:           currentUser.TenantID,
	}

	if req.Username != nil {
		updateParams.Username = *req.Username
	}
	if req.Email != nil {
		updateParams.Email = *req.Email
		updateParams.EmailVerified = false // Require re-verification
	}
	if req.IsLocked != nil {
		updateParams.IsLocked = *req.IsLocked
		if *req.IsLocked {
			updateParams.LoginAttempts = 0
		}
	}

	updatedUser, err := h.db.UpdateUser(r.Context(), updateParams)
	if err != nil {
		utils.HandleDBError(w, err, "UpdateUser")
		return
	}

	utils.SuccessResponse(w, http.StatusOK, updatedUser, "User updated successfully")
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := getTenantID(r)
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Get user
	user, err := h.db.GetUserByTenantAndID(r.Context(), database.GetUserByTenantAndIDParams{
		ID:       userID,
		TenantID: tenantID,
	})
	if err != nil {
		utils.HandleDBError(w, err, "GetUserByID", "User not found")
		return
	}

	// Prevent deleting tenant_owner
	if user.UserType == "tenant_owner" {
		utils.ErrorResponse(w, http.StatusForbidden, "Cannot delete tenant owner account")
		return
	}

	// Unlink from employee if staff
	if user.UserType == "staff" {
		employee, err := h.db.GetEmployeeByUserID(r.Context(), database.GetEmployeeByUserIDParams{
			UserID:   &user.ID,
			TenantID: tenantID,
		})
		if err == nil && employee.ID != uuid.Nil {
			err = h.db.UnlinkUserFromEmployee(r.Context(), database.UnlinkUserFromEmployeeParams{
				ID:       employee.ID,
				TenantID: tenantID,
			})
			if err != nil {
				utils.HandleDBError(w, err, "UnlinkUserFromEmployee")
				return
			}
		}
	}

	// Soft delete user
	err = h.db.SoftDeleteUser(r.Context(), database.SoftDeleteUserParams{
		ID:       userID,
		TenantID: tenantID,
	})
	if err != nil {
		utils.HandleDBError(w, err, "SoftDeleteUser")
		return
	}

	utils.SuccessResponse(w, http.StatusOK, nil, "User deleted successfully")
}

func (h *Handler) GetEligibleEmployees(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := getTenantID(r)
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	employees, err := h.db.ListEmployeesEligibleForUserCreation(r.Context(), tenantID)
	if err != nil {
		utils.HandleDBError(w, err, "ListEligibleEmployees")
		return
	}

	// Convert to response format
	response := make([]EmployeeEligibleResponse, 0, len(employees))
	for _, emp := range employees {
		response = append(response, EmployeeEligibleResponse{
			ID:        emp.ID,
			FirstName: emp.FirstName,
			LastName:  emp.LastName,
			Email:     emp.Email,
			FullName:  emp.FirstName + " " + emp.LastName,
		})
	}

	utils.SuccessResponse(w, http.StatusOK, response, "Eligible employees retrieved successfully")
}
