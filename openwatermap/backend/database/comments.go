package database

import (
	"fmt"
	"time"

	"openwatermap/models"
)

// migrateComments добавляет таблицы комментариев и фото
func (db *DB) migrateMedia() error {
	query := `
	CREATE TABLE IF NOT EXISTS comments (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		point_id   INTEGER NOT NULL REFERENCES water_points(id) ON DELETE CASCADE,
		user_id    INTEGER NOT NULL REFERENCES users(id),
		text       TEXT    NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_comments_point ON comments(point_id);

	CREATE TABLE IF NOT EXISTS photos (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		point_id   INTEGER NOT NULL REFERENCES water_points(id) ON DELETE CASCADE,
		user_id    INTEGER NOT NULL REFERENCES users(id),
		type       TEXT    NOT NULL DEFAULT 'photo'
		           CHECK(type IN ('photo', 'certificate')),
		file_name  TEXT    NOT NULL,
		file_path  TEXT    NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_photos_point ON photos(point_id);
	`
	_, err := db.conn.Exec(query)
	return err
}

// GetComments возвращает комментарии к точке
func (db *DB) GetComments(pointID int) ([]models.Comment, error) {
	rows, err := db.conn.Query(`
		SELECT c.id, c.point_id, c.user_id, u.name, u.role, c.text, c.created_at
		FROM comments c JOIN users u ON u.id = c.user_id
		WHERE c.point_id = ? ORDER BY c.created_at ASC
	`, pointID)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса комментариев: %w", err)
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var c models.Comment
		var createdAt string
		if err := rows.Scan(&c.ID, &c.PointID, &c.UserID, &c.UserName, &c.UserRole, &c.Text, &createdAt); err != nil {
			return nil, err
		}
		c.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		comments = append(comments, c)
	}
	if comments == nil {
		comments = []models.Comment{}
	}
	return comments, nil
}

// CreateComment добавляет комментарий
func (db *DB) CreateComment(pointID, userID int, text string) (*models.Comment, error) {
	result, err := db.conn.Exec(
		`INSERT INTO comments (point_id, user_id, text) VALUES (?, ?, ?)`,
		pointID, userID, text,
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания комментария: %w", err)
	}
	id, _ := result.LastInsertId()

	var c models.Comment
	var createdAt string
	err = db.conn.QueryRow(`
		SELECT c.id, c.point_id, c.user_id, u.name, u.role, c.text, c.created_at
		FROM comments c JOIN users u ON u.id = c.user_id WHERE c.id = ?
	`, id).Scan(&c.ID, &c.PointID, &c.UserID, &c.UserName, &c.UserRole, &c.Text, &createdAt)
	if err != nil {
		return nil, err
	}
	c.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	return &c, nil
}

// DeleteComment удаляет комментарий
func (db *DB) DeleteComment(commentID, userID int, isAdmin bool) error {
	var ownerID int
	if err := db.conn.QueryRow("SELECT user_id FROM comments WHERE id = ?", commentID).Scan(&ownerID); err != nil {
		return fmt.Errorf("комментарий не найден")
	}
	if !isAdmin && ownerID != userID {
		return fmt.Errorf("нет прав на удаление")
	}
	_, err := db.conn.Exec("DELETE FROM comments WHERE id = ?", commentID)
	return err
}

// SavePhoto сохраняет запись о фото
func (db *DB) SavePhoto(pointID, userID int, photoType, fileName, filePath string) (*models.Photo, error) {
	result, err := db.conn.Exec(
		`INSERT INTO photos (point_id, user_id, type, file_name, file_path) VALUES (?, ?, ?, ?, ?)`,
		pointID, userID, photoType, fileName, filePath,
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка сохранения фото: %w", err)
	}
	id, _ := result.LastInsertId()

	var p models.Photo
	var createdAt string
	err = db.conn.QueryRow(`
		SELECT ph.id, ph.point_id, ph.user_id, u.name, ph.type, ph.file_name, ph.file_path, ph.created_at
		FROM photos ph JOIN users u ON u.id = ph.user_id WHERE ph.id = ?
	`, id).Scan(&p.ID, &p.PointID, &p.UserID, &p.UserName, &p.Type, &p.FileName, &p.FilePath, &createdAt)
	if err != nil {
		return nil, err
	}
	p.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	return &p, nil
}

// GetPhotos возвращает фото точки
func (db *DB) GetPhotos(pointID int) ([]models.Photo, error) {
	rows, err := db.conn.Query(`
		SELECT ph.id, ph.point_id, ph.user_id, u.name, ph.type, ph.file_name, ph.file_path, ph.created_at
		FROM photos ph JOIN users u ON u.id = ph.user_id
		WHERE ph.point_id = ? ORDER BY ph.created_at DESC
	`, pointID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var photos []models.Photo
	for rows.Next() {
		var p models.Photo
		var createdAt string
		if err := rows.Scan(&p.ID, &p.PointID, &p.UserID, &p.UserName, &p.Type, &p.FileName, &p.FilePath, &createdAt); err != nil {
			return nil, err
		}
		p.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		photos = append(photos, p)
	}
	if photos == nil {
		photos = []models.Photo{}
	}
	return photos, nil
}
