package requests

import (
	"time"

	"encoding/json"
	"fmt"
	"net/http"

	"ccms.com/api/internal/constants"
	"ccms.com/api/internal/database"
	"ccms.com/api/internal/utils"
	"github.com/google/uuid"
)

type CreateRequest struct {
	RequestType string     `json:"requestType" validate:"required,oneof=task support approval query"`
	Category    string     `json:"category" validate:"omitempty,oneof=it hr finance general"`
	Title       string     `json:"title" validate:"required,min=5,max=200"`
	Description string     `json:"description" validate:"omitempty,max=2000"`
	AssigneeID  *uuid.UUID `json:"assigneeId" validate:"omitempty,uuid"`
	Priority    string     `json:"priority" validate:"required,oneof=low medium high urgent"`
	DueDate     *time.Time `json:"dueDate" validate:"omitempty"`
}

type UpdateRequestStatus struct {
	Status string `json:"status" validate:"required,oneof=new acknowledged in_progress pending completed declined"`
}

type UpdateRequestAssignee struct {
	AssigneeID uuid.UUID `json:"assigneeId" validate:"required,uuid"`
}

type AddComment struct {
	Comment string `json:"comment" validate:"required,min=1,max=1000"`
}

type RequestResponse struct {
	ID            uuid.UUID  `json:"id"`
	RequestNumber string     `json:"requestNumber"`
	RequestType   string     `json:"requestType"`
	Category      string     `json:"category,omitempty"`
	Title         string     `json:"title"`
	Description   string     `json:"description,omitempty"`
	RequesterID   uuid.UUID  `json:"requesterId"`
	RequesterName string     `json:"requesterName"`
	AssigneeID    *uuid.UUID `json:"assigneeId,omitempty"`
	AssigneeName  string     `json:"assigneeName,omitempty"`
	ManagerID     *uuid.UUID `json:"managerId,omitempty"`
	ManagerName   string     `json:"managerName,omitempty"`
	Priority      string     `json:"priority"`
	Status        string     `json:"status"`
	DueDate       *time.Time `json:"dueDate,omitempty"`
	CompletedAt   *time.Time `json:"completedAt,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

// CreateRequest handles new request creation with auto-routing
func (h *Handler) CreateRequest(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := getTenantID(r)
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID, _ := r.Context().Value(constants.ContextKeyUserID).(uuid.UUID)

	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if validationErrors := utils.ValidateStruct(req); validationErrors != nil {
		utils.ValidationErrorResponse(w, validationErrors)
		return
	}

	// Get employee record for the requester
	employee, err := h.getEmployeeByUserID(r, userID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusForbidden, "User not linked to employee")
		return
	}

	// Auto-route based on hierarchy
	var assigneeID *uuid.UUID
	var managerID *uuid.UUID

	// Set manager as requester's reports_to
	if employee.ReportsTo != nil {
		managerID = employee.ReportsTo
	}

	// If specific assignee provided, use it
	if req.AssigneeID != nil {
		assigneeID = req.AssigneeID
	}

	// Generate request number
	nextNum, _ := h.db.GetNextRequestNumber(r.Context(), tenantID)
	requestNumber := fmt.Sprintf("REQ-%d", nextNum)

	// Create request - Fix: Convert string to *string for nullable fields
	var categoryPtr *string
	if req.Category != "" {
		categoryPtr = &req.Category
	}

	var descriptionPtr *string
	if req.Description != "" {
		descriptionPtr = &req.Description
	}

	created, err := h.db.CreateRequest(r.Context(), database.CreateRequestParams{
		TenantID:          tenantID,
		RequestNumber:     requestNumber,
		RequestType:       req.RequestType,
		Category:          categoryPtr, // Fixed: pointer
		Title:             req.Title,
		Description:       descriptionPtr, // Fixed: pointer
		RequesterID:       employee.ID,
		AssigneeID:        assigneeID,
		RoutedToManagerID: managerID,
		Priority:          req.Priority,
		DueDate:           req.DueDate,
	})
	if err != nil {
		utils.HandleDBError(w, err, "CreateRequest")
		return
	}

	utils.SuccessResponse(w, http.StatusCreated, created, "Request created successfully")
}

// GetMyRequests returns requests created by current user
func (h *Handler) GetMyRequests(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := getTenantID(r)
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID, _ := r.Context().Value(constants.ContextKeyUserID).(uuid.UUID)
	employee, err := h.getEmployeeByUserID(r, userID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusForbidden, "User not linked to employee")
		return
	}

	status := r.URL.Query().Get("status")
	requestType := r.URL.Query().Get("type")
	page, pageSize := getPagination(r)

	requests, err := h.db.GetRequestsByRequester(r.Context(),
		database.GetRequestsByRequesterParams{
			TenantID:    tenantID,
			RequesterID: employee.ID,
			Status:      status,
			RequestType: requestType,
			PageSize:    int32(pageSize),
			PageOffset:  int32((page - 1) * pageSize),
		})
	if err != nil {
		utils.HandleDBError(w, err, "GetRequestsByRequester")
		return
	}

	utils.SuccessResponse(w, http.StatusOK, requests, "Requests retrieved")
}

// GetAssignedToMe returns requests assigned to current user
func (h *Handler) GetAssignedToMe(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := getTenantID(r)
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID, _ := r.Context().Value(constants.ContextKeyUserID).(uuid.UUID)
	employee, err := h.getEmployeeByUserID(r, userID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusForbidden, "User not linked to employee")
		return
	}

	page, pageSize := getPagination(r)

	requests, err := h.db.GetRequestsByAssignee(r.Context(),
		database.GetRequestsByAssigneeParams{
			TenantID:   tenantID,
			AssigneeID: &employee.ID,
			Limit:      int32(pageSize),
			Offset:     int32((page - 1) * pageSize),
		})
	if err != nil {
		utils.HandleDBError(w, err, "GetRequestsByAssignee")
		return
	}

	utils.SuccessResponse(w, http.StatusOK, requests, "Assigned requests retrieved")
}

// GetInbox returns requests from direct reports (for managers)
func (h *Handler) GetInbox(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := getTenantID(r)
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID, _ := r.Context().Value(constants.ContextKeyUserID).(uuid.UUID)
	employee, err := h.getEmployeeByUserID(r, userID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusForbidden, "User not linked to employee")
		return
	}

	page, pageSize := getPagination(r)

	// Get requests where this employee is the manager
	requests, err := h.db.GetRequestsForManager(r.Context(),
		database.GetRequestsForManagerParams{
			TenantID:          tenantID,
			RoutedToManagerID: &employee.ID,
			Limit:             int32(pageSize),
			Offset:            int32((page - 1) * pageSize),
		})
	if err != nil {
		utils.HandleDBError(w, err, "GetRequestsForManager")
		return
	}

	utils.SuccessResponse(w, http.StatusOK, requests, "Inbox retrieved")
}

// UpdateStatus updates request status
func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := getTenantID(r)
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	requestID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request ID")
		return
	}

	var req UpdateRequestStatus
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if validationErrors := utils.ValidateStruct(req); validationErrors != nil {
		utils.ValidationErrorResponse(w, validationErrors)
		return
	}

	updated, err := h.db.UpdateRequestStatus(r.Context(),
		database.UpdateRequestStatusParams{
			ID:       requestID,
			TenantID: tenantID,
			Status:   req.Status,
		})
	if err != nil {
		utils.HandleDBError(w, err, "UpdateRequestStatus")
		return
	}

	utils.SuccessResponse(w, http.StatusOK, updated, "Status updated")
}

// AddComment adds a comment to a request
func (h *Handler) AddComment(w http.ResponseWriter, r *http.Request) {
	requestID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request ID")
		return
	}

	userID, _ := r.Context().Value(constants.ContextKeyUserID).(uuid.UUID)

	var req AddComment
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if validationErrors := utils.ValidateStruct(req); validationErrors != nil {
		utils.ValidationErrorResponse(w, validationErrors)
		return
	}

	comment, err := h.db.AddRequestComment(r.Context(),
		database.AddRequestCommentParams{
			RequestID: requestID,
			UserID:    userID,
			Comment:   req.Comment,
		})
	if err != nil {
		utils.HandleDBError(w, err, "AddRequestComment")
		return
	}

	utils.SuccessResponse(w, http.StatusCreated, comment, "Comment added")
}
