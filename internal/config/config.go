package config

import (
	"flag"
	"log/slog"
	"os"
	"strconv"
)

type Config struct {
	ServerAddress   string
	BaseURLAddress  string
	LoggerLevel     string
	FileStoragePath string
	DatabaseDSN     string
	AuditFile       string
	AuditURL        string
	Secure          bool
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

	slog.Info("Configurations:::",
		slog.String("Server address", config.ServerAddress),
		slog.String("Base short URL", config.BaseURLAddress),
		slog.String("Log level", config.LoggerLevel),
	)

	return &config
}
