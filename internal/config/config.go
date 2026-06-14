package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strconv"
)

type Config struct {
	ServerAddress   string `json:"server_address"`
	BaseURLAddress  string `json:"base_url"`
	LoggerLevel     string `json:"logger_level"`
	FileStoragePath string `json:"file_storage_path"`
	DatabaseDSN     string `json:"database_dsn"`
	AuditFile       string `json:"audit_file"`
	AuditURL        string `json:"audit_url"`
	Secure          bool   `json:"enable_https"`
	ConfigFilePath  string `json:"-"`
}

func GetConfig() *Config {
	var config Config

	flag.StringVar(&config.ServerAddress, "a", "localhost:8080", "server address")
	flag.StringVar(&config.BaseURLAddress, "b", "http://localhost:8080", "base shorted URL")
	flag.StringVar(&config.LoggerLevel, "l", "info", "log level")
	flag.StringVar(&config.FileStoragePath, "f", "storage.json", "path to storage file")
	flag.StringVar(&config.DatabaseDSN, "d", "", "postgress url")
	flag.StringVar(&config.AuditFile, "audit-file", "", "path to audit log file")
	flag.StringVar(&config.AuditURL, "audit-url", "", "remote audit server url")
	flag.BoolVar(&config.Secure, "s", false, "use https")
	flag.StringVar(&config.ConfigFilePath, "c", "", "configuration file path")

	flag.Parse()

	if evnServerAddres, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		config.ServerAddress = evnServerAddres
	}

	if evnBaseURLAddress, ok := os.LookupEnv("BASE_URL"); ok {
		config.BaseURLAddress = evnBaseURLAddress
	}

	if evnLoggerLevel, ok := os.LookupEnv("LOG_LEVEL"); ok {
		config.LoggerLevel = evnLoggerLevel
	}

	if fileStoragePath, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		config.FileStoragePath = fileStoragePath
	}

	if databaseDSN, ok := os.LookupEnv("DATABASE_DSN"); ok {
		config.DatabaseDSN = databaseDSN
	}

	if auditFile, ok := os.LookupEnv("AUDIT_FILE"); ok {
		config.AuditFile = auditFile
	}

	if auditFile, ok := os.LookupEnv("AUDIT_URL"); ok {
		config.AuditFile = auditFile
	}

	if secure, ok := os.LookupEnv("ENABLE_HTTPS"); ok {
		if value, err := strconv.ParseBool(secure); err == nil {
			config.Secure = value
		} else {
			slog.Error("Cant parse ENABLE_HTTPS env", "ENABLE_HTTPS", secure)
		}
	}

	if configPath, ok := os.LookupEnv("CONFIG"); ok {
		config.ConfigFilePath = configPath
	}

	if config.ConfigFilePath != "" {
		fileConfigs, err := getFileConfigs(config.ConfigFilePath)
		if err != nil {
			slog.Error("FileConfig", "error", err)
		} else {
			mergeConfigs(&config, fileConfigs)
		}
	}

	slog.Info("Configurations:::",
		slog.String("Server address", config.ServerAddress),
		slog.String("Base short URL", config.BaseURLAddress),
		slog.String("Log level", config.LoggerLevel),
	)

	return &config
}

func getFileConfigs(path string) (*Config, error) {
	f, err := os.Open(path)

	if err != nil {
		return nil, fmt.Errorf("fail to open file: %w", err)
	}
	defer f.Close()

	var fileConfigs Config

	if err := json.NewDecoder(f).Decode(&fileConfigs); err != nil {
		return nil, fmt.Errorf("fail to decode json config: %w", err)
	}

	return &fileConfigs, nil
}

// перенести в генератор
func mergeConfigs(scfg *Config, fcfg *Config) {
	if scfg.ServerAddress == "" {
		scfg.ServerAddress = fcfg.ServerAddress
	}
	if scfg.BaseURLAddress == "" {
		scfg.BaseURLAddress = fcfg.BaseURLAddress
	}
	if scfg.LoggerLevel == "" {
		scfg.LoggerLevel = fcfg.LoggerLevel
	}
	if scfg.FileStoragePath == "" {
		scfg.FileStoragePath = fcfg.FileStoragePath
	}
	if scfg.DatabaseDSN == "" {
		scfg.DatabaseDSN = fcfg.DatabaseDSN
	}
	if scfg.AuditFile == "" {
		scfg.AuditFile = fcfg.AuditFile
	}
	if scfg.AuditURL == "" {
		scfg.AuditURL = fcfg.AuditURL
	}
	if !scfg.Secure {
		scfg.Secure = fcfg.Secure
	}
}
