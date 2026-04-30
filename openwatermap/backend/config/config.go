package config

import (
	"os"
)

// Config хранит все настройки приложения
type Config struct {
	// Адрес сервера
	ServerPort string

	// Путь к файлу базы данных SQLite
	DatabasePath string

	// Разрешённые источники для CORS (фронтенд)
	AllowedOrigins []string
}

// Load загружает настройки из переменных окружения
// Если переменная не задана — используется значение по умолчанию
func Load() *Config {
	return &Config{
		ServerPort:   getEnv("PORT", "8080"),
		DatabasePath: getEnv("DB_PATH", "./data/water.db"),
		AllowedOrigins: []string{
			getEnv("FRONTEND_URL", "http://localhost:5173"),
			"http://localhost:4173",
			"http://127.0.0.1:5173",
			"http://192.168.0.108:5173", // локальная сеть
		},
	}
}

// getEnv возвращает значение переменной окружения или defaultVal
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
