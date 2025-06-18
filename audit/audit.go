package audit

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/gin-gonic/gin"
)

// AuditLog represents an audit log entry
type AuditLog struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	UserID      string                 `json:"user_id"`
	Username    string                 `json:"username"`
	Action      string                 `json:"action"`
	Resource    string                 `json:"resource"`
	ResourceID  string                 `json:"resource_id,omitempty"`
	ClientIP    string                 `json:"client_ip"`
	UserAgent   string                 `json:"user_agent"`
	Success     bool                   `json:"success"`
	Error       string                 `json:"error,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
	SessionID   string                 `json:"session_id,omitempty"`
}

// AuditService handles audit logging
type AuditService struct {
	db *badger.DB
}

// NewAuditService creates a new audit service
func NewAuditService(db *badger.DB) *AuditService {
	return &AuditService{
		db: db,
	}
}

// LogEvent logs an audit event
func (a *AuditService) LogEvent(c *gin.Context, action, resource, resourceID string, success bool, err error, details map[string]interface{}) {
	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")
	sessionID, _ := c.Get("session_id")

	var errorMsg string
	if err != nil {
		errorMsg = err.Error()
	}

	auditLog := AuditLog{
		ID:         fmt.Sprintf("audit_%d", time.Now().UnixNano()),
		Timestamp:  time.Now(),
		UserID:     GetStringValue(userID),
		Username:   GetStringValue(username),
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		ClientIP:   c.ClientIP(),
		UserAgent:  c.GetHeader("User-Agent"),
		Success:    success,
		Error:      errorMsg,
		Details:    details,
		SessionID:  GetStringValue(sessionID),
	}

	// Store in database
	data, _ := json.Marshal(auditLog)
	a.db.Update(func(txn *badger.Txn) error {
		key := fmt.Sprintf("audit:%s", auditLog.ID)
		return txn.Set([]byte(key), data)
	})
}

// GetAuditLogs retrieves audit logs with filtering
func (a *AuditService) GetAuditLogs(userID, action, resource string, startTime, endTime time.Time, offset, limit int) ([]AuditLog, error) {
	var logs []AuditLog

	err := a.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte("audit:")
		count := 0
		skipped := 0
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var log AuditLog
				if err := json.Unmarshal(val, &log); err != nil {
					return err
				}

				// Apply filters
				if userID != "" && log.UserID != userID {
					return nil
				}
				if action != "" && log.Action != action {
					return nil
				}
				if resource != "" && log.Resource != resource {
					return nil
				}
				if !startTime.IsZero() && log.Timestamp.Before(startTime) {
					return nil
				}
				if !endTime.IsZero() && log.Timestamp.After(endTime) {
					return nil
				}

				if skipped < offset {
					skipped++
					return nil
				}

				if limit > 0 && count >= limit {
					return nil
				}

				logs = append(logs, log)
				count++
				return nil
			})
			if err != nil {
				return err
			}
			if limit > 0 && count >= limit {
				break
			}
		}
		return nil
	})

	// Sort logs by Timestamp descending
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].Timestamp.After(logs[j].Timestamp)
	})
	return logs, err
}

// GetAuditLogsByIncident retrieves audit logs for a specific incident/session
func (a *AuditService) GetAuditLogsByIncident(sessionID string) ([]AuditLog, error) {
	var logs []AuditLog

	err := a.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte("audit:")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var log AuditLog
				if err := json.Unmarshal(val, &log); err != nil {
					return err
				}

				if log.SessionID == sessionID {
					logs = append(logs, log)
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	return logs, err
}

// Helper function to safely convert interface{} to string
func GetStringValue(value interface{}) string {
	if value == nil {
		return ""
	}
	if str, ok := value.(string); ok {
		return str
	}
	return fmt.Sprintf("%v", value)
}
