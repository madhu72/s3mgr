package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/minio/madmin-go/v3"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOAdminConfig struct {
	URL       string
	AccessKey string
	SecretKey string
}

type MinIODefaultConfig struct {
	Endpoint string
	Bucket   string
	Region   string
	SSL      bool
}

func getMinIOAdminConfig() *MinIOAdminConfig {
	return &MinIOAdminConfig{
		URL:       getEnvWithDefault("MINIO_ADMIN_URL", "http://localhost:9000"),
		AccessKey: getEnvWithDefault("MINIO_ADMIN_ACCESS_KEY", "minioadmin"),
		SecretKey: getEnvWithDefault("MINIO_ADMIN_SECRET_KEY", "minioadmin"),
	}
}

func getMinIODefaultConfig() *MinIODefaultConfig {
	return &MinIODefaultConfig{
		Endpoint: getEnvWithDefault("MINIO_DEFAULT_ENDPOINT", "localhost:9000"),
		Bucket:   getEnvWithDefault("MINIO_DEFAULT_BUCKET", "s3manager-default"),
		Region:   getEnvWithDefault("MINIO_DEFAULT_REGION", "us-east-1"),
		SSL:      getEnvWithDefault("MINIO_DEFAULT_SSL", "false") == "true",
	}
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// CreateMinIOUserAndBucket creates a MinIO user and bucket for the S3Manager user
func CreateMinIOUserAndBucket(username, userID string) (*S3Config, error) {
	log.Printf("Starting MinIO auto-configuration for user: %s (ID: %s)", username, userID)
	
	adminConfig := getMinIOAdminConfig()
	defaultConfig := getMinIODefaultConfig()

	log.Printf("MinIO Admin Config - URL: %s, AccessKey: %s", adminConfig.URL, adminConfig.AccessKey)

	// Create MinIO admin client
	adminURL := strings.TrimPrefix(adminConfig.URL, "http://")
	adminURL = strings.TrimPrefix(adminURL, "https://")
	madmClnt, err := madmin.New(adminURL, adminConfig.AccessKey, adminConfig.SecretKey, false)
	if err != nil {
		log.Printf("Failed to create MinIO admin client: %v", err)
		return nil, fmt.Errorf("failed to create MinIO admin client: %v", err)
	}

	log.Printf("MinIO admin client created successfully")

	// Test admin connection
	_, err = madmClnt.ServerInfo(context.Background())
	if err != nil {
		log.Printf("Failed to connect to MinIO admin API: %v", err)
		return nil, fmt.Errorf("failed to connect to MinIO admin API: %v", err)
	}

	log.Printf("MinIO admin connection verified")

	// Generate user credentials
	userIDSuffix := userID
	if len(userID) > 8 {
		userIDSuffix = userID[:8]
	}
	userAccessKey := fmt.Sprintf("s3mgr_%s", userIDSuffix)
	userSecretKey := generateRandomString(32)
	userBucket := fmt.Sprintf("s3mgr-%s", userIDSuffix)

	// Create MinIO user
	err = madmClnt.AddUser(context.Background(), userAccessKey, userSecretKey)
	if err != nil {
		log.Printf("Warning: Failed to create MinIO user (may already exist): %v", err)
	}

	// Create policy for the user
	policyName := fmt.Sprintf("s3mgr-policy-%s", userIDSuffix)
	policy := fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Action": [
					"s3:GetObject",
					"s3:PutObject",
					"s3:DeleteObject",
					"s3:ListBucket"
				],
				"Resource": [
					"arn:aws:s3:::%s",
					"arn:aws:s3:::%s/*"
				]
			}
		]
	}`, userBucket, userBucket)

	err = madmClnt.AddCannedPolicy(context.Background(), policyName, []byte(policy))
	if err != nil {
		log.Printf("Warning: Failed to create policy (may already exist): %v", err)
	}

	// Attach policy to user
	err = madmClnt.SetPolicy(context.Background(), policyName, userAccessKey, false)
	if err != nil {
		log.Printf("Warning: Failed to attach policy to user: %v", err)
	}

	// Create bucket using admin credentials first
	adminS3Client, err := minio.New(defaultConfig.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(adminConfig.AccessKey, adminConfig.SecretKey, ""),
		Secure: defaultConfig.SSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create admin S3 client: %v", err)
	}

	// Create bucket with admin credentials
	err = adminS3Client.MakeBucket(context.Background(), userBucket, minio.MakeBucketOptions{
		Region: defaultConfig.Region,
	})
	if err != nil {
		// Check if bucket already exists
		exists, errBucketExists := adminS3Client.BucketExists(context.Background(), userBucket)
		if errBucketExists == nil && exists {
			log.Printf("Bucket %s already exists", userBucket)
		} else {
			log.Printf("Warning: Failed to create bucket with admin credentials: %v", err)
		}
	} else {
		log.Printf("Successfully created bucket %s", userBucket)
	}

	// Test user access to the bucket
	s3Client, err := minio.New(defaultConfig.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(userAccessKey, userSecretKey, ""),
		Secure: defaultConfig.SSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user S3 client: %v", err)
	}

	// Verify user can access the bucket
	exists, err := s3Client.BucketExists(context.Background(), userBucket)
	if err != nil {
		log.Printf("Warning: User cannot access bucket: %v", err)
	} else if exists {
		log.Printf("User successfully verified access to bucket %s", userBucket)
	}

	// Create S3Config for the user
	config := &S3Config{
		ID:          generateID(),
		UserID:      userID,
		Name:        fmt.Sprintf("MinIO Default (%s)", username),
		StorageType: "minio",
		EndpointURL: defaultConfig.Endpoint,
		Region:      defaultConfig.Region,
		AccessKey:   userAccessKey,
		SecretKey:   userSecretKey,
		BucketName:  userBucket,
		UseSSL:      defaultConfig.SSL,
		IsDefault:   true,
		CreatedAt:   getCurrentTime().Format(time.RFC3339),
		UpdatedAt:   getCurrentTime().Format(time.RFC3339),
	}

	return config, nil
}

// generateRandomString generates a random string of specified length
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func generateID() string {
	return fmt.Sprintf("%x", rand.Int63())
}

func getCurrentTime() time.Time {
	return time.Now().UTC()
}
