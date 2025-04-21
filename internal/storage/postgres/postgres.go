package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/aventhis/avito_pvz_service/internal/models"
)

// PostgresStorage реализация интерфейса Storage для PostgreSQL
type PostgresStorage struct {
	db *sql.DB
}

// New создает новый экземпляр PostgresStorage
func New(connStr string) (*PostgresStorage, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStorage{db: db}, nil
}

// CreateUser создает нового пользователя в базе данных
func (s *PostgresStorage) CreateUser(user *models.User) error {
	user.ID = uuid.New().String()
	query := `INSERT INTO users (id, email, password, role) VALUES ($1, $2, $3, $4)`
	_, err := s.db.Exec(query, user.ID, user.Email, user.Password, user.Role)
	return err
}

// GetUserByEmail получает пользователя по email
func (s *PostgresStorage) GetUserByEmail(email string) (*models.User, error) {
	query := `SELECT id, email, password, role FROM users WHERE email = $1`
	var user models.User
	err := s.db.QueryRow(query, email).Scan(&user.ID, &user.Email, &user.Password, &user.Role)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// CreatePVZ создает новый ПВЗ в базе данных
func (s *PostgresStorage) CreatePVZ(pvz *models.PVZ) error {
	pvz.ID = uuid.New().String()
	pvz.RegistrationDate = time.Now()
	query := `INSERT INTO pvz (id, registration_date, city) VALUES ($1, $2, $3)`
	_, err := s.db.Exec(query, pvz.ID, pvz.RegistrationDate, pvz.City)
	return err
}

// GetPVZByID получает ПВЗ по ID
func (s *PostgresStorage) GetPVZByID(id string) (*models.PVZ, error) {
	query := `SELECT id, registration_date, city FROM pvz WHERE id = $1`
	var pvz models.PVZ
	err := s.db.QueryRow(query, id).Scan(&pvz.ID, &pvz.RegistrationDate, &pvz.City)
	if err != nil {
		return nil, err
	}
	return &pvz, nil
}

// GetPVZList получает список ПВЗ с фильтрацией по дате приемки и пагинацией
func (s *PostgresStorage) GetPVZList(startDate, endDate *time.Time, page, limit int) ([]models.PVZListItem, error) {
	offset := (page - 1) * limit

	var query string
	var args []interface{}

	if startDate != nil && endDate != nil {
		query = `
			SELECT p.id, p.registration_date, p.city 
			FROM pvz p
			INNER JOIN receptions r ON p.id = r.pvz_id
			WHERE r.date_time BETWEEN $1 AND $2
			GROUP BY p.id, p.registration_date, p.city
			ORDER BY p.registration_date DESC
			LIMIT $3 OFFSET $4
		`
		args = []interface{}{startDate, endDate, limit, offset}
	} else {
		query = `
			SELECT id, registration_date, city 
			FROM pvz
			ORDER BY registration_date DESC
			LIMIT $1 OFFSET $2
		`
		args = []interface{}{limit, offset}
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.PVZListItem
	for rows.Next() {
		var pvz models.PVZ
		if err := rows.Scan(&pvz.ID, &pvz.RegistrationDate, &pvz.City); err != nil {
			return nil, err
		}

		receptions, err := s.getReceptionsWithProductsByPVZID(pvz.ID)
		if err != nil {
			return nil, err
		}

		result = append(result, models.PVZListItem{
			PVZ:        pvz,
			Receptions: receptions,
		})
	}

	return result, nil
}

// getReceptionsWithProductsByPVZID получает приемки с товарами для ПВЗ
func (s *PostgresStorage) getReceptionsWithProductsByPVZID(pvzID string) ([]models.ReceptionWithProducts, error) {
	query := `
		SELECT id, date_time, pvz_id, status
		FROM receptions
		WHERE pvz_id = $1
		ORDER BY date_time DESC
	`
	rows, err := s.db.Query(query, pvzID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.ReceptionWithProducts
	for rows.Next() {
		var reception models.Reception
		if err := rows.Scan(&reception.ID, &reception.DateTime, &reception.PVZID, &reception.Status); err != nil {
			return nil, err
		}

		products, err := s.GetProductsByReceptionID(reception.ID)
		if err != nil {
			return nil, err
		}

		result = append(result, models.ReceptionWithProducts{
			Reception: reception,
			Products:  products,
		})
	}

	return result, nil
}

// CreateReception создает новую приемку в базе данных
func (s *PostgresStorage) CreateReception(reception *models.Reception) error {
	reception.ID = uuid.New().String()
	reception.DateTime = time.Now()
	reception.Status = "in_progress"

	// Проверяем, нет ли незакрытой приемки
	lastReception, err := s.GetLastReceptionByPVZID(reception.PVZID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if lastReception != nil && lastReception.Status == "in_progress" {
		return fmt.Errorf("уже есть незакрытая приемка для этого ПВЗ")
	}

	query := `INSERT INTO receptions (id, date_time, pvz_id, status) VALUES ($1, $2, $3, $4)`
	_, err = s.db.Exec(query, reception.ID, reception.DateTime, reception.PVZID, reception.Status)
	return err
}

// GetLastReceptionByPVZID получает последнюю приемку для ПВЗ
func (s *PostgresStorage) GetLastReceptionByPVZID(pvzID string) (*models.Reception, error) {
	query := `
		SELECT id, date_time, pvz_id, status
		FROM receptions
		WHERE pvz_id = $1
		ORDER BY date_time DESC
		LIMIT 1
	`
	var reception models.Reception
	err := s.db.QueryRow(query, pvzID).Scan(&reception.ID, &reception.DateTime, &reception.PVZID, &reception.Status)
	if err != nil {
		return nil, err
	}
	return &reception, nil
}

// CloseReception закрывает приемку
func (s *PostgresStorage) CloseReception(receptionID string) error {
	query := `UPDATE receptions SET status = 'close' WHERE id = $1 AND status = 'in_progress'`
	result, err := s.db.Exec(query, receptionID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("приемка уже закрыта или не существует")
	}

	return nil
}

// CreateProduct создает новый товар в базе данных
func (s *PostgresStorage) CreateProduct(product *models.Product) error {
	product.ID = uuid.New().String()
	product.DateTime = time.Now()

	query := `INSERT INTO products (id, date_time, type, reception_id) VALUES ($1, $2, $3, $4)`
	_, err := s.db.Exec(query, product.ID, product.DateTime, product.Type, product.ReceptionID)
	return err
}

// GetProductsByReceptionID получает товары по ID приемки
func (s *PostgresStorage) GetProductsByReceptionID(receptionID string) ([]models.Product, error) {
	query := `
		SELECT id, date_time, type, reception_id
		FROM products
		WHERE reception_id = $1
		ORDER BY date_time ASC
	`
	rows, err := s.db.Query(query, receptionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var product models.Product
		if err := rows.Scan(&product.ID, &product.DateTime, &product.Type, &product.ReceptionID); err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	return products, nil
}

// DeleteLastProductInReception удаляет последний добавленный товар в приемке
func (s *PostgresStorage) DeleteLastProductInReception(receptionID string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Получаем ID последнего добавленного товара
	query := `
		SELECT id
		FROM products
		WHERE reception_id = $1
		ORDER BY date_time DESC
		LIMIT 1
	`
	var productID string
	err = tx.QueryRow(query, receptionID).Scan(&productID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("нет товаров для удаления")
		}
		return err
	}

	// Удаляем товар
	_, err = tx.Exec(`DELETE FROM products WHERE id = $1`, productID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// InitDB инициализирует базу данных
func (s *PostgresStorage) InitDB() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			role TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS pvz (
			id UUID PRIMARY KEY,
			registration_date TIMESTAMP NOT NULL,
			city TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS receptions (
			id UUID PRIMARY KEY,
			date_time TIMESTAMP NOT NULL,
			pvz_id UUID NOT NULL,
			status TEXT NOT NULL,
			FOREIGN KEY (pvz_id) REFERENCES pvz (id)
		)`,
		`CREATE TABLE IF NOT EXISTS products (
			id UUID PRIMARY KEY,
			date_time TIMESTAMP NOT NULL,
			type TEXT NOT NULL,
			reception_id UUID NOT NULL,
			FOREIGN KEY (reception_id) REFERENCES receptions (id)
		)`,
	}

	for _, query := range queries {
		_, err := s.db.Exec(query)
		if err != nil {
			return err
		}
	}

	return nil
}

// Close закрывает соединение с базой данных
func (s *PostgresStorage) Close() error {
	return s.db.Close()
} 