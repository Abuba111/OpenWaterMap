package config

import "os"

type Config struct {
	ServerPort     string
	DatabaseURL    string
	AllowedOrigins []string
}

func Load() *Config {
	return &Config{
		ServerPort:  getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", ""),
		AllowedOrigins: []string{
			getEnv("FRONTEND_URL", "http://localhost:5173"),
			"http://localhost:4173",
			"http://127.0.0.1:5173",
			"http://192.168.0.108:5173",
			"https://openwatermap-9c9x.onrender.com",
			"https://openwatermap.onrender.com",
		},
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
