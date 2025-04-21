package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aventhis/avito_pvz_service/pkg/auth"
	"github.com/aventhis/avito_pvz_service/pkg/models"
	"github.com/aventhis/avito_pvz_service/pkg/storage/mock"
	"github.com/stretchr/testify/assert"
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

// TestGetPVZList проверяет получение списка ПВЗ
func TestGetPVZList(t *testing.T) {
	mockStorage := mock.New()
	authService := auth.New("test-secret")
	api := New(mockStorage, authService)

	// Создаем тестовый токен с ролью сотрудника
	token, _ := authService.GenerateDummyToken("employee")

	// Создаем тестовые ПВЗ
	pvz1 := &models.PVZ{
		ID:               "pvz-id-1",
		RegistrationDate: time.Now(),
		City:             "Москва",
	}
	mockStorage.CreatePVZ(pvz1)
	
	pvz2 := &models.PVZ{
		ID:               "pvz-id-2", 
		RegistrationDate: time.Now(),
		City:             "Санкт-Петербург",
	}
	mockStorage.CreatePVZ(pvz2)

	// Создаем запрос
	req := httptest.NewRequest(http.MethodGet, "/pvz?page=1&limit=10", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Создаем recorder для записи ответа
	rr := httptest.NewRecorder()

	// Обрабатываем запрос
	api.ServeHTTP(rr, req)

	// Проверяем статус код
	assert.Equal(t, http.StatusOK, rr.Code)

	// Проверяем ответ
	var response []models.PVZListItem
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 2)
}

// TestGetPVZList_WithPagination проверяет получение списка ПВЗ с пагинацией
func TestGetPVZList_WithPagination(t *testing.T) {
	mockStorage := mock.New()
	authService := auth.New("test-secret")
	api := New(mockStorage, authService)

	// Создаем тестовый токен с ролью сотрудника
	token, _ := authService.GenerateDummyToken("employee")

	// Создаем тестовые ПВЗ
	for i := 1; i <= 15; i++ {
		pvz := &models.PVZ{
			ID:               fmt.Sprintf("pvz-id-%d", i),
			RegistrationDate: time.Now(),
			City:             "Москва",
		}
		mockStorage.CreatePVZ(pvz)
	}

	// Создаем запрос с пагинацией - страница 2, лимит 5
	req := httptest.NewRequest(http.MethodGet, "/pvz?page=2&limit=5", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Создаем recorder для записи ответа
	rr := httptest.NewRecorder()

	// Обрабатываем запрос
	api.ServeHTTP(rr, req)

	// Проверяем статус код
	assert.Equal(t, http.StatusOK, rr.Code)

	// Проверяем ответ
	var response []models.PVZListItem
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 5) // Вторая страница должна содержать 5 элементов
}

// TestGetPVZList_InvalidParams проверяет обработку некорректных параметров пагинации
func TestGetPVZList_InvalidParams(t *testing.T) {
	mockStorage := mock.New()
	authService := auth.New("test-secret")
	api := New(mockStorage, authService)

	// Создаем тестовый токен с ролью сотрудника
	token, _ := authService.GenerateDummyToken("employee")

	// Создаем запрос с некорректными параметрами
	req := httptest.NewRequest(http.MethodGet, "/pvz?page=invalid&limit=invalid", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Создаем recorder для записи ответа
	rr := httptest.NewRecorder()

	// Обрабатываем запрос
	api.ServeHTTP(rr, req)

	// Проверяем статус код - должен быть 200, так как используются значения по умолчанию
	assert.Equal(t, http.StatusOK, rr.Code)
}

// TestGetPVZList_WithDateFilter проверяет получение списка ПВЗ с фильтрацией по дате
func TestGetPVZList_WithDateFilter(t *testing.T) {
	mockStorage := mock.New()
	authService := auth.New("test-secret")
	api := New(mockStorage, authService)

	// Создаем тестовый токен с ролью сотрудника
	token, _ := authService.GenerateDummyToken("employee")

	// Создаем тестовый ПВЗ
	pvz := &models.PVZ{
		ID:               "pvz-id-1",
		RegistrationDate: time.Now(),
		City:             "Москва",
	}
	mockStorage.CreatePVZ(pvz)

	// Создаем тестовую приемку
	reception := &models.Reception{
		ID:       "reception-id",
		DateTime: time.Now(),
		PVZID:    pvz.ID,
		Status:   "in_progress",
	}
	mockStorage.CreateReception(reception)

	// Создаем запрос с фильтрацией по дате - последние 7 дней
	startDate := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
	endDate := time.Now().Format("2006-01-02")
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/pvz?page=1&limit=10&startDate=%s&endDate=%s", startDate, endDate), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Создаем recorder для записи ответа
	rr := httptest.NewRecorder()

	// Обрабатываем запрос
	api.ServeHTTP(rr, req)

	// Проверяем статус код
	assert.Equal(t, http.StatusOK, rr.Code)
}

// TestCreatePVZWithModerator проверяет создание ПВЗ модератором
func TestCreatePVZWithModerator(t *testing.T) {
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
	assert.Equal(t, http.StatusCreated, rr.Code)

	// Проверяем ответ
	var response models.PVZ
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response.ID)
	assert.Equal(t, "Москва", response.City)
}

// TestCreatePVZWithEmployee проверяет невозможность создания ПВЗ сотрудником
func TestCreatePVZWithEmployee(t *testing.T) {
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

	// Проверяем статус код - должен быть 403 Forbidden
	assert.Equal(t, http.StatusForbidden, rr.Code)
}

// TestCreatePVZInvalidCityAdditional проверяет ошибку при создании ПВЗ с неверным городом
func TestCreatePVZInvalidCityAdditional(t *testing.T) {
	mockStorage := mock.New()
	authService := auth.New("test-secret")
	api := New(mockStorage, authService)

	// Создаем тестовый токен с ролью модератора
	token, _ := authService.GenerateDummyToken("moderator")

	// Создаем запрос с неверным городом
	reqBody := models.PVZ{
		City: "Новосибирск", // Неверный город, не входит в список допустимых
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/pvz", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Создаем recorder для записи ответа
	rr := httptest.NewRecorder()

	// Обрабатываем запрос
	api.ServeHTTP(rr, req)

	// Проверяем статус код - должен быть 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// TestCreatePVZInvalidJSON проверяет ошибку при невалидном JSON
func TestCreatePVZInvalidJSON(t *testing.T) {
	mockStorage := mock.New()
	authService := auth.New("test-secret")
	api := New(mockStorage, authService)

	// Создаем тестовый токен с ролью модератора
	token, _ := authService.GenerateDummyToken("moderator")

	// Создаем запрос с неверным JSON
	invalidJSON := []byte(`{"city": "Москва"`) // Незакрытая скобка
	req := httptest.NewRequest(http.MethodPost, "/pvz", bytes.NewReader(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Создаем recorder для записи ответа
	rr := httptest.NewRecorder()

	// Обрабатываем запрос
	api.ServeHTTP(rr, req)

	// Проверяем статус код - должен быть 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// TestCreateReception проверяет создание приемки
func TestCreateReception(t *testing.T) {
	mockStorage := mock.New()
	authService := auth.New("test-secret")
	api := New(mockStorage, authService)

	// Создаем тестовый токен с ролью сотрудника
	token, _ := authService.GenerateDummyToken("employee")

	// Создаем тестовый ПВЗ
	pvz := &models.PVZ{
		ID:               "test-pvz-id",
		RegistrationDate: time.Now(),
		City:             "Москва",
	}
	mockStorage.CreatePVZ(pvz)

	// Создаем запрос
	reqBody := models.ReceptionRequest{
		PVZID: pvz.ID,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/receptions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Создаем recorder для записи ответа
	rr := httptest.NewRecorder()

	// Обрабатываем запрос
	api.ServeHTTP(rr, req)

	// Проверяем статус код
	assert.Equal(t, http.StatusCreated, rr.Code)

	// Проверяем ответ
	var response models.Reception
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response.ID)
	assert.Equal(t, pvz.ID, response.PVZID)
	assert.Equal(t, "in_progress", response.Status)
}

// TestCreateReception_Forbidden проверяет запрет создания приемки для неавторизованного пользователя
func TestCreateReception_Forbidden(t *testing.T) {
	mockStorage := mock.New()
	authService := auth.New("test-secret")
	api := New(mockStorage, authService)

	// Создаем тестовый токен с ролью модератора (не сотрудника)
	token, _ := authService.GenerateDummyToken("moderator")

	// Создаем тестовый ПВЗ
	pvz := &models.PVZ{
		ID:               "test-pvz-id",
		RegistrationDate: time.Now(),
		City:             "Москва",
	}
	mockStorage.CreatePVZ(pvz)

	// Создаем запрос
	reqBody := models.ReceptionRequest{
		PVZID: pvz.ID,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/receptions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Создаем recorder для записи ответа
	rr := httptest.NewRecorder()

	// Обрабатываем запрос
	api.ServeHTTP(rr, req)

	// Проверяем статус код
	assert.Equal(t, http.StatusForbidden, rr.Code)
}

// TestCreateReception_InvalidPVZ проверяет ошибку при создании приемки с неверным PVZ ID
func TestCreateReception_InvalidPVZ(t *testing.T) {
	mockStorage := mock.New()
	authService := auth.New("test-secret")
	api := New(mockStorage, authService)

	// Создаем тестовый токен с ролью сотрудника
	token, _ := authService.GenerateDummyToken("employee")

	// Создаем запрос с несуществующим PVZ ID
	reqBody := models.ReceptionRequest{
		PVZID: "nonexistent-pvz-id",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/receptions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Создаем recorder для записи ответа
	rr := httptest.NewRecorder()

	// Обрабатываем запрос
	api.ServeHTTP(rr, req)

	// Проверяем статус код - должен быть 400 или 404
	assert.Contains(t, []int{http.StatusBadRequest, http.StatusNotFound}, rr.Code)
}

// TestCloseLastReception2 проверяет закрытие последней приемки
func TestCloseLastReception2(t *testing.T) {
	mockStorage := mock.New()
	authService := auth.New("test-secret")
	api := New(mockStorage, authService)

	// Создаем тестовый токен с ролью сотрудника
	token, _ := authService.GenerateDummyToken("employee")

	// Создаем тестовый ПВЗ
	pvz := &models.PVZ{
		ID:               "test-pvz-id-2",
		RegistrationDate: time.Now(),
		City:             "Москва",
	}
	mockStorage.CreatePVZ(pvz)

	// Создаем приемку
	reception := &models.Reception{
		PVZID:  pvz.ID,
		Status: "in_progress",
	}
	mockStorage.CreateReception(reception)

	// Создаем запрос
	req := httptest.NewRequest(http.MethodPost, "/pvz/"+pvz.ID+"/close_last_reception", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Создаем recorder для записи ответа
	rr := httptest.NewRecorder()

	// Обрабатываем запрос
	api.ServeHTTP(rr, req)

	// Проверяем статус код
	assert.Equal(t, http.StatusOK, rr.Code)

	// Проверяем ответ
	var response models.Reception
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "close", response.Status)
}

// TestCloseLastReception_Forbidden2 проверяет запрет закрытия приемки для неавторизованного пользователя
func TestCloseLastReception_Forbidden2(t *testing.T) {
	mockStorage := mock.New()
	authService := auth.New("test-secret")
	api := New(mockStorage, authService)

	// Создаем тестовый токен с ролью модератора (не сотрудника)
	token, _ := authService.GenerateDummyToken("moderator")

	// Создаем тестовый ПВЗ
	pvz := &models.PVZ{
		ID:               "test-pvz-id-2",
		RegistrationDate: time.Now(),
		City:             "Москва",
	}
	mockStorage.CreatePVZ(pvz)

	// Создаем приемку
	reception := &models.Reception{
		PVZID:  pvz.ID,
		Status: "in_progress",
	}
	mockStorage.CreateReception(reception)

	// Создаем запрос
	req := httptest.NewRequest(http.MethodPost, "/pvz/"+pvz.ID+"/close_last_reception", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Создаем recorder для записи ответа
	rr := httptest.NewRecorder()

	// Обрабатываем запрос
	api.ServeHTTP(rr, req)

	// Проверяем статус код
	assert.Equal(t, http.StatusForbidden, rr.Code)
}

// TestCreateProduct проверяет создание товара
func TestCreateProduct(t *testing.T) {
	mockStorage := mock.New()
	authService := auth.New("test-secret")
	api := New(mockStorage, authService)

	// Создаем тестовый токен с ролью сотрудника
	token, _ := authService.GenerateDummyToken("employee")

	// Создаем тестовый ПВЗ
	pvz := &models.PVZ{
		ID:               "test-pvz-id",
		RegistrationDate: time.Now(),
		City:             "Москва",
	}
	mockStorage.CreatePVZ(pvz)

	// Создаем приемку
	reception := &models.Reception{
		PVZID:  pvz.ID,
		Status: "in_progress",
	}
	mockStorage.CreateReception(reception)

	// Создаем запрос
	reqBody := models.ProductRequest{
		PVZID: pvz.ID,
		Type:  "электроника",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Создаем recorder для записи ответа
	rr := httptest.NewRecorder()

	// Обрабатываем запрос
	api.ServeHTTP(rr, req)

	// Проверяем статус код
	assert.Equal(t, http.StatusCreated, rr.Code)

	// Проверяем ответ
	var response models.Product
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response.ID)
	assert.Equal(t, "электроника", response.Type)
	assert.Equal(t, reception.ID, response.ReceptionID)
}

// TestCreateProduct_InvalidType проверяет ошибку при создании товара с недопустимым типом
func TestCreateProduct_InvalidType(t *testing.T) {
	mockStorage := mock.New()
	authService := auth.New("test-secret")
	api := New(mockStorage, authService)

	// Создаем тестовый токен с ролью сотрудника
	token, _ := authService.GenerateDummyToken("employee")

	// Создаем тестовый ПВЗ
	pvz := &models.PVZ{
		ID:               "test-pvz-id",
		RegistrationDate: time.Now(),
		City:             "Москва",
	}
	mockStorage.CreatePVZ(pvz)

	// Создаем приемку
	reception := &models.Reception{
		PVZID:  pvz.ID,
		Status: "in_progress",
	}
	mockStorage.CreateReception(reception)

	// Создаем запрос с неверным типом товара
	reqBody := models.ProductRequest{
		PVZID: pvz.ID,
		Type:  "продукты", // Недопустимый тип
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Создаем recorder для записи ответа
	rr := httptest.NewRecorder()

	// Обрабатываем запрос
	api.ServeHTTP(rr, req)

	// Проверяем статус код
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	// Проверяем сообщение об ошибке
	var response models.Error
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response.Message, "Недопустимый тип товара")
}

// TestCreateProduct_ClosedReception проверяет ошибку при добавлении товара в закрытую приемку
func TestCreateProduct_ClosedReception(t *testing.T) {
	mockStorage := mock.New()
	authService := auth.New("test-secret")
	api := New(mockStorage, authService)

	// Создаем тестовый токен с ролью сотрудника
	token, _ := authService.GenerateDummyToken("employee")

	// Создаем тестовый ПВЗ
	pvz := &models.PVZ{
		ID:               "test-pvz-id",
		RegistrationDate: time.Now(),
		City:             "Москва",
	}
	mockStorage.CreatePVZ(pvz)

	// Создаем приемку в статусе in_progress
	reception := &models.Reception{
		PVZID:  pvz.ID,
		Status: "in_progress",
	}
	mockStorage.CreateReception(reception)
	
	// Закрываем приемку вручную перед созданием товара
	mockStorage.CloseReception(reception.ID)

	// Создаем запрос на добавление товара
	reqBody := models.ProductRequest{
		PVZID: pvz.ID,
		Type:  "электроника",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Создаем recorder для записи ответа
	rr := httptest.NewRecorder()

	// Обрабатываем запрос
	api.ServeHTTP(rr, req)

	// Проверяем статус код
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	// Проверяем сообщение об ошибке
	var response models.Error
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response.Message, "закрыта")
}

// TestDeleteLastProduct проверяет удаление последнего товара из приемки
func TestDeleteLastProduct(t *testing.T) {
	mockStorage := mock.New()
	authService := auth.New("test-secret")
	api := New(mockStorage, authService)

	// Создаем тестовый токен с ролью сотрудника
	token, _ := authService.GenerateDummyToken("employee")

	// Создаем тестовый ПВЗ
	pvz := &models.PVZ{
		ID:               "test-pvz-id",
		RegistrationDate: time.Now(),
		City:             "Москва",
	}
	mockStorage.CreatePVZ(pvz)

	// Создаем приемку
	reception := &models.Reception{
		PVZID:  pvz.ID,
		Status: "in_progress",
	}
	mockStorage.CreateReception(reception)

	// Создаем товар
	product := &models.Product{
		Type:        "электроника",
		ReceptionID: reception.ID,
	}
	mockStorage.CreateProduct(product)

	// Создаем запрос
	req := httptest.NewRequest(http.MethodPost, "/pvz/"+pvz.ID+"/delete_last_product", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Создаем recorder для записи ответа
	rr := httptest.NewRecorder()

	// Обрабатываем запрос
	api.ServeHTTP(rr, req)

	// Проверяем статус код
	assert.Equal(t, http.StatusOK, rr.Code)
}

// TestDeleteLastProduct_Forbidden проверяет запрет удаления товара для неавторизованного пользователя
func TestDeleteLastProduct_Forbidden(t *testing.T) {
	mockStorage := mock.New()
	authService := auth.New("test-secret")
	api := New(mockStorage, authService)

	// Создаем тестовый токен с ролью модератора (не сотрудника)
	token, _ := authService.GenerateDummyToken("moderator")

	// Создаем тестовый ПВЗ
	pvz := &models.PVZ{
		ID:               "test-pvz-id",
		RegistrationDate: time.Now(),
		City:             "Москва",
	}
	mockStorage.CreatePVZ(pvz)

	// Создаем приемку
	reception := &models.Reception{
		PVZID:  pvz.ID,
		Status: "in_progress",
	}
	mockStorage.CreateReception(reception)

	// Создаем товар
	product := &models.Product{
		Type:        "электроника",
		ReceptionID: reception.ID,
	}
	mockStorage.CreateProduct(product)

	// Создаем запрос
	req := httptest.NewRequest(http.MethodPost, "/pvz/"+pvz.ID+"/delete_last_product", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Создаем recorder для записи ответа
	rr := httptest.NewRecorder()

	// Обрабатываем запрос
	api.ServeHTTP(rr, req)

	// Проверяем статус код
	assert.Equal(t, http.StatusForbidden, rr.Code)
}

// TestDeleteLastProduct_ClosedReception проверяет ошибку при удалении товара из закрытой приемки
func TestDeleteLastProduct_ClosedReception(t *testing.T) {
	mockStorage := mock.New()
	authService := auth.New("test-secret")
	api := New(mockStorage, authService)

	// Создаем тестовый токен с ролью сотрудника
	token, _ := authService.GenerateDummyToken("employee")

	// Создаем тестовый ПВЗ
	pvz := &models.PVZ{
		ID:               "test-pvz-id",
		RegistrationDate: time.Now(),
		City:             "Москва",
	}
	mockStorage.CreatePVZ(pvz)

	// Создаем приемку
	reception := &models.Reception{
		PVZID:  pvz.ID,
		Status: "in_progress",
	}
	mockStorage.CreateReception(reception)

	// Создаем товар
	product := &models.Product{
		Type:        "электроника",
		ReceptionID: reception.ID,
	}
	mockStorage.CreateProduct(product)
	
	// Закрываем приемку
	mockStorage.CloseReception(reception.ID)

	// Создаем запрос
	req := httptest.NewRequest(http.MethodPost, "/pvz/"+pvz.ID+"/delete_last_product", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Создаем recorder для записи ответа
	rr := httptest.NewRecorder()

	// Обрабатываем запрос
	api.ServeHTTP(rr, req)

	// Проверяем статус код
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	
	// Проверяем сообщение об ошибке
	var response models.Error
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response.Message, "закрыта")
}

// TestDeleteLastProduct_NoProducts проверяет ошибку при удалении товара, когда их нет
func TestDeleteLastProduct_NoProducts(t *testing.T) {
	mockStorage := mock.New()
	authService := auth.New("test-secret")
	api := New(mockStorage, authService)

	// Создаем тестовый токен с ролью сотрудника
	token, _ := authService.GenerateDummyToken("employee")

	// Создаем тестовый ПВЗ
	pvz := &models.PVZ{
		ID:               "test-pvz-id",
		RegistrationDate: time.Now(),
		City:             "Москва",
	}
	mockStorage.CreatePVZ(pvz)

	// Создаем приемку без товаров
	reception := &models.Reception{
		PVZID:  pvz.ID,
		Status: "in_progress",
	}
	mockStorage.CreateReception(reception)

	// Создаем запрос
	req := httptest.NewRequest(http.MethodPost, "/pvz/"+pvz.ID+"/delete_last_product", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Создаем recorder для записи ответа
	rr := httptest.NewRecorder()

	// Обрабатываем запрос
	api.ServeHTTP(rr, req)

	// Проверяем статус код
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	
	// Проверяем сообщение об ошибке
	var response models.Error
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response.Message, "нет товаров")
}

// TestRegister проверяет регистрацию пользователя
func TestRegister(t *testing.T) {
	mockStorage := mock.New()
	authService := auth.New("test-secret")
	api := New(mockStorage, authService)

	// Создаем запрос
	reqBody := models.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Role:     "employee",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Создаем recorder для записи ответа
	rr := httptest.NewRecorder()

	// Обрабатываем запрос
	api.ServeHTTP(rr, req)

	// Проверяем статус код
	assert.Equal(t, http.StatusCreated, rr.Code)

	// Проверяем ответ
	var response models.User
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response.ID)
	assert.Equal(t, reqBody.Email, response.Email)
	assert.Equal(t, reqBody.Role, response.Role)
	assert.Empty(t, response.Password) // Пароль не должен возвращаться
}

// TestRegister_InvalidData проверяет ошибку при регистрации с неверными данными
func TestRegister_InvalidData(t *testing.T) {
	mockStorage := mock.New()
	authService := auth.New("test-secret")
	api := New(mockStorage, authService)

	// Случаи с неверными данными
	testCases := []struct {
		name     string
		reqBody  models.RegisterRequest
		expected int
	}{
		{
			name: "Empty Email",
			reqBody: models.RegisterRequest{
				Email:    "",
				Password: "password123",
				Role:     "employee",
			},
			expected: http.StatusBadRequest,
		},
		{
			name: "Empty Password",
			reqBody: models.RegisterRequest{
				Email:    "test@example.com",
				Password: "",
				Role:     "employee",
			},
			expected: http.StatusBadRequest,
		},
		{
			name: "Invalid Role",
			reqBody: models.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
				Role:     "invalid-role",
			},
			expected: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.reqBody)
			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			api.ServeHTTP(rr, req)

			assert.Equal(t, tc.expected, rr.Code)
		})
	}
}

// TestLogin проверяет авторизацию пользователя
func TestLogin(t *testing.T) {
	mockStorage := mock.New()
	authService := auth.New("test-secret")
	api := New(mockStorage, authService)

	// Создаем пользователя
	user := &models.User{
		Email:    "test@example.com",
		Password: authService.HashPassword("password123"),
		Role:     "employee",
	}
	mockStorage.CreateUser(user)

	// Создаем запрос на логин
	reqBody := models.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Создаем recorder для записи ответа
	rr := httptest.NewRecorder()

	// Обрабатываем запрос
	api.ServeHTTP(rr, req)

	// Проверяем статус код
	assert.Equal(t, http.StatusOK, rr.Code)

	// Проверяем, что в ответе есть токен
	var response string
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response)
}

// TestLogin_InvalidCredentials проверяет ошибку при авторизации с неверными учетными данными
func TestLogin_InvalidCredentials(t *testing.T) {
	mockStorage := mock.New()
	authService := auth.New("test-secret")
	api := New(mockStorage, authService)

	// Создаем пользователя
	user := &models.User{
		Email:    "test@example.com",
		Password: authService.HashPassword("password123"),
		Role:     "employee",
	}
	mockStorage.CreateUser(user)

	// Тестовые случаи
	testCases := []struct {
		name     string
		email    string
		password string
		expected int
	}{
		{
			name:     "Wrong Email",
			email:    "wrong@example.com",
			password: "password123",
			expected: http.StatusUnauthorized,
		},
		{
			name:     "Wrong Password",
			email:    "test@example.com",
			password: "wrongpassword",
			expected: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reqBody := models.LoginRequest{
				Email:    tc.email,
				Password: tc.password,
			}
			body, _ := json.Marshal(reqBody)
			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			api.ServeHTTP(rr, req)

			assert.Equal(t, tc.expected, rr.Code)
		})
	}
} 