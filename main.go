package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"s3mgr/audit"
	"s3mgr/config"
	"s3mgr/logger"
	"s3mgr/middleware"
)

func main() {
	// Command line flags
	createAdmin := flag.Bool("create-admin", false, "Create admin user interactively")
	flag.Parse()

	// Handle admin creation
	if *createAdmin {
		fmt.Println("Please use the separate create-admin tool:")
		fmt.Println("go run cmd/create-admin.go -interactive")
		return
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Initialize logger
	err = logger.Initialize(cfg.Logging)
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}

	logger.Info("Starting S3 Manager server...")
	logger.Info("Configuration loaded")

	// Initialize database
	db, err := InitDB(cfg)
	if err != nil {
		logger.Error("Failed to initialize database", err)
		log.Fatal(err)
	}
	defer db.Close()

	// Initialize services
	authService := NewAuthService(db)
	s3Service := NewS3Service(db)
	auditService := audit.NewAuditService(db)

	// Set Gin mode based on log level
	if cfg.Logging.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	r := gin.New()

	// Add middleware
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger()) // Custom request logger
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"timestamp": time.Now().UTC(),
			"version": "1.0.0",
		})
	})

	// Debug endpoint to change log level (only in debug mode)
	if cfg.Logging.Level == "debug" {
		r.POST("/debug/log-level", func(c *gin.Context) {
			var req struct {
				Level string `json:"level"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			
			if err := logger.SetLogLevel(req.Level); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			
			c.JSON(http.StatusOK, gin.H{"message": "Log level updated", "level": req.Level})
		})
	}

	// API routes
	api := r.Group("/api")

	// Authentication routes
	auth := api.Group("/auth")
	{
		auth.POST("/register", authService.Register)
		auth.POST("/login", authService.Login)
	}

	// Protected routes
	protected := api.Group("")
	protected.Use(AuthMiddleware(authService))
	{
		// User profile routes
		protected.POST("/auth/change-password", authService.ChangePassword)

		// Configuration routes
		protected.GET("/configs", s3Service.GetConfigs)
		protected.GET("/configs/:id", s3Service.GetConfigByID)
		protected.POST("/configs", s3Service.CreateConfig)
		protected.PUT("/configs/:id", s3Service.UpdateConfig)
		protected.DELETE("/configs/:id", s3Service.DeleteConfig)
		protected.POST("/configs/:id/set-default", s3Service.SetDefaultConfig)
		protected.POST("/configs/auto-minio", s3Service.AutoConfigureMinIO)

		// File operation routes
		protected.POST("/files/upload", s3Service.UploadFile)
		protected.GET("/files/download/:key", s3Service.DownloadFile)
		protected.DELETE("/files/:key", s3Service.DeleteFile)
		protected.GET("/files", s3Service.ListFiles)
	}

	// Admin-only routes
	admin := api.Group("/admin")
	admin.Use(AuthMiddleware(authService))
	admin.Use(AdminMiddleware(authService)) // Custom middleware to check admin status
	{
		// Bulk user import/export
		admin.GET("/users/export", authService.ExportUsersHandler)
		admin.POST("/users/import", authService.ImportUsersHandler)

		// User management list
		admin.GET("/users", authService.ListUsersHandler)

		// Bulk config import/export
		admin.GET("/configs/export", s3Service.ExportConfigsHandler)
		admin.POST("/configs/import", s3Service.ImportConfigsHandler)

		// User management routes
		admin.PUT("/users/:username", authService.UpdateUser)
		admin.DELETE("/users/:username", authService.DeleteUser)
		admin.GET("/users/:username/config", authService.GetUserConfig)

		// Audit log routes
		admin.GET("/audit-logs", auditService.GetAuditLogsHandler)
		admin.POST("/audit-logs/filter", auditService.PostAuditLogsFilterHandler)
		admin.GET("/audit-logs/incident/:session_id", auditService.GetAuditLogsByIncidentHandler)
	}

	// Start server
	port := fmt.Sprintf("%d", cfg.Server.Port)
	logger.Info("Server starting", map[string]interface{}{
		"port": port,
		"host": cfg.Server.Host,
	})
	
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, port),
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}
	
	log.Fatal(server.ListenAndServe())
}

// AdminMiddleware checks if the user is an admin
func AdminMiddleware(authService *AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		username, exists := c.Get("username")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		user, err := authService.GetUserByUsername(username.(string))
		if err != nil || !user.IsAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin privileges required"})
			c.Abort()
			return
		}

		c.Next()
	}
}
