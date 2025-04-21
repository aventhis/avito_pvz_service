package mock

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/aventhis/avito_pvz_service/pkg/models"
)

// MockStorage реализует интерфейс Storage для тестирования
type MockStorage struct {
	users      map[string]*models.User
	usersByEmail map[string]*models.User
	pvzs       map[string]*models.PVZ
	receptions map[string]*models.Reception
	products   map[string]*models.Product
}

// New создает новый экземпляр MockStorage
func New() *MockStorage {
	return &MockStorage{
		users:      make(map[string]*models.User),
		usersByEmail: make(map[string]*models.User),
		pvzs:       make(map[string]*models.PVZ),
		receptions: make(map[string]*models.Reception),
		products:   make(map[string]*models.Product),
	}
}

// CreateUser создает нового пользователя
func (s *MockStorage) CreateUser(user *models.User) error {
	user.ID = uuid.New().String()
	s.users[user.ID] = user
	s.usersByEmail[user.Email] = user
	return nil
}

// GetUserByEmail получает пользователя по email
func (s *MockStorage) GetUserByEmail(email string) (*models.User, error) {
	user, exists := s.usersByEmail[email]
	if !exists {
		return nil, errors.New("пользователь не найден")
	}
	return user, nil
}

// CreatePVZ создает новый ПВЗ
func (s *MockStorage) CreatePVZ(pvz *models.PVZ) error {
	pvz.ID = uuid.New().String()
	pvz.RegistrationDate = time.Now()
	s.pvzs[pvz.ID] = pvz
	return nil
}

// GetPVZByID получает ПВЗ по ID
func (s *MockStorage) GetPVZByID(id string) (*models.PVZ, error) {
	pvz, exists := s.pvzs[id]
	if !exists {
		return nil, errors.New("ПВЗ не найден")
	}
	return pvz, nil
}

// GetPVZList получает список ПВЗ с фильтрацией по дате приемки и пагинацией
func (s *MockStorage) GetPVZList(startDate, endDate *time.Time, page, limit int) ([]models.PVZListItem, error) {
	var result []models.PVZListItem

	for _, pvz := range s.pvzs {
		item := models.PVZListItem{
			PVZ:        *pvz,
			Receptions: []models.ReceptionWithProducts{},
		}

		for _, reception := range s.receptions {
			if reception.PVZID == pvz.ID {
				// Проверяем фильтр по дате
				if startDate != nil && reception.DateTime.Before(*startDate) {
					continue
				}
				if endDate != nil && reception.DateTime.After(*endDate) {
					continue
				}

				var products []models.Product
				for _, product := range s.products {
					if product.ReceptionID == reception.ID {
						products = append(products, *product)
					}
				}

				item.Receptions = append(item.Receptions, models.ReceptionWithProducts{
					Reception: *reception,
					Products:  products,
				})
			}
		}

		result = append(result, item)
	}

	// Применяем пагинацию
	if len(result) == 0 {
		return []models.PVZListItem{}, nil
	}

	startIndex := (page - 1) * limit
	endIndex := startIndex + limit

	if startIndex >= len(result) {
		return []models.PVZListItem{}, nil
	}

	if endIndex > len(result) {
		endIndex = len(result)
	}

	return result[startIndex:endIndex], nil
}

// CreateReception создает новую приемку
func (s *MockStorage) CreateReception(reception *models.Reception) error {
	// Проверяем существование ПВЗ
	_, exists := s.pvzs[reception.PVZID]
	if !exists {
		return errors.New("ПВЗ не найден")
	}

	// Проверяем, нет ли незакрытой приемки
	for _, r := range s.receptions {
		if r.PVZID == reception.PVZID && r.Status == "in_progress" {
			return errors.New("уже есть незакрытая приемка для этого ПВЗ")
		}
	}

	reception.ID = uuid.New().String()
	reception.DateTime = time.Now()
	reception.Status = "in_progress"
	s.receptions[reception.ID] = reception
	return nil
}

// GetLastReceptionByPVZID получает последнюю приемку для ПВЗ
func (s *MockStorage) GetLastReceptionByPVZID(pvzID string) (*models.Reception, error) {
	var lastReception *models.Reception
	var lastTime time.Time

	for _, reception := range s.receptions {
		if reception.PVZID == pvzID && (lastReception == nil || reception.DateTime.After(lastTime)) {
			lastReception = reception
			lastTime = reception.DateTime
		}
	}

	if lastReception == nil {
		return nil, errors.New("приемка не найдена")
	}

	return lastReception, nil
}

// CloseReception закрывает приемку
func (s *MockStorage) CloseReception(receptionID string) error {
	reception, exists := s.receptions[receptionID]
	if !exists {
		return errors.New("приемка не найдена")
	}

	if reception.Status == "close" {
		return errors.New("приемка уже закрыта")
	}

	reception.Status = "close"
	return nil
}

// CreateProduct создает новый товар
func (s *MockStorage) CreateProduct(product *models.Product) error {
	// Проверяем существование приемки
	reception, exists := s.receptions[product.ReceptionID]
	if !exists {
		return errors.New("приемка не найдена")
	}

	// Проверяем, что приемка не закрыта
	if reception.Status == "close" {
		return errors.New("приемка уже закрыта")
	}

	product.ID = uuid.New().String()
	product.DateTime = time.Now()
	s.products[product.ID] = product
	return nil
}

// GetProductsByReceptionID получает товары по ID приемки
func (s *MockStorage) GetProductsByReceptionID(receptionID string) ([]models.Product, error) {
	var result []models.Product

	for _, product := range s.products {
		if product.ReceptionID == receptionID {
			result = append(result, *product)
		}
	}

	return result, nil
}

// DeleteLastProductInReception удаляет последний добавленный товар в приемке
func (s *MockStorage) DeleteLastProductInReception(receptionID string) error {
	var lastProduct *models.Product
	var lastTime time.Time

	for _, product := range s.products {
		if product.ReceptionID == receptionID && (lastProduct == nil || product.DateTime.After(lastTime)) {
			lastProduct = product
			lastTime = product.DateTime
		}
	}

	if lastProduct == nil {
		return errors.New("нет товаров для удаления")
	}

	delete(s.products, lastProduct.ID)
	return nil
} 