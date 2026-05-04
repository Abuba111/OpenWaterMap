package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"openwatermap/config"
	"openwatermap/database"
	"openwatermap/handlers"
	"openwatermap/models"
)

func main() {
	// Загрузить конфигурацию
	cfg := config.Load()

	// Проверить что DATABASE_URL задан
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL не задан! Добавь переменную окружения.")
	}

	// Подключиться к PostgreSQL
	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Ошибка базы данных: %v", err)
	}
	defer db.Close()

	log.Println("База данных PostgreSQL подключена ✓")

	// Создать хендлеры
	waterHandler := handlers.NewWaterHandler(db)
	authHandler  := handlers.NewAuthHandler(db)
	mediaHandler := handlers.NewMediaHandler(db, "./uploads")

	// Настроить роутер
	r := chi.NewRouter()

	// Middleware — логи, восстановление после паники, таймаут
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(middleware.RealIP)

	// CORS — разрешить запросы с фронтенда
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Accept", "Authorization"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Роуты
	r.Get("/health", handlers.Health)

	r.Route("/api", func(r chi.Router) {

		// Авторизация — публичные роуты
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register) // регистрация
			r.Post("/login",    authHandler.Login)    // вход

			// Защищённые роуты авторизации
			r.Group(func(r chi.Router) {
				r.Use(handlers.AuthMiddleware)
				r.Get("/me", authHandler.Me) // мой профиль

				// Только админ
				r.Group(func(r chi.Router) {
					r.Use(handlers.RequireRole(models.RoleAdmin))
					r.Get("/users",               authHandler.GetUsers)   // все пользователи
					r.Put("/users/{id}/role",     authHandler.UpdateRole) // сменить роль
				})
			})
		})

		// Точки воды
		r.Route("/points", func(r chi.Router) {
			r.Get("/",     waterHandler.GetPoints)
			r.Get("/{id}", waterHandler.GetPointByID)

			// Публичное чтение комментариев и фото
			r.Get("/{id}/comments", mediaHandler.GetComments)
			r.Get("/{id}/photos",   mediaHandler.GetPhotos)

			// Добавить точку — только ДЛ-1, ДЛ-2, Админ
			r.Group(func(r chi.Router) {
				r.Use(handlers.AuthMiddleware)
				r.Use(handlers.RequireRole(
					models.RoleAdmin,
					models.RoleDL1,
					models.RoleDL2,
				))
				r.Post("/", waterHandler.CreatePoint)
				r.Post("/{id}/photos", mediaHandler.UploadPhoto)

				// Редактировать — Админ, ДЛ-1, ДЛ-2
				r.Put("/{id}", waterHandler.UpdatePoint)
			})

			// Удалить точку — только Админ и ДЛ-1
			r.Group(func(r chi.Router) {
				r.Use(handlers.AuthMiddleware)
				r.Use(handlers.RequireRole(models.RoleAdmin, models.RoleDL1))
				r.Delete("/{id}", waterHandler.DeletePoint)
			})

			// Комментарии — любой авторизованный
			r.Group(func(r chi.Router) {
				r.Use(handlers.AuthMiddleware)
				r.Post("/{id}/comments", mediaHandler.CreateComment)
			})
		})

		// Удаление комментария
		r.Group(func(r chi.Router) {
			r.Use(handlers.AuthMiddleware)
			r.Delete("/comments/{id}", mediaHandler.DeleteComment)
		})
	})

	// Отдать загруженные файлы
	r.Get("/uploads/{type}/{filename}", mediaHandler.ServeFile)

	// Запустить сервер
	server := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown — корректное завершение при Ctrl+C
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		fmt.Printf("\n🌊 OpenWaterMap запущен на http://localhost:%s\n", cfg.ServerPort)
		fmt.Printf("📊 API: http://localhost:%s/api/points\n", cfg.ServerPort)
		fmt.Printf("❤️  Health: http://localhost:%s/health\n\n", cfg.ServerPort)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка сервера: %v", err)
		}
	}()

	// Ждать сигнала завершения
	<-quit
	log.Println("Завершаю работу...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Ошибка завершения: %v", err)
	}

	log.Println("Сервер остановлен")
}
