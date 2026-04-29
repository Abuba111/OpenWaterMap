package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"openwatermap/database"
	"openwatermap/models"
)

// AuthHandler держит зависимости для авторизации
type AuthHandler struct {
	db *database.DB
}

// NewAuthHandler создаёт хендлер авторизации
func NewAuthHandler(db *database.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

// Register — регистрация нового пользователя
// POST /api/auth/register
// Новые пользователи получают роль "user" по умолчанию
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "некорректный JSON")
		return
	}

	// Валидация
	if errMsg := req.Validate(); errMsg != "" {
		respondError(w, http.StatusBadRequest, errMsg)
		return
	}

	// Проверить что email не занят
	exists, err := h.db.EmailExists(req.Email)
	if err != nil {
		log.Printf("Register EmailExists ошибка: %v", err)
		respondError(w, http.StatusInternalServerError, "ошибка сервера")
		return
	}
	if exists {
		respondError(w, http.StatusConflict, "email уже зарегистрирован")
		return
	}

	// Создать пользователя с ролью "user"
	user, err := h.db.CreateUser(&req, models.RoleUser)
	if err != nil {
		log.Printf("Register CreateUser ошибка: %v", err)
		respondError(w, http.StatusInternalServerError, "ошибка создания пользователя")
		return
	}

	// Выдать токен
	token, err := GenerateToken(user)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "ошибка создания токена")
		return
	}

	respondJSON(w, http.StatusCreated, models.AuthResponse{
		Token: token,
		User:  user,
	})
}

// Login — вход в систему
// POST /api/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "некорректный JSON")
		return
	}

	if req.Email == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "email и пароль обязательны")
		return
	}

	// Найти пользователя
	user, err := h.db.GetUserByEmail(req.Email)
	if err != nil {
		log.Printf("Login GetUserByEmail ошибка: %v", err)
		respondError(w, http.StatusInternalServerError, "ошибка сервера")
		return
	}

	// Не говорим что именно неверно — email или пароль (безопасность)
	if user == nil || !h.db.CheckPassword(user, req.Password) {
		respondError(w, http.StatusUnauthorized, "неверный email или пароль")
		return
	}

	if !user.IsActive {
		respondError(w, http.StatusForbidden, "аккаунт заблокирован")
		return
	}

	// Выдать токен
	token, err := GenerateToken(user)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "ошибка создания токена")
		return
	}

	respondJSON(w, http.StatusOK, models.AuthResponse{
		Token: token,
		User:  user,
	})
}

// Me — получить данные текущего пользователя
// GET /api/auth/me
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "не авторизован")
		return
	}

	user, err := h.db.GetUserByID(claims.UserID)
	if err != nil || user == nil {
		respondError(w, http.StatusNotFound, "пользователь не найден")
		return
	}

	respondJSON(w, http.StatusOK, user)
}

// GetUsers — список всех пользователей (только админ)
// GET /api/auth/users
func (h *AuthHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.db.GetAllUsers()
	if err != nil {
		log.Printf("GetUsers ошибка: %v", err)
		respondError(w, http.StatusInternalServerError, "ошибка получения пользователей")
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{
		"data":  users,
		"count": len(users),
	})
}

// UpdateRole — изменить роль пользователя (только админ)
// PUT /api/auth/users/{id}/role
func (h *AuthHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	// Достать ID из chi параметра
	idStr := chi.URLParam(r, "id")
	id, err := parseID(idStr)
	if err != nil || id <= 0 {
		respondError(w, http.StatusBadRequest, "некорректный ID")
		return
	}

	var body struct {
		Role models.Role `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "некорректный JSON")
		return
	}

	// Проверить допустимость роли
	switch body.Role {
	case models.RoleAdmin, models.RoleDL1, models.RoleDL2, models.RoleUser:
	default:
		respondError(w, http.StatusBadRequest, "неверная роль")
		return
	}

	if err := h.db.UpdateUserRole(id, body.Role); err != nil {
		respondError(w, http.StatusInternalServerError, "ошибка обновления роли")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "роль обновлена"})
}

// parseID парсит числовой ID
func parseID(s string) (int, error) {
	var id int
	_, err := fmt.Sscanf(s, "%d", &id)
	if err != nil || id <= 0 {
		return 0, err
	}
	return id, nil
}
