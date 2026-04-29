package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
	"openwatermap/models"
)

// seedAdminIfEmpty создаёт дефолтного админа если пользователей нет
func (db *DB) seedAdminIfEmpty() error {
	var count int
	if err := db.conn.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	// Хэшировать пароль
	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = db.conn.Exec(`
		INSERT INTO users (name, email, password_hash, role)
		VALUES (?, ?, ?, ?)`,
		"Администратор", "admin@openwatermap.kz", string(hash), "admin",
	)
	if err != nil {
		return err
	}

	log.Println("Создан дефолтный админ: admin@openwatermap.kz / admin123")
	log.Println("ВАЖНО: Смени пароль после первого входа!")
	return nil
}

// CreateUser регистрирует нового пользователя
func (db *DB) CreateUser(req *models.RegisterRequest, role models.Role) (*models.User, error) {
	// Хэшировать пароль
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("ошибка хэширования пароля: %w", err)
	}

	result, err := db.conn.Exec(`
		INSERT INTO users (name, email, password_hash, role)
		VALUES (?, ?, ?, ?)`,
		req.Name, req.Email, string(hash), string(role),
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания пользователя: %w", err)
	}

	id, _ := result.LastInsertId()
	return db.GetUserByID(int(id))
}

// GetUserByEmail ищет пользователя по email
func (db *DB) GetUserByEmail(email string) (*models.User, error) {
	var u models.User
	var createdAt string

	err := db.conn.QueryRow(`
		SELECT id, name, email, password_hash, role, is_active, created_at
		FROM users WHERE email = ?`, email,
	).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.IsActive, &createdAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	u.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	return &u, nil
}

// GetUserByID ищет пользователя по ID
func (db *DB) GetUserByID(id int) (*models.User, error) {
	var u models.User
	var createdAt string

	err := db.conn.QueryRow(`
		SELECT id, name, email, password_hash, role, is_active, created_at
		FROM users WHERE id = ?`, id,
	).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.IsActive, &createdAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	u.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	return &u, nil
}

// CheckPassword проверяет пароль пользователя
func (db *DB) CheckPassword(user *models.User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	return err == nil
}

// GetAllUsers возвращает всех пользователей (только для админа)
func (db *DB) GetAllUsers() ([]models.User, error) {
	rows, err := db.conn.Query(`
		SELECT id, name, email, password_hash, role, is_active, created_at
		FROM users ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		var createdAt string
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.IsActive, &createdAt); err != nil {
			return nil, err
		}
		u.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		users = append(users, u)
	}
	if users == nil {
		users = []models.User{}
	}
	return users, nil
}

// UpdateUserRole меняет роль пользователя (только админ)
func (db *DB) UpdateUserRole(userID int, role models.Role) error {
	_, err := db.conn.Exec(`UPDATE users SET role = ? WHERE id = ?`, string(role), userID)
	return err
}

// EmailExists проверяет занят ли email
func (db *DB) EmailExists(email string) (bool, error) {
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", email).Scan(&count)
	return count > 0, err
}
