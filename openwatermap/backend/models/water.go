package models

import "time"

// WaterStatus — статус качества воды
type WaterStatus string

const (
	StatusGood    WaterStatus = "good"
	StatusWarning WaterStatus = "warning"
	StatusDanger  WaterStatus = "danger"
)

// ReviewStatus — статус верификации точки
type ReviewStatus string

const (
	ReviewPending  ReviewStatus = "pending"  // Ожидает проверки ДЛ-1
	ReviewApproved ReviewStatus = "approved" // Подтверждено — видно на карте
	ReviewRejected ReviewStatus = "rejected" // Отклонено
)

// WaterPoint — одна точка воды на карте
type WaterPoint struct {
	ID           int          `json:"id"`
	Name         string       `json:"name"`
	Lat          float64      `json:"lat"`
	Lng          float64      `json:"lng"`
	Ph           float64      `json:"ph"`
	Turbidity    float64      `json:"turbidity"`
	Chlorine     float64      `json:"chlorine"`
	TDS          float64      `json:"tds"`
	Status       WaterStatus  `json:"status"`
	ReviewStatus ReviewStatus `json:"review_status"`
	SubmittedBy  int          `json:"submitted_by"`
	ReviewedBy   *int         `json:"reviewed_by,omitempty"`
	RejectReason string       `json:"reject_reason,omitempty"`
	Source       string       `json:"source"`
	CheckedAt    string       `json:"checked_at"`
	CreatedAt    time.Time    `json:"created_at"`
}

// CreateWaterPointRequest — входные данные при создании точки
type CreateWaterPointRequest struct {
	Name      string  `json:"name"`
	Lat       float64 `json:"lat"`
	Lng       float64 `json:"lng"`
	Ph        float64 `json:"ph"`
	Turbidity float64 `json:"turbidity"`
	Chlorine  float64 `json:"chlorine"`
	TDS       float64 `json:"tds"`
	Source    string  `json:"source"`
	CheckedAt string  `json:"checked_at"`
}

// ReviewRequest — запрос на верификацию
type ReviewRequest struct {
	Action string `json:"action"` // "approve" или "reject"
	Reason string `json:"reason"` // причина отклонения
}

// FilterRequest — параметры фильтрации
type FilterRequest struct {
	Status       string `json:"status"`
	ReviewStatus string `json:"review_status"`
	Limit        int    `json:"limit"`
	Offset       int    `json:"offset"`
}

// CalcStatus автоматически определяет статус по показателям воды
func CalcStatus(ph, turbidity float64) WaterStatus {
	if ph < 6.0 || ph > 9.0 || turbidity > 10.0 {
		return StatusDanger
	}
	if ph < 6.5 || ph > 8.5 || turbidity > 3.0 {
		return StatusWarning
	}
	return StatusGood
}

// Validate проверяет корректность входных данных
func (r *CreateWaterPointRequest) Validate() string {
	if r.Name == "" {
		return "name обязателен"
	}
	if len(r.Name) > 200 {
		return "name слишком длинный (макс 200 символов)"
	}
	if r.Lat < 35.0 || r.Lat > 56.0 {
		return "lat вне диапазона Центральной Азии (35–56)"
	}
	if r.Lng < 46.0 || r.Lng > 90.0 {
		return "lng вне диапазона Центральной Азии (46–90)"
	}
	if r.Ph < 0 || r.Ph > 14 {
		return "ph должен быть от 0 до 14"
	}
	if r.Turbidity < 0 {
		return "turbidity не может быть отрицательным"
	}
	if r.Chlorine < 0 {
		return "chlorine не может быть отрицательным"
	}
	if r.TDS < 0 {
		return "tds не может быть отрицательным"
	}
	return ""
}
