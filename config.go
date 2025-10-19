package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost         string
	DBPort         string
	DBUser         string
	DBPassword     string
	DBName         string
	OllamaHost     string
	OllamaModel    string
	GeminiAPIKey   string
	DeepSeekAPIKey string
}

func loadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("คำเตือน: ไม่พบไฟล์ .env:", err)
	}

	return &Config{
		DBHost:         getEnv("DB_HOST", "localhost"),
		DBPort:         getEnv("DB_PORT", "5432"),
		DBUser:         getEnv("DB_USER", "postgres"),
		DBPassword:     getEnv("DB_PASSWORD", ""),
		DBName:         getEnv("DB_NAME", "testvector"),
		OllamaHost:     getEnv("OLLAMA_HOST", "http://localhost:11434"),
		OllamaModel:    getEnv("OLLAMA_MODEL", "bge-m3"),
		GeminiAPIKey:   getEnv("GEMINI_API_KEY", ""),
		DeepSeekAPIKey: getEnv("DEEPSEEK_API_KEY", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
