package utils

import (
	"encoding/json"
)

// PermissionNode represents the full permission structure from database
type PermissionNode struct {
	Name               string           `json:"name"`
	Label              string           `json:"label"`
	Enabled            bool             `json:"enabled"`
	Nested             []PermissionNode `json:"nested,omitempty"`
	RequiredSelections []string         `json:"requiredSelections,omitempty"`
}

// FilteredPermission represents the response structure (no enabled field)
type FilteredPermission struct {
	Name   string               `json:"name"`
	Label  string               `json:"label"`
	Nested []FilteredPermission `json:"nested,omitempty"`
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

// FilterEnabledPermissions returns only enabled permissions without the enabled field
func FilterEnabledPermissions(permissions []PermissionNode) []FilteredPermission {
	var result []FilteredPermission

	for _, perm := range permissions {
		// Skip disabled permissions
		if !perm.Enabled {
			continue
		}

		filtered := FilteredPermission{
			Name:  perm.Name,
			Label: perm.Label,
		}

		// Process nested permissions
		if len(perm.Nested) > 0 {
			nested := FilterEnabledPermissions(perm.Nested)
			if len(nested) > 0 {
				filtered.Nested = nested
			}
		}

		result = append(result, filtered)
	}

	return result
}

// FilterPermissionsFromJSON parses JSON and returns filtered permissions
func FilterPermissionsFromJSON(jsonData []byte) ([]FilteredPermission, error) {
	var permissions []PermissionNode
	if err := json.Unmarshal(jsonData, &permissions); err != nil {
		return nil, err
	}
	return FilterEnabledPermissions(permissions), nil
}

// FlattenPermissions converts filtered permissions to flat array
func FlattenPermissions(permissions []FilteredPermission) []string {
	var result []string

	var flatten func(nodes []FilteredPermission, prefix string)
	flatten = func(nodes []FilteredPermission, prefix string) {
		for _, node := range nodes {
			path := node.Name
			if prefix != "" {
				path = prefix + "." + node.Name
			}
			result = append(result, path)

			if len(node.Nested) > 0 {
				flatten(node.Nested, path)
			}
		}
	}

	flatten(permissions, "")
	return result
}
