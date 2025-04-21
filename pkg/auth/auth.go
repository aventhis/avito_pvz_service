package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/aventhis/avito_pvz_service/pkg/models"
)

var (
	ErrInvalidCredentials = errors.New("неверные учетные данные")
	ErrInvalidToken       = errors.New("неверный токен")
	ErrUnauthorized       = errors.New("неавторизован")
	ErrForbidden          = errors.New("доступ запрещен")
)

// Auth представляет сервис аутентификации
type Auth struct {
	secret string
}

// TokenClaims представляет JWT-токен
type TokenClaims struct {
	jwt.StandardClaims
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

// New создает новый экземпляр Auth
func New(secret string) *Auth {
	return &Auth{secret: secret}
}

// HashPassword хеширует пароль
func (a *Auth) HashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

// GenerateToken генерирует JWT-токен для пользователя
func (a *Auth) GenerateToken(user *models.User) (string, error) {
	claims := TokenClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		UserID: user.ID,
		Role:   user.Role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(a.secret))
}

// GenerateDummyToken генерирует тестовый JWT-токен с указанной ролью
func (a *Auth) GenerateDummyToken(role string) (string, error) {
	if role != "employee" && role != "moderator" {
		return "", fmt.Errorf("недопустимая роль: %s", role)
	}

	claims := TokenClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		UserID: "dummy-user",
		Role:   role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(a.secret))
}

// ValidateToken проверяет JWT-токен и возвращает claims
func (a *Auth) ValidateToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&TokenClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("неожиданный метод подписи: %v", token.Header["alg"])
			}
			return []byte(a.secret), nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// CheckRole проверяет, имеет ли пользователь указанную роль
func (a *Auth) CheckRole(tokenString, requiredRole string) error {
	claims, err := a.ValidateToken(tokenString)
	if err != nil {
		return ErrUnauthorized
	}

	if claims.Role != requiredRole {
		return ErrForbidden
	}

	return nil
}

// CheckRoleAny проверяет, имеет ли пользователь одну из указанных ролей
func (a *Auth) CheckRoleAny(tokenString string, roles ...string) error {
	claims, err := a.ValidateToken(tokenString)
	if err != nil {
		return ErrUnauthorized
	}

	for _, role := range roles {
		if claims.Role == role {
			return nil
		}
	}

	return ErrForbidden
} 