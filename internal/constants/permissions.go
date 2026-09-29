package constants

import "ccms.com/api/internal/utils"

// TENANT OWNER PERMISSIONS (Full Access)
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
				{Name: "view", Label: "View Users", Enabled: true},
				{Name: "add", Label: "Add User", Enabled: true},
				{Name: "edit", Label: "Edit User", Enabled: true},
				{Name: "delete", Label: "Delete User", Enabled: true},
				{Name: "archive", Label: "Archive User", Enabled: true},
				{Name: "roles", Label: "Roles", Enabled: true},
			},
		},
		{
			Name:               "employees",
			Label:              "Employees",
			Enabled:            true,
			RequiredSelections: []string{},
			Nested: []utils.PermissionNode{
				{Name: "view", Label: "View Employees", Enabled: true},
				{Name: "add", Label: "Add Employee", Enabled: true},
				{Name: "edit", Label: "Edit Employee", Enabled: true},
				{Name: "delete", Label: "Delete Employee", Enabled: true},
				{Name: "departments", Label: "Departments", Enabled: true},
				{Name: "positions", Label: "Positions", Enabled: true},
				{Name: "hierarchy", Label: "Reporting Tree", Enabled: true},
			},
		},
		{
			Name:    "operations-hub",
			Label:   "Operations Hub",
			Enabled: true,
			Nested: []utils.PermissionNode{
				{Name: "view", Label: "View Requests", Enabled: true},
				{Name: "create", Label: "Create Requests", Enabled: true},
				{Name: "assign", Label: "Assign Requests", Enabled: true},
				{Name: "respond", Label: "Respond to Requests", Enabled: true},
				{Name: "complete", Label: "Complete Requests", Enabled: true},
				{Name: "approve", Label: "Approve Requests", Enabled: true},
			},
		},
		{
			Name:    "support-center",
			Label:   "Support Center",
			Enabled: true,
			Nested: []utils.PermissionNode{
				{Name: "view", Label: "View Tickets", Enabled: true},
				{Name: "raise", Label: "Raise Tickets", Enabled: true},
				{Name: "chat", Label: "Use Support Chat", Enabled: true},
				{Name: "accept", Label: "Accept Tickets", Enabled: true},
				{Name: "resolve", Label: "Resolve Tickets", Enabled: true},
				{Name: "manage", Label: "Manage Support", Enabled: true},
			},
		},
		{
			Name:    "cc-specials",
			Label:   "CC Specials",
			Enabled: true,
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

// STAFF PERMISSIONS (Limited Access)
func DefaultStaffPermissions() []utils.PermissionNode {
	return []utils.PermissionNode{
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
				{Name: "view", Label: "View Employees", Enabled: true},
				{Name: "add", Label: "Add Employee", Enabled: false},
				{Name: "edit", Label: "Edit Employee", Enabled: false},
				{Name: "delete", Label: "Delete Employee", Enabled: false},
				{Name: "departments", Label: "Departments", Enabled: false},
				{Name: "positions", Label: "Positions", Enabled: false},
				{Name: "hierarchy", Label: "Reporting Tree", Enabled: false},
			},
		},
		{
			Name:    "operations-hub",
			Label:   "Operations Hub",
			Enabled: true,
			Nested: []utils.PermissionNode{
				{Name: "view", Label: "View Requests", Enabled: true},
				{Name: "create", Label: "Create Requests", Enabled: true},
				{Name: "assign", Label: "Assign Requests", Enabled: false},
				{Name: "respond", Label: "Respond to Requests", Enabled: true},
				{Name: "complete", Label: "Complete Requests", Enabled: true},
				{Name: "approve", Label: "Approve Requests", Enabled: false},
			},
		},
		{
			Name:    "support-center",
			Label:   "Support Center",
			Enabled: true,
			Nested: []utils.PermissionNode{
				{Name: "view", Label: "View Tickets", Enabled: true},
				{Name: "raise", Label: "Raise Tickets", Enabled: true},
				{Name: "chat", Label: "Use Support Chat", Enabled: true},
				{Name: "accept", Label: "Accept Tickets", Enabled: true},
				{Name: "resolve", Label: "Resolve Tickets", Enabled: true},
				{Name: "manage", Label: "Manage Support", Enabled: false},
			},
		},
		{
			Name:               "cc-specials",
			Label:              "CC Specials",
			Enabled:            true,
			RequiredSelections: []string{},
			Nested: []utils.PermissionNode{
				{Name: "add", Label: "Add Menu", Enabled: false},
				{Name: "edit", Label: "Edit Menu", Enabled: false},
				{Name: "delete", Label: "Delete Menu", Enabled: false},
				{Name: "category-template", Label: "Category Template", Enabled: false},
			},
		},
	}
}

// External PERMISSIONS (Minimal Access)
func DefaultExternalPermissions() []utils.PermissionNode {
	return []utils.PermissionNode{
		{
			Name:    "my-dashboard",
			Label:   "My Dashboard",
			Enabled: true,
		},
	}
}
