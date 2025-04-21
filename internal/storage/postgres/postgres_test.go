package postgres

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/aventhis/avito_pvz_service/internal/models"
	"github.com/stretchr/testify/assert"
)

// TestNew проверяет создание нового экземпляра PostgresStorage
func TestNew(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Ошибка при создании mock DB: %v", err)
	}
	defer db.Close()

	// Тест не может напрямую проверить создание через New(),
	// так как он пытается подключиться к реальной базе данных.
	// Вместо этого мы проверяем, что структура создается корректно.
	storage := &PostgresStorage{db: db}
	assert.NotNil(t, storage)
	assert.NotNil(t, storage.db)
}

// TestCreateUser проверяет создание пользователя
func TestCreateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Ошибка при создании mock DB: %v", err)
	}
	defer db.Close()

	storage := &PostgresStorage{db: db}

	user := &models.User{
		Email:    "test@example.com",
		Password: "password123",
		Role:     "employee",
	}

	// ID будет сгенерирован автоматически, поэтому используем регулярное выражение для проверки
	mock.ExpectExec("INSERT INTO users").
		WithArgs(sqlmock.AnyArg(), user.Email, user.Password, user.Role).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = storage.CreateUser(user)
	assert.NoError(t, err)
	assert.NotEmpty(t, user.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestGetUserByEmail проверяет получение пользователя по email
func TestGetUserByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Ошибка при создании mock DB: %v", err)
	}
	defer db.Close()

	storage := &PostgresStorage{db: db}

	// Подготавливаем ожидаемые данные и ответ mock БД
	expectedUser := &models.User{
		ID:       "test-id",
		Email:    "test@example.com",
		Password: "password123",
		Role:     "employee",
	}

	mock.ExpectQuery("SELECT id, email, password, role FROM users WHERE email = \\$1").
		WithArgs(expectedUser.Email).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password", "role"}).
			AddRow(expectedUser.ID, expectedUser.Email, expectedUser.Password, expectedUser.Role))

	// Вызываем тестируемый метод
	user, err := storage.GetUserByEmail(expectedUser.Email)

	// Проверяем результаты
	assert.NoError(t, err)
	assert.Equal(t, expectedUser.ID, user.ID)
	assert.Equal(t, expectedUser.Email, user.Email)
	assert.Equal(t, expectedUser.Password, user.Password)
	assert.Equal(t, expectedUser.Role, user.Role)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestGetUserByEmail_NotFound проверяет получение пользователя по email, когда он не найден
func TestGetUserByEmail_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Ошибка при создании mock DB: %v", err)
	}
	defer db.Close()

	storage := &PostgresStorage{db: db}

	email := "nonexistent@example.com"

	mock.ExpectQuery("SELECT id, email, password, role FROM users WHERE email = \\$1").
		WithArgs(email).
		WillReturnError(sql.ErrNoRows)

	// Вызываем тестируемый метод
	user, err := storage.GetUserByEmail(email)

	// Проверяем результаты
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestCreatePVZ проверяет создание ПВЗ
func TestCreatePVZ(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Ошибка при создании mock DB: %v", err)
	}
	defer db.Close()

	storage := &PostgresStorage{db: db}

	pvz := &models.PVZ{
		City: "Москва",
	}

	mock.ExpectExec("INSERT INTO pvz").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), pvz.City).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = storage.CreatePVZ(pvz)
	assert.NoError(t, err)
	assert.NotEmpty(t, pvz.ID)
	assert.NotEmpty(t, pvz.RegistrationDate)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestGetPVZByID проверяет получение ПВЗ по ID
func TestGetPVZByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Ошибка при создании mock DB: %v", err)
	}
	defer db.Close()

	storage := &PostgresStorage{db: db}

	// Подготавливаем ожидаемые данные и ответ mock БД
	registrationDate := time.Now()
	expectedPVZ := &models.PVZ{
		ID:               "test-id",
		RegistrationDate: registrationDate,
		City:             "Москва",
	}

	mock.ExpectQuery("SELECT id, registration_date, city FROM pvz WHERE id = \\$1").
		WithArgs(expectedPVZ.ID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "registration_date", "city"}).
			AddRow(expectedPVZ.ID, expectedPVZ.RegistrationDate, expectedPVZ.City))

	// Вызываем тестируемый метод
	pvz, err := storage.GetPVZByID(expectedPVZ.ID)

	// Проверяем результаты
	assert.NoError(t, err)
	assert.Equal(t, expectedPVZ.ID, pvz.ID)
	assert.Equal(t, expectedPVZ.City, pvz.City)
	assert.Equal(t, expectedPVZ.RegistrationDate.Unix(), pvz.RegistrationDate.Unix())
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestGetPVZByID_NotFound проверяет получение ПВЗ по ID, когда он не найден
func TestGetPVZByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Ошибка при создании mock DB: %v", err)
	}
	defer db.Close()

	storage := &PostgresStorage{db: db}

	pvzID := "nonexistent-id"

	mock.ExpectQuery("SELECT id, registration_date, city FROM pvz WHERE id = \\$1").
		WithArgs(pvzID).
		WillReturnError(sql.ErrNoRows)

	// Вызываем тестируемый метод
	pvz, err := storage.GetPVZByID(pvzID)

	// Проверяем результаты
	assert.Error(t, err)
	assert.Nil(t, pvz)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestGetPVZList проверяет получение списка ПВЗ без фильтрации по дате
func TestGetPVZList(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Ошибка при создании mock DB: %v", err)
	}
	defer db.Close()

	storage := &PostgresStorage{db: db}

	// Подготавливаем ожидаемые данные
	now := time.Now()
	page := 1
	limit := 10
	
	// Запрос на получение ПВЗ без фильтрации по дате
	mock.ExpectQuery("SELECT id, registration_date, city FROM pvz").
		WithArgs(limit, (page-1)*limit).
		WillReturnRows(sqlmock.NewRows([]string{"id", "registration_date", "city"}).
			AddRow("pvz-id-1", now, "Москва").
			AddRow("pvz-id-2", now, "Санкт-Петербург"))
			
	// Запрос на получение приемок для первого ПВЗ
	mock.ExpectQuery("SELECT id, date_time, pvz_id, status FROM receptions").
		WithArgs("pvz-id-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "date_time", "pvz_id", "status"}).
			AddRow("reception-id-1", now, "pvz-id-1", "in_progress"))
	
	// Запрос на получение товаров для приемки первого ПВЗ
	mock.ExpectQuery("SELECT id, date_time, type, reception_id FROM products").
		WithArgs("reception-id-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "date_time", "type", "reception_id"}).
			AddRow("product-id-1", now, "электроника", "reception-id-1"))
			
	// Запрос на получение приемок для второго ПВЗ
	mock.ExpectQuery("SELECT id, date_time, pvz_id, status FROM receptions").
		WithArgs("pvz-id-2").
		WillReturnRows(sqlmock.NewRows([]string{"id", "date_time", "pvz_id", "status"}))

	// Вызываем тестируемый метод
	pvzList, err := storage.GetPVZList(nil, nil, page, limit)

	// Проверяем результаты
	assert.NoError(t, err)
	assert.Len(t, pvzList, 2)
	assert.Equal(t, "pvz-id-1", pvzList[0].PVZ.ID)
	assert.Equal(t, "pvz-id-2", pvzList[1].PVZ.ID)
	assert.Len(t, pvzList[0].Receptions, 1)
	assert.Len(t, pvzList[1].Receptions, 0)
	assert.Equal(t, "reception-id-1", pvzList[0].Receptions[0].Reception.ID)
	assert.Len(t, pvzList[0].Receptions[0].Products, 1)
	assert.Equal(t, "product-id-1", pvzList[0].Receptions[0].Products[0].ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestGetPVZList_WithDateFilter проверяет получение списка ПВЗ с фильтрацией по дате
func TestGetPVZList_WithDateFilter(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Ошибка при создании mock DB: %v", err)
	}
	defer db.Close()

	storage := &PostgresStorage{db: db}

	// Подготавливаем ожидаемые данные
	now := time.Now()
	startDate := now.AddDate(0, -1, 0) // 1 месяц назад
	endDate := now
	page := 1
	limit := 10
	
	// Запрос на получение ПВЗ с фильтрацией по дате
	mock.ExpectQuery("SELECT p.id, p.registration_date, p.city FROM pvz p").
		WithArgs(startDate, endDate, limit, (page-1)*limit).
		WillReturnRows(sqlmock.NewRows([]string{"id", "registration_date", "city"}).
			AddRow("pvz-id-1", now, "Москва"))
			
	// Запрос на получение приемок для ПВЗ
	mock.ExpectQuery("SELECT id, date_time, pvz_id, status FROM receptions").
		WithArgs("pvz-id-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "date_time", "pvz_id", "status"}).
			AddRow("reception-id-1", now, "pvz-id-1", "in_progress"))
	
	// Запрос на получение товаров для приемки
	mock.ExpectQuery("SELECT id, date_time, type, reception_id FROM products").
		WithArgs("reception-id-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "date_time", "type", "reception_id"}))

	// Вызываем тестируемый метод
	pvzList, err := storage.GetPVZList(&startDate, &endDate, page, limit)

	// Проверяем результаты
	assert.NoError(t, err)
	assert.Len(t, pvzList, 1)
	assert.Equal(t, "pvz-id-1", pvzList[0].PVZ.ID)
	assert.Len(t, pvzList[0].Receptions, 1)
	assert.Equal(t, "reception-id-1", pvzList[0].Receptions[0].Reception.ID)
	assert.Len(t, pvzList[0].Receptions[0].Products, 0)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestCreateReception проверяет создание приемки
func TestCreateReception(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Ошибка при создании mock DB: %v", err)
	}
	defer db.Close()

	storage := &PostgresStorage{db: db}

	reception := &models.Reception{
		PVZID: "pvz-id",
	}

	// Получение последней приемки для проверки, нет ли незакрытой
	mock.ExpectQuery("SELECT id, date_time, pvz_id, status FROM receptions").
		WithArgs(reception.PVZID).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectExec("INSERT INTO receptions").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), reception.PVZID, "in_progress").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = storage.CreateReception(reception)
	assert.NoError(t, err)
	assert.NotEmpty(t, reception.ID)
	assert.NotEmpty(t, reception.DateTime)
	assert.Equal(t, "in_progress", reception.Status)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestCreateReception_ExistingOpenReception проверяет невозможность создания приемки, когда уже есть открытая
func TestCreateReception_ExistingOpenReception(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Ошибка при создании mock DB: %v", err)
	}
	defer db.Close()

	storage := &PostgresStorage{db: db}

	reception := &models.Reception{
		PVZID: "pvz-id",
	}

	// Получение последней приемки - уже есть незакрытая
	lastDateTime := time.Now()
	mock.ExpectQuery("SELECT id, date_time, pvz_id, status FROM receptions").
		WithArgs(reception.PVZID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "date_time", "pvz_id", "status"}).
			AddRow("last-reception-id", lastDateTime, reception.PVZID, "in_progress"))

	err = storage.CreateReception(reception)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "уже есть незакрытая приемка")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestGetLastReceptionByPVZID проверяет получение последней приемки для ПВЗ
func TestGetLastReceptionByPVZID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Ошибка при создании mock DB: %v", err)
	}
	defer db.Close()

	storage := &PostgresStorage{db: db}

	pvzID := "pvz-id"
	now := time.Now()

	mock.ExpectQuery("SELECT id, date_time, pvz_id, status FROM receptions").
		WithArgs(pvzID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "date_time", "pvz_id", "status"}).
			AddRow("reception-id", now, pvzID, "in_progress"))

	reception, err := storage.GetLastReceptionByPVZID(pvzID)
	assert.NoError(t, err)
	assert.Equal(t, "reception-id", reception.ID)
	assert.Equal(t, pvzID, reception.PVZID)
	assert.Equal(t, "in_progress", reception.Status)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestGetLastReceptionByPVZID_NotFound проверяет получение последней приемки для ПВЗ, когда она не найдена
func TestGetLastReceptionByPVZID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Ошибка при создании mock DB: %v", err)
	}
	defer db.Close()

	storage := &PostgresStorage{db: db}

	pvzID := "pvz-id"

	mock.ExpectQuery("SELECT id, date_time, pvz_id, status FROM receptions").
		WithArgs(pvzID).
		WillReturnError(sql.ErrNoRows)

	reception, err := storage.GetLastReceptionByPVZID(pvzID)
	assert.Error(t, err)
	assert.Nil(t, reception)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestCloseReception проверяет закрытие приемки
func TestCloseReception(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Ошибка при создании mock DB: %v", err)
	}
	defer db.Close()

	storage := &PostgresStorage{db: db}

	receptionID := "reception-id"

	mock.ExpectExec("UPDATE receptions SET status = 'close' WHERE id = \\$1 AND status = 'in_progress'").
		WithArgs(receptionID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = storage.CloseReception(receptionID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestCloseReception_AlreadyClosed проверяет ошибку при попытке закрыть уже закрытую приемку
func TestCloseReception_AlreadyClosed(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Ошибка при создании mock DB: %v", err)
	}
	defer db.Close()

	storage := &PostgresStorage{db: db}

	receptionID := "reception-id"

	mock.ExpectExec("UPDATE receptions SET status = 'close' WHERE id = \\$1 AND status = 'in_progress'").
		WithArgs(receptionID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = storage.CloseReception(receptionID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "приемка уже закрыта или не существует")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestCreateProduct проверяет создание товара
func TestCreateProduct(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Ошибка при создании mock DB: %v", err)
	}
	defer db.Close()

	storage := &PostgresStorage{db: db}

	product := &models.Product{
		Type:        "электроника",
		ReceptionID: "reception-id",
	}

	mock.ExpectExec("INSERT INTO products").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), product.Type, product.ReceptionID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = storage.CreateProduct(product)
	assert.NoError(t, err)
	assert.NotEmpty(t, product.ID)
	assert.NotEmpty(t, product.DateTime)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestGetProductsByReceptionID проверяет получение товаров по ID приемки
func TestGetProductsByReceptionID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Ошибка при создании mock DB: %v", err)
	}
	defer db.Close()

	storage := &PostgresStorage{db: db}

	receptionID := "reception-id"
	dateTime := time.Now()

	mock.ExpectQuery("SELECT id, date_time, type, reception_id FROM products").
		WithArgs(receptionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "date_time", "type", "reception_id"}).
			AddRow("product-id-1", dateTime, "электроника", receptionID).
			AddRow("product-id-2", dateTime, "одежда", receptionID))

	products, err := storage.GetProductsByReceptionID(receptionID)
	assert.NoError(t, err)
	assert.Len(t, products, 2)
	assert.Equal(t, "product-id-1", products[0].ID)
	assert.Equal(t, "электроника", products[0].Type)
	assert.Equal(t, "product-id-2", products[1].ID)
	assert.Equal(t, "одежда", products[1].Type)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestDeleteLastProductInReception проверяет удаление последнего товара в приемке
func TestDeleteLastProductInReception(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Ошибка при создании mock DB: %v", err)
	}
	defer db.Close()

	storage := &PostgresStorage{db: db}

	receptionID := "reception-id"
	productID := "product-id-2"
	
	// Начало транзакции
	mock.ExpectBegin()
	
	// Получаем ID последнего добавленного товара
	rows := sqlmock.NewRows([]string{"id"}).AddRow(productID)
	mock.ExpectQuery("SELECT id FROM products WHERE reception_id = \\$1 ORDER BY date_time DESC LIMIT 1").
		WithArgs(receptionID).
		WillReturnRows(rows)

	// Удаляем товар
	mock.ExpectExec("DELETE FROM products WHERE id = \\$1").
		WithArgs(productID).
		WillReturnResult(sqlmock.NewResult(0, 1))
		
	// Коммит транзакции
	mock.ExpectCommit()

	err = storage.DeleteLastProductInReception(receptionID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestDeleteLastProductInReception_NoProducts проверяет ошибку при попытке удалить товар, когда их нет
func TestDeleteLastProductInReception_NoProducts(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Ошибка при создании mock DB: %v", err)
	}
	defer db.Close()

	storage := &PostgresStorage{db: db}

	receptionID := "reception-id"
	
	// Начало транзакции
	mock.ExpectBegin()
	
	// Пустой результат запроса - товаров нет
	mock.ExpectQuery("SELECT id FROM products WHERE reception_id = \\$1 ORDER BY date_time DESC LIMIT 1").
		WithArgs(receptionID).
		WillReturnError(sql.ErrNoRows)
		
	// Откат транзакции при ошибке
	mock.ExpectRollback()

	err = storage.DeleteLastProductInReception(receptionID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "нет товаров для удаления")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestInitDB проверяет инициализацию базы данных
func TestInitDB(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Ошибка при создании mock DB: %v", err)
	}
	defer db.Close()

	storage := &PostgresStorage{db: db}

	// Ожидаем выполнение запросов для создания таблиц
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS users").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS pvz").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS receptions").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS products").WillReturnResult(sqlmock.NewResult(0, 0))

	err = storage.InitDB()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestClose проверяет закрытие соединения с базой данных
func TestClose(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Ошибка при создании mock DB: %v", err)
	}

	storage := &PostgresStorage{db: db}

	// Ожидаем закрытие соединения
	mock.ExpectClose()

	err = storage.Close()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
} 