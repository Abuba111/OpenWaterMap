package models

import "time"

// WaterStatus — статус качества воды
type WaterStatus string

const (
	StatusGood    WaterStatus = "good"    // Чистая
	StatusWarning WaterStatus = "warning" // Внимание
	StatusDanger  WaterStatus = "danger"  // Опасная
)

// WaterPoint — одна точка воды на карте
type WaterPoint struct {
	ID         int         `json:"id"`
	Name       string      `json:"name"`
	Lat        float64     `json:"lat"`
	Lng        float64     `json:"lng"`
	Ph         float64     `json:"ph"`
	Turbidity  float64     `json:"turbidity"`  // Мутность (NTU)
	Chlorine   float64     `json:"chlorine"`   // Хлор (мг/л)
	TDS        float64     `json:"tds"`        // Растворённые вещества (мг/л)
	Status     WaterStatus `json:"status"`
	Source     string      `json:"source"`     // Источник данных
	CheckedAt  string      `json:"checked_at"` // Дата проверки
	CreatedAt  time.Time   `json:"created_at"`
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

// FilterRequest — параметры фильтрации
type FilterRequest struct {
	Status string `json:"status"` // good | warning | danger | "" (все)
	Limit  int    `json:"limit"`  // максимум точек (по умолчанию 500)
	Offset int    `json:"offset"` // пагинация
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
	if r.Lat < 40.0 || r.Lat > 56.0 {
		return "lat вне диапазона Казахстана (40–56)"
	}
	if r.Lng < 50.0 || r.Lng > 88.0 {
		return "lng вне диапазона Казахстана (50–88)"
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
	return "" // всё ок
}
