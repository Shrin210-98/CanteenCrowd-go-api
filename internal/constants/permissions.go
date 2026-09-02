package constants

import (
	"ccms.com/api/internal/utils"
)

func DefaultTenantOwnerPermissions() []utils.PermissionNode {
	return []utils.PermissionNode{
		{
			Name:    "dashboard",
			Label:   "Dashboard",
			Enabled: true,
		},
		{
			Name:    "users",
			Label:   "Users",
			Enabled: true,
			Nested: []utils.PermissionNode{
				{Name: "add", Label: "Add User", Enabled: true},
				{Name: "edit", Label: "Edit User", Enabled: true},
				{Name: "delete", Label: "Delete User", Enabled: true},
				{Name: "roles", Label: "Roles", Enabled: true},
			},
		},
		{
			Name:               "employees",
			Label:              "Employees",
			Enabled:            true,
			RequiredSelections: []string{},
			Nested: []utils.PermissionNode{
				{Name: "add", Label: "Add Employee", Enabled: true},
				{Name: "edit", Label: "Edit Employee", Enabled: true},
				{Name: "delete", Label: "Delete Employee", Enabled: true},
				{Name: "departments", Label: "Departments", Enabled: true},
				{Name: "positions", Label: "Positions", Enabled: true},
			},
		},
		{
			Name:    "cc-specials",
			Label:   "CC Specials",
			Enabled: true,
			Nested: []utils.PermissionNode{
				{Name: "add", Label: "Add Menu", Enabled: true},
				{Name: "category-template", Label: "Category Template", Enabled: true},
			},
		},
		{
			Name:               "settings",
			Label:              "Settings",
			Enabled:            true,
			RequiredSelections: []string{"profile"},
			Nested: []utils.PermissionNode{
				{Name: "profile", Label: "Profile", Enabled: true},
				{Name: "firewall", Label: "Firewall", Enabled: true},
				{Name: "faq-support", Label: "FAQ and Support", Enabled: true},
			},
		},
	}
}

func DefaultStaffPermissions() []utils.PermissionNode {
	return []utils.PermissionNode{
		{
			Name:    "dashboard",
			Label:   "Dashboard",
			Enabled: true,
		},
		{
			Name:    "my-dashboard",
			Label:   "My Dashboard",
			Enabled: true,
		},
		{
			Name:               "employees",
			Label:              "Employees",
			Enabled:            true,
			RequiredSelections: []string{},
			Nested: []utils.PermissionNode{
				{Name: "add", Label: "Add Employee", Enabled: false},
				{Name: "edit", Label: "Edit Employee", Enabled: false},
				{Name: "delete", Label: "Delete Employee", Enabled: false},
				{Name: "departments", Label: "Departments", Enabled: true},
			},
		},
	}
}

func DefaultGuestPermissions() []utils.PermissionNode {
	return []utils.PermissionNode{
		{
			Name:    "dashboard",
			Label:   "Dashboard",
			Enabled: true,
		},
		{
			Name:               "cc-specials",
			Label:              "CC Specials",
			Enabled:            true,
			RequiredSelections: []string{},
			Nested: []utils.PermissionNode{
				{Name: "add", Label: "Add Menu", Enabled: false},
				{Name: "category-template", Label: "Category Template", Enabled: false},
			},
		},
	}
}
