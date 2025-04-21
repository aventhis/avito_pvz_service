package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aventhis/avito_pvz_service/internal/api"
	"github.com/aventhis/avito_pvz_service/internal/auth"
	"github.com/aventhis/avito_pvz_service/internal/models"
	"github.com/aventhis/avito_pvz_service/internal/storage/mock"
)

func TestIntegration(t *testing.T) {
	// Создаем мок-хранилище и API
	mockStorage := mock.New()
	authService := auth.New("test-secret")
	apiService := api.New(mockStorage, authService)

	// Получаем токен модератора
	token, err := authService.GenerateDummyToken("moderator")
	if err != nil {
		t.Fatalf("Ошибка при генерации токена: %v", err)
	}

	// 1. Создаем ПВЗ
	pvz, err := createPVZ(apiService, token, "Москва")
	if err != nil {
		t.Fatalf("Ошибка при создании ПВЗ: %v", err)
	}

	// Получаем токен сотрудника
	employeeToken, err := authService.GenerateDummyToken("employee")
	if err != nil {
		t.Fatalf("Ошибка при генерации токена: %v", err)
	}

	// 2. Создаем приемку
	_, err = createReception(apiService, employeeToken, pvz.ID)
	if err != nil {
		t.Fatalf("Ошибка при создании приемки: %v", err)
	}

	// 3. Добавляем 50 товаров
	productTypes := []string{"электроника", "одежда", "обувь"}
	for i := 0; i < 50; i++ {
		productType := productTypes[i%len(productTypes)]
		_, err := createProduct(apiService, employeeToken, pvz.ID, productType)
		if err != nil {
			t.Fatalf("Ошибка при добавлении товара №%d: %v", i+1, err)
		}
	}

	// 4. Закрываем приемку
	err = closeReception(apiService, employeeToken, pvz.ID)
	if err != nil {
		t.Fatalf("Ошибка при закрытии приемки: %v", err)
	}

	// Проверяем, что приемка действительно закрыта
	pvzList, err := getPVZList(apiService, employeeToken)
	if err != nil {
		t.Fatalf("Ошибка при получении списка ПВЗ: %v", err)
	}

	if len(pvzList) == 0 {
		t.Fatalf("Пустой список ПВЗ")
	}

	if len(pvzList[0].Receptions) == 0 {
		t.Fatalf("Пустой список приемок")
	}

	if pvzList[0].Receptions[0].Reception.Status != "close" {
		t.Errorf("Неверный статус приемки: %s", pvzList[0].Receptions[0].Reception.Status)
	}

	if len(pvzList[0].Receptions[0].Products) != 50 {
		t.Errorf("Неверное количество товаров: %d", len(pvzList[0].Receptions[0].Products))
	}
}

// createPVZ создает ПВЗ
func createPVZ(api *api.API, token, city string) (*models.PVZ, error) {
	reqBody := models.PVZ{
		City: city,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/pvz", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	api.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		return nil, fmt.Errorf("неверный статус код: %d", rr.Code)
	}

	var response models.PVZ
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// createReception создает приемку
func createReception(api *api.API, token, pvzID string) (*models.Reception, error) {
	reqBody := models.ReceptionRequest{
		PVZID: pvzID,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/receptions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	api.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		return nil, fmt.Errorf("неверный статус код: %d", rr.Code)
	}

	var response models.Reception
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// createProduct добавляет товар
func createProduct(api *api.API, token, pvzID, productType string) (*models.Product, error) {
	reqBody := models.ProductRequest{
		PVZID: pvzID,
		Type:  productType,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	api.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		return nil, fmt.Errorf("неверный статус код: %d", rr.Code)
	}

	var response models.Product
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// closeReception закрывает приемку
func closeReception(api *api.API, token, pvzID string) error {
	req := httptest.NewRequest(http.MethodPost, "/pvz/"+pvzID+"/close_last_reception", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	api.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		return fmt.Errorf("неверный статус код: %d", rr.Code)
	}

	return nil
}

// getPVZList получает список ПВЗ
func getPVZList(api *api.API, token string) ([]models.PVZListItem, error) {
	req := httptest.NewRequest(http.MethodGet, "/pvz", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	api.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		return nil, fmt.Errorf("неверный статус код: %d", rr.Code)
	}

	var response []models.PVZListItem
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		return nil, err
	}

	return response, nil
} 