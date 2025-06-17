package config

import (
	"flag"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
	"s3mgr/logger"
)

type Config struct {
	Logging     logger.LogConfig `yaml:"logging"`
	Server      ServerConfig     `yaml:"server"`
	Database    DatabaseConfig   `yaml:"database"`
	JWT         JWTConfig        `yaml:"jwt"`
	MinIOAdmin  MinIOAdminConfig `yaml:"minio_admin"`
	MinIODefault MinIODefaultConfig `yaml:"minio_default"`
}

type ServerConfig struct {
	Port         int    `yaml:"port"`
	Host         string `yaml:"host"`
	ReadTimeout  int    `yaml:"read_timeout"`
	WriteTimeout int    `yaml:"write_timeout"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

type JWTConfig struct {
	Secret      string `yaml:"secret"`
	ExpiryHours int    `yaml:"expiry_hours"`
}

type MinIOAdminConfig struct {
	URL       string `yaml:"url"`
	AccessKey string `yaml:"access_key"`
	SecretKey string `yaml:"secret_key"`
}

type MinIODefaultConfig struct {
	Endpoint string `yaml:"endpoint"`
	Bucket   string `yaml:"bucket"`
	Region   string `yaml:"region"`
	SSL      bool   `yaml:"ssl"`
}

var (
	AppConfig *Config
	configFile string
)

// LoadConfig loads configuration from file and environment variables
func LoadConfig() (*Config, error) {
	// Parse command line flags
	flag.StringVar(&configFile, "config", "config.yaml", "Path to configuration file")
	flag.Parse()

	// Load configuration from file
	config, err := loadFromFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load config from file: %v", err)
	}

	// Override with environment variables if present
	overrideWithEnv(config)

	AppConfig = config
	return config, nil
}

func loadFromFile(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %v", filename, err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	// Set defaults
	setDefaults(&config)

	return &config, nil
}

func setDefaults(config *Config) {
	// Logging defaults
	if config.Logging.Level == "" {
		config.Logging.Level = "info"
	}
	if config.Logging.File == "" {
		config.Logging.File = "logs/s3mgr.log"
	}
	if config.Logging.MaxSize == 0 {
		config.Logging.MaxSize = 100
	}
	if config.Logging.MaxBackups == 0 {
		config.Logging.MaxBackups = 30
	}
	if config.Logging.MaxAge == 0 {
		config.Logging.MaxAge = 30
	}
	if config.Logging.Format == "" {
		config.Logging.Format = "json"
	}

	// Server defaults
	if config.Server.Port == 0 {
		config.Server.Port = 8081
	}
	if config.Server.Host == "" {
		config.Server.Host = "0.0.0.0"
	}
	if config.Server.ReadTimeout == 0 {
		config.Server.ReadTimeout = 30
	}
	if config.Server.WriteTimeout == 0 {
		config.Server.WriteTimeout = 30
	}

	// Database defaults
	if config.Database.Path == "" {
		config.Database.Path = "s3mgr.db"
	}

	// JWT defaults
	if config.JWT.ExpiryHours == 0 {
		config.JWT.ExpiryHours = 24
	}
}

func overrideWithEnv(config *Config) {
	// Override with environment variables
	if val := os.Getenv("LOG_LEVEL"); val != "" {
		config.Logging.Level = val
	}
	if val := os.Getenv("LOG_FILE"); val != "" {
		config.Logging.File = val
	}
	if val := os.Getenv("SERVER_PORT"); val != "" {
		fmt.Sscanf(val, "%d", &config.Server.Port)
	}
	if val := os.Getenv("JWT_SECRET"); val != "" {
		config.JWT.Secret = val
	}
	if val := os.Getenv("MINIO_ADMIN_URL"); val != "" {
		config.MinIOAdmin.URL = val
	}
	if val := os.Getenv("MINIO_ADMIN_ACCESS_KEY"); val != "" {
		config.MinIOAdmin.AccessKey = val
	}
	if val := os.Getenv("MINIO_ADMIN_SECRET_KEY"); val != "" {
		config.MinIOAdmin.SecretKey = val
	}
	if val := os.Getenv("MINIO_DEFAULT_ENDPOINT"); val != "" {
		config.MinIODefault.Endpoint = val
	}
	if val := os.Getenv("MINIO_DEFAULT_BUCKET"); val != "" {
		config.MinIODefault.Bucket = val
	}
	if val := os.Getenv("MINIO_DEFAULT_REGION"); val != "" {
		config.MinIODefault.Region = val
	}
}

// GetConfigFile returns the path to the configuration file
func GetConfigFile() string {
	return configFile
}

// ReloadConfig reloads the configuration from file
func ReloadConfig() error {
	config, err := loadFromFile(configFile)
	if err != nil {
		return err
	}
	overrideWithEnv(config)
	AppConfig = config
	
	// Reinitialize logger with new config
	return logger.Initialize(config.Logging)
}
