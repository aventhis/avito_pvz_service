package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aventhis/avito_pvz_service/pkg/auth"
	"github.com/aventhis/avito_pvz_service/pkg/models"
	"github.com/aventhis/avito_pvz_service/pkg/storage/mock"
)

func TestDummyLogin(t *testing.T) {
	mockStorage := mock.New()
	authService := auth.New("test-secret")
	api := New(mockStorage, authService)

	// Создаем запрос
	reqBody := models.DummyLoginRequest{
		Role: "employee",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/dummyLogin", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Создаем recorder для записи ответа
	rr := httptest.NewRecorder()

	// Обрабатываем запрос
	api.ServeHTTP(rr, req)

	// Проверяем статус код
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Неверный статус код: получен %v, ожидался %v", status, http.StatusOK)
	}

	// Проверяем, что в ответе есть токен
	var response string
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Ошибка при декодировании ответа: %v", err)
	}

	if response == "" {
		t.Errorf("Пустой токен в ответе")
	}
}

func TestCreatePVZ(t *testing.T) {
	mockStorage := mock.New()
	authService := auth.New("test-secret")
	api := New(mockStorage, authService)

	// Создаем тестовый токен с ролью модератора
	token, _ := authService.GenerateDummyToken("moderator")

	// Создаем запрос
	reqBody := models.PVZ{
		City: "Москва",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/pvz", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Создаем recorder для записи ответа
	rr := httptest.NewRecorder()

	// Обрабатываем запрос
	api.ServeHTTP(rr, req)

	// Проверяем статус код
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Неверный статус код: получен %v, ожидался %v", status, http.StatusCreated)
	}

	// Проверяем ответ
	var response models.PVZ
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Ошибка при декодировании ответа: %v", err)
	}

	if response.ID == "" {
		t.Errorf("Пустой ID в ответе")
	}
	if response.City != "Москва" {
		t.Errorf("Неверный город: получен %v, ожидался %v", response.City, "Москва")
	}
}

func TestCreatePVZForbidden(t *testing.T) {
	mockStorage := mock.New()
	authService := auth.New("test-secret")
	api := New(mockStorage, authService)

	// Создаем тестовый токен с ролью сотрудника (не модератора)
	token, _ := authService.GenerateDummyToken("employee")

	// Создаем запрос
	reqBody := models.PVZ{
		City: "Москва",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/pvz", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Создаем recorder для записи ответа
	rr := httptest.NewRecorder()

	// Обрабатываем запрос
	api.ServeHTTP(rr, req)

	// Проверяем статус код
	if status := rr.Code; status != http.StatusForbidden {
		t.Errorf("Неверный статус код: получен %v, ожидался %v", status, http.StatusForbidden)
	}
}

func TestCreatePVZInvalidCity(t *testing.T) {
	mockStorage := mock.New()
	authService := auth.New("test-secret")
	api := New(mockStorage, authService)

	// Создаем тестовый токен с ролью модератора
	token, _ := authService.GenerateDummyToken("moderator")

	// Создаем запрос с неверным городом
	reqBody := models.PVZ{
		City: "Новосибирск", // Неверный город
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/pvz", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Создаем recorder для записи ответа
	rr := httptest.NewRecorder()

	// Обрабатываем запрос
	api.ServeHTTP(rr, req)

	// Проверяем статус код
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Неверный статус код: получен %v, ожидался %v", status, http.StatusBadRequest)
	}
} 