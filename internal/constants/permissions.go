package constants

import (
	"encoding/json"
)

// PermissionNode represents a permission in the tree structure
type PermissionNode struct {
	Name    string           `json:"name"`
	Label   string           `json:"label"`
	Enabled bool             `json:"enabled"`
	Nested  []PermissionNode `json:"nested,omitempty"`
}

// DefaultTenantOwnerPermissions returns the default permissions for tenant owner
func DefaultTenantOwnerPermissions() []PermissionNode {
	return []PermissionNode{
		{
			Name:    "dashboard",
			Label:   "Dashboard",
			Enabled: true,
		},
		{
			Name:    "employees",
			Label:   "Employees",
			Enabled: true,
			Nested: []PermissionNode{
				{Name: "view", Label: "View Details", Enabled: true},
				{Name: "add", Label: "Add Employee", Enabled: true},
				{Name: "edit", Label: "Edit Employee", Enabled: true},
				{Name: "delete", Label: "Delete Employee", Enabled: true},
				{Name: "departments", Label: "Departments", Enabled: true},
				{Name: "positions", Label: "Positions", Enabled: true},
			},
		},
		{
			Name:    "users",
			Label:   "Users",
			Enabled: true,
			Nested: []PermissionNode{
				{Name: "view", Label: "View Users", Enabled: true},
				{Name: "add", Label: "Add User", Enabled: true},
				{Name: "edit", Label: "Edit User", Enabled: true},
				{Name: "delete", Label: "Delete User", Enabled: true},
			},
		},
		{
			Name:    "cc-specials",
			Label:   "CC Specials",
			Enabled: true,
			Nested: []PermissionNode{
				{Name: "index", Label: "View List", Enabled: true},
				{Name: "add", Label: "Add Menu", Enabled: true},
				{Name: "category-template", Label: "Category Template", Enabled: true},
			},
		},
		{
			Name:    "permissions",
			Label:   "Permissions",
			Enabled: true,
		},
	}
}

// DefaultStaffPermissions returns the default permissions for staff users
func DefaultStaffPermissions() []PermissionNode {
	return []PermissionNode{
		{
			Name:    "dashboard",
			Label:   "Dashboard",
			Enabled: true,
		},
		{
			Name:    "employees",
			Label:   "Employees",
			Enabled: true,
			Nested: []PermissionNode{
				{Name: "view", Label: "View Details", Enabled: true},
				{Name: "add", Label: "Add Employee", Enabled: false},
				{Name: "edit", Label: "Edit Employee", Enabled: false},
				{Name: "delete", Label: "Delete Employee", Enabled: false},
				{Name: "departments", Label: "Departments", Enabled: true},
			},
		},
	}
}

// DefaultGuestPermissions returns the default permissions for guest users
func DefaultGuestPermissions() []PermissionNode {
	return []PermissionNode{
		{
			Name:    "dashboard",
			Label:   "Dashboard",
			Enabled: true,
		},
		{
			Name:    "cc-specials",
			Label:   "CC Specials",
			Enabled: true,
			Nested: []PermissionNode{
				{Name: "index", Label: "View List", Enabled: true},
			},
		},
	}
}

// ToJSON converts permissions to JSON for database storage
func ToJSON(permissions []PermissionNode) (json.RawMessage, error) {
	data, err := json.Marshal(permissions)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

// MustToJSON converts permissions to JSON, panics on error (use for defaults)
func MustToJSON(permissions []PermissionNode) json.RawMessage {
	data, err := json.Marshal(permissions)
	if err != nil {
		panic(err)
	}
	return json.RawMessage(data)
}
