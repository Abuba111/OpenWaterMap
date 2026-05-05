package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
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
// ДЛ-2 → pending, Админ/ДЛ-1 → approved
func (h *WaterHandler) CreatePoint(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req models.CreateWaterPointRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "некорректный JSON: "+err.Error())
		return
	}

	if errMsg := req.Validate(); errMsg != "" {
		respondError(w, http.StatusBadRequest, errMsg)
		return
	}

	// ДЛ-2 → pending, Админ/ДЛ-1 → approved сразу
	reviewStatus := models.ReviewApproved
	if claims != nil && claims.Role == models.RoleDL2 {
		reviewStatus = models.ReviewPending
	}

	userID := 0
	if claims != nil {
		userID = claims.UserID
	}

	point, err := h.db.CreatePointWithReview(&req, userID, reviewStatus)
	if err != nil {
		log.Printf("CreatePoint ошибка: %v", err)
		respondError(w, http.StatusInternalServerError, "ошибка сохранения")
		return
	}

	respondJSON(w, http.StatusCreated, point)
}

// GetPendingPoints — список точек на проверке (ДЛ-1 и Админ)
// GET /api/points/pending
func (h *WaterHandler) GetPendingPoints(w http.ResponseWriter, r *http.Request) {
	points, err := h.db.GetPendingPoints()
	if err != nil {
		log.Printf("GetPendingPoints ошибка: %v", err)
		respondError(w, http.StatusInternalServerError, "ошибка получения данных")
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{
		"data":  points,
		"count": len(points),
	})
}

// ReviewPoint — подтвердить или отклонить точку
// POST /api/points/{id}/review
// Body: { "action": "approve" | "reject", "reason": "..." }
func (h *WaterHandler) ReviewPoint(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "требуется авторизация")
		return
	}

	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		respondError(w, http.StatusBadRequest, "некорректный ID")
		return
	}

	var req models.ReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "некорректный JSON")
		return
	}

	if req.Action != "approve" && req.Action != "reject" {
		respondError(w, http.StatusBadRequest, "action должен быть approve или reject")
		return
	}

	if req.Action == "reject" && req.Reason == "" {
		respondError(w, http.StatusBadRequest, "укажите причину отклонения")
		return
	}

	point, err := h.db.ReviewPoint(id, claims.UserID, req.Action, req.Reason)
	if err != nil {
		log.Printf("ReviewPoint ошибка: %v", err)
		respondError(w, http.StatusInternalServerError, "ошибка обновления")
		return
	}
	if point == nil {
		respondError(w, http.StatusNotFound, "точка не найдена")
		return
	}

	action := "подтверждена"
	if req.Action == "reject" {
		action = "отклонена"
	}
	respondJSON(w, http.StatusOK, map[string]any{
		"message": "точка " + action,
		"point":   point,
	})
}

// DeletePoint удаляет точку воды
// DELETE /api/points/{id}
// Только Админ и ДЛ-1
func (h *WaterHandler) DeletePoint(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "требуется авторизация")
		return
	}

	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		respondError(w, http.StatusBadRequest, "некорректный ID")
		return
	}

	if err := h.db.DeletePoint(id); err != nil {
		log.Printf("DeletePoint ошибка: %v", err)
		respondError(w, http.StatusInternalServerError, "ошибка удаления")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "точка удалена"})
}

// UpdatePoint обновляет точку воды
// PUT /api/points/{id}
// Админ, ДЛ-1, ДЛ-2
func (h *WaterHandler) UpdatePoint(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "требуется авторизация")
		return
	}

	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		respondError(w, http.StatusBadRequest, "некорректный ID")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req models.CreateWaterPointRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "некорректный JSON")
		return
	}

	if errMsg := req.Validate(); errMsg != "" {
		respondError(w, http.StatusBadRequest, errMsg)
		return
	}

	point, err := h.db.UpdatePoint(id, &req)
	if err != nil {
		log.Printf("UpdatePoint ошибка: %v", err)
		respondError(w, http.StatusInternalServerError, "ошибка обновления")
		return
	}
	if point == nil {
		respondError(w, http.StatusNotFound, "точка не найдена")
		return
	}

	respondJSON(w, http.StatusOK, point)
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
