package handlers

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"
)

// startTime фиксирует время запуска сервера
var startTime = time.Now()

// HealthResponse — ответ на /health
type HealthResponse struct {
	Status  string `json:"status"`
	Uptime  string `json:"uptime"`
	Version string `json:"version"`
	OS      string `json:"os"`
}

// Health возвращает статус сервера
// GET /health
func Health(w http.ResponseWriter, r *http.Request) {
	uptime := time.Since(startTime).Round(time.Second)

	resp := HealthResponse{
		Status:  "ok",
		Uptime:  uptime.String(),
		Version: "1.0.0",
		OS:      runtime.GOOS,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
