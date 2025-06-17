package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	Logger *logrus.Logger
	config LogConfig
)

type LogConfig struct {
	Level       string `yaml:"level"`
	File        string `yaml:"file"`
	MaxSize     int    `yaml:"max_size"`
	MaxBackups  int    `yaml:"max_backups"`
	MaxAge      int    `yaml:"max_age"`
	Compress    bool   `yaml:"compress"`
	Console     bool   `yaml:"console"`
	Format      string `yaml:"format"`
}

type RequestLog struct {
	Timestamp    time.Time `json:"timestamp"`
	Method       string    `json:"method"`
	Path         string    `json:"path"`
	StatusCode   int       `json:"status_code"`
	Duration     string    `json:"duration"`
	ClientIP     string    `json:"client_ip"`
	UserAgent    string    `json:"user_agent"`
	UserID       string    `json:"user_id,omitempty"`
	Username     string    `json:"username,omitempty"`
	RequestSize  int64     `json:"request_size"`
	ResponseSize int       `json:"response_size"`
	Error        string    `json:"error,omitempty"`
}

type AuthLog struct {
	Timestamp time.Time `json:"timestamp"`
	Action    string    `json:"action"` // login, logout, register
	Username  string    `json:"username"`
	UserID    string    `json:"user_id,omitempty"`
	ClientIP  string    `json:"client_ip"`
	UserAgent string    `json:"user_agent"`
	Success   bool      `json:"success"`
	Error     string    `json:"error,omitempty"`
	SessionID string    `json:"session_id,omitempty"`
}

type ConfigLog struct {
	Timestamp time.Time `json:"timestamp"`
	Action    string    `json:"action"` // create, update, delete, set_default
	ConfigID  string    `json:"config_id"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	ClientIP  string    `json:"client_ip"`
	Details   string    `json:"details,omitempty"`
	Success   bool      `json:"success"`
	Error     string    `json:"error,omitempty"`
}

type FileLog struct {
	Timestamp time.Time `json:"timestamp"`
	Action    string    `json:"action"` // upload, download, delete, list
	FileName  string    `json:"file_name"`
	FileSize  int64     `json:"file_size,omitempty"`
	ConfigID  string    `json:"config_id"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	ClientIP  string    `json:"client_ip"`
	Success   bool      `json:"success"`
	Error     string    `json:"error,omitempty"`
	Duration  string    `json:"duration,omitempty"`
}

// Initialize sets up the logger with the given configuration
func Initialize(cfg LogConfig) error {
	config = cfg
	Logger = logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		return fmt.Errorf("invalid log level: %v", err)
	}
	Logger.SetLevel(level)

	// Set formatter
	if cfg.Format == "json" {
		Logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	} else {
		Logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
		})
	}

	// Set up file logging with rotation
	var writers []io.Writer

	if cfg.File != "" {
		// Create log directory if it doesn't exist
		logDir := filepath.Dir(cfg.File)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %v", err)
		}

		// Set up log rotation
		fileWriter := &lumberjack.Logger{
			Filename:   cfg.File,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		}
		writers = append(writers, fileWriter)
	}

	// Add console output if enabled
	if cfg.Console {
		writers = append(writers, os.Stdout)
	}

	if len(writers) > 0 {
		Logger.SetOutput(io.MultiWriter(writers...))
	}

	return nil
}

// LogRequest logs HTTP request details
func LogRequest(req RequestLog) {
	Logger.WithFields(logrus.Fields{
		"type":          "request",
		"method":        req.Method,
		"path":          req.Path,
		"status_code":   req.StatusCode,
		"duration":      req.Duration,
		"client_ip":     req.ClientIP,
		"user_agent":    req.UserAgent,
		"user_id":       req.UserID,
		"username":      req.Username,
		"request_size":  req.RequestSize,
		"response_size": req.ResponseSize,
		"error":         req.Error,
	}).Info("HTTP Request")
}

// LogAuth logs authentication events
func LogAuth(auth AuthLog) {
	level := logrus.InfoLevel
	if !auth.Success {
		level = logrus.WarnLevel
	}

	Logger.WithFields(logrus.Fields{
		"type":       "auth",
		"action":     auth.Action,
		"username":   auth.Username,
		"user_id":    auth.UserID,
		"client_ip":  auth.ClientIP,
		"user_agent": auth.UserAgent,
		"success":    auth.Success,
		"error":      auth.Error,
		"session_id": auth.SessionID,
	}).Log(level, fmt.Sprintf("Auth %s", auth.Action))
}

// LogConfigEvent logs configuration management events
func LogConfigEvent(cfg ConfigLog) {
	level := logrus.InfoLevel
	if !cfg.Success {
		level = logrus.ErrorLevel
	}

	Logger.WithFields(logrus.Fields{
		"type":      "config",
		"action":    cfg.Action,
		"config_id": cfg.ConfigID,
		"user_id":   cfg.UserID,
		"username":  cfg.Username,
		"client_ip": cfg.ClientIP,
		"details":   cfg.Details,
		"success":   cfg.Success,
		"error":     cfg.Error,
	}).Log(level, fmt.Sprintf("Config %s", cfg.Action))
}

// LogFile logs file operation events
func LogFile(file FileLog) {
	level := logrus.InfoLevel
	if !file.Success {
		level = logrus.ErrorLevel
	}

	Logger.WithFields(logrus.Fields{
		"type":      "file",
		"action":    file.Action,
		"file_name": file.FileName,
		"file_size": file.FileSize,
		"config_id": file.ConfigID,
		"user_id":   file.UserID,
		"username":  file.Username,
		"client_ip": file.ClientIP,
		"success":   file.Success,
		"error":     file.Error,
		"duration":  file.Duration,
	}).Log(level, fmt.Sprintf("File %s", file.Action))
}

// Debug logs debug messages (only if debug level is enabled)
func Debug(msg string, fields ...logrus.Fields) {
	entry := Logger.WithField("type", "debug")
	if len(fields) > 0 {
		entry = entry.WithFields(fields[0])
	}
	entry.Debug(msg)
}

// Info logs info messages
func Info(msg string, fields ...logrus.Fields) {
	entry := Logger.WithField("type", "info")
	if len(fields) > 0 {
		entry = entry.WithFields(fields[0])
	}
	entry.Info(msg)
}

// Warn logs warning messages
func Warn(msg string, fields ...logrus.Fields) {
	entry := Logger.WithField("type", "warning")
	if len(fields) > 0 {
		entry = entry.WithFields(fields[0])
	}
	entry.Warn(msg)
}

// Error logs error messages
func Error(msg string, err error, fields ...logrus.Fields) {
	entry := Logger.WithField("type", "error")
	if len(fields) > 0 {
		entry = entry.WithFields(fields[0])
	}
	if err != nil {
		entry = entry.WithField("error", err.Error())
	}
	entry.Error(msg)
}

// GetGinLogger returns a Gin middleware logger
func GetGinLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Log to our custom logger instead of Gin's default
		LogRequest(RequestLog{
			Timestamp:    param.TimeStamp,
			Method:       param.Method,
			Path:         param.Path,
			StatusCode:   param.StatusCode,
			Duration:     param.Latency.String(),
			ClientIP:     param.ClientIP,
			UserAgent:    param.Request.UserAgent(),
			RequestSize:  param.Request.ContentLength,
			ResponseSize: param.BodySize,
			Error:        param.ErrorMessage,
		})
		return ""
	})
}

// SetLogLevel dynamically changes the log level
func SetLogLevel(level string) error {
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}
	Logger.SetLevel(logLevel)
	Info(fmt.Sprintf("Log level changed to: %s", level))
	return nil
}

// GetLogLevel returns the current log level
func GetLogLevel() string {
	return Logger.GetLevel().String()
}
