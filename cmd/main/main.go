package main

import (
	"log"
	"net/http"
	"os"

	"github.com/aventhis/avito_pvz_service/internal/api"
	"github.com/aventhis/avito_pvz_service/internal/auth"
	"github.com/aventhis/avito_pvz_service/internal/storage/postgres"
)

func main() {
	// Получаем переменные окружения
	dbURL := getEnv("DB_URL", "postgres://postgres:postgres@db:5432/postgres?sslmode=disable")
	jwtSecret := getEnv("JWT_SECRET", "your-secret-key")
	port := getEnv("PORT", "8080")

	// Инициализируем хранилище
	storage, err := postgres.New(dbURL)
	if err != nil {
		log.Fatalf("Ошибка при инициализации хранилища: %v", err)
	}
	defer storage.Close()

	// Инициализируем базу данных
	if err := storage.InitDB(); err != nil {
		log.Fatalf("Ошибка при инициализации базы данных: %v", err)
	}

	// Инициализируем сервис аутентификации
	authService := auth.New(jwtSecret)

	// Инициализируем API
	apiService := api.New(storage, authService)

	// Запускаем сервер
	log.Printf("Сервер запущен на http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, apiService); err != nil {
		log.Fatalf("Ошибка при запуске сервера: %v", err)
	}
}

// getEnv получает значение переменной окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
} 