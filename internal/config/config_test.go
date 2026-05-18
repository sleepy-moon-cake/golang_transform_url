package config

import (
	"flag"
	"os"
	"testing"
)

// Тест проверяет, что если переменные окружения не заданы,
// конфигурация корректно заполняется дефолтными значениями
func TestGetConfig_Defaults(t *testing.T) {
	// Подменяем аргументы командной строки на пустые на время теста,
	// чтобы flag.Parse() не пытался читать системные флаги `go test`
	oldArgs := os.Args
	os.Args = []string{"cmd"}
	defer func() { os.Args = oldArgs }()

	// Пересоздаем CommandLine, чтобы очистить старые регистрации из других тестов
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	// Очищаем переменные окружения на время теста
	os.Unsetenv("SERVER_ADDRESS")
	os.Unsetenv("BASE_URL")
	os.Unsetenv("FILE_STORAGE_PATH")

	cfg := GetConfig()

	if cfg.ServerAddress != "localhost:8080" {
		t.Errorf("expected default ServerAddress 'localhost:8080', got '%s'", cfg.ServerAddress)
	}
	if cfg.FileStoragePath != "storage.json" {
		t.Errorf("expected default FileStoragePath 'storage.json', got '%s'", cfg.FileStoragePath)
	}
}

// Тест проверяет приоритет: переменные окружения ДОЛЖНЫ перекрывать значения флагов
func TestGetConfig_EnvPriority(t *testing.T) {
	oldArgs := os.Args
	os.Args = []string{"cmd"}
	defer func() { os.Args = oldArgs }()

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	// Задаем переменные окружения
	expectedAddr := "127.0.0.1:9090"
	expectedBase := "http://my-domain.com"
	expectedPath := "/tmp/test.json"

	os.Setenv("SERVER_ADDRESS", expectedAddr)
	os.Setenv("BASE_URL", expectedBase)
	os.Setenv("FILE_STORAGE_PATH", expectedPath)

	defer func() {
		os.Unsetenv("SERVER_ADDRESS")
		os.Unsetenv("BASE_URL")
		os.Unsetenv("FILE_STORAGE_PATH")
	}()

	cfg := GetConfig()

	if cfg.ServerAddress != expectedAddr {
		t.Errorf("expected ServerAddress to be overridden by ENV to '%s', got '%s'", expectedAddr, cfg.ServerAddress)
	}
	if cfg.BaseURLAddress != expectedBase {
		t.Errorf("expected BaseURLAddress to be overridden by ENV to '%s', got '%s'", expectedBase, cfg.BaseURLAddress)
	}
	if cfg.FileStoragePath != expectedPath {
		t.Errorf("expected FileStoragePath to be overridden by ENV to '%s', got '%s'", expectedPath, cfg.FileStoragePath)
	}
}
