package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"social-network/pkg/middleware"
	"social-network/pkg/models"
	"social-network/pkg/repository"
	"social-network/pkg/repository/group"
	"social-network/pkg/utils"
	"social-network/pkg/websocket"
)

// validChoices is the single source of truth for allowed RSVP values.
// The DB CHECK constraint mirrors this, but we validate here so the error
// message is clear and we never hit the database with garbage.
var validChoices = map[string]bool{
	"going":     true,
	"not going": true,
	"maybe":     true,
}

// GroupEventHandler handles group event and RSVP operations
type GroupEventHandler struct {
	EventRepo        *group.GroupEventRepository
	MemberRepo       *group.GroupMemberRepository
	GroupRepo        *group.GroupRepository
	NotificationRepo *repository.NotificationRepository
	Hub              *websocket.Hub
}

// NewGroupEventHandler creates a new GroupEventHandler
func NewGroupEventHandler(
	eventRepo *group.GroupEventRepository,
	memberRepo *group.GroupMemberRepository,
	groupRepo *group.GroupRepository,
	notificationRepo *repository.NotificationRepository,
	hub *websocket.Hub,
) *GroupEventHandler {
	return &GroupEventHandler{
		EventRepo:        eventRepo,
		MemberRepo:       memberRepo,
		GroupRepo:        groupRepo,
		NotificationRepo: notificationRepo,
		Hub:              hub,
	}
}

// CreateEvent creates a new group event
// @Summary      Create group event
// @Description  Creates a new event in a group (members only)
// @Tags         Groups
// @Security     CookieAuth
// @Accept       json
// @Produce      json
// @Param        id    path      string  true  "Group ID"
// @Param        data  body      object  true  "Event details"
// @Success      201   {object}  models.GroupEventWithDetails
// @Failure      400   {object}  models.ErrorResponse
// @Failure      401   {object}  models.ErrorResponse
// @Failure      403   {object}  models.ErrorResponse
// @Router       /api/v1/groups/{id}/events [post]
func (h *GroupEventHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract group ID from URL
	groupID := strings.TrimPrefix(r.URL.Path, "/api/v1/groups/events/create/")

	if groupID == "" {
		utils.ErrorResponse(w, "Group ID is required", http.StatusBadRequest)
		return
	}

	// Check if user is a member
	isMember, err := h.MemberRepo.IsMember(groupID, user.ID)
	if err != nil {
		log.Printf("Failed to check membership: %v", err)
		utils.ErrorResponse(w, "Failed to verify membership", http.StatusInternalServerError)
		return
	}

	if !isMember {
		utils.ErrorResponse(w, "Only group members can create events", http.StatusForbidden)
		return
	}

	// Parse request body
	var req struct {
		Title       string    `json:"title"`
		Description string    `json:"description"`
		EventTime   time.Time `json:"event_time"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate inputs
	req.Title = strings.TrimSpace(req.Title)
	req.Description = strings.TrimSpace(req.Description)

	if req.Title == "" {
		utils.ErrorResponse(w, "Event title is required", http.StatusBadRequest)
		return
	}

	if req.EventTime.IsZero() {
		utils.ErrorResponse(w, "Event time is required", http.StatusBadRequest)
		return
	}

	// Check if event time is in the future
	if req.EventTime.Before(time.Now()) {
		utils.ErrorResponse(w, "Event time must be in the future", http.StatusBadRequest)
		return
	}

	var description *string
	if req.Description != "" {
		description = &req.Description
	}

	// Create event
	event := models.GroupEvent{
		GroupID:     groupID,
		CreatorID:   user.ID,
		Title:       req.Title,
		Description: description,
		EventTime:   req.EventTime,
	}

	createdEvent, err := h.EventRepo.CreateEvent(event)
	if err != nil {
		log.Printf("Failed to create event: %v", err)
		utils.ErrorResponse(w, "Failed to create event", http.StatusInternalServerError)
		return
	}

	// Notify all other members that a new group event was created.
	members, err := h.MemberRepo.GetGroupMembers(groupID)
	if err != nil {
		log.Printf("Failed to load group members for notifications: %v", err)
	} else {
		for _, member := range members {
			if member.UserID == "" || member.UserID == user.ID {
				continue
			}
			createAndPushNotification(h.NotificationRepo, h.Hub, member.UserID, user.ID, user.Nickname, "group_event")
		}
	}

	// Return event with details
	eventWithDetails, err := h.EventRepo.GetEventWithDetails(createdEvent.ID, user.ID)
	if err != nil {
		log.Printf("Failed to get event details: %v", err)
		utils.ErrorResponse(w, "Failed to retrieve event details", http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, eventWithDetails, http.StatusCreated)
}

// GetGroupEvents retrieves all events for a group
// @Summary      Get group events
// @Description  Retrieves all events for a group (members only)
// @Tags         Groups
// @Security     CookieAuth
// @Produce      json
// @Param        id   path      string  true  "Group ID"
// @Success      200  {array}   models.GroupEvent
// @Failure      401  {object}  models.ErrorResponse
// @Failure      403  {object}  models.ErrorResponse
// @Router       /api/v1/groups/{id}/events [get]
func (h *GroupEventHandler) GetGroupEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract group ID from URL
	groupID := strings.TrimPrefix(r.URL.Path, "/api/v1/groups/events/")

	if groupID == "" {
		utils.ErrorResponse(w, "Group ID is required", http.StatusBadRequest)
		return
	}

	// Check if user is a member
	isMember, err := h.MemberRepo.IsMember(groupID, user.ID)
	if err != nil {
		log.Printf("Failed to check membership: %v", err)
		utils.ErrorResponse(w, "Failed to verify membership", http.StatusInternalServerError)
		return
	}

	if !isMember {
		utils.ErrorResponse(w, "Only group members can view events", http.StatusForbidden)
		return
	}

	// Get events
	events, err := h.EventRepo.GetGroupEvents(groupID)
	if err != nil {
		log.Printf("Failed to get events: %v", err)
		utils.ErrorResponse(w, "Failed to retrieve events", http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, map[string]interface{}{
		"events": events,
	}, http.StatusOK)
}

// GetEventDetails retrieves detailed event information with RSVP stats
// @Summary      Get event details
// @Description  Retrieves event details with RSVP options and vote counts
// @Tags         Groups
// @Security     CookieAuth
// @Produce      json
// @Param        id   path      string  true  "Event ID"
// @Success      200  {object}  models.GroupEventWithDetails
// @Failure      401  {object}  models.ErrorResponse
// @Failure      403  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Router       /api/v1/events/{id} [get]
func (h *GroupEventHandler) GetEventDetails(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract event ID from URL
	eventID := strings.TrimPrefix(r.URL.Path, "/api/v1/events/")

	if eventID == "" {
		utils.ErrorResponse(w, "Event ID is required", http.StatusBadRequest)
		return
	}

	// Get basic event info to check group membership
	event, err := h.EventRepo.GetEventByID(eventID)
	if err != nil {
		log.Printf("Failed to get event: %v", err)
		utils.ErrorResponse(w, "Failed to retrieve event", http.StatusInternalServerError)
		return
	}

	if event == nil {
		utils.ErrorResponse(w, "Event not found", http.StatusNotFound)
		return
	}

	// Check if user is a member of the group
	isMember, err := h.MemberRepo.IsMember(event.GroupID, user.ID)
	if err != nil {
		log.Printf("Failed to check membership: %v", err)
		utils.ErrorResponse(w, "Failed to verify membership", http.StatusInternalServerError)
		return
	}

	if !isMember {
		utils.ErrorResponse(w, "Only group members can view events", http.StatusForbidden)
		return
	}

	// Get event with details
	eventWithDetails, err := h.EventRepo.GetEventWithDetails(eventID, user.ID)
	if err != nil {
		log.Printf("Failed to get event details: %v", err)
		utils.ErrorResponse(w, "Failed to retrieve event details", http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, eventWithDetails, http.StatusOK)
}

// VoteOnEvent records or updates a user's RSVP for an event
// @Summary      Vote on event
// @Description  Records user's RSVP choice for an event.  Valid choices: "going", "not going", "maybe".
// @Tags         Groups
// @Security     CookieAuth
// @Accept       json
// @Produce      json
// @Param        id    path      string  true   "Event ID"
// @Param        data  body      object  true   "Choice to record"
// @Success      200   {object}  map[string]string
// @Failure      400   {object}  models.ErrorResponse
// @Failure      401   {object}  models.ErrorResponse
// @Failure      403   {object}  models.ErrorResponse
// @Router       /api/v1/events/{id}/vote [post]
func (h *GroupEventHandler) VoteOnEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract event ID from URL
	eventID := strings.TrimPrefix(r.URL.Path, "/api/v1/events/vote/")

	if eventID == "" {
		utils.ErrorResponse(w, "Event ID is required", http.StatusBadRequest)
		return
	}

	// Parse request body  ───────────────────────────────────────────────────
	var req struct {
		Choice string `json:"choice"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	req.Choice = strings.TrimSpace(req.Choice)

	if req.Choice == "" {
		utils.ErrorResponse(w, "Choice is required", http.StatusBadRequest)
		return
	}

	// Validate before hitting the database.  The CHECK constraint on the table
	// is the last line of defence; this gives the caller an actionable message.
	if !validChoices[req.Choice] {
		utils.ErrorResponse(w, `Invalid choice. Must be one of: "going", "not going", "maybe"`, http.StatusBadRequest)
		return
	}

	// Get event to verify group membership ──────────────────────────────────
	event, err := h.EventRepo.GetEventByID(eventID)
	if err != nil {
		log.Printf("Failed to get event: %v", err)
		utils.ErrorResponse(w, "Failed to retrieve event", http.StatusInternalServerError)
		return
	}

	if event == nil {
		utils.ErrorResponse(w, "Event not found", http.StatusNotFound)
		return
	}

	// Check if user is a member
	isMember, err := h.MemberRepo.IsMember(event.GroupID, user.ID)
	if err != nil {
		log.Printf("Failed to check membership: %v", err)
		utils.ErrorResponse(w, "Failed to verify membership", http.StatusInternalServerError)
		return
	}

	if !isMember {
		utils.ErrorResponse(w, "Only group members can vote on events", http.StatusForbidden)
		return
	}

	// Record vote (replaces any existing vote for this user + event) ─────────
	if err := h.EventRepo.VoteOnEvent(eventID, user.ID, req.Choice); err != nil {
		log.Printf("Failed to record vote: %v", err)
		utils.ErrorResponse(w, "Failed to record vote", http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, map[string]string{
		"message": "Vote recorded successfully",
	}, http.StatusOK)
}

// DeleteEvent deletes an event
// @Summary      Delete event
// @Description  Deletes an event (creator or group owner only)
// @Tags         Groups
// @Security     CookieAuth
// @Produce      json
// @Param        id   path      string  true  "Event ID"
// @Success      200  {object}  map[string]string
// @Failure      401  {object}  models.ErrorResponse
// @Failure      403  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Router       /api/v1/events/{id} [delete]
func (h *GroupEventHandler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract event ID from URL
	eventID := strings.TrimPrefix(r.URL.Path, "/api/v1/events/delete/")

	if eventID == "" {
		utils.ErrorResponse(w, "Event ID is required", http.StatusBadRequest)
		return
	}

	// Get event details
	event, err := h.EventRepo.GetEventByID(eventID)
	if err != nil {
		log.Printf("Failed to get event: %v", err)
		utils.ErrorResponse(w, "Failed to retrieve event", http.StatusInternalServerError)
		return
	}

	if event == nil {
		utils.ErrorResponse(w, "Event not found", http.StatusNotFound)
		return
	}

	// Check if user is the creator or group owner
	isCreator := event.CreatorID == user.ID
	isGroupOwner, err := h.GroupRepo.IsUserOwner(event.GroupID, user.ID)
	if err != nil {
		log.Printf("Failed to check ownership: %v", err)
		utils.ErrorResponse(w, "Failed to verify permissions", http.StatusInternalServerError)
		return
	}

	if !isCreator && !isGroupOwner {
		utils.ErrorResponse(w, "Only the event creator or group owner can delete this event", http.StatusForbidden)
		return
	}

	// Delete event
	if err := h.EventRepo.DeleteEvent(eventID); err != nil {
		log.Printf("Failed to delete event: %v", err)
		utils.ErrorResponse(w, "Failed to delete event", http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, map[string]string{
		"message": "Event deleted successfully",
	}, http.StatusOK)
}
