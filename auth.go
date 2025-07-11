package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"s3mgr/audit"
	"s3mgr/middleware"
)

type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"password,omitempty"` // Omit from JSON responses
	Email     string    `json:"email,omitempty"`
	IsAdmin   bool      `json:"is_admin"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	LastLogin time.Time `json:"last_login,omitempty"`
}

type UserResponse struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email,omitempty"`
	IsAdmin   bool      `json:"is_admin"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	LastLogin time.Time `json:"last_login,omitempty"`
}

type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
	Email    string `json:"email"`
	IsAdmin  bool   `json:"is_admin"`
}

type UpdateUserRequest struct {
	Email    string `json:"email"`
	IsAdmin  bool   `json:"is_admin"`
	IsActive bool   `json:"is_active"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

type Claims struct {
	Username string `json:"username"`
	IsAdmin  bool   `json:"is_admin"`
	jwt.RegisteredClaims
}

type AuthService struct {
	db           *badger.DB
	jwtSecret    []byte
	auditService *audit.AuditService
}

// Logout handler
func (a *AuthService) Logout(c *gin.Context) {
	username := c.GetString("username")
	if username == "" {
		// Try to extract from JWT or fallback to user_id
		username = c.GetString("user_id")
	}
	// audit log removed(c, "logout", "user", username, true, nil, nil)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func NewAuthService(db *badger.DB, auditService *audit.AuditService) *AuthService {
	return &AuthService{
		db:           db,
		jwtSecret:    []byte("your-secret-key"), // In production, use environment variable
		auditService: auditService,
	}
}

func (a *AuthService) hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func (a *AuthService) checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (a *AuthService) generateToken(username string, isAdmin bool) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: username,
		IsAdmin:  isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(a.jwtSecret)
}

func (a *AuthService) validateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return a.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return claims, nil
}

func (a *AuthService) Login(c *gin.Context) {
	// For audit logging

	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		// audit log removed(c, "login", "user", user.Username, false, err, map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var storedUser User
	err := a.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("user:" + user.Username))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &storedUser)
		})
	})

	if err != nil {
		// audit log removed(c, "login", "user", user.Username, false, err, map[string]interface{}{"error": "Invalid credentials"})
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if !storedUser.IsActive {
		// audit log removed(c, "login", "user", storedUser.Username, false, fmt.Errorf("user account is inactive"), map[string]interface{}{"error": "Account is inactive"})
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account is inactive"})
		return
	}

	if !a.checkPasswordHash(user.Password, storedUser.Password) {
		// audit log removed(c, "login", "user", storedUser.Username, false, fmt.Errorf("invalid password"), map[string]interface{}{"error": "Invalid credentials"})
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Update last login time
	storedUser.LastLogin = time.Now()
	userData, _ := json.Marshal(storedUser)
	a.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte("user:"+storedUser.Username), userData)
	})

	token, err := a.generateToken(storedUser.Username, storedUser.IsAdmin)
	if err != nil {
		// audit log removed(c, "login", "user", storedUser.Username, false, err, map[string]interface{}{"error": "Failed to generate token"})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Set username, user_id, session_id in context for audit logging
	c.Set("username", storedUser.Username)
	c.Set("user_id", storedUser.Username)
	// session_id can be set here if available (e.g., from JWT or generated)

	// audit log removed(c, "login", "user", storedUser.Username, true, nil, map[string]interface{}{"status": c.Writer.Status()})
	c.JSON(http.StatusOK, gin.H{
		"token":    token,
		"username": storedUser.Username,
		"is_admin": storedUser.IsAdmin,
	})
}

func (a *AuthService) Register(c *gin.Context) {
	var createUserRequest CreateUserRequest
	if err := c.ShouldBindJSON(&createUserRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if createUserRequest.Username == "" || createUserRequest.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username and password are required"})
		return
	}

	// Check if user already exists
	err := a.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte("user:" + createUserRequest.Username))
		return err
	})

	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}

	// Hash password
	hashedPassword, err := a.hashPassword(createUserRequest.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Save user
	userData, _ := json.Marshal(User{
		ID:       "",
		Username: createUserRequest.Username,
		Password: hashedPassword,
		Email:    createUserRequest.Email,
		IsAdmin:  createUserRequest.IsAdmin,
		IsActive: true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	err = a.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte("user:"+createUserRequest.Username), userData)
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
}

func (a *AuthService) GetUserByUsername(username string) (*User, error) {
	var user User
	err := a.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("user:" + username))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &user)
		})
	})

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (a *AuthService) GetAllUsers() ([]UserResponse, error) {
	var users []UserResponse

	err := a.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte("user:")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var user User
				if err := json.Unmarshal(val, &user); err != nil {
					return err
				}

				users = append(users, UserResponse{
					ID:        user.ID,
					Username:  user.Username,
					Email:     user.Email,
					IsAdmin:   user.IsAdmin,
					IsActive:  user.IsActive,
					CreatedAt: user.CreatedAt,
					UpdatedAt: user.UpdatedAt,
					LastLogin: user.LastLogin,
				})
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	return users, err
}

// ListUsersHandler returns all users as JSON (admin only)
func (a *AuthService) ListUsersHandler(c *gin.Context) {
	users, err := a.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get users"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": users})
}

// ExportUsersHandler returns all users as CSV or JSON (admin only)
func (a *AuthService) ExportUsersHandler(c *gin.Context) {
	// Audit logging helper
	logAudit := func(success bool, err error, details map[string]interface{}) {
		if a.auditService != nil {
			a.auditService.LogEvent(c, "export_users", "user", "", success, err, details)
		}
	}

	defer func() {
		// audit log removed(c, "export_users", "user", "", c.Writer.Status() == http.StatusOK, nil, map[string]interface{}{"format": c.DefaultQuery("format", "csv")})
	}()

	format := c.DefaultQuery("format", "csv")
	users, err := a.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get users"})
		return
	}
	if format == "json" {
		logAudit(true, nil, map[string]interface{}{"format": format, "count": len(users)})
		c.Header("Content-Disposition", "attachment; filename=users.json")
		c.JSON(http.StatusOK, users)
		return
	}
	// Default: CSV
	c.Header("Content-Disposition", "attachment; filename=users.csv")
	c.Header("Content-Type", "text/csv")
	w := csv.NewWriter(c.Writer)
	defer w.Flush()
	w.Write([]string{"id","username","email","is_admin","is_active","created_at","updated_at","last_login"})
	for _, u := range users {
		w.Write([]string{
			u.ID,
			u.Username,
			u.Email,
			fmt.Sprintf("%v", u.IsAdmin),
			fmt.Sprintf("%v", u.IsActive),
			u.CreatedAt.Format(time.RFC3339),
			u.UpdatedAt.Format(time.RFC3339),
			u.LastLogin.Format(time.RFC3339),
		})
	}
	logAudit(true, nil, map[string]interface{}{"format": format, "count": len(users)})
}

// ImportUsersHandler accepts CSV or JSON and creates/updates users (admin only)
func (a *AuthService) ImportUsersHandler(c *gin.Context) {
	// Audit logging helper
	logAudit := func(success bool, err error, details map[string]interface{}) {
		if a.auditService != nil {
			a.auditService.LogEvent(c, "import_users", "user", "", success, err, details)
		}
	}

	defer func() {
		// audit log removed(c, "import_users", "user", "", c.Writer.Status() == http.StatusOK, nil, map[string]interface{}{"format": c.DefaultQuery("format", "csv")})
	}()

	format := c.DefaultQuery("format", "csv")
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		logAudit(false, err, map[string]interface{}{"stage": "parse_form_file"})
		c.JSON(http.StatusBadRequest, gin.H{"error": "File required"})
		return
	}
	defer file.Close()
	var users []User
	if format == "json" {
		dec := json.NewDecoder(file)
		if err := dec.Decode(&users); err != nil {
			logAudit(false, err, map[string]interface{}{"stage": "decode_json"})
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			return
		}
	} else {
		r := csv.NewReader(file)
		records, err := r.ReadAll()
		if err != nil || len(records) < 2 {
			logAudit(false, err, map[string]interface{}{"stage": "decode_csv"})
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid CSV"})
			return
		}
		for i, rec := range records {
			if i == 0 { continue } // skip header
			if len(rec) < 8 { continue }
			createdAt, _ := time.Parse(time.RFC3339, rec[5])
			updatedAt, _ := time.Parse(time.RFC3339, rec[6])
			lastLogin, _ := time.Parse(time.RFC3339, rec[7])
			users = append(users, User{
				ID: rec[0], Username: rec[1], Email: rec[2],
				IsAdmin: rec[3] == "true", IsActive: rec[4] == "true",
				CreatedAt: createdAt, UpdatedAt: updatedAt, LastLogin: lastLogin,
			})
		}
	}
	// Save users (create or update)
	for _, u := range users {
		userData, _ := json.Marshal(u)
		a.db.Update(func(txn *badger.Txn) error {
			return txn.Set([]byte("user:"+u.Username), userData)
		})
	}
	logAudit(true, nil, map[string]interface{}{"format": format, "count": len(users)})
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Imported %d users", len(users))})
}

func (a *AuthService) CreateUser(c *gin.Context) {
	// Check if current user is admin
	currentUser, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	user, err := a.GetUserByUsername(currentUser.(string))
	if err != nil || !user.IsAdmin {
		middleware.LogAuthEvent(c, "create_user", currentUser.(string), false, fmt.Errorf("insufficient privileges"))
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin privileges required"})
		return
	}

	var createUserRequest CreateUserRequest
	if err := c.ShouldBindJSON(&createUserRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user already exists
	_, err = a.GetUserByUsername(createUserRequest.Username)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}

	// Hash password
	hashedPassword, err := a.hashPassword(createUserRequest.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create new user
	newUser := User{
		ID:        fmt.Sprintf("user_%d", time.Now().UnixNano()),
		Username:  createUserRequest.Username,
		Password:  hashedPassword,
		Email:     createUserRequest.Email,
		IsAdmin:   createUserRequest.IsAdmin,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	userData, _ := json.Marshal(newUser)
	err = a.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte("user:"+newUser.Username), userData)
	})

	if err != nil {
		middleware.LogAuthEvent(c, "create_user", currentUser.(string), false, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	middleware.LogAuthEvent(c, "create_user", currentUser.(string), true, nil)
	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"user": UserResponse{
			ID:        newUser.ID,
			Username:  newUser.Username,
			Email:     newUser.Email,
			IsAdmin:   newUser.IsAdmin,
			IsActive:  newUser.IsActive,
			CreatedAt: newUser.CreatedAt,
			UpdatedAt: newUser.UpdatedAt,
		},
	})
}

func (a *AuthService) GetUsers(c *gin.Context) {
	// Check if current user is admin
	currentUser, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	user, err := a.GetUserByUsername(currentUser.(string))
	if err != nil || !user.IsAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin privileges required"})
		return
	}

	users, err := a.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

func (a *AuthService) UpdateUser(c *gin.Context) {
	// Check if current user is admin
	currentUser, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	adminUser, err := a.GetUserByUsername(currentUser.(string))
	if err != nil || !adminUser.IsAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin privileges required"})
		return
	}

	username := c.Param("username")
	
	// Get target user
	targetUser, err := a.GetUserByUsername(username)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Get update request
	var updateRequest UpdateUserRequest
	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update user fields
	targetUser.Email = updateRequest.Email
	targetUser.IsAdmin = updateRequest.IsAdmin
	targetUser.IsActive = updateRequest.IsActive
	targetUser.UpdatedAt = time.Now()

	userData, _ := json.Marshal(targetUser)
	err = a.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte("user:"+targetUser.Username), userData)
	})

	if err != nil {
		middleware.LogAuthEvent(c, "update_user", currentUser.(string), false, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	middleware.LogAuthEvent(c, "update_user", currentUser.(string), true, nil)
	c.JSON(http.StatusOK, gin.H{
		"message": "User updated successfully",
		"user": UserResponse{
			ID:        targetUser.ID,
			Username:  targetUser.Username,
			Email:     targetUser.Email,
			IsAdmin:   targetUser.IsAdmin,
			IsActive:  targetUser.IsActive,
			CreatedAt: targetUser.CreatedAt,
			UpdatedAt: targetUser.UpdatedAt,
			LastLogin: targetUser.LastLogin,
		},
	})
}

func (a *AuthService) DeleteUser(c *gin.Context) {
	// Check if current user is admin
	currentUser, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	adminUser, err := a.GetUserByUsername(currentUser.(string))
	if err != nil || !adminUser.IsAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin privileges required"})
		return
	}

	username := c.Param("username")
	
	// Prevent admin from deleting themselves
	if username == currentUser.(string) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete your own account"})
		return
	}

	// Check if user exists
	_, err = a.GetUserByUsername(username)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Delete user
	err = a.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte("user:" + username))
	})

	if err != nil {
		middleware.LogAuthEvent(c, "delete_user", currentUser.(string), false, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	middleware.LogAuthEvent(c, "delete_user", currentUser.(string), true, nil)
	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

func (a *AuthService) ChangePassword(c *gin.Context) {
	currentUser, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var changePasswordRequest ChangePasswordRequest
	if err := c.ShouldBindJSON(&changePasswordRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current user
	user, err := a.GetUserByUsername(currentUser.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Verify current password
	if !a.checkPasswordHash(changePasswordRequest.CurrentPassword, user.Password) {
		middleware.LogAuthEvent(c, "change_password", currentUser.(string), false, fmt.Errorf("invalid current password"))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Current password is incorrect"})
		return
	}

	// Hash new password
	hashedPassword, err := a.hashPassword(changePasswordRequest.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Update password
	user.Password = hashedPassword
	user.UpdatedAt = time.Now()

	userData, _ := json.Marshal(user)
	err = a.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte("user:"+user.Username), userData)
	})

	if err != nil {
		middleware.LogAuthEvent(c, "change_password", currentUser.(string), false, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	middleware.LogAuthEvent(c, "change_password", currentUser.(string), true, nil)
	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

func (a *AuthService) GetUserConfig(c *gin.Context) {
	// Check if current user is admin
	currentUser, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	adminUser, err := a.GetUserByUsername(currentUser.(string))
	if err != nil || !adminUser.IsAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin privileges required"})
		return
	}

	username := c.Param("username")
	
	// Get target user
	targetUser, err := a.GetUserByUsername(username)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Get user's default configuration from database
	var userConfig map[string]interface{}
	err = a.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("config:default:" + username))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &userConfig)
		})
	})

	if err != nil {
		// If no config found, return empty config
		userConfig = map[string]interface{}{
			"access_key": "",
			"secret_key": "",
			"endpoint":   "",
			"bucket":     "",
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"user": UserResponse{
			ID:        targetUser.ID,
			Username:  targetUser.Username,
			Email:     targetUser.Email,
			IsAdmin:   targetUser.IsAdmin,
			IsActive:  targetUser.IsActive,
			CreatedAt: targetUser.CreatedAt,
			UpdatedAt: targetUser.UpdatedAt,
			LastLogin: targetUser.LastLogin,
		},
		"config": userConfig,
	})
}

func AuthMiddleware(authService *AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		claims, err := authService.validateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Set("username", claims.Username)
		c.Set("is_admin", claims.IsAdmin)
		c.Set("user_id", claims.Username) // Set user_id to username for compatibility
		c.Next()
	}
}
