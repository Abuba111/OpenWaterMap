package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
	"openwatermap/models"
)

func (db *DB) seedAdminIfEmpty() error {
	var count int
	if err := db.conn.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = db.conn.Exec(`
		INSERT INTO users (name, email, password_hash, role)
		VALUES ($1, $2, $3, $4)`,
		"Администратор", "admin@openwatermap.kz", string(hash), "admin",
	)
	if err != nil {
		return err
	}
	log.Println("Создан дефолтный админ: admin@openwatermap.kz / admin123")
	return nil
}

func (db *DB) CreateUser(req *models.RegisterRequest, role models.Role) (*models.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("ошибка хэширования: %w", err)
	}
	var id int
	err = db.conn.QueryRow(`
		INSERT INTO users (name, email, password_hash, role)
		VALUES ($1, $2, $3, $4) RETURNING id`,
		req.Name, req.Email, string(hash), string(role),
	).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания пользователя: %w", err)
	}
	return db.GetUserByID(id)
}

func (db *DB) GetUserByEmail(email string) (*models.User, error) {
	var u models.User
	err := db.conn.QueryRow(`
		SELECT id, name, email, password_hash, role, is_active, created_at
		FROM users WHERE email = $1`, email,
	).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.IsActive, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (db *DB) GetUserByID(id int) (*models.User, error) {
	var u models.User
	err := db.conn.QueryRow(`
		SELECT id, name, email, password_hash, role, is_active, created_at
		FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.IsActive, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (db *DB) CheckPassword(user *models.User, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) == nil
}

func (db *DB) GetAllUsers() ([]models.User, error) {
	rows, err := db.conn.Query(`
		SELECT id, name, email, password_hash, role, is_active, created_at
		FROM users ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.IsActive, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	if users == nil {
		users = []models.User{}
	}
	return users, nil
}

func (db *DB) UpdateUserRole(userID int, role models.Role) error {
	_, err := db.conn.Exec(`UPDATE users SET role = $1 WHERE id = $2`, string(role), userID)
	return err
}

func (db *DB) EmailExists(email string) (bool, error) {
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM users WHERE email = $1", email).Scan(&count)
	return count > 0, err
}

// Unused but kept for compatibility
var _ = time.Now
