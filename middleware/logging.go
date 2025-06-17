package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"s3mgr/logger"
)

type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// RequestLogger creates a middleware that logs all HTTP requests with detailed information
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Capture request body size
		var requestSize int64
		if c.Request.Body != nil {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			requestSize = int64(len(bodyBytes))
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// Create custom response writer to capture response size
		blw := &responseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = blw

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Get user information from context (set by auth middleware)
		userID, _ := c.Get("user_id")
		username, _ := c.Get("username")

		// Determine if there was an error
		var errorMsg string
		if len(c.Errors) > 0 {
			errorMsg = c.Errors.String()
		}

		// Log the request
		logger.LogRequest(logger.RequestLog{
			Timestamp:    start,
			Method:       c.Request.Method,
			Path:         c.Request.URL.Path,
			StatusCode:   c.Writer.Status(),
			Duration:     duration.String(),
			ClientIP:     c.ClientIP(),
			UserAgent:    c.Request.UserAgent(),
			UserID:       getStringValue(userID),
			Username:     getStringValue(username),
			RequestSize:  requestSize,
			ResponseSize: blw.body.Len(),
			Error:        errorMsg,
		})
	}
}

// AuthLogger logs authentication events
func LogAuthEvent(c *gin.Context, action string, username string, success bool, err error, sessionID ...string) {
	var errorMsg string
	if err != nil {
		errorMsg = err.Error()
	}

	var sid string
	if len(sessionID) > 0 {
		sid = sessionID[0]
	}

	userID, _ := c.Get("user_id")

	logger.LogAuth(logger.AuthLog{
		Timestamp: time.Now(),
		Action:    action,
		Username:  username,
		UserID:    getStringValue(userID),
		ClientIP:  c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Success:   success,
		Error:     errorMsg,
		SessionID: sid,
	})
}

// ConfigLogger logs configuration management events
func LogConfigEvent(c *gin.Context, action string, configID string, details string, success bool, err error) {
	var errorMsg string
	if err != nil {
		errorMsg = err.Error()
	}

	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")

	logger.LogConfigEvent(logger.ConfigLog{
		Timestamp: time.Now(),
		Action:    action,
		ConfigID:  configID,
		UserID:    getStringValue(userID),
		Username:  getStringValue(username),
		ClientIP:  c.ClientIP(),
		Details:   details,
		Success:   success,
		Error:     errorMsg,
	})
}

// FileLogger logs file operation events
func LogFileEvent(c *gin.Context, action string, fileName string, fileSize int64, configID string, success bool, duration time.Duration, err error) {
	var errorMsg string
	if err != nil {
		errorMsg = err.Error()
	}

	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")

	logger.LogFile(logger.FileLog{
		Timestamp: time.Now(),
		Action:    action,
		FileName:  fileName,
		FileSize:  fileSize,
		ConfigID:  configID,
		UserID:    getStringValue(userID),
		Username:  getStringValue(username),
		ClientIP:  c.ClientIP(),
		Success:   success,
		Error:     errorMsg,
		Duration:  duration.String(),
	})
}

func getStringValue(value interface{}) string {
	if value == nil {
		return ""
	}
	if str, ok := value.(string); ok {
		return str
	}
	return ""
}
