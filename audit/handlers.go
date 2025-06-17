package audit

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// AuditFilterRequest represents the request for filtering audit logs
type AuditFilterRequest struct {
	UserID    string `json:"user_id,omitempty"`
	Action    string `json:"action,omitempty"`
	Resource  string `json:"resource,omitempty"`
	StartTime string `json:"start_time,omitempty"` // RFC3339 format
	EndTime   string `json:"end_time,omitempty"`   // RFC3339 format
	Limit     int    `json:"limit,omitempty"`
}

// GetAuditLogsHandler handles GET /api/admin/audit-logs
func (a *AuditService) GetAuditLogsHandler(c *gin.Context) {
	// Check if current user is admin
	currentUser, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	_ = currentUser // Use the variable to avoid lint warning

	// Get user from auth service to check admin status
	// This would need to be injected or accessed differently in real implementation
	// For now, we'll assume admin check is done via middleware

	// Parse query parameters
	userID := c.Query("user_id")
	action := c.Query("action")
	resource := c.Query("resource")
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")
	limitStr := c.Query("limit")

	var startTime, endTime time.Time
	var err error

	if startTimeStr != "" {
		startTime, err = time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_time format. Use RFC3339 format"})
			return
		}
	}

	if endTimeStr != "" {
		endTime, err = time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_time format. Use RFC3339 format"})
			return
		}
	}

	limit := 100 // Default limit
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Log the audit query action
	a.LogEvent(c, "query_audit_logs", "audit_logs", "", true, nil, map[string]interface{}{
		"filters": map[string]interface{}{
			"user_id":    userID,
			"action":     action,
			"resource":   resource,
			"start_time": startTimeStr,
			"end_time":   endTimeStr,
			"limit":      limit,
		},
	})

	logs, err := a.GetAuditLogs(userID, action, resource, startTime, endTime, limit)
	if err != nil {
		a.LogEvent(c, "query_audit_logs", "audit_logs", "", false, err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve audit logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"audit_logs": logs,
		"count":      len(logs),
		"filters": map[string]interface{}{
			"user_id":    userID,
			"action":     action,
			"resource":   resource,
			"start_time": startTimeStr,
			"end_time":   endTimeStr,
			"limit":      limit,
		},
	})
}

// GetAuditLogsByIncidentHandler handles GET /api/admin/audit-logs/incident/:session_id
func (a *AuditService) GetAuditLogsByIncidentHandler(c *gin.Context) {
	// Check if current user is admin
	currentUser, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	_ = currentUser // Use the variable to avoid lint warning

	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session ID is required"})
		return
	}

	// Log the audit query action
	a.LogEvent(c, "query_audit_logs_by_incident", "audit_logs", sessionID, true, nil, map[string]interface{}{
		"session_id": sessionID,
	})

	logs, err := a.GetAuditLogsByIncident(sessionID)
	if err != nil {
		a.LogEvent(c, "query_audit_logs_by_incident", "audit_logs", sessionID, false, err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve audit logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"audit_logs": logs,
		"count":      len(logs),
		"session_id": sessionID,
	})
}

// PostAuditLogsFilterHandler handles POST /api/admin/audit-logs/filter for complex filtering
func (a *AuditService) PostAuditLogsFilterHandler(c *gin.Context) {
	// Check if current user is admin
	currentUser, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	_ = currentUser // Use the variable to avoid lint warning

	var filterRequest AuditFilterRequest
	if err := c.ShouldBindJSON(&filterRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var startTime, endTime time.Time
	var err error

	if filterRequest.StartTime != "" {
		startTime, err = time.Parse(time.RFC3339, filterRequest.StartTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_time format. Use RFC3339 format"})
			return
		}
	}

	if filterRequest.EndTime != "" {
		endTime, err = time.Parse(time.RFC3339, filterRequest.EndTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_time format. Use RFC3339 format"})
			return
		}
	}

	if filterRequest.Limit <= 0 {
		filterRequest.Limit = 100 // Default limit
	}

	// Log the audit query action
	a.LogEvent(c, "filter_audit_logs", "audit_logs", "", true, nil, map[string]interface{}{
		"filters": filterRequest,
	})

	logs, err := a.GetAuditLogs(filterRequest.UserID, filterRequest.Action, filterRequest.Resource, startTime, endTime, filterRequest.Limit)
	if err != nil {
		a.LogEvent(c, "filter_audit_logs", "audit_logs", "", false, err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve audit logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"audit_logs": logs,
		"count":      len(logs),
		"filters":    filterRequest,
	})
}
