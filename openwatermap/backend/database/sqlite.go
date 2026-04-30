package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite" // SQLite драйвер без CGO

	"openwatermap/models"
)

// DB обёртка над sql.DB
type DB struct {
	conn *sql.DB
}

// New открывает (или создаёт) SQLite базу данных
func New(dbPath string) (*DB, error) {
	// Создать директорию если не существует
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("не удалось создать директорию %s: %w", dir, err)
	}

	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть базу данных: %w", err)
	}

	// Проверить соединение
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("не удалось подключиться к базе данных: %w", err)
	}

	// Настройки SQLite для лучшей производительности
	pragmas := []string{
		"PRAGMA journal_mode=WAL",       // Лучше при одновременных запросах
		"PRAGMA synchronous=NORMAL",     // Баланс скорости и надёжности
		"PRAGMA foreign_keys=ON",        // Включить внешние ключи
		"PRAGMA cache_size=-64000",      // 64MB кэш
	}

	for _, pragma := range pragmas {
		if _, err := conn.Exec(pragma); err != nil {
			log.Printf("Предупреждение PRAGMA: %s - %v", pragma, err)
		}
	}

	db := &DB{conn: conn}

	// Запустить миграции
	if err := db.migrate(); err != nil {
		return nil, fmt.Errorf("ошибка миграции: %w", err)
	}

	return db, nil
}

// migrate создаёт таблицы если их нет
func (db *DB) migrate() error {
	query := `
	CREATE TABLE IF NOT EXISTS water_points (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		name        TEXT    NOT NULL,
		lat         REAL    NOT NULL,
		lng         REAL    NOT NULL,
		ph          REAL    NOT NULL,
		turbidity   REAL    NOT NULL DEFAULT 0,
		chlorine    REAL    NOT NULL DEFAULT 0,
		tds         REAL    NOT NULL DEFAULT 0,
		status      TEXT    NOT NULL CHECK(status IN ('good', 'warning', 'danger')),
		source      TEXT    NOT NULL DEFAULT 'Неизвестно',
		checked_at  TEXT    NOT NULL,
		created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	-- Индексы для быстрой фильтрации
	CREATE INDEX IF NOT EXISTS idx_water_points_status ON water_points(status);
	CREATE INDEX IF NOT EXISTS idx_water_points_lat    ON water_points(lat);
	CREATE INDEX IF NOT EXISTS idx_water_points_lng    ON water_points(lng);

	-- Таблица пользователей
	CREATE TABLE IF NOT EXISTS users (
		id            INTEGER PRIMARY KEY AUTOINCREMENT,
		name          TEXT    NOT NULL,
		email         TEXT    NOT NULL UNIQUE,
		password_hash TEXT    NOT NULL,
		role          TEXT    NOT NULL DEFAULT 'user'
		              CHECK(role IN ('admin', 'dl1', 'dl2', 'user')),
		is_active     INTEGER NOT NULL DEFAULT 1,
		created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

	-- Таблица сессий (JWT токены)
	CREATE TABLE IF NOT EXISTS sessions (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id    INTEGER NOT NULL REFERENCES users(id),
		token      TEXT    NOT NULL UNIQUE,
		expires_at DATETIME NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token);
	`

	if _, err := db.conn.Exec(query); err != nil {
		return err
	}

	// Добавить тестовые данные если таблица пустая
	if err := db.seedIfEmpty(); err != nil {
		return err
	}

	// Создать дефолтного админа если нет пользователей
	if err := db.seedAdminIfEmpty(); err != nil {
		return err
	}

	// Создать таблицы для фото и комментариев
	return db.migrateMedia()
}

// seedIfEmpty добавляет тестовые данные по Казахстану
func (db *DB) seedIfEmpty() error {
	var count int
	if err := db.conn.QueryRow("SELECT COUNT(*) FROM water_points").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil // данные уже есть
	}

	log.Println("База пустая — добавляю тестовые данные...")

	seedData := []models.CreateWaterPointRequest{
		{Name: "Река Или — г. Балхаш",          Lat: 46.8477, Lng: 74.9854, Ph: 7.8, Turbidity: 2.1,  Chlorine: 0.20, TDS: 320,  Source: "КазГидромет",        CheckedAt: "2025-04-10"},
		{Name: "Водохранилище Капшагай",          Lat: 43.8631, Lng: 77.0768, Ph: 7.2, Turbidity: 0.8,  Chlorine: 0.15, TDS: 210,  Source: "МЭ РК",              CheckedAt: "2025-04-08"},
		{Name: "Река Сырдарья — Кызылорда",      Lat: 44.8479, Lng: 65.5093, Ph: 8.9, Turbidity: 12.4, Chlorine: 0.05, TDS: 890,  Source: "Лаборатория ЖКХ",    CheckedAt: "2025-03-22"},
		{Name: "Арал — Аральск",                 Lat: 46.7897, Lng: 61.6638, Ph: 9.4, Turbidity: 45.2, Chlorine: 0.00, TDS: 3200, Source: "ПРООН / Арал",        CheckedAt: "2025-02-15"},
		{Name: "Иртыш — Усть-Каменогорск",       Lat: 49.9597, Lng: 82.6155, Ph: 7.1, Turbidity: 1.2,  Chlorine: 0.25, TDS: 190,  Source: "Горводоканал",        CheckedAt: "2025-04-12"},
		{Name: "Тобол — Костанай",               Lat: 53.1956, Lng: 63.6198, Ph: 6.3, Turbidity: 3.8,  Chlorine: 0.10, TDS: 420,  Source: "Горводоканал",        CheckedAt: "2025-04-05"},
		{Name: "Ишим — Астана",                  Lat: 51.1801, Lng: 71.4460, Ph: 7.5, Turbidity: 0.6,  Chlorine: 0.30, TDS: 280,  Source: "СЭС г.Астана",        CheckedAt: "2025-04-14"},
		{Name: "Река Урал — Атырау",             Lat: 47.1151, Lng: 51.8891, Ph: 8.6, Turbidity: 5.7,  Chlorine: 0.08, TDS: 650,  Source: "Нефтехим контроль",   CheckedAt: "2025-03-30"},
		{Name: "Каратал — Талдыкорган",          Lat: 45.0165, Lng: 78.3762, Ph: 7.4, Turbidity: 1.5,  Chlorine: 0.20, TDS: 230,  Source: "Областная лаборатория",CheckedAt: "2025-04-11"},
		{Name: "Зайсан — ВКО",                  Lat: 47.4789, Lng: 84.2654, Ph: 6.8, Turbidity: 0.9,  Chlorine: 0.18, TDS: 175,  Source: "КазГидромет",         CheckedAt: "2025-04-09"},
		{Name: "Шу — Жамбыл",                   Lat: 42.9000, Lng: 71.3667, Ph: 8.3, Turbidity: 4.2,  Chlorine: 0.12, TDS: 560,  Source: "ЖКХ Тараз",           CheckedAt: "2025-03-28"},
		{Name: "Нура — Карагандинская обл.",     Lat: 50.1692, Lng: 73.1375, Ph: 9.1, Turbidity: 18.3, Chlorine: 0.03, TDS: 1100, Source: "АО АрселорМиттал",    CheckedAt: "2025-03-15"},
	}

	for _, point := range seedData {
		if _, err := db.CreatePoint(&point); err != nil {
			log.Printf("Ошибка seed: %v", err)
		}
	}

	log.Printf("Добавлено %d тестовых точек", len(seedData))
	return nil
}

// GetPoints возвращает список точек с фильтрацией и пагинацией
func (db *DB) GetPoints(filter models.FilterRequest) ([]models.WaterPoint, error) {
	// Установить лимит по умолчанию
	if filter.Limit <= 0 || filter.Limit > 500 {
		filter.Limit = 500
	}

	query := `
		SELECT id, name, lat, lng, ph, turbidity, chlorine, tds,
		       status, source, checked_at, created_at
		FROM water_points
		WHERE (? = '' OR status = ?)
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := db.conn.Query(query, filter.Status, filter.Status, filter.Limit, filter.Offset)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса: %w", err)
	}
	defer rows.Close()

	var points []models.WaterPoint
	for rows.Next() {
		var p models.WaterPoint
		var createdAt string
		err := rows.Scan(
			&p.ID, &p.Name, &p.Lat, &p.Lng,
			&p.Ph, &p.Turbidity, &p.Chlorine, &p.TDS,
			&p.Status, &p.Source, &p.CheckedAt, &createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка чтения строки: %w", err)
		}
		p.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		points = append(points, p)
	}

	return points, rows.Err()
}

// GetPointByID возвращает одну точку по ID
func (db *DB) GetPointByID(id int) (*models.WaterPoint, error) {
	query := `
		SELECT id, name, lat, lng, ph, turbidity, chlorine, tds,
		       status, source, checked_at, created_at
		FROM water_points WHERE id = ?
	`

	var p models.WaterPoint
	var createdAt string
	err := db.conn.QueryRow(query, id).Scan(
		&p.ID, &p.Name, &p.Lat, &p.Lng,
		&p.Ph, &p.Turbidity, &p.Chlorine, &p.TDS,
		&p.Status, &p.Source, &p.CheckedAt, &createdAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil // не найдено
	}
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса по ID: %w", err)
	}
	p.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	return &p, nil
}

// CreatePoint добавляет новую точку воды
func (db *DB) CreatePoint(req *models.CreateWaterPointRequest) (*models.WaterPoint, error) {
	// Автоматически вычислить статус
	status := models.CalcStatus(req.Ph, req.Turbidity)

	// Дата проверки по умолчанию — сегодня
	checkedAt := req.CheckedAt
	if checkedAt == "" {
		checkedAt = time.Now().Format("2006-01-02")
	}

	source := req.Source
	if source == "" {
		source = "Пользователь"
	}

	query := `
		INSERT INTO water_points (name, lat, lng, ph, turbidity, chlorine, tds, status, source, checked_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := db.conn.Exec(query,
		req.Name, req.Lat, req.Lng,
		req.Ph, req.Turbidity, req.Chlorine, req.TDS,
		string(status), source, checkedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка вставки: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("ошибка получения ID: %w", err)
	}

	return db.GetPointByID(int(id))
}

// Close закрывает соединение с базой данных
func (db *DB) Close() error {
	return db.conn.Close()
}
