package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/dgraph-io/badger/v4"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
)

type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"password,omitempty"`
	Email     string    `json:"email,omitempty"`
	IsAdmin   bool      `json:"is_admin"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	LastLogin time.Time `json:"last_login,omitempty"`
}

func main() {
	var (
		interactive = flag.Bool("interactive", false, "Interactive mode")
		username    = flag.String("username", "", "Admin username")
		email       = flag.String("email", "", "Admin email")
		password    = flag.String("password", "", "Admin password (for non-interactive mode)")
		dbPath      = flag.String("db", "s3mgr.db", "Path to the database file")
	)
	flag.Parse()

	// Open database
	opts := badger.DefaultOptions(*dbPath)
	opts.Logger = nil // Disable badger logging
	db, err := badger.Open(opts)
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	var adminUsername, adminEmail, adminPassword string

	if *interactive {
		// Interactive mode
		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Enter admin username: ")
		adminUsername, _ = reader.ReadString('\n')
		adminUsername = strings.TrimSpace(adminUsername)

		fmt.Print("Enter admin email (optional): ")
		adminEmail, _ = reader.ReadString('\n')
		adminEmail = strings.TrimSpace(adminEmail)

		fmt.Print("Enter admin password: ")
		passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			log.Fatal("Failed to read password:", err)
		}
		adminPassword = string(passwordBytes)
		fmt.Println() // New line after password input

		fmt.Print("Confirm admin password: ")
		confirmPasswordBytes, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			log.Fatal("Failed to read password confirmation:", err)
		}
		confirmPassword := string(confirmPasswordBytes)
		fmt.Println() // New line after password input

		if adminPassword != confirmPassword {
			log.Fatal("Passwords do not match")
		}
	} else {
		// Non-interactive mode
		if *username == "" {
			log.Fatal("Username is required. Use -username flag or -interactive mode")
		}

		if *password == "" {
			log.Fatal("Password is required for non-interactive mode. Use -password flag")
		}

		adminUsername = *username
		adminEmail = *email
		adminPassword = *password
	}

	if adminUsername == "" {
		log.Fatal("Username cannot be empty")
	}

	if len(adminPassword) < 8 {
		log.Fatal("Password must be at least 8 characters long")
	}

	// Check if user already exists
	err = db.View(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte("user:" + adminUsername))
		return err
	})

	if err == nil {
		log.Fatal("User already exists:", adminUsername)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Failed to hash password:", err)
	}

	// Create admin user
	adminUser := User{
		ID:        fmt.Sprintf("user_%d", time.Now().UnixNano()),
		Username:  adminUsername,
		Password:  string(hashedPassword),
		Email:     adminEmail,
		IsAdmin:   true,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	userData, err := json.Marshal(adminUser)
	if err != nil {
		log.Fatal("Failed to marshal user data:", err)
	}

	// Save to database
	err = db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte("user:"+adminUser.Username), userData)
	})

	if err != nil {
		log.Fatal("Failed to save admin user:", err)
	}

	fmt.Printf("✅ Admin user '%s' created successfully!\n", adminUsername)
	if adminEmail != "" {
		fmt.Printf("📧 Email: %s\n", adminEmail)
	}
	fmt.Printf("🔑 User ID: %s\n", adminUser.ID)
	fmt.Printf("⏰ Created at: %s\n", adminUser.CreatedAt.Format(time.RFC3339))
	fmt.Println("\nYou can now use this admin account to log in and manage users.")
}
