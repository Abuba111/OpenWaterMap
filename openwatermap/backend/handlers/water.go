package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"openwatermap/database"
	"openwatermap/models"
)

// WaterHandler держит зависимости для всех хендлеров
type WaterHandler struct {
	db *database.DB
}

// NewWaterHandler создаёт хендлер с базой данных
func NewWaterHandler(db *database.DB) *WaterHandler {
	return &WaterHandler{db: db}
}

// respondJSON отправляет JSON ответ
func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Ошибка кодирования JSON: %v", err)
	}
}

// respondError отправляет ошибку в JSON формате
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

// GetPoints возвращает все точки воды с фильтрацией
// GET /api/points?status=good&limit=100&offset=0
func (h *WaterHandler) GetPoints(w http.ResponseWriter, r *http.Request) {
	// Параметры фильтрации из URL
	filter := models.FilterRequest{
		Status: r.URL.Query().Get("status"),
		Limit:  parseIntParam(r.URL.Query().Get("limit"), 500),
		Offset: parseIntParam(r.URL.Query().Get("offset"), 0),
	}

	// Проверить допустимость статуса
	if filter.Status != "" &&
		filter.Status != "good" &&
		filter.Status != "warning" &&
		filter.Status != "danger" {
		respondError(w, http.StatusBadRequest, "status должен быть: good, warning, danger или пустым")
		return
	}

	points, err := h.db.GetPoints(filter)
	if err != nil {
		log.Printf("GetPoints ошибка: %v", err)
		respondError(w, http.StatusInternalServerError, "ошибка получения данных")
		return
	}

	// Вернуть пустой массив вместо null
	if points == nil {
		points = []models.WaterPoint{}
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"data":   points,
		"count":  len(points),
		"filter": filter.Status,
	})
}

// GetPointByID возвращает одну точку по ID
// GET /api/points/{id}
func (h *WaterHandler) GetPointByID(w http.ResponseWriter, r *http.Request) {
	// Достать ID из URL пути
	idStr := extractID(r.URL.Path)
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		respondError(w, http.StatusBadRequest, "некорректный ID")
		return
	}

	point, err := h.db.GetPointByID(id)
	if err != nil {
		log.Printf("GetPointByID ошибка: %v", err)
		respondError(w, http.StatusInternalServerError, "ошибка получения данных")
		return
	}
	if point == nil {
		respondError(w, http.StatusNotFound, "точка не найдена")
		return
	}

	respondJSON(w, http.StatusOK, point)
}

// CreatePoint добавляет новую точку воды
// POST /api/points
// Body: { "name": "...", "lat": 43.2, "lng": 76.8, "ph": 7.2, ... }
func (h *WaterHandler) CreatePoint(w http.ResponseWriter, r *http.Request) {
	// Ограничить размер тела запроса (1MB)
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req models.CreateWaterPointRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "некорректный JSON: "+err.Error())
		return
	}

	// Валидация входных данных
	if errMsg := req.Validate(); errMsg != "" {
		respondError(w, http.StatusBadRequest, errMsg)
		return
	}

	point, err := h.db.CreatePoint(&req)
	if err != nil {
		log.Printf("CreatePoint ошибка: %v", err)
		respondError(w, http.StatusInternalServerError, "ошибка сохранения")
		return
	}

	respondJSON(w, http.StatusCreated, point)
}

// parseIntParam парсит числовой параметр из URL
func parseIntParam(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil || v < 0 {
		return defaultVal
	}
	return v
}

// extractID достаёт последний сегмент из URL пути
// /api/points/42 → "42"
func extractID(path string) string {
	parts := strings.Split(strings.TrimSuffix(path, "/"), "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}
