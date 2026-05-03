package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"

	"openwatermap/models"
)

// DB обёртка над sql.DB
type DB struct {
	conn *sql.DB
}

// New открывает подключение к PostgreSQL
func New(dbURL string) (*DB, error) {
	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть базу данных: %w", err)
	}

	conn.SetMaxOpenConns(10)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(time.Hour)

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("не удалось подключиться: %w", err)
	}

	db := &DB{conn: conn}

	if err := db.migrate(); err != nil {
		return nil, fmt.Errorf("ошибка миграции: %w", err)
	}

	return db, nil
}

// migrate создаёт все таблицы
func (db *DB) migrate() error {
	query := `
	CREATE TABLE IF NOT EXISTS water_points (
		id         SERIAL PRIMARY KEY,
		name       TEXT NOT NULL,
		lat        REAL NOT NULL,
		lng        REAL NOT NULL,
		ph         REAL NOT NULL,
		turbidity  REAL NOT NULL DEFAULT 0,
		chlorine   REAL NOT NULL DEFAULT 0,
		tds        REAL NOT NULL DEFAULT 0,
		status     TEXT NOT NULL CHECK(status IN ('good','warning','danger')),
		source     TEXT NOT NULL DEFAULT 'Неизвестно',
		checked_at TEXT NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_wp_status ON water_points(status);
	CREATE INDEX IF NOT EXISTS idx_wp_lat    ON water_points(lat);
	CREATE INDEX IF NOT EXISTS idx_wp_lng    ON water_points(lng);

	CREATE TABLE IF NOT EXISTS users (
		id            SERIAL PRIMARY KEY,
		name          TEXT    NOT NULL,
		email         TEXT    NOT NULL UNIQUE,
		password_hash TEXT    NOT NULL,
		role          TEXT    NOT NULL DEFAULT 'user'
		              CHECK(role IN ('admin','dl1','dl2','user')),
		is_active     BOOLEAN NOT NULL DEFAULT TRUE,
		created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

	CREATE TABLE IF NOT EXISTS comments (
		id         SERIAL PRIMARY KEY,
		point_id   INTEGER NOT NULL REFERENCES water_points(id) ON DELETE CASCADE,
		user_id    INTEGER NOT NULL REFERENCES users(id),
		text       TEXT    NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_comments_point ON comments(point_id);

	CREATE TABLE IF NOT EXISTS photos (
		id         SERIAL PRIMARY KEY,
		point_id   INTEGER NOT NULL REFERENCES water_points(id) ON DELETE CASCADE,
		user_id    INTEGER NOT NULL REFERENCES users(id),
		type       TEXT NOT NULL DEFAULT 'photo' CHECK(type IN ('photo','certificate')),
		file_name  TEXT NOT NULL,
		file_path  TEXT NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_photos_point ON photos(point_id);
	`

	if _, err := db.conn.Exec(query); err != nil {
		return err
	}

	if err := db.seedIfEmpty(); err != nil {
		return err
	}
	return db.seedAdminIfEmpty()
}

// migrateMedia — совместимость (таблицы уже созданы в migrate)
func (db *DB) migrateMedia() error { return nil }

// seedIfEmpty добавляет тестовые данные
func (db *DB) seedIfEmpty() error {
	var count int
	if err := db.conn.QueryRow("SELECT COUNT(*) FROM water_points").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	log.Println("База пустая — добавляю тестовые данные...")

	points := []models.CreateWaterPointRequest{
		{Name: "Река Или — г. Балхаш",      Lat: 46.8477, Lng: 74.9854, Ph: 7.8, Turbidity: 2.1,  Chlorine: 0.20, TDS: 320,  Source: "КазГидромет",          CheckedAt: "2025-04-10"},
		{Name: "Водохранилище Капшагай",      Lat: 43.8631, Lng: 77.0768, Ph: 7.2, Turbidity: 0.8,  Chlorine: 0.15, TDS: 210,  Source: "МЭ РК",                CheckedAt: "2025-04-08"},
		{Name: "Река Сырдарья — Кызылорда",  Lat: 44.8479, Lng: 65.5093, Ph: 8.9, Turbidity: 12.4, Chlorine: 0.05, TDS: 890,  Source: "Лаборатория ЖКХ",      CheckedAt: "2025-03-22"},
		{Name: "Арал — Аральск",             Lat: 46.7897, Lng: 61.6638, Ph: 9.4, Turbidity: 45.2, Chlorine: 0.00, TDS: 3200, Source: "ПРООН / Арал",          CheckedAt: "2025-02-15"},
		{Name: "Иртыш — Усть-Каменогорск",   Lat: 49.9597, Lng: 82.6155, Ph: 7.1, Turbidity: 1.2,  Chlorine: 0.25, TDS: 190,  Source: "Горводоканал",          CheckedAt: "2025-04-12"},
		{Name: "Тобол — Костанай",           Lat: 53.1956, Lng: 63.6198, Ph: 6.3, Turbidity: 3.8,  Chlorine: 0.10, TDS: 420,  Source: "Горводоканал",          CheckedAt: "2025-04-05"},
		{Name: "Ишим — Астана",              Lat: 51.1801, Lng: 71.4460, Ph: 7.5, Turbidity: 0.6,  Chlorine: 0.30, TDS: 280,  Source: "СЭС г.Астана",          CheckedAt: "2025-04-14"},
		{Name: "Река Урал — Атырау",         Lat: 47.1151, Lng: 51.8891, Ph: 8.6, Turbidity: 5.7,  Chlorine: 0.08, TDS: 650,  Source: "Нефтехим контроль",     CheckedAt: "2025-03-30"},
		{Name: "Каратал — Талдыкорган",      Lat: 45.0165, Lng: 78.3762, Ph: 7.4, Turbidity: 1.5,  Chlorine: 0.20, TDS: 230,  Source: "Областная лаборатория", CheckedAt: "2025-04-11"},
		{Name: "Зайсан — ВКО",              Lat: 47.4789, Lng: 84.2654, Ph: 6.8, Turbidity: 0.9,  Chlorine: 0.18, TDS: 175,  Source: "КазГидромет",           CheckedAt: "2025-04-09"},
		{Name: "Шу — Жамбыл",               Lat: 42.9000, Lng: 71.3667, Ph: 8.3, Turbidity: 4.2,  Chlorine: 0.12, TDS: 560,  Source: "ЖКХ Тараз",             CheckedAt: "2025-03-28"},
		{Name: "Нура — Карагандинская обл.", Lat: 50.1692, Lng: 73.1375, Ph: 9.1, Turbidity: 18.3, Chlorine: 0.03, TDS: 1100, Source: "АО АрселорМиттал",      CheckedAt: "2025-03-15"},
	}

	for _, p := range points {
		if _, err := db.CreatePoint(&p); err != nil {
			log.Printf("Ошибка seed: %v", err)
		}
	}
	log.Printf("Добавлено %d точек", len(points))
	return nil
}

// GetPoints возвращает точки с фильтрацией
func (db *DB) GetPoints(filter models.FilterRequest) ([]models.WaterPoint, error) {
	if filter.Limit <= 0 || filter.Limit > 500 {
		filter.Limit = 500
	}
	rows, err := db.conn.Query(`
		SELECT id, name, lat, lng, ph, turbidity, chlorine, tds,
		       status, source, checked_at, created_at
		FROM water_points
		WHERE ($1 = '' OR status = $1)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, filter.Status, filter.Limit, filter.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []models.WaterPoint
	for rows.Next() {
		var p models.WaterPoint
		if err := rows.Scan(&p.ID, &p.Name, &p.Lat, &p.Lng,
			&p.Ph, &p.Turbidity, &p.Chlorine, &p.TDS,
			&p.Status, &p.Source, &p.CheckedAt, &p.CreatedAt); err != nil {
			return nil, err
		}
		points = append(points, p)
	}
	if points == nil {
		points = []models.WaterPoint{}
	}
	return points, rows.Err()
}

// GetPointByID возвращает одну точку
func (db *DB) GetPointByID(id int) (*models.WaterPoint, error) {
	var p models.WaterPoint
	err := db.conn.QueryRow(`
		SELECT id, name, lat, lng, ph, turbidity, chlorine, tds,
		       status, source, checked_at, created_at
		FROM water_points WHERE id = $1
	`, id).Scan(&p.ID, &p.Name, &p.Lat, &p.Lng,
		&p.Ph, &p.Turbidity, &p.Chlorine, &p.TDS,
		&p.Status, &p.Source, &p.CheckedAt, &p.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// CreatePoint добавляет точку
func (db *DB) CreatePoint(req *models.CreateWaterPointRequest) (*models.WaterPoint, error) {
	status    := models.CalcStatus(req.Ph, req.Turbidity)
	checkedAt := req.CheckedAt
	if checkedAt == "" {
		checkedAt = time.Now().Format("2006-01-02")
	}
	source := req.Source
	if source == "" {
		source = "Пользователь"
	}

	var id int
	err := db.conn.QueryRow(`
		INSERT INTO water_points (name, lat, lng, ph, turbidity, chlorine, tds, status, source, checked_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id
	`, req.Name, req.Lat, req.Lng, req.Ph, req.Turbidity, req.Chlorine, req.TDS,
		string(status), source, checkedAt,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return db.GetPointByID(id)
}

// Close закрывает соединение
func (db *DB) Close() error {
	return db.conn.Close()
}
