 # 🌊 OpenWaterMap — Качество воды Казахстана

Интерактивная карта качества воды. Стек: Go + SQLite + Leaflet + Docker.

---

## 🚀 Быстрый старт (одна команда)

```bash
docker compose up --build
```

Открой: **http://localhost:5173**

Остановить: `docker compose down`

---

## 📁 Структура

```
openwatermap/
├── docker-compose.yml    # запуск одной командой
├── nginx.conf            # прокси фронтенда
├── openwatermap.html     # карта (фронтенд)
└── backend/
    ├── Dockerfile
    ├── go.mod
    ├── main.go
    ├── config/config.go
    ├── models/water.go
    ├── database/sqlite.go
    └── handlers/
        ├── water.go
        └── health.go
```

---

## 🔌 API

| Метод | URL | Описание |
|-------|-----|----------|
| GET | /health | Статус сервера |
| GET | /api/points | Все точки |
| GET | /api/points?status=good | Фильтр |
| GET | /api/points/{id} | Одна точка |
| POST | /api/points | Добавить точку |

---

## 🎯 Статусы воды

| Статус | Цвет | Условие |
|--------|------|---------|
| good | 🟢 | pH 6.5–8.5, мутность < 3 NTU |
| warning | 🟡 | pH 6.0–6.5 или 8.5–9.0 |
| danger | 🔴 | pH < 6 или > 9, мутность > 10 NTU |


# 🌊 OpenWaterMap — Kazakhstan Water Quality

An interactive map for monitoring and visualizing water quality across Kazakhstan. This project provides a lightweight, containerized solution for environmental data visualization.

**Tech Stack:** Go (Backend) + SQLite (Database) + Leaflet (Map) + Docker (Deployment).

---

## 🚀 Quick Start

Launch the entire infrastructure with a single command:

```bash
docker compose up --build
