package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"openwatermap/database"
	"openwatermap/models"
)

// MediaHandler хендлер для комментариев и фото
type MediaHandler struct {
	db        *database.DB
	uploadDir string
}

// NewMediaHandler создаёт хендлер и папки для файлов
func NewMediaHandler(db *database.DB, uploadDir string) *MediaHandler {
	os.MkdirAll(filepath.Join(uploadDir, "photos"),       0755)
	os.MkdirAll(filepath.Join(uploadDir, "certificates"), 0755)
	return &MediaHandler{db: db, uploadDir: uploadDir}
}

// GetComments — GET /api/points/{id}/comments
func (h *MediaHandler) GetComments(w http.ResponseWriter, r *http.Request) {
	pointID, err := parseID(chi.URLParam(r, "id"))
	if err != nil || pointID <= 0 {
		respondError(w, http.StatusBadRequest, "некорректный ID точки")
		return
	}
	comments, err := h.db.GetComments(pointID)
	if err != nil {
		log.Printf("GetComments: %v", err)
		respondError(w, http.StatusInternalServerError, "ошибка получения комментариев")
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"data": comments, "count": len(comments)})
}

// CreateComment — POST /api/points/{id}/comments
func (h *MediaHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "требуется авторизация")
		return
	}
	pointID, err := parseID(chi.URLParam(r, "id"))
	if err != nil || pointID <= 0 {
		respondError(w, http.StatusBadRequest, "некорректный ID точки")
		return
	}
	var req models.CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "некорректный JSON")
		return
	}
	if errMsg := req.Validate(); errMsg != "" {
		respondError(w, http.StatusBadRequest, errMsg)
		return
	}
	comment, err := h.db.CreateComment(pointID, claims.UserID, req.Text)
	if err != nil {
		log.Printf("CreateComment: %v", err)
		respondError(w, http.StatusInternalServerError, "ошибка создания комментария")
		return
	}
	respondJSON(w, http.StatusCreated, comment)
}

// DeleteComment — DELETE /api/comments/{id}
func (h *MediaHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	commentID, err := parseID(chi.URLParam(r, "id"))
	if err != nil || commentID <= 0 {
		respondError(w, http.StatusBadRequest, "некорректный ID")
		return
	}
	isAdmin := claims.Role == models.RoleAdmin
	if err := h.db.DeleteComment(commentID, claims.UserID, isAdmin); err != nil {
		respondError(w, http.StatusForbidden, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"message": "комментарий удалён"})
}

// GetPhotos — GET /api/points/{id}/photos
func (h *MediaHandler) GetPhotos(w http.ResponseWriter, r *http.Request) {
	pointID, err := parseID(chi.URLParam(r, "id"))
	if err != nil || pointID <= 0 {
		respondError(w, http.StatusBadRequest, "некорректный ID точки")
		return
	}
	photos, err := h.db.GetPhotos(pointID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "ошибка получения фото")
		return
	}
	for i := range photos {
		photos[i].FileURL = fmt.Sprintf("/uploads/%ss/%s", photos[i].Type, photos[i].FileName)
	}
	respondJSON(w, http.StatusOK, map[string]any{"data": photos, "count": len(photos)})
}

// UploadPhoto — POST /api/points/{id}/photos
func (h *MediaHandler) UploadPhoto(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "требуется авторизация")
		return
	}
	pointID, err := parseID(chi.URLParam(r, "id"))
	if err != nil || pointID <= 0 {
		respondError(w, http.StatusBadRequest, "некорректный ID точки")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		respondError(w, http.StatusBadRequest, "файл слишком большой (макс 10MB)")
		return
	}

	photoType := r.FormValue("type")
	if photoType != "photo" && photoType != "certificate" {
		photoType = "photo"
	}

	// Сертификаты только для ДЛ и выше
	if photoType == "certificate" {
		allowed := claims.Role == models.RoleAdmin ||
			claims.Role == models.RoleDL1 ||
			claims.Role == models.RoleDL2
		if !allowed {
			respondError(w, http.StatusForbidden, "только ДЛ-2 и выше могут загружать сертификаты")
			return
		}
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "файл не найден")
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	validExt := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".pdf": true}
	if !validExt[ext] {
		respondError(w, http.StatusBadRequest, "допустимые форматы: JPG, PNG, PDF")
		return
	}

	fileName := fmt.Sprintf("%d_%d%s", pointID, time.Now().UnixNano(), ext)
	subDir   := photoType + "s"
	filePath := filepath.Join(h.uploadDir, subDir, fileName)

	dst, err := os.Create(filePath)
	if err != nil {
		log.Printf("UploadPhoto os.Create: %v", err)
		respondError(w, http.StatusInternalServerError, "ошибка сохранения файла")
		return
	}
	defer dst.Close()
	io.Copy(dst, file)

	photo, err := h.db.SavePhoto(pointID, claims.UserID, photoType, fileName, filePath)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "ошибка сохранения в базу")
		return
	}
	photo.FileURL = fmt.Sprintf("/uploads/%s/%s", subDir, fileName)
	respondJSON(w, http.StatusCreated, photo)
}

// ServeFile — GET /uploads/{type}/{filename}
func (h *MediaHandler) ServeFile(w http.ResponseWriter, r *http.Request) {
	photoType := chi.URLParam(r, "type")
	fileName  := chi.URLParam(r, "filename")
	if strings.Contains(fileName, "..") || strings.Contains(photoType, "..") {
		respondError(w, http.StatusBadRequest, "недопустимый путь")
		return
	}
	http.ServeFile(w, r, filepath.Join(h.uploadDir, photoType, fileName))
}
