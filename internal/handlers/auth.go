package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"ccms.com/api/internal/constants"
	"ccms.com/api/internal/database"
	"ccms.com/api/internal/utils"
	"github.com/google/uuid"
	// "github.com/jackc/pgx/v5/pgtype"
)

func getTenantID(r *http.Request) (uuid.UUID, bool) {
	tenantID, ok := r.Context().Value(constants.ContextKeyTenantID).(uuid.UUID)
	return tenantID, ok
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username   string `json:"username" validate:"required,min=3,max=50"`
		Email      string `json:"email" validate:"required,email"`
		Password   string `json:"password" validate:"required,min=8"`
		TenantName string `json:"tenantName" validate:"required,min=2,max=100"`
		TenantSlug string `json:"tenantSlug" validate:"required,min=2,max=50"`
		TenantPlan string `json:"tenantPlan,omitempty" validate:"omitempty,oneof=free basic premium enterprise"`
		FullName   string `json:"fullName,omitempty" validate:"omitempty,min=2,max=100"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request")
		return
	}

	// Validate request
	if validationErrors := utils.ValidateStruct(req); validationErrors != nil {
		utils.ValidationErrorResponse(w, validationErrors)
		return
	}

	// Set default plan
	if req.TenantPlan == "" {
		req.TenantPlan = constants.DefaultTenantPlan
	}

	// Check if username or email already exists
	usernameExists, err := h.db.CheckUsernameExists(r.Context(), req.Username)
	if err != nil {
		utils.HandleDBError(w, err, "CheckUsernameExists")
		return
	}
	if usernameExists {
		utils.ErrorResponse(w, http.StatusConflict, "Username already exists")
		return
	}

	emailExists, err := h.db.CheckEmailExists(r.Context(), req.Email)
	if err != nil {
		utils.HandleDBError(w, err, "CheckEmailExists")
		return
	}
	if emailExists {
		utils.ErrorResponse(w, http.StatusConflict, "Email already exists")
		return
	}

	// Check if tenant slug exists
	slugExists, err := h.db.CheckTenantSlugExists(r.Context(), req.TenantSlug)
	if err != nil {
		utils.HandleDBError(w, err, "CheckTenantSlugExists")
		return
	}
	if slugExists {
		utils.ErrorResponse(w, http.StatusConflict, "Tenant slug already exists")
		return
	}

	if h.Pool == nil {
		utils.ErrorResponse(w, http.StatusServiceUnavailable, "Database not available")
		return
	}

	// Start transaction
	tx, err := h.Pool.Begin(r.Context())
	if err != nil {
		utils.HandleDBError(w, err, "BeginTransaction")
		return
	}
	defer tx.Rollback(r.Context())

	// Create queries with transaction
	qtx := database.New(h.Pool).WithTx(tx)

	// Create tenant
	tenant, err := qtx.CreateTenant(r.Context(), database.CreateTenantParams{
		Name:     req.TenantName,
		Slug:     req.TenantSlug,
		Plan:     req.TenantPlan,
		Status:   constants.DefaultTenantStatus,
		Settings: json.RawMessage(`{}`),
	})
	if err != nil {
		utils.HandleDBError(w, err, "CreateTenant")
		return
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	defaultPermissions := utils.MustToJSON(constants.DefaultTenantOwnerPermissions())

	adminRoleDescription := "Default administrator role with full permissions"
	role, err := qtx.CreateRole(r.Context(), database.CreateRoleParams{
		TenantID:    tenant.ID,
		Name:        "Administrator",
		Description: &adminRoleDescription,
		Permissions: defaultPermissions,
	})
	if err != nil {
		utils.HandleDBError(w, err, "CreateRole")
		return
	}

	user, err := qtx.CreateUser(r.Context(), database.CreateUserParams{
		TenantID:            tenant.ID,
		Username:            req.Username,
		Email:               req.Email,
		PasswordHash:        hashedPassword,
		UserType:            constants.UserTypeTenantOwner,
		RoleID:              &role.ID,
		PermissionsOverride: nil,
	})
	if err != nil {
		utils.HandleDBError(w, err, "CreateUser")
		return
	}

	var fullName *string
	if req.FullName != "" {
		fullName = &req.FullName
	}

	profile, err := qtx.CreateUserProfile(r.Context(), database.CreateUserProfileParams{
		UserID:   user.ID,
		TenantID: tenant.ID,
		FullName: fullName,
	})
	if err != nil {
		utils.HandleDBError(w, err, "CreateUserProfile")
		return
	}

	// Commit transaction
	if err := tx.Commit(r.Context()); err != nil {
		utils.HandleDBError(w, err, "CommitTransaction")
		return
	}

	// Generate JWT token
	token, err := utils.GenerateTokens(user.ID, tenant.ID, user.UserType, h.jwtSecret)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	responseData := map[string]any{
		"token":     token,
		"expiresIn": int(constants.TokenExpiry.Seconds()),
		"user": map[string]any{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"userType": user.UserType,
			"tenantId": user.TenantID,
			"roleId":   user.RoleID,
		},
		"tenant": map[string]any{
			"id":   tenant.ID,
			"name": tenant.Name,
			"slug": tenant.Slug,
			"plan": tenant.Plan,
		},
		"role": map[string]any{
			"id":        role.ID,
			"name":      role.Name,
			"tenantId":  role.TenantID,
			"isDefault": role.IsDefault,
		},
		"profile": map[string]any{
			"id":       profile.ID,
			"userId":   profile.UserID,
			"tenantId": profile.TenantID,
			"fullName": profile.FullName,
		},
	}

	utils.SuccessResponse(w, http.StatusCreated, responseData, "Registration successful")
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=8"`
	}

	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid Request")
		return
	}

	// Validate input
	if validationErrors := utils.ValidateStruct(creds); validationErrors != nil {
		utils.ValidationErrorResponse(w, validationErrors)
		return
	}

	user, err := h.db.GetUserByEmail(r.Context(), creds.Email)
	if err != nil {
		log.Printf("Login failed - user not found: %v", err)
		utils.ErrorResponse(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Check if account is locked
	if user.IsLocked {
		if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
			log.Printf("Login failed - account locked for user: %s", user.Email)
			utils.ErrorResponse(w, http.StatusForbidden, "Account is locked. Please try again later.")
			return
		}
		// If lock has expired, reset lock status
		_, err = h.db.UpdateLoginAttempts(r.Context(), database.UpdateLoginAttemptsParams{
			ID:            user.ID,
			LoginAttempts: 0,
			IsLocked:      false,
			LockedUntil:   nil,
		})
		if err != nil {
			log.Printf("Failed to reset login attempts: %v", err)
		}
	}

	// Check password
	if !utils.CheckPasswordHash(creds.Password, user.PasswordHash) {
		// Increment login attempts
		newAttempts := user.LoginAttempts + 1
		isLocked := false
		var lockedUntil *time.Time

		// Lock account after 5 failed attempts
		if newAttempts >= constants.MaxLoginAttempts {
			lockTime := time.Now().Add(constants.LockoutDuration)
			lockedUntil = &lockTime
			log.Printf("Account locked for user: %s after %d failed attempts", user.Email, newAttempts)
		}

		_, err = h.db.UpdateLoginAttempts(r.Context(), database.UpdateLoginAttemptsParams{
			ID:            user.ID,
			LoginAttempts: newAttempts,
			IsLocked:      isLocked,
			LockedUntil:   lockedUntil,
		})
		if err != nil {
			log.Printf("Failed to update login attempts: %v", err)
		}

		attemptsLeft := constants.MaxLoginAttempts - newAttempts
		if attemptsLeft < 0 {
			attemptsLeft = 0
		}

		utils.ErrorResponse(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Reset login attempts on successful login
	_, err = h.db.UpdateLoginAttempts(r.Context(), database.UpdateLoginAttemptsParams{
		ID:            user.ID,
		LoginAttempts: 0,
		IsLocked:      false,
		LockedUntil:   nil,
	})
	if err != nil {
		log.Printf("Failed to reset login attempts: %v", err)
	}

	// Update last login
	now := time.Now()
	_, err = h.db.UpdateUser(r.Context(), database.UpdateUserParams{
		ID:                  user.ID,
		Username:            user.Username,
		Email:               user.Email,
		LastLogin:           &now,
		LoginAttempts:       0,
		IsLocked:            false,
		LockedUntil:         nil,
		PasswordResetToken:  user.PasswordResetToken,
		TokenExpiry:         user.TokenExpiry,
		MustChangePassword:  user.MustChangePassword,
		EmailVerified:       user.EmailVerified,
		MfaEnabled:          user.MfaEnabled,
		LastPasswordChange:  user.LastPasswordChange,
		UserType:            user.UserType,
		TenantID:            user.TenantID,
		RoleID:              user.RoleID,
		PermissionsOverride: user.PermissionsOverride,
	})
	if err != nil {
		log.Printf("Failed to update last login: %v", err)
		// Continue anyway - this is not critical
	}

	// Generate JWT token
	token, err := utils.GenerateTokens(user.ID, user.TenantID, user.UserType, h.jwtSecret)
	if err != nil {
		log.Printf("Failed to generate token: %v", err)
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	userWithRole, err := h.db.GetUserWithRoleAndPermissions(r.Context(), database.GetUserWithRoleAndPermissionsParams{
		ID:       user.ID,
		TenantID: user.TenantID,
	})
	if err != nil {
		log.Printf("Failed to get user role: %v", err)
		// Continue without role info - not critical
	}

	responseData := map[string]any{
		"token":     token,
		"tokenType": constants.TokenType,
		"expiresIn": int(constants.TokenExpiry.Seconds()),
		"user": map[string]any{
			"id":       user.ID,
			"tenantId": user.TenantID,
			"username": user.Username,
			"email":    user.Email,
			"userType": user.UserType,
		},
	}

	if userWithRole.ID != uuid.Nil {
		responseData["user"].(map[string]any)["roleId"] = userWithRole.RoleID
		responseData["user"].(map[string]any)["roleName"] = userWithRole.RoleName
		responseData["user"].(map[string]any)["permissions"] = userWithRole.EffectivePermissions
		responseData["user"].(map[string]any)["hasPermissionsOverride"] = userWithRole.HasPermissionsOverride
	}

	// Include mustChangePassword flag if needed
	if user.MustChangePassword {
		responseData["mustChangePassword"] = true
	}

	utils.SuccessResponse(w, http.StatusOK, responseData, "Login successful")
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	// Since JWT is stateless, logout is primarily client-side
	// The client should discard the token

	userID, ok := r.Context().Value(constants.ContextKeyUserID).(uuid.UUID)
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Token versioning to invalidate token
	// Increment token version
	// db.Update("UPDATE users SET token_version = token_version + 1 WHERE id = $1", userID)
	// All existing tokens for this user are now invalid
	// Old tokens have version 5, but DB now has version 6

	// Log the logout event
	log.Printf("User %s logged out", userID)

	// Return success - client will remove token
	utils.SuccessResponse(w, http.StatusOK, map[string]any{"loggedOut": true}, "Logout successful")
}

func (h *Handler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := getTenantID(r)
	if !ok {
		utils.JsonResponse(w, http.StatusUnauthorized, map[string]any{"message": "Unauthorized"})
		return
	}

	// Get user ID from context
	userID, ok := r.Context().Value(constants.ContextKeyUserID).(uuid.UUID)
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// >>> CHANGED: Get user with role and permissions instead of GetUserByID <<<
	userWithRole, err := h.db.GetUserWithRoleAndPermissions(r.Context(), database.GetUserWithRoleAndPermissionsParams{
		ID:       userID,
		TenantID: tenantID,
	})
	if err != nil {
		utils.HandleDBError(w, err, "GetUserWithRoleAndPermissions", "User not found")
		return
	}

	// Build full user response
	response := map[string]any{
		"id":            userWithRole.ID,
		"tenantId":      userWithRole.TenantID,
		"username":      userWithRole.Username,
		"email":         userWithRole.Email,
		"userType":      userWithRole.UserType,
		"emailVerified": userWithRole.EmailVerified,
		"mfaEnabled":    userWithRole.MfaEnabled,
		"lastLogin":     userWithRole.LastLogin,
		"createdAt":     userWithRole.CreatedAt,
		"updatedAt":     userWithRole.UpdatedAt,
	}

	if userWithRole.RoleID != nil {
		response["roleId"] = *userWithRole.RoleID
		response["roleName"] = userWithRole.RoleName
		response["hasPermissionsOverride"] = userWithRole.HasPermissionsOverride
	}

	if userWithRole.EffectivePermissions != nil {
		filteredPermissions, err := utils.FilterPermissionsFromJSON(userWithRole.EffectivePermissions)
		if err == nil {
			// Send as flat array
			flatPermissions := utils.FlattenPermissions(filteredPermissions)
			response["permissions"] = flatPermissions
		}
	}

	// Add employee info for staff users
	if userWithRole.UserType == constants.UserTypeStaff { // >>> CHANGED: Use constant <<<
		employee, err := h.db.GetEmployeeByUserID(r.Context(), database.GetEmployeeByUserIDParams{
			UserID:   &userWithRole.ID,
			TenantID: userWithRole.TenantID,
		})
		if err == nil && employee.ID != uuid.Nil {
			response["employeeId"] = employee.ID
			response["employeeName"] = employee.FirstName + " " + employee.LastName
		}
	}

	// Add profile info if exists
	profile, err := h.db.GetUserProfile(r.Context(), userWithRole.ID)
	if err == nil && profile.ID != uuid.Nil {
		if profile.FullName != nil {
			response["fullName"] = *profile.FullName
		}
		if profile.Phone != nil {
			response["phone"] = *profile.Phone
		}
		if profile.AvatarUrl != nil {
			response["avatarUrl"] = *profile.AvatarUrl
		}
		if profile.Timezone != nil {
			response["timezone"] = *profile.Timezone
		}
	}

	utils.SuccessResponse(w, http.StatusOK, response, "Profile retrieved successfully")
}

func (h *Handler) UpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := getTenantID(r)
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get user ID from context
	userID, ok := r.Context().Value(constants.ContextKeyUserID).(uuid.UUID)
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req struct {
		Username *string `json:"username,omitempty" validate:"omitempty,min=3,max=50"`
		Email    *string `json:"email,omitempty" validate:"omitempty,email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Username == nil && req.Email == nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "At least one field must be provided for update")
		return
	}

	if validationErrors := utils.ValidateStruct(req); validationErrors != nil {
		utils.ValidationErrorResponse(w, validationErrors)
		return
	}

	currentUser, err := h.db.GetUserByTenantAndID(r.Context(), database.GetUserByTenantAndIDParams{
		ID:       userID,
		TenantID: tenantID,
	})
	if err != nil {
		utils.HandleDBError(w, err, "GetUserByTenantAndID", "User not found")
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
		ID:                  userID,
		Username:            currentUser.Username,
		Email:               currentUser.Email,
		LastLogin:           currentUser.LastLogin,
		LoginAttempts:       currentUser.LoginAttempts,
		IsLocked:            currentUser.IsLocked,
		LockedUntil:         currentUser.LockedUntil,
		PasswordResetToken:  currentUser.PasswordResetToken,
		TokenExpiry:         currentUser.TokenExpiry,
		MustChangePassword:  currentUser.MustChangePassword,
		EmailVerified:       currentUser.EmailVerified,
		MfaEnabled:          currentUser.MfaEnabled,
		LastPasswordChange:  currentUser.LastPasswordChange,
		UserType:            currentUser.UserType,
		TenantID:            currentUser.TenantID,
		RoleID:              currentUser.RoleID,
		PermissionsOverride: currentUser.PermissionsOverride,
	}

	if req.Username != nil {
		updateParams.Username = *req.Username
	}
	if req.Email != nil {
		updateParams.Email = *req.Email
		updateParams.EmailVerified = false
	}

	updatedUser, err := h.db.UpdateUser(r.Context(), updateParams)
	if err != nil {
		utils.HandleDBError(w, err, "UpdateUser")
		return
	}

	response := map[string]any{
		"id":            updatedUser.ID,
		"username":      updatedUser.Username,
		"email":         updatedUser.Email,
		"tenantId":      updatedUser.TenantID,
		"userType":      updatedUser.UserType,
		"emailVerified": updatedUser.EmailVerified,
		"mfaEnabled":    updatedUser.MfaEnabled,
		"lastLogin":     updatedUser.LastLogin,
		"createdAt":     updatedUser.CreatedAt,
		"updatedAt":     updatedUser.UpdatedAt,
	}

	if updatedUser.RoleID != nil {
		response["roleId"] = *updatedUser.RoleID
	}

	utils.SuccessResponse(w, http.StatusOK, response, "Profile updated successfully")
}
