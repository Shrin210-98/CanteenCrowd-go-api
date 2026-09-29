package employees

import (
	"encoding/json"
	"fmt"
	"net/http"

	"ccms.com/api/internal/constants"
	"ccms.com/api/internal/database"
	"ccms.com/api/internal/utils"
	"github.com/google/uuid"
)

// Hierarchy types
type HierarchyNode struct {
	ID            uuid.UUID       `json:"id"`
	ReportsTo     *uuid.UUID      `json:"reportsTo,omitempty"`
	FirstName     string          `json:"firstName"`
	LastName      string          `json:"lastName"`
	PositionTitle string          `json:"positionTitle"`
	Level         int             `json:"level"`
	Children      []HierarchyNode `json:"children,omitempty"`
}

type UpdateHierarchyRequest struct {
	Changes []HierarchyChange `json:"changes" validate:"required,dive"`
}

type HierarchyChange struct {
	EmployeeID   uuid.UUID  `json:"employeeId" validate:"required,uuid"`
	NewReportsTo *uuid.UUID `json:"newReportsTo" validate:"omitempty,uuid"`
}

// GetHierarchy returns the complete org hierarchy
func (h *Handler) GetHierarchy(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := getTenantID(r)
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	rows, err := h.db.GetEmployeeHierarchy(r.Context(), tenantID)
	if err != nil {
		utils.HandleDBError(w, err, "GetEmployeeHierarchy")
		return
	}

	// Build tree from flat rows
	tree := buildHierarchyTree(rows)

	utils.SuccessResponse(w, http.StatusOK, tree, "Hierarchy retrieved successfully")
}

// UpdateHierarchy handles drag-and-drop changes
func (h *Handler) UpdateHierarchy(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := getTenantID(r)
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req UpdateHierarchyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if validationErrors := utils.ValidateStruct(req); validationErrors != nil {
		utils.ValidationErrorResponse(w, validationErrors)
		return
	}

	// Process each change
	for _, change := range req.Changes {
		// Validate no circular reference
		if change.NewReportsTo != nil {
			// Check for self-reference
			if change.EmployeeID == *change.NewReportsTo {
				utils.ErrorResponse(w, http.StatusBadRequest,
					"Cannot set self as manager")
				return
			}

			// Check for circular reference
			isCycle, err := h.wouldCreateCycle(r, change.EmployeeID, *change.NewReportsTo)
			if err != nil {
				utils.HandleDBError(w, err, "CheckCycle")
				return
			}
			if isCycle {
				utils.ErrorResponse(w, http.StatusBadRequest,
					"Cannot create circular reporting structure")
				return
			}
		}

		// Update reports_to
		_, err := h.db.UpdateEmployeeReportsTo(r.Context(),
			database.UpdateEmployeeReportsToParams{
				ID:        change.EmployeeID,
				TenantID:  tenantID,
				ReportsTo: change.NewReportsTo,
			})
		if err != nil {
			utils.HandleDBError(w, err, "UpdateEmployeeReportsTo")
			return
		}
	}

	utils.SuccessResponse(w, http.StatusOK, nil, "Hierarchy updated successfully")
}

// wouldCreateCycle checks if setting newReportsTo would create a cycle
func (h *Handler) wouldCreateCycle(r *http.Request, employeeID, newReportsTo uuid.UUID) (bool, error) {
	tenantID, ok := getTenantID(r)
	if !ok {
		return false, fmt.Errorf("unauthorized")
	}

	// Walk up from newReportsTo to see if we reach employeeID
	currentID := newReportsTo
	visited := make(map[uuid.UUID]bool)

	for currentID != uuid.Nil {
		// If we've seen this employee before, there's already a cycle
		if visited[currentID] {
			return true, nil
		}
		visited[currentID] = true

		// If we reach the employee being moved, it's a cycle
		if currentID == employeeID {
			return true, nil
		}

		// Get the current employee's manager (following your CreateUser pattern)
		employee, err := h.db.GetEmployeeById(r.Context(), database.GetEmployeeByIdParams{
			ID:       currentID, // Fixed: pass currentID directly
			TenantID: tenantID,  // Added: need tenantID
		})
		if err != nil {
			return false, err
		}

		// Move up to the manager
		if employee.ReportsTo == nil {
			break // Reached top of hierarchy
		}
		currentID = *employee.ReportsTo
	}

	return false, nil
}

// GetDirectReports returns employees directly under a manager
func (h *Handler) GetDirectReports(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := getTenantID(r)
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	managerID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid manager ID")
		return
	}

	reports, err := h.db.GetDirectReports(r.Context(), database.GetDirectReportsParams{
		TenantID:  tenantID,
		ReportsTo: &managerID, // Use managerID directly (not &managerID)
	})
	if err != nil {
		utils.HandleDBError(w, err, "GetDirectReports")
		return
	}

	utils.SuccessResponse(w, http.StatusOK, reports, "Direct reports retrieved")
}

// GetUnassignedEmployees returns employees without a manager
func (h *Handler) GetUnassignedEmployees(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := getTenantID(r)
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get current user's ID from context
	userID, _ := r.Context().Value(constants.ContextKeyUserID).(uuid.UUID)

	// Get employee record for the user (following DeleteUser pattern)
	employee, err := h.db.GetEmployeeByUserID(r.Context(), database.GetEmployeeByUserIDParams{
		UserID:   &userID,
		TenantID: tenantID,
	})

	var excludeEmployeeID uuid.UUID
	if err == nil && employee.ID != uuid.Nil {
		excludeEmployeeID = employee.ID
	} else {
		excludeEmployeeID = uuid.Nil
	}

	employees, err := h.db.GetUnassignedEmployees(r.Context(),
		database.GetUnassignedEmployeesParams{
			TenantID: tenantID,
			ID:       excludeEmployeeID,
		})
	if err != nil {
		utils.HandleDBError(w, err, "GetUnassignedEmployees")
		return
	}

	utils.SuccessResponse(w, http.StatusOK, employees, "Unassigned employees retrieved")
}

// Helper to build tree from flat hierarchy rows
func buildHierarchyTree(rows []database.GetEmployeeHierarchyRow) []HierarchyNode {
	nodeMap := make(map[uuid.UUID]*HierarchyNode)
	var roots []HierarchyNode

	// Create all nodes
	for _, row := range rows {
		// Handle nullable PositionTitle
		positionTitle := ""
		if row.PositionTitle != nil {
			positionTitle = *row.PositionTitle
		}

		node := HierarchyNode{
			ID:            row.ID,
			ReportsTo:     row.ReportsTo,
			FirstName:     row.FirstName,
			LastName:      row.LastName,
			PositionTitle: positionTitle, // Fixed: dereference pointer
			Level:         int(row.HierarchyLevel),
			Children:      []HierarchyNode{},
		}
		nodeMap[row.ID] = &node
	}

	// Build tree
	for _, row := range rows {
		node := nodeMap[row.ID]
		if row.ReportsTo == nil {
			roots = append(roots, *node)
		} else {
			if parent, exists := nodeMap[*row.ReportsTo]; exists {
				parent.Children = append(parent.Children, *node)
			}
		}
	}

	return roots
}
