package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"ccms.com/api/internal/database"
	"ccms.com/api/internal/utils"
	"github.com/google/uuid"
	// "github.com/jackc/pgx/v5/pgtype"
)

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username   string `json:"username"`
		Email      string `json:"email"`
		Password   string `json:"password"`
		TenantName string `json:"tenant_name"`
		TenantSlug string `json:"tenant_slug"`
		TenantPlan string `json:"tenant_plan,omitempty"`
		FullName   string `json:"full_name,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, map[string]any{"message": "Invalid request"})
		return
	}

	// Validate required fields
	if req.Username == "" || req.Email == "" || req.Password == "" {
		utils.JsonResponse(w, http.StatusBadRequest, map[string]any{"message": "Username, email, and password are required"})
		return
	}

	if req.TenantName == "" || req.TenantSlug == "" {
		utils.JsonResponse(w, http.StatusBadRequest, map[string]any{"message": "Tenant name and slug are required"})
		return
	}

	// Set default plan
	if req.TenantPlan == "" {
		req.TenantPlan = "free"
	}

	// Check if username or email already exists
	usernameExists, err := h.db.CheckUsernameExists(r.Context(), req.Username)
	if err != nil {
		log.Printf("Register username check failed: %v", err)
		utils.JsonResponse(w, http.StatusInternalServerError, map[string]any{"message": "Failed to check username"})
		return
	}
	if usernameExists {
		utils.JsonResponse(w, http.StatusConflict, map[string]any{"message": "Username already exists"})
		return
	}

	emailExists, err := h.db.CheckEmailExists(r.Context(), req.Email)
	if err != nil {
		log.Printf("Register email check failed: %v", err)
		utils.JsonResponse(w, http.StatusInternalServerError, map[string]any{"message": "Failed to check email"})
		return
	}
	if emailExists {
		utils.JsonResponse(w, http.StatusConflict, map[string]any{"message": "Email already exists"})
		return
	}

	// Check if tenant slug exists
	slugExists, err := h.db.CheckTenantSlugExists(r.Context(), req.TenantSlug)
	if err != nil {
		log.Printf("Register tenant slug check failed: %v", err)
		utils.JsonResponse(w, http.StatusInternalServerError, map[string]any{"message": "Failed to check tenant slug"})
		return
	}
	if slugExists {
		utils.JsonResponse(w, http.StatusConflict, map[string]any{"message": "Tenant slug already exists"})
		return
	}

	// Start transaction
	tx, err := h.Pool.Begin(r.Context())
	if err != nil {
		log.Printf("Register transaction start failed: %v", err)
		utils.JsonResponse(w, http.StatusInternalServerError, map[string]any{"message": "Failed to start transaction"})
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
		Status:   "active",
		Settings: json.RawMessage(`{}`),
	})

	if err != nil {
		log.Printf("Register tenant creation failed: %v", err)
		utils.JsonResponse(w, http.StatusInternalServerError, map[string]any{"message": fmt.Sprintf("Failed to create tenant: %v", err)})
		return
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		log.Printf("Register password hashing failed: %v", err)
		utils.JsonResponse(w, http.StatusInternalServerError, map[string]any{"message": "Failed to hash password"})
		return
	}

	// Create user with tenant_id (now NOT NULL)
	user, err := qtx.CreateUser(r.Context(), database.CreateUserParams{
		TenantID:     tenant.ID, // Direct UUID, not pointer
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		UserType:     "tenant_owner", // The main account holder
	})

	if err != nil {
		log.Printf("Register user creation failed: %v", err)
		utils.JsonResponse(w, http.StatusInternalServerError, map[string]any{"message": fmt.Sprintf("Failed to register user: %v", err)})
		return
	}

	// Create user profile with default permissions
	defaultPermissions := json.RawMessage(`{
        "system": { "index": true, "restart": true, "backup": true, "shutdown": true },
        "dashboard": { "index": true },
        "users": { "index": true, "view": true, "add": true, "edit": true, "delete": true },
        "employees": { "index": true, "view": true, "add": true, "archive": true },
        "clients": { "index": true, "view": true, "add": true, "edit": true, "delete": true }
    }`)

	var fullName *string
	if req.FullName != "" {
		fullName = &req.FullName
	}

	profile, err := qtx.CreateUserProfile(r.Context(), database.CreateUserProfileParams{
		UserID:      user.ID,
		TenantID:    tenant.ID,
		FullName:    fullName,
		Permissions: defaultPermissions,
	})

	if err != nil {
		log.Printf("Register user profile creation failed: %v", err)
		utils.JsonResponse(w, http.StatusInternalServerError, map[string]any{"message": fmt.Sprintf("Failed to create user profile: %v", err)})
		return
	}

	// Create default admin role for the tenant
	adminRoleDescription := "Default administrator role with full permissions"
	role, err := qtx.CreateRole(r.Context(), database.CreateRoleParams{
		TenantID:           tenant.ID,
		Name:               "admin",
		Description:        &adminRoleDescription,
		PermissionTemplate: defaultPermissions,
		IsDefault:          true,
	})

	if err != nil {
		log.Printf("Register default role creation failed: %v", err)
		// Don't fail the registration if role creation fails
		// You can create it later or handle it differently
	}

	// Commit transaction
	if err := tx.Commit(r.Context()); err != nil {
		log.Printf("Register transaction commit failed: %v", err)
		utils.JsonResponse(w, http.StatusInternalServerError, map[string]any{"message": "Failed to commit transaction"})
		return
	}

	// Prepare response
	responseData := map[string]any{
		"user": map[string]any{
			"id":        user.ID,
			"username":  user.Username,
			"email":     user.Email,
			"userType":  user.UserType,
			"tenantId":  user.TenantID,
			"createdAt": user.CreatedAt,
		},
		"tenant": map[string]any{
			"id":        tenant.ID,
			"name":      tenant.Name,
			"slug":      tenant.Slug,
			"plan":      tenant.Plan,
			"status":    tenant.Status,
			"createdAt": tenant.CreatedAt,
		},
		"profile": map[string]any{
			"id":        profile.ID,
			"userId":    profile.UserID,
			"tenantId":  profile.TenantID,
			"createdAt": profile.CreatedAt,
		},
	}

	// Include role in response if it was created
	if role.ID != uuid.Nil {
		responseData["role"] = map[string]any{
			"id":        role.ID,
			"name":      role.Name,
			"tenantId":  role.TenantID,
			"isDefault": role.IsDefault,
		}
	}

	utils.JsonResponse(w, http.StatusCreated, map[string]any{
		"data":    responseData,
		"message": "Successfully Registered",
	})

	// Example Request:
	// 	{
	//     "username": "john_doe",
	//     "email": "john@example.com",
	//     "password": "SecurePass123!",
	//     "tenant_name": "Acme Corporation",
	//     "tenant_slug": "acme-corp",
	//     "tenant_plan": "premium",
	//     "full_name": "John Doe"
	//  }
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, map[string]any{"message": "Invalid Request"})
		return
	}

	// Validate input
	if creds.Email == "" || creds.Password == "" {
		utils.JsonResponse(w, http.StatusBadRequest, map[string]any{"message": "Email and password are required"})
		return
	}

	// Get user by email
	user, err := h.db.GetUserByEmail(r.Context(), creds.Email)
	if err != nil {
		log.Printf("Login failed - user not found: %v", err)
		utils.JsonResponse(w, http.StatusUnauthorized, map[string]any{"message": "Invalid credentials"})
		return
	}

	// Check if account is locked
	if user.IsLocked {
		if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
			log.Printf("Login failed - account locked for user: %s", user.Email)
			utils.JsonResponse(w, http.StatusForbidden, map[string]any{
				"message":     "Account is locked. Please try again later.",
				"lockedUntil": user.LockedUntil,
			})
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
		if newAttempts >= 5 {
			isLocked = true
			lockTime := time.Now().Add(15 * time.Minute) // Lock for 15 minutes
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

		utils.JsonResponse(w, http.StatusUnauthorized, map[string]any{
			"message":      "Invalid credentials",
			"attemptsLeft": 5 - newAttempts,
		})
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
		ID:                 user.ID,
		Username:           user.Username,
		Email:              user.Email,
		LastLogin:          &now,
		LoginAttempts:      0,
		IsLocked:           false,
		LockedUntil:        nil,
		PasswordResetToken: user.PasswordResetToken,
		TokenExpiry:        user.TokenExpiry,
		MustChangePassword: user.MustChangePassword,
		EmailVerified:      user.EmailVerified,
		MfaEnabled:         user.MfaEnabled,
		LastPasswordChange: user.LastPasswordChange,
		UserType:           user.UserType,
		TenantID:           user.TenantID,
	})
	if err != nil {
		log.Printf("Failed to update last login: %v", err)
		// Continue anyway - this is not critical
	}

	// Generate JWT token
	token, err := utils.GenerateTokens(user.ID, user.TenantID, h.jwtSecret)
	if err != nil {
		log.Printf("Failed to generate token: %v", err)
		utils.JsonResponse(w, http.StatusInternalServerError, map[string]any{"message": "Failed to generate token"})
		return
	}

	// Get user profile (optional)
	profile, err := h.db.GetUserProfile(r.Context(), user.ID)
	if err != nil {
		log.Printf("Profile not found for user: %v", err)
		// Continue without profile - not critical
	}

	// Prepare response
	responseData := map[string]any{
		"user": map[string]any{
			"id":            user.ID,
			"tenantId":      user.TenantID,
			"username":      user.Username,
			"email":         user.Email,
			"lastLogin":     now,
			"userType":      user.UserType,
			"emailVerified": user.EmailVerified,
			"mfaEnabled":    user.MfaEnabled,
			"createdAt":     user.CreatedAt,
			"updatedAt":     user.UpdatedAt,
		},
		"token": token,
	}

	// Include profile if found
	if profile.ID != uuid.Nil {
		responseData["profile"] = map[string]any{
			"id":          profile.ID,
			"fullName":    profile.FullName,
			"phone":       profile.Phone,
			"avatarUrl":   profile.AvatarUrl,
			"timezone":    profile.Timezone,
			"permissions": profile.Permissions,
		}
	}

	// Check if password change is required
	if user.MustChangePassword {
		responseData["mustChangePassword"] = true
	}

	utils.JsonResponse(w, http.StatusOK, map[string]any{
		"data":    responseData,
		"message": "Successfully Logged In",
	})
	// Example Login Request:
	//
	//	{
	//	    "email": "john@example.com",
	//	    "password": "SecurePass123!"
	//	}
}

// func (h *Handler) Register_v1(w http.ResponseWriter, r *http.Request) {
// 	var req struct {
// 		Username string `json:"username"`
// 		Email    string `json:"email"`
// 		Password string `json:"password"`
// 	}

// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		utils.JsonResponse(w, http.StatusBadRequest, map[string]any{"message": "Invalid request"})
// 		return
// 	}

// 	log.Printf("Register request received: username=%s email=%s", req.Username, req.Email)

// 	hashedPassword, err := utils.HashPassword(req.Password)
// 	if err != nil {
// 		log.Printf("Register password hashing failed: %v", err)
// 		utils.JsonResponse(w, http.StatusInternalServerError, map[string]any{"message": "Failed to hash password"})
// 		return
// 	}

// 	log.Printf("Register attempting DB insert for user: %s", req.Username)
// 	user, err := h.db.CreateUser(r.Context(), database.CreateUserParams{
// 		Username:     req.Username,
// 		Email:        req.Email,
// 		PasswordHash: hashedPassword,
// 		UserType:     "owner",
// 	})

// 	if err != nil {
// 		log.Printf("Register DB insert failed: %v", err)
// 		utils.JsonResponse(w, http.StatusInternalServerError, map[string]any{"data": user, "message": fmt.Sprintf("Failed to Register owner user: %v", err)})
// 		return
// 	}

// 	log.Printf("Register success: %+v", user)

// 	responseData := map[string]any{
// 		"id":        user.ID,
// 		"username":  user.Username,
// 		"email":     user.Email,
// 		"lastLogin": user.LastLogin,
// 		"userType":  user.UserType,
// 		// "emailVerified": user.EmailVerified,
// 		// "mfaEnabled":    user.MfaEnabled,
// 		"createdAt": user.CreatedAt,
// 		// "updatedAt":     user.UpdatedAt,
// 	}
// 	utils.JsonResponse(w, http.StatusCreated, map[string]any{"data": responseData, "message": "Successfully Registered the Owner"})
// }

// func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
// 	var creds struct {
// 		Email    string `json:"email"`
// 		Password string `json:"password"`
// 	}

// 	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
// 		utils.JsonResponse(w, http.StatusBadRequest, map[string]any{"message": "Invalid Request"})
// 		return
// 	}

// 	user, err := h.db.GetUserByEmail(r.Context(), creds.Email)
// 	if err != nil {
// 		utils.JsonResponse(w, http.StatusUnauthorized, map[string]any{"message": "Invalid credentials"})
// 		return
// 	}

// 	if user.IsLocked.Bool && user.LockedUntil.Time.After(time.Now()) {
// 		utils.JsonResponse(w, http.StatusForbidden, map[string]any{"message": "Account locked"})
// 		return
// 	}

// 	if !utils.CheckPasswordHash(creds.Password, user.PasswordHash) {
// 		// h.db.IncrementLoginAttempts(r.Context(), user.ID)
// 		utils.JsonResponse(w, http.StatusBadRequest, map[string]any{"message": "Invalid credentials"})
// 		return
// 	}

// 	// h.db.ResetLoginAttempts(r.Context(), user.ID)

// 	token, err := utils.GenerateTokens(uuid.UUID(user.ID.Bytes), h.jwtSecret)
// 	if err != nil {
// 		utils.JsonResponse(w, http.StatusInternalServerError, map[string]any{"message": "Failed to generate token"})
// 		return
// 	}

// 	responseData := map[string]any{
// 		"id":        user.ID,
// 		"username":  user.Username,
// 		"email":     user.Email,
// 		"lastLogin": user.LastLogin,
// 		"userType":  user.UserType,
// 		"token":     token,
// 		// "emailVerified": user.EmailVerified,
// 		// "mfaEnabled":    user.MfaEnabled,
// 		"createdAt": user.CreatedAt,
// 		// "updatedAt":     user.UpdatedAt,
// 	}
// 	utils.JsonResponse(w, http.StatusOK, map[string]any{"data": responseData, "message": "Succesfully Logged In"})
// }
