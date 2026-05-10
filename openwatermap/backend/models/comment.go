package models

import "time"

// Comment — комментарий к точке воды
type Comment struct {
	ID        int       `json:"id"`
	PointID   int       `json:"point_id"`
	UserID    int       `json:"user_id"`
	UserName  string    `json:"user_name"`
	UserRole  Role      `json:"user_role"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateCommentRequest — данные для создания комментария
type CreateCommentRequest struct {
	Text string `json:"text"`
}

// Validate проверяет комментарий
func (r *CreateCommentRequest) Validate() string {
	if r.Text == "" {
		return "текст комментария обязателен"
	}
	if len(r.Text) > 1000 {
		return "комментарий слишком длинный (макс 1000 символов)"
	}
	return ""
}

// Photo — фото точки воды
type Photo struct {
	ID        int       `json:"id"`
	PointID   int       `json:"point_id"`
	UserID    int       `json:"user_id"`
	UserName  string    `json:"user_name"`
	Type      string    `json:"type"`     // "photo" или "certificate"
	FileName  string    `json:"file_name"`
	FilePath  string    `json:"-"`        // не отдаём путь клиенту
	FileURL   string    `json:"file_url"` // публичный URL
	CreatedAt time.Time `json:"created_at"`
}
