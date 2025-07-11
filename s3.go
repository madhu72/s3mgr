package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/dgraph-io/badger/v4"
	"github.com/gin-gonic/gin"

	"s3mgr/audit"
)

type S3Config struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	Name        string `json:"name"`
	AccessKey   string `json:"access_key"`
	SecretKey   string `json:"secret_key"`
	Region      string `json:"region"`
	BucketName  string `json:"bucket_name"`
	EndpointURL string `json:"endpoint_url,omitempty"`
	UseSSL      bool   `json:"use_ssl"`
	StorageType string `json:"storage_type"`
	IsDefault   bool   `json:"is_default"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type S3Service struct {
	db           *badger.DB
	auditService *audit.AuditService
}

func NewS3Service(db *badger.DB, auditService *audit.AuditService) *S3Service {
	return &S3Service{db: db, auditService: auditService}
}

func (s *S3Service) generateConfigID() string {
	return fmt.Sprintf("config_%d", time.Now().UnixNano())
}

func (s *S3Service) createS3Client(config S3Config) *s3.S3 {
	if config.StorageType == "minio" {
		sess, err := session.NewSession(&aws.Config{
			Region:           aws.String(config.Region),
			Endpoint:         aws.String(config.EndpointURL),
			S3ForcePathStyle: aws.Bool(true),
			Credentials:      credentials.NewStaticCredentials(config.AccessKey, config.SecretKey, ""),
			DisableSSL:       aws.Bool(!config.UseSSL),
		})
		if err != nil {
			return nil
		}
		return s3.New(sess)
	} else {
		sess := session.Must(session.NewSession(&aws.Config{
			Region: aws.String(config.Region),
			Credentials: credentials.NewStaticCredentials(
				config.AccessKey,
				config.SecretKey,
				"",
			),
		}))
		return s3.New(sess)
	}
}

func (s *S3Service) getUserConfigs(userID string) ([]S3Config, error) {
	var configs []S3Config

	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte(fmt.Sprintf("user_config_%s_", userID))
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var config S3Config
				if err := json.Unmarshal(val, &config); err != nil {
					return err
				}
				configs = append(configs, config)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	return configs, err
}

func (s *S3Service) getConfigByID(userID, configID string) (*S3Config, error) {
	var config S3Config

	err := s.db.View(func(txn *badger.Txn) error {
		key := fmt.Sprintf("user_config_%s_%s", userID, configID)
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &config)
		})
	})

	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (s *S3Service) saveConfig(config S3Config) error {
	config.UpdatedAt = time.Now().Format(time.RFC3339)
	if config.CreatedAt == "" {
		config.CreatedAt = config.UpdatedAt
	}

	data, err := json.Marshal(config)
	if err != nil {
		return err
	}

	return s.db.Update(func(txn *badger.Txn) error {
		key := fmt.Sprintf("user_config_%s_%s", config.UserID, config.ID)
		return txn.Set([]byte(key), data)
	})
}

// DeleteConfig is a Gin handler for deleting a user config
func (s *S3Service) DeleteConfig(c *gin.Context) {
	userID := c.GetString("user_id")
	configID := c.Param("id")

	// Check if there are other configs
	configs, err := s.getUserConfigs(userID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to check configurations"})
		return
	}
	if len(configs) <= 1 {
		c.JSON(400, gin.H{"error": "Cannot delete the last configuration"})
		return
	}

	if err := s.deleteConfig(userID, configID); err != nil {
		c.JSON(500, gin.H{"error": "Failed to delete configuration"})
		return
	}

	// If this was the default, set another as default
	var deletedWasDefault bool
	for _, cfg := range configs {
		if cfg.ID == configID && cfg.IsDefault {
			deletedWasDefault = true
			break
		}
	}
	if deletedWasDefault && len(configs) > 1 {
		for _, cfg := range configs {
			if cfg.ID != configID {
				s.setDefaultConfig(userID, cfg.ID)
				break
			}
		}
	}

	c.JSON(200, gin.H{"message": "Configuration deleted successfully"})
}

// SetDefaultConfig is a Gin handler for setting a config as default
func (s *S3Service) SetDefaultConfig(c *gin.Context) {
	userID := c.GetString("user_id")
	configID := c.Param("id")

	if err := s.setDefaultConfig(userID, configID); err != nil {
		c.JSON(500, gin.H{"error": "Failed to set default configuration"})
		return
	}
	c.JSON(200, gin.H{"message": "Default configuration set"})
}

// Internal utility for deleting a config
func (s *S3Service) deleteConfig(userID, configID string) error {
	return s.db.Update(func(txn *badger.Txn) error {
		key := fmt.Sprintf("user_config_%s_%s", userID, configID)
		return txn.Delete([]byte(key))
	})
}

// Internal utility for setting a config as default
func (s *S3Service) setDefaultConfig(userID, configID string) error {
	configs, err := s.getUserConfigs(userID)
	if err != nil {
		return err
	}

	for _, config := range configs {
		if config.IsDefault {
			config.IsDefault = false
			if err := s.saveConfig(config); err != nil {
				return err
			}
		}
	}
	for _, config := range configs {
		if config.ID == configID {
			config.IsDefault = true
			if err := s.saveConfig(config); err != nil {
				return err
			}
			break
		}
	}
	return nil
}

func (s *S3Service) getDefaultConfig(userID string) (*S3Config, error) {
	configs, err := s.getUserConfigs(userID)
	if err != nil {
		return nil, err
	}

	for _, config := range configs {
		if config.IsDefault {
			return &config, nil
		}
	}

	// If no default, return the first config
	if len(configs) > 0 {
		return &configs[0], nil
	}

	return nil, fmt.Errorf("no configurations found")
}

// API Handlers

// UploadFile handles file upload to S3
func (s *S3Service) UploadFile(c *gin.Context) {
	// Audit logging helper
	logAudit := func(success bool, err error, details map[string]interface{}) {
		if s.auditService != nil {
			s.auditService.LogEvent(c, "upload_file", "file", "", success, err, details)
		}
	}

	userID := c.GetString("user_id")
	configID := c.Query("config_id")

	var config *S3Config
	var err error
	if configID != "" {
		config, err = s.getConfigByID(userID, configID)
	} else {
		config, err = s.getDefaultConfig(userID)
	}
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Configuration not found"})
		return
	}
	client := s.createS3Client(*config)
	if client == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create storage client"})
		return
	}
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File required"})
		return
	}
	defer file.Close()
	userPrefix := fmt.Sprintf("users/%s/", userID)
	key := userPrefix + header.Filename

	// Detect file size
	fileSize := header.Size
	const multipartThreshold = 5 * 1024 * 1024 // 5MB

	if fileSize > multipartThreshold {
		// --- Multipart upload for large files ---
		createResp, err := client.CreateMultipartUpload(&s3.CreateMultipartUploadInput{
			Bucket: aws.String(config.BucketName),
			Key:    aws.String(key),
		})
		if err != nil {
			logAudit(false, err, map[string]interface{}{
				"stage": "initiate_multipart",
				"filename": header.Filename,
				"size": fileSize,
			})
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initiate multipart upload: " + err.Error()})
			return
		}

		var completedParts []*s3.CompletedPart
		const partSize = 5 * 1024 * 1024 // 5MB
		buffer := make([]byte, partSize)
		partNumber := int64(1)
		for {
			n, readErr := file.Read(buffer)
			if n == 0 && readErr == io.EOF {
				break
			}
			if n == 0 && readErr != nil {
				logAudit(false, readErr, map[string]interface{}{
					"stage": "read_part",
					"filename": header.Filename,
					"size": fileSize,
					"part_number": partNumber,
				})
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file part: " + readErr.Error()})
				return
			}
			partInput := &s3.UploadPartInput{
				Bucket:     aws.String(config.BucketName),
				Key:        aws.String(key),
				PartNumber: aws.Int64(partNumber),
				UploadId:   createResp.UploadId,
				Body:       strings.NewReader(string(buffer[:n])),
			}
			partResp, uploadErr := client.UploadPart(partInput)
			if uploadErr != nil {
				client.AbortMultipartUpload(&s3.AbortMultipartUploadInput{
					Bucket:   aws.String(config.BucketName),
					Key:      aws.String(key),
					UploadId: createResp.UploadId,
				})
				logAudit(false, uploadErr, map[string]interface{}{
					"stage": "upload_part",
					"filename": header.Filename,
					"size": fileSize,
					"part_number": partNumber,
				})
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload part: " + uploadErr.Error()})
				return
			}
			completedParts = append(completedParts, &s3.CompletedPart{
				ETag:       partResp.ETag,
				PartNumber: aws.Int64(partNumber),
			})
			partNumber++
			if readErr == io.EOF {
				break
			}
		}
		// Complete multipart upload
		_, err = client.CompleteMultipartUpload(&s3.CompleteMultipartUploadInput{
			Bucket:   aws.String(config.BucketName),
			Key:      aws.String(key),
			UploadId: createResp.UploadId,
			MultipartUpload: &s3.CompletedMultipartUpload{
				Parts: completedParts,
			},
		})
		if err != nil {
			logAudit(false, err, map[string]interface{}{
				"stage": "complete_multipart",
				"filename": header.Filename,
				"size": fileSize,
			})
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete multipart upload: " + err.Error()})
			return
		}
		logAudit(true, nil, map[string]interface{}{
			"stage": "multipart_upload",
			"filename": header.Filename,
			"size": fileSize,
			"parts": len(completedParts),
		})
		c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully (multipart)", "key": header.Filename})
		return
	}

	// --- Small file: use PutObject ---
	_, err = client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(config.BucketName),
		Key:    aws.String(key),
		Body:   file,
	})
	if err != nil {
		logAudit(false, err, map[string]interface{}{
			"stage": "put_object",
			"filename": header.Filename,
			"size": fileSize,
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file: " + err.Error()})
		return
	}
	logAudit(true, nil, map[string]interface{}{
		"stage": "put_object",
		"filename": header.Filename,
		"size": fileSize,
	})
	c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully", "key": header.Filename})
}


// DownloadFile handles file download from S3
func (s *S3Service) DownloadFile(c *gin.Context) {
	// Audit logging helper
	logAudit := func(success bool, err error, details map[string]interface{}) {
		if s.auditService != nil {
			s.auditService.LogEvent(c, "download_file", "file", "", success, err, details)
		}
	}

	userID := c.GetString("user_id")
	configID := c.Query("config_id")
	key := c.Param("key")

	var config *S3Config
	var err error
	if configID != "" {
		config, err = s.getConfigByID(userID, configID)
	} else {
		config, err = s.getDefaultConfig(userID)
	}
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Configuration not found"})
		return
	}
	client := s.createS3Client(*config)
	if client == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create storage client"})
		return
	}
	userPrefix := fmt.Sprintf("users/%s/", userID)
	fullKey := userPrefix + key
	resp, err := client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(config.BucketName),
		Key:    aws.String(fullKey),
	})
	if err != nil {
		logAudit(false, err, map[string]interface{}{
			"filename": key,
			"full_key": fullKey,
			"stage": "get_object",
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to download file: " + err.Error()})
		return
	}
	defer resp.Body.Close()
	c.Header("Content-Disposition", "attachment; filename="+key)
	c.Header("Content-Type", *resp.ContentType)
	c.Status(http.StatusOK)
	_, _ = io.Copy(c.Writer, resp.Body)
	// Log success (content length may be nil for some S3 backends)
	var size int64 = 0
	if resp.ContentLength != nil {
		size = *resp.ContentLength
	}
	logAudit(true, nil, map[string]interface{}{
		"filename": key,
		"full_key": fullKey,
		"size": size,
	})
}

// ListFiles lists files in S3 with pagination
func (s *S3Service) ListFiles(c *gin.Context) {
	userID := c.GetString("user_id")
	configID := c.Query("config_id")
	page := 1
	pageSize := 10
	if p := c.Query("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	if ps := c.Query("page_size"); ps != "" {
		fmt.Sscanf(ps, "%d", &pageSize)
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	var config *S3Config
	var err error
	if configID != "" {
		config, err = s.getConfigByID(userID, configID)
	} else {
		config, err = s.getDefaultConfig(userID)
	}
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Configuration not found"})
		return
	}
	client := s.createS3Client(*config)
	if client == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create storage client"})
		return
	}
	userPrefix := fmt.Sprintf("users/%s/", userID)
	result, err := client.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(config.BucketName),
		Prefix: aws.String(userPrefix),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list files: " + err.Error()})
		return
	}
	var files []map[string]interface{}
	for _, obj := range result.Contents {
		displayKey := strings.TrimPrefix(*obj.Key, userPrefix)
		if displayKey == "" {
			continue
		}
		files = append(files, map[string]interface{}{
			"key":           displayKey,
			"full_key":      *obj.Key,
			"size":          *obj.Size,
			"last_modified": obj.LastModified.Format(time.RFC3339),
		})
	}
	total := len(files)
	start := (page - 1) * pageSize
	end := start + pageSize
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	paginated := files[start:end]
	c.JSON(http.StatusOK, gin.H{
		"files":       paginated,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"config_id":   config.ID,
		"config_name": config.Name,
	})
}

// DeleteFile deletes a file from S3
func (s *S3Service) DeleteFile(c *gin.Context) {
	// Audit logging helper
	logAudit := func(success bool, err error, details map[string]interface{}) {
		if s.auditService != nil {
			s.auditService.LogEvent(c, "delete_file", "file", "", success, err, details)
		}
	}

	userID := c.GetString("user_id")
	configID := c.Query("config_id")
	key := c.Param("key")

	var config *S3Config
	var err error
	if configID != "" {
		config, err = s.getConfigByID(userID, configID)
	} else {
		config, err = s.getDefaultConfig(userID)
	}
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Configuration not found"})
		return
	}
	client := s.createS3Client(*config)
	if client == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create storage client"})
		return
	}
	userPrefix := fmt.Sprintf("users/%s/", userID)
	fullKey := userPrefix + key
	_, err = client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(config.BucketName),
		Key:    aws.String(fullKey),
	})
	if err != nil {
		logAudit(false, err, map[string]interface{}{
			"filename": key,
			"full_key": fullKey,
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file: " + err.Error()})
		return
	}
	logAudit(true, nil, map[string]interface{}{
		"filename": key,
		"full_key": fullKey,
	})
	c.JSON(http.StatusOK, gin.H{"message": "File deleted successfully"})
}


// ExportConfigsHandler returns all configs as CSV or JSON (admin only)
func (s *S3Service) ExportConfigsHandler(c *gin.Context) {
	// Audit logging helper
	logAudit := func(success bool, err error, details map[string]interface{}) {
		if s.auditService != nil {
			s.auditService.LogEvent(c, "export_configs", "config", "", success, err, details)
		}
	}

	defer func() {
	}()

	format := c.DefaultQuery("format", "csv")
	var configs []S3Config
	// For admin: get all configs for all users
	err := s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte("config:")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var cfg S3Config
				if err := json.Unmarshal(val, &cfg); err != nil {
					return err
				}
				configs = append(configs, cfg)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		logAudit(false, err, map[string]interface{}{"stage": "get_configs"})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get configs"})
		return
	}
	if format == "json" {
		logAudit(true, nil, map[string]interface{}{"format": format, "count": len(configs)})
		c.Header("Content-Disposition", "attachment; filename=configs.json")
		c.JSON(http.StatusOK, configs)
		return
	}
	// Default: CSV
	c.Header("Content-Disposition", "attachment; filename=configs.csv")
	c.Header("Content-Type", "text/csv")
	w := csv.NewWriter(c.Writer)
	defer w.Flush()
	w.Write([]string{"id", "user_id", "name", "access_key", "secret_key", "region", "bucket_name", "endpoint_url", "use_ssl", "storage_type", "is_default", "created_at", "updated_at"})
	for _, cfg := range configs {
		w.Write([]string{
			cfg.ID,
			cfg.UserID,
			cfg.Name,
			cfg.AccessKey,
			cfg.SecretKey,
			cfg.Region,
			cfg.BucketName,
			cfg.EndpointURL,
			fmt.Sprintf("%v", cfg.UseSSL),
			cfg.StorageType,
			fmt.Sprintf("%v", cfg.IsDefault),
			cfg.CreatedAt,
			cfg.UpdatedAt,
		})
	}
	logAudit(true, nil, map[string]interface{}{"format": format, "count": len(configs)})
}

// ImportConfigsHandler accepts CSV or JSON and creates/updates configs (admin only)
func (s *S3Service) ImportConfigsHandler(c *gin.Context) {
	// Audit logging helper
	logAudit := func(success bool, err error, details map[string]interface{}) {
		if s.auditService != nil {
			s.auditService.LogEvent(c, "import_configs", "config", "", success, err, details)
		}
	}

	defer func() {
	}()

	format := c.DefaultQuery("format", "csv")
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		logAudit(false, err, map[string]interface{}{"stage": "parse_form_file"})
		c.JSON(http.StatusBadRequest, gin.H{"error": "File required"})
		return
	}
	defer file.Close()
	var configs []S3Config
	if format == "json" {
		dec := json.NewDecoder(file)
		if err := dec.Decode(&configs); err != nil {
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
			if i == 0 {
				continue
			}
			if len(rec) < 13 {
				continue
			}
			configs = append(configs, S3Config{
				ID: rec[0], UserID: rec[1], Name: rec[2], AccessKey: rec[3], SecretKey: rec[4],
				Region: rec[5], BucketName: rec[6], EndpointURL: rec[7],
				UseSSL: rec[8] == "true", StorageType: rec[9], IsDefault: rec[10] == "true",
				CreatedAt: rec[11], UpdatedAt: rec[12],
			})
		}
	}
	// Save configs (create or update)
	for _, cfg := range configs {
		cfgData, _ := json.Marshal(cfg)
		s.db.Update(func(txn *badger.Txn) error {
			return txn.Set([]byte("config:"+cfg.ID), cfgData)
		})
	}
	logAudit(true, nil, map[string]interface{}{"format": format, "count": len(configs)})
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Imported %d configs", len(configs))})
}

// GetConfigs returns a list of configs with redacted secrets
func (s *S3Service) GetConfigs(c *gin.Context) {
	userID := c.GetString("user_id")
	configs, err := s.getUserConfigs(userID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to get configurations"})
		return
	}
	var safeConfigs []map[string]interface{}
	for _, config := range configs {
		safeConfig := map[string]interface{}{
			"id":           config.ID,
			"name":         config.Name,
			"region":       config.Region,
			"bucket_name":  config.BucketName,
			"access_key":   config.AccessKey[:min(4, len(config.AccessKey))] + "****",
			"endpoint_url": config.EndpointURL,
			"use_ssl":      config.UseSSL,
			"storage_type": config.StorageType,
			"is_default":   config.IsDefault,
			"created_at":   config.CreatedAt,
			"updated_at":   config.UpdatedAt,
		}
		safeConfigs = append(safeConfigs, safeConfig)
	}
	c.JSON(200, gin.H{"configurations": safeConfigs})
}

// GetConfigByID returns the full config including secret_key if the user is owner or admin
func (s *S3Service) GetConfigByID(c *gin.Context) {
	userID := c.GetString("user_id")
	isAdmin := c.GetBool("is_admin")
	configID := c.Param("id")
	config, err := s.getConfigByID(userID, configID)
	if err != nil {
		c.JSON(404, gin.H{"error": "Configuration not found"})
		return
	}
	if config.UserID != userID && !isAdmin {
		c.JSON(403, gin.H{"error": "Forbidden"})
		return
	}
	c.JSON(200, config)
}

func (s *S3Service) CreateConfig(c *gin.Context) {
	userID := c.GetString("user_id")

	var config S3Config
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid configuration data"})
		return
	}

	// Generate ID and set user
	config.ID = s.generateConfigID()
	config.UserID = userID

	// Validate configuration by testing connection
	client := s.createS3Client(config)
	if client == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create storage client"})
		return
	}

	_, err := client.ListObjects(&s3.ListObjectsInput{
		Bucket:  aws.String(config.BucketName),
		MaxKeys: aws.Int64(1),
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to connect to storage: " + err.Error()})
		return
	}

	// If this is the first config, make it default
	existingConfigs, _ := s.getUserConfigs(userID)
	if len(existingConfigs) == 0 {
		config.IsDefault = true
	}

	if err := s.saveConfig(config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save configuration"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Configuration created successfully",
		"id":      config.ID,
	})
}

func (s *S3Service) UpdateConfig(c *gin.Context) {
	userID := c.GetString("user_id")
	configID := c.Param("id")

	existingConfig, err := s.getConfigByID(userID, configID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Configuration not found"})
		return
	}

	var updateData S3Config
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid configuration data"})
		return
	}

	// Preserve ID, UserID, and timestamps
	updateData.ID = existingConfig.ID
	updateData.UserID = existingConfig.UserID
	updateData.CreatedAt = existingConfig.CreatedAt
	updateData.IsDefault = existingConfig.IsDefault

	// Validate configuration
	client := s.createS3Client(updateData)
	if client == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create storage client"})
		return
	}

	_, err = client.ListObjects(&s3.ListObjectsInput{
		Bucket:  aws.String(updateData.BucketName),
		MaxKeys: aws.Int64(1),
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to connect to storage: " + err.Error()})
		return
	}

	if err := s.saveConfig(updateData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update configuration"})
		return
	}
	userID = c.GetString("user_id")
	configID = c.Param("id")

	config, err := s.getConfigByID(userID, configID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Configuration not found"})
		return
	}

	// Check if there are other configs
	configs, err := s.getUserConfigs(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check configurations"})
		return
	}

	if len(configs) <= 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete the last configuration"})
		return
	}

	if err := s.deleteConfig(userID, configID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete configuration"})
		return
	}

	// If this was the default, set another as default
	if config.IsDefault && len(configs) > 1 {
		for _, cfg := range configs {
			if cfg.ID != configID {
				s.setDefaultConfig(userID, cfg.ID)
				break
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Configuration deleted successfully"})
}



func (s *S3Service) AutoConfigureMinIO(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req struct {
		Username string `json:"username" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
		return
	}

	// Create MinIO user and bucket using admin credentials
	config, err := CreateMinIOUserAndBucket(req.Username, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create MinIO configuration: " + err.Error()})
		return
	}

	// Save configuration to database
	err = s.saveConfig(*config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save configuration: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "MinIO configuration created successfully",
		"config":  config,
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
