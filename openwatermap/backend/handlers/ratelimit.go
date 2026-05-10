package handlers

import (
	"net/http"
	"sync"
	"time"
)

// rateLimiter хранит счётчики запросов по IP
type rateLimiter struct {
	mu       sync.Mutex
	requests map[string]*clientInfo
	limit    int           // максимум запросов
	window   time.Duration // за какой период
}

type clientInfo struct {
	count    int
	resetAt  time.Time
}

// newRateLimiter создаёт новый лимитер
// limit — максимум запросов за window
func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		requests: make(map[string]*clientInfo),
		limit:    limit,
		window:   window,
	}

	// Очищать устаревшие записи каждую минуту
	go func() {
		for range time.Tick(time.Minute) {
			rl.cleanup()
		}
	}()

	return rl
}

// allow проверяет можно ли пропустить запрос с этого IP
func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	client, exists := rl.requests[ip]

	if !exists || now.After(client.resetAt) {
		// Новый клиент или окно сбросилось
		rl.requests[ip] = &clientInfo{
			count:   1,
			resetAt: now.Add(rl.window),
		}
		return true
	}

	if client.count >= rl.limit {
		return false // превышен лимит
	}

	client.count++
	return true
}

// cleanup удаляет устаревшие записи
func (rl *rateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	for ip, client := range rl.requests {
		if now.After(client.resetAt) {
			delete(rl.requests, ip)
		}
	}
}

// Глобальные лимитеры
var (
	// Общий лимит — 60 запросов в минуту для всех
	globalLimiter = newRateLimiter(60, time.Minute)

	// Строгий лимит для авторизации — 10 попыток в минуту
	authLimiter = newRateLimiter(10, time.Minute)
)

// RateLimit — middleware для общих запросов (60/мин)
func RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getIP(r)
		if !globalLimiter.allow(ip) {
			w.Header().Set("Retry-After", "60")
			respondError(w, http.StatusTooManyRequests,
				"слишком много запросов — подождите минуту")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RateLimitAuth — строгий лимит для логина/регистрации (10/мин)
func RateLimitAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getIP(r)
		if !authLimiter.allow(ip) {
			w.Header().Set("Retry-After", "60")
			respondError(w, http.StatusTooManyRequests,
				"слишком много попыток входа — подождите минуту")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// getIP достаёт реальный IP клиента
func getIP(r *http.Request) string {
	// Проверить заголовки прокси
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	return r.RemoteAddr
}
