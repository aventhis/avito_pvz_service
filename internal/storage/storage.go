package storage

import (
	"time"

	"github.com/aventhis/avito_pvz_service/internal/models"
)

// Storage интерфейс для работы с хранилищем данных
type Storage interface {
	// Пользователи
	CreateUser(user *models.User) error
	GetUserByEmail(email string) (*models.User, error)

	// ПВЗ
	CreatePVZ(pvz *models.PVZ) error
	GetPVZByID(id string) (*models.PVZ, error)
	GetPVZList(startDate, endDate *time.Time, page, limit int) ([]models.PVZListItem, error)

	// Приемки
	CreateReception(reception *models.Reception) error
	GetLastReceptionByPVZID(pvzID string) (*models.Reception, error)
	CloseReception(receptionID string) error

	// Товары
	CreateProduct(product *models.Product) error
	GetProductsByReceptionID(receptionID string) ([]models.Product, error)
	DeleteLastProductInReception(receptionID string) error
} 