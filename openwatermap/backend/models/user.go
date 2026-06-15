package models

import "time"

// Role — роль пользователя в системе
type Role string

const (
	RoleAdmin   Role = "admin"   // Админ — полный доступ
	RoleDL1     Role = "dl1"     // Доверенное лицо 1 — верификатор
	RoleDL2     Role = "dl2"     // Доверенное лицо 2 — сборщик данных
	RoleUser    Role = "user"    // Обычный пользователь — только просмотр
)

// User — пользователь системы
type User struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // никогда не отдаём в JSON
	Role         Role      `json:"role"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}

// RegisterRequest — данные для регистрации
type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest — данные для входа
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse — ответ после логина/регистрации
type AuthResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

// Validate проверяет данные регистрации
func (r *RegisterRequest) Validate() string {
	if r.Name == "" {
		return "name обязателен"
	}
	if len(r.Name) > 100 {
		return "name слишком длинный"
	}
	if r.Email == "" {
		return "email обязателен"
	}
	if len(r.Email) > 200 {
		return "email слишком длинный"
	}
	if len(r.Password) < 6 {
		return "пароль минимум 6 символов"
	}
	return ""
}
