package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/aventhis/avito_pvz_service/pkg/auth"
	"github.com/aventhis/avito_pvz_service/pkg/models"
	"github.com/aventhis/avito_pvz_service/pkg/storage"
)

// API представляет API-сервис
type API struct {
	router  *mux.Router
	storage storage.Storage
	auth    *auth.Auth
}

// New создает новый экземпляр API
func New(storage storage.Storage, auth *auth.Auth) *API {
	api := &API{
		router:  mux.NewRouter(),
		storage: storage,
		auth:    auth,
	}

	api.setupRoutes()
	return api
}

// setupRoutes настраивает маршруты API
func (a *API) setupRoutes() {
	// Аутентификация
	a.router.HandleFunc("/dummyLogin", a.handleDummyLogin).Methods(http.MethodPost)
	a.router.HandleFunc("/register", a.handleRegister).Methods(http.MethodPost)
	a.router.HandleFunc("/login", a.handleLogin).Methods(http.MethodPost)

	// ПВЗ
	a.router.HandleFunc("/pvz", a.handleCreatePVZ).Methods(http.MethodPost)
	a.router.HandleFunc("/pvz", a.handleGetPVZList).Methods(http.MethodGet)
	a.router.HandleFunc("/pvz/{pvzId}/close_last_reception", a.handleCloseLastReception).Methods(http.MethodPost)
	a.router.HandleFunc("/pvz/{pvzId}/delete_last_product", a.handleDeleteLastProduct).Methods(http.MethodPost)

	// Приемки и товары
	a.router.HandleFunc("/receptions", a.handleCreateReception).Methods(http.MethodPost)
	a.router.HandleFunc("/products", a.handleCreateProduct).Methods(http.MethodPost)
}

// ServeHTTP обслуживает HTTP-запросы
func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.router.ServeHTTP(w, r)
}

// getTokenFromHeader извлекает токен из заголовка Authorization
func (a *API) getTokenFromHeader(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}

// respondWithJSON отправляет JSON-ответ
func (a *API) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Ошибка при маршалинге JSON: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// respondWithError отправляет JSON-ответ с ошибкой
func (a *API) respondWithError(w http.ResponseWriter, code int, message string) {
	a.respondWithJSON(w, code, models.Error{Message: message})
}

// handleDummyLogin обрабатывает запрос на тестовую авторизацию
func (a *API) handleDummyLogin(w http.ResponseWriter, r *http.Request) {
	var req models.DummyLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.respondWithError(w, http.StatusBadRequest, "Неверный запрос")
		return
	}

	token, err := a.auth.GenerateDummyToken(req.Role)
	if err != nil {
		a.respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	a.respondWithJSON(w, http.StatusOK, token)
}

// handleRegister обрабатывает запрос на регистрацию
func (a *API) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.respondWithError(w, http.StatusBadRequest, "Неверный запрос")
		return
	}

	// Проверяем валидность данных
	if req.Email == "" || req.Password == "" || (req.Role != "employee" && req.Role != "moderator") {
		a.respondWithError(w, http.StatusBadRequest, "Неверные данные")
		return
	}

	// Создаем пользователя
	user := &models.User{
		Email:    req.Email,
		Password: a.auth.HashPassword(req.Password),
		Role:     req.Role,
	}

	if err := a.storage.CreateUser(user); err != nil {
		a.respondWithError(w, http.StatusInternalServerError, "Ошибка при создании пользователя")
		return
	}

	// Не отправляем пароль в ответе
	user.Password = ""
	a.respondWithJSON(w, http.StatusCreated, user)
}

// handleLogin обрабатывает запрос на авторизацию
func (a *API) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.respondWithError(w, http.StatusBadRequest, "Неверный запрос")
		return
	}

	// Получаем пользователя по email
	user, err := a.storage.GetUserByEmail(req.Email)
	if err != nil {
		a.respondWithError(w, http.StatusUnauthorized, "Неверные учетные данные")
		return
	}

	// Проверяем пароль
	if user.Password != a.auth.HashPassword(req.Password) {
		a.respondWithError(w, http.StatusUnauthorized, "Неверные учетные данные")
		return
	}

	// Генерируем токен
	token, err := a.auth.GenerateToken(user)
	if err != nil {
		a.respondWithError(w, http.StatusInternalServerError, "Ошибка при генерации токена")
		return
	}

	a.respondWithJSON(w, http.StatusOK, token)
}

// handleCreatePVZ обрабатывает запрос на создание ПВЗ
func (a *API) handleCreatePVZ(w http.ResponseWriter, r *http.Request) {
	// Проверяем роль
	token := a.getTokenFromHeader(r)
	if err := a.auth.CheckRole(token, "moderator"); err != nil {
		a.respondWithError(w, http.StatusForbidden, "Доступ запрещен")
		return
	}

	var pvz models.PVZ
	if err := json.NewDecoder(r.Body).Decode(&pvz); err != nil {
		a.respondWithError(w, http.StatusBadRequest, "Неверный запрос")
		return
	}

	// Проверяем город
	if pvz.City != "Москва" && pvz.City != "Санкт-Петербург" && pvz.City != "Казань" {
		a.respondWithError(w, http.StatusBadRequest, "ПВЗ можно создать только в городах: Москва, Санкт-Петербург, Казань")
		return
	}

	// Создаем ПВЗ
	if err := a.storage.CreatePVZ(&pvz); err != nil {
		a.respondWithError(w, http.StatusInternalServerError, "Ошибка при создании ПВЗ")
		return
	}

	a.respondWithJSON(w, http.StatusCreated, pvz)
}

// handleGetPVZList обрабатывает запрос на получение списка ПВЗ
func (a *API) handleGetPVZList(w http.ResponseWriter, r *http.Request) {
	// Проверяем роль
	token := a.getTokenFromHeader(r)
	if err := a.auth.CheckRoleAny(token, "employee", "moderator"); err != nil {
		a.respondWithError(w, http.StatusForbidden, "Доступ запрещен")
		return
	}

	// Параметры пагинации и фильтрации
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	startDateStr := r.URL.Query().Get("startDate")
	endDateStr := r.URL.Query().Get("endDate")

	page := 1
	limit := 10

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 30 {
			limit = l
		}
	}

	var startDate, endDate *time.Time
	if startDateStr != "" {
		if t, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			startDate = &t
		}
	}

	if endDateStr != "" {
		if t, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			endDate = &t
		}
	}

	// Получаем список ПВЗ
	pvzList, err := a.storage.GetPVZList(startDate, endDate, page, limit)
	if err != nil {
		a.respondWithError(w, http.StatusInternalServerError, "Ошибка при получении списка ПВЗ")
		return
	}

	a.respondWithJSON(w, http.StatusOK, pvzList)
}

// handleCreateReception обрабатывает запрос на создание приемки
func (a *API) handleCreateReception(w http.ResponseWriter, r *http.Request) {
	// Проверяем роль
	token := a.getTokenFromHeader(r)
	if err := a.auth.CheckRole(token, "employee"); err != nil {
		a.respondWithError(w, http.StatusForbidden, "Доступ запрещен")
		return
	}

	var req models.ReceptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.respondWithError(w, http.StatusBadRequest, "Неверный запрос")
		return
	}

	// Проверяем существование ПВЗ
	pvz, err := a.storage.GetPVZByID(req.PVZID)
	if err != nil {
		a.respondWithError(w, http.StatusBadRequest, "ПВЗ не найден")
		return
	}

	// Создаем приемку
	reception := &models.Reception{
		PVZID: pvz.ID,
	}

	if err := a.storage.CreateReception(reception); err != nil {
		a.respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	a.respondWithJSON(w, http.StatusCreated, reception)
}

// handleCloseLastReception обрабатывает запрос на закрытие последней приемки
func (a *API) handleCloseLastReception(w http.ResponseWriter, r *http.Request) {
	// Проверяем роль
	token := a.getTokenFromHeader(r)
	if err := a.auth.CheckRole(token, "employee"); err != nil {
		a.respondWithError(w, http.StatusForbidden, "Доступ запрещен")
		return
	}

	// Получаем ID ПВЗ
	vars := mux.Vars(r)
	pvzID := vars["pvzId"]

	// Получаем последнюю приемку
	reception, err := a.storage.GetLastReceptionByPVZID(pvzID)
	if err != nil {
		a.respondWithError(w, http.StatusBadRequest, "Приемка не найдена")
		return
	}

	// Закрываем приемку
	if err := a.storage.CloseReception(reception.ID); err != nil {
		a.respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Обновляем статус в объекте
	reception.Status = "close"
	a.respondWithJSON(w, http.StatusOK, reception)
}

// handleCreateProduct обрабатывает запрос на добавление товара
func (a *API) handleCreateProduct(w http.ResponseWriter, r *http.Request) {
	// Проверяем роль
	token := a.getTokenFromHeader(r)
	if err := a.auth.CheckRole(token, "employee"); err != nil {
		a.respondWithError(w, http.StatusForbidden, "Доступ запрещен")
		return
	}

	var req models.ProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.respondWithError(w, http.StatusBadRequest, "Неверный запрос")
		return
	}

	// Проверяем тип товара
	if req.Type != "электроника" && req.Type != "одежда" && req.Type != "обувь" {
		a.respondWithError(w, http.StatusBadRequest, "Недопустимый тип товара")
		return
	}

	// Получаем последнюю приемку для ПВЗ
	reception, err := a.storage.GetLastReceptionByPVZID(req.PVZID)
	if err != nil {
		a.respondWithError(w, http.StatusBadRequest, "Активная приемка не найдена")
		return
	}

	// Проверяем, что приемка не закрыта
	if reception.Status != "in_progress" {
		a.respondWithError(w, http.StatusBadRequest, "Приемка уже закрыта")
		return
	}

	// Создаем товар
	product := &models.Product{
		Type:        req.Type,
		ReceptionID: reception.ID,
	}

	if err := a.storage.CreateProduct(product); err != nil {
		a.respondWithError(w, http.StatusInternalServerError, "Ошибка при добавлении товара")
		return
	}

	a.respondWithJSON(w, http.StatusCreated, product)
}

// handleDeleteLastProduct обрабатывает запрос на удаление последнего товара
func (a *API) handleDeleteLastProduct(w http.ResponseWriter, r *http.Request) {
	// Проверяем роль
	token := a.getTokenFromHeader(r)
	if err := a.auth.CheckRole(token, "employee"); err != nil {
		a.respondWithError(w, http.StatusForbidden, "Доступ запрещен")
		return
	}

	// Получаем ID ПВЗ
	vars := mux.Vars(r)
	pvzID := vars["pvzId"]

	// Получаем последнюю приемку
	reception, err := a.storage.GetLastReceptionByPVZID(pvzID)
	if err != nil {
		a.respondWithError(w, http.StatusBadRequest, "Приемка не найдена")
		return
	}

	// Проверяем, что приемка не закрыта
	if reception.Status != "in_progress" {
		a.respondWithError(w, http.StatusBadRequest, "Приемка уже закрыта")
		return
	}

	// Удаляем последний товар
	if err := a.storage.DeleteLastProductInReception(reception.ID); err != nil {
		a.respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	a.respondWithJSON(w, http.StatusOK, struct{}{})
} 