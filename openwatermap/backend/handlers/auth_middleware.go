package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"openwatermap/models"
)

// jwtSecret — секретный ключ для подписи токенов
// В продакшене берётся из переменной окружения JWT_SECRET
var jwtSecret = []byte("openwatermap-secret-key-change-in-production")

// contextKey тип для ключей контекста
type contextKey string

const userContextKey contextKey = "user"

// Claims — данные внутри JWT токена
type Claims struct {
	UserID int         `json:"user_id"`
	Role   models.Role `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken создаёт JWT токен для пользователя
func GenerateToken(user *models.User) (string, error) {
	claims := &Claims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ParseToken парсит и проверяет JWT токен
func ParseToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неверный метод подписи")
		}
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("невалидный токен")
	}
	return claims, nil
}

// AuthMiddleware — проверяет JWT токен в заголовке
// Добавляет данные пользователя в контекст запроса
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Достать токен из заголовка: Authorization: Bearer <token>
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondError(w, http.StatusUnauthorized, "требуется авторизация")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			respondError(w, http.StatusUnauthorized, "неверный формат токена")
			return
		}

		claims, err := ParseToken(parts[1])
		if err != nil {
			respondError(w, http.StatusUnauthorized, "токен недействителен или истёк")
			return
		}

		// Добавить claims в контекст
		ctx := context.WithValue(r.Context(), userContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole — проверяет что пользователь имеет нужную роль
func RequireRole(roles ...models.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetClaims(r)
			if claims == nil {
				respondError(w, http.StatusUnauthorized, "требуется авторизация")
				return
			}

			for _, role := range roles {
				if claims.Role == role {
					next.ServeHTTP(w, r)
					return
				}
			}

			respondError(w, http.StatusForbidden, "недостаточно прав")
		})
	}
}

// GetClaims достаёт данные пользователя из контекста
func GetClaims(r *http.Request) *Claims {
	claims, _ := r.Context().Value(userContextKey).(*Claims)
	return claims
}
