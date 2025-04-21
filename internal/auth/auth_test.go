package auth

import (
	"testing"

	"github.com/aventhis/avito_pvz_service/internal/models"
)

func TestHashPassword(t *testing.T) {
	auth := New("test-secret")
	
	password := "test-password"
	hash1 := auth.HashPassword(password)
	hash2 := auth.HashPassword(password)
	
	// Один и тот же пароль должен давать одинаковый хеш
	if hash1 != hash2 {
		t.Errorf("Хеши для одинаковых паролей не совпадают: %s != %s", hash1, hash2)
	}
	
	// Разные пароли должны давать разные хеши
	hash3 := auth.HashPassword("other-password")
	if hash1 == hash3 {
		t.Errorf("Хеши для разных паролей совпадают")
	}
}

func TestGenerateDummyToken(t *testing.T) {
	auth := New("test-secret")
	
	// Проверка с правильной ролью
	token, err := auth.GenerateDummyToken("employee")
	if err != nil {
		t.Errorf("Ошибка при генерации токена: %v", err)
	}
	if token == "" {
		t.Errorf("Пустой токен")
	}
	
	// Проверка с неправильной ролью
	_, err = auth.GenerateDummyToken("invalid-role")
	if err == nil {
		t.Errorf("Должна быть ошибка при неправильной роли")
	}
}

func TestGenerateToken(t *testing.T) {
	auth := New("test-secret")
	
	user := &models.User{
		ID:   "test-id",
		Role: "employee",
	}
	
	token, err := auth.GenerateToken(user)
	if err != nil {
		t.Errorf("Ошибка при генерации токена: %v", err)
	}
	if token == "" {
		t.Errorf("Пустой токен")
	}
}

func TestValidateToken(t *testing.T) {
	auth := New("test-secret")
	
	// Создаем тестовый токен
	user := &models.User{
		ID:   "test-id",
		Role: "employee",
	}
	token, _ := auth.GenerateToken(user)
	
	// Проверяем валидацию токена
	claims, err := auth.ValidateToken(token)
	if err != nil {
		t.Errorf("Ошибка при валидации токена: %v", err)
	}
	if claims.UserID != user.ID || claims.Role != user.Role {
		t.Errorf("Неверные данные в токене")
	}
	
	// Проверяем валидацию неверного токена
	_, err = auth.ValidateToken("invalid-token")
	if err == nil {
		t.Errorf("Должна быть ошибка при неверном токене")
	}
}

func TestCheckRole(t *testing.T) {
	auth := New("test-secret")
	
	// Создаем тестовый токен с ролью employee
	employeeToken, _ := auth.GenerateDummyToken("employee")
	
	// Проверяем с правильной ролью
	err := auth.CheckRole(employeeToken, "employee")
	if err != nil {
		t.Errorf("Ошибка при проверке правильной роли: %v", err)
	}
	
	// Проверяем с неправильной ролью
	err = auth.CheckRole(employeeToken, "moderator")
	if err == nil {
		t.Errorf("Должна быть ошибка при неправильной роли")
	}
}

func TestCheckRoleAny(t *testing.T) {
	auth := New("test-secret")
	
	// Создаем тестовый токен с ролью employee
	employeeToken, _ := auth.GenerateDummyToken("employee")
	
	// Проверяем с правильной ролью
	err := auth.CheckRoleAny(employeeToken, "employee", "moderator")
	if err != nil {
		t.Errorf("Ошибка при проверке правильной роли: %v", err)
	}
	
	// Проверяем с неправильной ролью
	err = auth.CheckRoleAny(employeeToken, "moderator", "admin")
	if err == nil {
		t.Errorf("Должна быть ошибка при неправильных ролях")
	}
} 