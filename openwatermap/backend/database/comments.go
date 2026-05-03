package database

import (
	"database/sql"
	"fmt"
	"time"

	"openwatermap/models"
)

func (db *DB) migrateMedia() error { return nil }

func (db *DB) GetComments(pointID int) ([]models.Comment, error) {
	rows, err := db.conn.Query(`
		SELECT c.id, c.point_id, c.user_id, u.name, u.role, c.text, c.created_at
		FROM comments c JOIN users u ON u.id = c.user_id
		WHERE c.point_id = $1 ORDER BY c.created_at ASC
	`, pointID)
	if err != nil {
		return nil, fmt.Errorf("ошибка комментариев: %w", err)
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var c models.Comment
		if err := rows.Scan(&c.ID, &c.PointID, &c.UserID, &c.UserName, &c.UserRole, &c.Text, &c.CreatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	if comments == nil {
		comments = []models.Comment{}
	}
	return comments, nil
}

func (db *DB) CreateComment(pointID, userID int, text string) (*models.Comment, error) {
	var c models.Comment
	err := db.conn.QueryRow(`
		WITH ins AS (
			INSERT INTO comments (point_id, user_id, text) VALUES ($1, $2, $3) RETURNING *
		)
		SELECT ins.id, ins.point_id, ins.user_id, u.name, u.role, ins.text, ins.created_at
		FROM ins JOIN users u ON u.id = ins.user_id
	`, pointID, userID, text).Scan(&c.ID, &c.PointID, &c.UserID, &c.UserName, &c.UserRole, &c.Text, &c.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания комментария: %w", err)
	}
	return &c, nil
}

func (db *DB) DeleteComment(commentID, userID int, isAdmin bool) error {
	var ownerID int
	if err := db.conn.QueryRow("SELECT user_id FROM comments WHERE id = $1", commentID).Scan(&ownerID); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("комментарий не найден")
		}
		return err
	}
	if !isAdmin && ownerID != userID {
		return fmt.Errorf("нет прав на удаление")
	}
	_, err := db.conn.Exec("DELETE FROM comments WHERE id = $1", commentID)
	return err
}

func (db *DB) SavePhoto(pointID, userID int, photoType, fileName, filePath string) (*models.Photo, error) {
	var p models.Photo
	err := db.conn.QueryRow(`
		WITH ins AS (
			INSERT INTO photos (point_id, user_id, type, file_name, file_path) VALUES ($1,$2,$3,$4,$5) RETURNING *
		)
		SELECT ins.id, ins.point_id, ins.user_id, u.name, ins.type, ins.file_name, ins.file_path, ins.created_at
		FROM ins JOIN users u ON u.id = ins.user_id
	`, pointID, userID, photoType, fileName, filePath).Scan(
		&p.ID, &p.PointID, &p.UserID, &p.UserName, &p.Type, &p.FileName, &p.FilePath, &p.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("ошибка сохранения фото: %w", err)
	}
	return &p, nil
}

func (db *DB) GetPhotos(pointID int) ([]models.Photo, error) {
	rows, err := db.conn.Query(`
		SELECT ph.id, ph.point_id, ph.user_id, u.name, ph.type, ph.file_name, ph.file_path, ph.created_at
		FROM photos ph JOIN users u ON u.id = ph.user_id
		WHERE ph.point_id = $1 ORDER BY ph.created_at DESC
	`, pointID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var photos []models.Photo
	for rows.Next() {
		var p models.Photo
		if err := rows.Scan(&p.ID, &p.PointID, &p.UserID, &p.UserName, &p.Type, &p.FileName, &p.FilePath, &p.CreatedAt); err != nil {
			return nil, err
		}
		photos = append(photos, p)
	}
	if photos == nil {
		photos = []models.Photo{}
	}
	return photos, nil
}

var _ = time.Now
