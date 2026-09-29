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
	UserTypeExternal    = "external"

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
	UserTypeExternal,
}

// Allowed tenant plans
var AllowedTenantPlans = []string{
	PlanFree,
	PlanBasic,
	PlanPremium,
	PlanEnterprise,
}

var DefaultDepartments = []struct {
	Name           string
	Description    string
	DepartmentType string // 'operational', 'support', 'management'
}{
	{
		Name:           "Management",
		Description:    "Management and administration",
		DepartmentType: "management",
	},
	{
		Name:           "Kitchen",
		Description:    "Food preparation and cooking",
		DepartmentType: "operational",
	},
	{
		Name:           "Service",
		Description:    "Customer service and front desk",
		DepartmentType: "operational",
	},
	{
		Name:           "IT Support",
		Description:    "Technical support and maintenance",
		DepartmentType: "support", // Support department
	},
	{
		Name:           "Human Resources",
		Description:    "HR, leave, and employee relations",
		DepartmentType: "support", // Support department
	},
	{
		Name:           "Finance",
		Description:    "Accounts, payroll, and expenses",
		DepartmentType: "support", // Support department
	},
	{
		Name:           "General Support",
		Description:    "General queries and assistance",
		DepartmentType: "support", // Support department
	},
}
