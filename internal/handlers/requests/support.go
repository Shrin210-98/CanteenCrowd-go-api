package requests

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"ccms.com/api/internal/constants"
	"ccms.com/api/internal/database"
	"ccms.com/api/internal/utils"
	"github.com/google/uuid"
)

type CreateSupportRequest struct {
	SupportCategory string `json:"supportCategory" validate:"required,oneof=it hr finance general"`
	Title           string `json:"title" validate:"required,min=5,max=200"`
	Description     string `json:"description" validate:"omitempty,max=2000"`
	Priority        string `json:"priority" validate:"required,oneof=low medium high urgent"`
}

type ChatMessageResponse struct {
	ID         uuid.UUID  `json:"id"`
	RequestID  uuid.UUID  `json:"requestId"`
	SenderID   uuid.UUID  `json:"senderId"`
	SenderName string     `json:"senderName"`
	Message    string     `json:"message"`
	IsRead     bool       `json:"isRead"`
	ReadAt     *time.Time `json:"readAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
}

// CreateSupportRequest creates a support request routed to department
func (h *Handler) CreateSupportRequest(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := getTenantID(r)
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID, _ := r.Context().Value(constants.ContextKeyUserID).(uuid.UUID)
	log.Println("🔥 CreateSupportRequest v2 (no auto-assign)")

	var req CreateSupportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if validationErrors := utils.ValidateStruct(req); validationErrors != nil {
		utils.ValidationErrorResponse(w, validationErrors)
		return
	}

	// Get employee record for requester
	employee, err := h.getEmployeeByUserID(r, userID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusForbidden, "User not linked to employeeee")
		return
	}

	// Find support department by category
	dept, err := h.db.GetSupportDepartmentByCategory(r.Context(), database.GetSupportDepartmentByCategoryParams{
		TenantID: tenantID,
		Category: &req.SupportCategory,
	})
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "No support department found for this category")
		return
	}

	// Find available support agent in department
	// agent, err := h.db.GetAvailableSupportAgent(r.Context(), database.GetAvailableSupportAgentParams{
	// 	DepartmentID: dept.ID,
	// 	TenantID:     tenantID,
	// })

	var assigneeID *uuid.UUID
	// if err == nil && agent.ID != uuid.Nil {
	// 	assigneeID = &agent.ID
	// }

	// Generate request number
	nextNum, _ := h.db.GetNextRequestNumber(r.Context(), tenantID)
	requestNumber := generateRequestNumber(nextNum)

	var categoryPtr *string
	if req.SupportCategory != "" {
		categoryPtr = &req.SupportCategory
	}

	var descriptionPtr *string
	if req.Description != "" {
		descriptionPtr = &req.Description
	}

	// Create request with support routing
	created, err := h.db.CreateRequest(r.Context(), database.CreateRequestParams{
		TenantID:            tenantID,
		RequestNumber:       requestNumber,
		RequestType:         "support",
		Category:            categoryPtr,
		Title:               req.Title,
		Description:         descriptionPtr,
		RequesterID:         employee.ID,
		AssigneeID:          assigneeID,
		RoutedToManagerID:   nil, // Support requests don't route to manager
		SupportDepartmentID: &dept.ID,
		SupportCategory:     categoryPtr,
		Priority:            req.Priority,
		DueDate:             nil,
	})
	if err != nil {
		utils.HandleDBError(w, err, "CreateSupportRequest")
		return
	}

	utils.SuccessResponse(w, http.StatusCreated, created, "Support request created successfully")
}

// GetDepartmentRequests returns filtered department requests
func (h *Handler) GetDepartmentRequests(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := getTenantID(r)
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID, _ := r.Context().Value(constants.ContextKeyUserID).(uuid.UUID)
	userType, _ := r.Context().Value(constants.ContextKeyUserType).(string)

	// Read all query params
	filter := r.URL.Query().Get("filter")     // unassigned | mine | team | all | resolved
	status := r.URL.Query().Get("status")     // new | in_progress | ...
	priority := r.URL.Query().Get("priority") // low | medium | high | urgent
	search := r.URL.Query().Get("search")     // text search
	page, pageSize := getPagination(r)

	// Default filter if empty
	if filter == "" {
		filter = "all"
	}

	if userType == constants.UserTypeTenantOwner {
		requests, err := h.db.GetAllTenantSupportRequests(r.Context(), database.GetAllTenantSupportRequestsParams{
			TenantID:   tenantID,
			Filter:     filter,
			EmployeeID: nil,
			Status:     status,
			Priority:   priority,
			Search:     search,
			PageSize:   int32(pageSize),
			PageOffset: int32((page - 1) * pageSize),
		})
		if err != nil {
			utils.HandleDBError(w, err, "GetAllTenantSupportRequests")
			return
		}

		// Counts across whole tenant
		counts, err := h.db.GetTenantSupportRequestCounts(r.Context(), database.GetTenantSupportRequestCountsParams{
			TenantID:   tenantID,
			EmployeeID: &userID, // owner has no employee, use userID (always non-null)
		})
		if err != nil {
			log.Printf("Failed to get counts: %v", err)
			counts = database.GetTenantSupportRequestCountsRow{}
		}

		utils.SuccessResponse(w, http.StatusOK, map[string]any{
			"requests": requests,
			"counts": map[string]int64{
				"unassigned": counts.UnassignedCount,
				"mine":       counts.MineCount,
				"team":       counts.TeamCount,
				"all":        counts.AllCount,
			},
		}, "All support requests retrieved")
		return
	}
	employee, err := h.getEmployeeByUserID(r, userID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusForbidden, "User not linked to employee")
		return
	}
	requests, err := h.db.GetDepartmentRequests(r.Context(), database.GetDepartmentRequestsParams{
		TenantID:            tenantID,
		SupportDepartmentID: &employee.DepartmentID,
		Filter:              filter,
		EmployeeID:          &employee.ID,
		Status:              status,
		Priority:            priority,
		Search:              search,
		PageSize:            int32(pageSize),
		PageOffset:          int32((page - 1) * pageSize),
	})
	if err != nil {
		utils.HandleDBError(w, err, "GetDepartmentRequests")
		return
	}

	counts, err := h.db.GetDepartmentRequestCounts(r.Context(), database.GetDepartmentRequestCountsParams{
		TenantID:            tenantID,
		SupportDepartmentID: &employee.DepartmentID,
		AssigneeID:          &employee.ID,
	})
	if err != nil {
		// Log but don't fail the request
		log.Printf("Failed to get counts: %v", err)
		counts = database.GetDepartmentRequestCountsRow{}
	}

	utils.SuccessResponse(w, http.StatusOK, map[string]any{
		"requests": requests,
		"counts": map[string]int64{
			"unassigned": counts.UnassignedCount,
			"mine":       counts.MineCount,
			"team":       counts.TeamCount,
			"all":        counts.AllCount,
		},
	}, "Department requests retrieved")
}

// AcceptRequest assigns the request to current user (support agent)
func (h *Handler) AcceptRequest(w http.ResponseWriter, r *http.Request) {
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

	userID, _ := r.Context().Value(constants.ContextKeyUserID).(uuid.UUID)
	employee, err := h.getEmployeeByUserID(r, userID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusForbidden, "User not linked to employee")
		return
	}

	log.Printf("AcceptRequest: requestID=%s, tenantID=%s", requestID, tenantID)
	log.Printf("  userID=%s", userID)
	log.Printf("  employee.ID=%s", employee.ID)
	log.Printf("  employee.departmentID=%s", employee.DepartmentID)
	// Update request assignee to current user
	updated, err := h.db.UpdateRequestAssignee(r.Context(), database.UpdateRequestAssigneeParams{
		ID:         requestID,
		TenantID:   tenantID,
		AssigneeID: &employee.ID,
	})
	if err != nil {
		utils.HandleDBError(w, err, "UpdateRequestAssignee")
		return
	}

	utils.SuccessResponse(w, http.StatusOK, updated, "Request accepted successfully")
}

func (h *Handler) GetSupportDepartments(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := getTenantID(r)
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	depts, err := h.db.GetSupportDepartments(r.Context(), tenantID)
	if err != nil {
		utils.HandleDBError(w, err, "GetSupportDepartments")
		return
	}

	utils.SuccessResponse(w, http.StatusOK, depts, "Support departments retrieved")
}

func (h *Handler) GetSupportTeamMembers(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := getTenantID(r)
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	deptID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid department ID")
		return
	}

	members, err := h.db.GetDepartmentMembers(r.Context(), database.GetDepartmentMembersParams{
		TenantID:     tenantID,
		DepartmentID: deptID,
	})
	if err != nil {
		utils.HandleDBError(w, err, "GetDepartmentMembers")
		return
	}

	utils.SuccessResponse(w, http.StatusOK, members, "Department members retrieved")
}

func (h *Handler) GetCurrentDepartmentMembers(w http.ResponseWriter, r *http.Request) {
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

	members, err := h.db.GetDepartmentMembers(r.Context(), database.GetDepartmentMembersParams{
		TenantID:     tenantID,
		DepartmentID: employee.DepartmentID,
	})
	if err != nil {
		utils.HandleDBError(w, err, "GetDepartmentMembers")
		return
	}

	utils.SuccessResponse(w, http.StatusOK, members, "Department members retrieved")
}

// ============ CHAT HANDLERS ============

// SendChatMessage sends a chat message on a request
// func (h *Handler) SendChatMessage(w http.ResponseWriter, r *http.Request) {
// 	requestID, err := uuid.Parse(r.PathValue("id"))
// 	if err != nil {
// 		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request ID")
// 		return
// 	}

// 	userID, _ := r.Context().Value(constants.ContextKeyUserID").(uuid.UUID)

// 	var req struct {
// 		Message string `json:"message" validate:"required,min=1,max=2000"`
// 	}
// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
// 		return
// 	}

// 	if validationErrors := utils.ValidateStruct(req); validationErrors != nil {
// 		utils.ValidationErrorResponse(w, validationErrors)
// 		return
// 	}

// 	// Create chat message
// 	message, err := h.db.CreateChatMessage(r.Context(), database.CreateChatMessageParams{
// 		RequestID: requestID,
// 		SenderID:  userID,
// 		Message:   req.Message,
// 	})
// 	if err != nil {
// 		utils.HandleDBError(w, err, "CreateChatMessage")
// 		return
// 	}

// 	utils.SuccessResponse(w, http.StatusCreated, message, "Message sent successfully")
// }

// GetChatMessages returns all chat messages for a request
func (h *Handler) GetChatMessages(w http.ResponseWriter, r *http.Request) {
	requestID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request ID")
		return
	}

	messages, err := h.db.GetChatMessages(r.Context(), requestID)
	if err != nil {
		utils.HandleDBError(w, err, "GetChatMessages")
		return
	}

	utils.SuccessResponse(w, http.StatusOK, messages, "Messages retrieved successfully")
}

// MarkChatRead marks all messages as read for current user
func (h *Handler) MarkChatRead(w http.ResponseWriter, r *http.Request) {
	requestID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request ID")
		return
	}

	userID, _ := r.Context().Value(constants.ContextKeyUserID).(uuid.UUID)

	err = h.db.MarkChatMessagesRead(r.Context(), database.MarkChatMessagesReadParams{
		RequestID: requestID,
		SenderID:  userID,
	})
	if err != nil {
		utils.HandleDBError(w, err, "MarkChatMessagesRead")
		return
	}

	utils.SuccessResponse(w, http.StatusOK, nil, "Messages marked as read")
}

// GetUnreadCount returns unread message count for a request
func (h *Handler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	requestID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request ID")
		return
	}

	userID, _ := r.Context().Value(constants.ContextKeyUserID).(uuid.UUID)

	count, err := h.db.GetUnreadChatCount(r.Context(), database.GetUnreadChatCountParams{
		RequestID: requestID,
		SenderID:  userID,
	})
	if err != nil {
		utils.HandleDBError(w, err, "GetUnreadChatCount")
		return
	}

	utils.SuccessResponse(w, http.StatusOK, map[string]interface{}{
		"unreadCount": count,
	}, "Unread count retrieved")
}

// Helper function to generate request number
func generateRequestNumber(nextNum int32) string {
	return "REQ-" + padNumber(int64(nextNum))
}

func padNumber(num int64) string {
	s := strconv.FormatInt(num, 10)
	for len(s) < 4 {
		s = "0" + s
	}
	return s
}

// SendChatMessage sends a chat message and broadcasts via WebSocket
func (h *Handler) SendChatMessage(w http.ResponseWriter, r *http.Request) {
	requestID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request ID")
		return
	}

	userID, _ := r.Context().Value(constants.ContextKeyUserID).(uuid.UUID)

	var req struct {
		Message string `json:"message" validate:"required,min=1,max=2000"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if validationErrors := utils.ValidateStruct(req); validationErrors != nil {
		utils.ValidationErrorResponse(w, validationErrors)
		return
	}

	// Save to database
	message, err := h.db.CreateChatMessage(r.Context(), database.CreateChatMessageParams{
		RequestID: requestID,
		SenderID:  userID,
		Message:   req.Message,
	})
	if err != nil {
		utils.HandleDBError(w, err, "CreateChatMessage")
		return
	}

	// Broadcast via WebSocket
	h.hub.broadcast <- &ChatMessage{
		ID:        message.ID,
		Type:      "message",
		RequestID: requestID,
		UserID:    userID,
		Message:   req.Message,
		CreatedAt: time.Now(),
	}

	utils.SuccessResponse(w, http.StatusCreated, message, "Message sent successfully")
}
