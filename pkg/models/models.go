package models

import (
	"time"
)

// User представляет пользователя системы
type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Password string `json:"-"`
	Role     string `json:"role"` // employee или moderator
}

// PVZ представляет пункт выдачи заказов
type PVZ struct {
	ID               string    `json:"id"`
	RegistrationDate time.Time `json:"registrationDate"`
	City             string    `json:"city"` // Москва, Санкт-Петербург или Казань
}

// Reception представляет приемку товаров
type Reception struct {
	ID       string    `json:"id"`
	DateTime time.Time `json:"dateTime"`
	PVZID    string    `json:"pvzId"`
	Status   string    `json:"status"` // in_progress или close
}

// Product представляет товар
type Product struct {
	ID          string    `json:"id"`
	DateTime    time.Time `json:"dateTime"`
	Type        string    `json:"type"` // электроника, одежда или обувь
	ReceptionID string    `json:"receptionId"`
}

// LoginRequest модель для запроса авторизации
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// DummyLoginRequest модель для тестовой авторизации
type DummyLoginRequest struct {
	Role string `json:"role"`
}

// RegisterRequest модель для запроса регистрации
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

// ReceptionRequest модель для создания приемки
type ReceptionRequest struct {
	PVZID string `json:"pvzId"`
}

// ProductRequest модель для добавления товара
type ProductRequest struct {
	Type  string `json:"type"`
	PVZID string `json:"pvzId"`
}

// Error модель для ошибки
type Error struct {
	Message string `json:"message"`
}

// PVZListItem представляет элемент списка ПВЗ с приемками и товарами
type PVZListItem struct {
	PVZ        PVZ                   `json:"pvz"`
	Receptions []ReceptionWithProducts `json:"receptions"`
}

// ReceptionWithProducts представляет приемку с товарами
type ReceptionWithProducts struct {
	Reception Reception `json:"reception"`
	Products  []Product `json:"products"`
} 