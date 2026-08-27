package constants

import "time"

// Default values for the application
const (
	// Default tenant settings
	DefaultTenantPlan   = "free"
	DefaultTenantStatus = "active"

	// Default pagination
	DefaultPageSize = 20
	MaxPageSize     = 100

	// Default lockout settings
	MaxLoginAttempts = 5
	LockoutDuration  = 15 * time.Minute

	// User types
	UserTypeTenantOwner = "tenant_owner"
	UserTypeStaff       = "staff"
	UserTypeGuest       = "guest"

	// Tenant plans
	PlanFree       = "free"
	PlanBasic      = "basic"
	PlanPremium    = "premium"
	PlanEnterprise = "enterprise"
)

// Allowed user types
var AllowedUserTypes = []string{
	UserTypeTenantOwner,
	UserTypeStaff,
	UserTypeGuest,
}

// Allowed tenant plans
var AllowedTenantPlans = []string{
	PlanFree,
	PlanBasic,
	PlanPremium,
	PlanEnterprise,
}
