 # 🌊 OpenWaterMap — Качество воды Казахстана

Интерактивная карта качества воды. Стек: Go + SQLite + Leaflet + Docker.

---

## 🚀 Быстрый старт (одна команда)

```bash
docker compose up --build
```

Открой: **http://localhost:5173**

Остановить: `docker compose down`



# 🌊 OpenWaterMap — Kazakhstan Water Quality

An interactive map for monitoring and visualizing water quality across Kazakhstan. 
Built with a focus on simplicity, performance, and containerization.

**Stack:** Go + SQLite + Leaflet + Docker.

---

## 🚀 Quick Start (Single Command)

Get the project up and running in seconds using Docker:

```bash
docker compose up --build
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
openwatermap/
├── docker-compose.yml    # Orchestration
├── nginx.conf            # Frontend proxy & static serving
├── openwatermap.html      # Leaflet-based map (Frontend)
└── backend/
    ├── Dockerfile
    ├── go.mod
    ├── main.go           # Entry point
    ├── config/           # Configuration management
    ├── models/           # Data structures (Water quality)
    ├── database/         # SQLite initialization & Queries
    └── handlers/         # API Route logic (Water & Health)

---

## 🔌 API

| Метод | URL | Описание |
|-------|-----|----------|
| GET | /health | Статус сервера |
| GET | /api/points | Все точки |
| GET | /api/points?status=good | Фильтр |
| GET | /api/points/{id} | Одна точка |
| POST | /api/points | Добавить точку |


Method,Endpoint,Description
GET,/health,Server status check
GET,/api/points,Fetch all water quality points
GET,/api/points?status=good,Filter points by status
GET,/api/points/{id},Get detailed data for a specific point
POST,/api/points,Submit new water quality data
---

## 🎯 Статусы воды

| Статус | Цвет | Условие |
|--------|------|---------|
| good | 🟢 | pH 6.5–8.5, мутность < 3 NTU |
| warning | 🟡 | pH 6.0–6.5 или 8.5–9.0 |
| danger | 🔴 | pH < 6 или > 9, мутность > 10 NTU |

Status,Marker,Conditions
Good,🟢,"pH 6.5–8.5, Turbidity < 3 NTU"
Warning,🟡,pH 6.0–6.5 or 8.5–9.0
Danger,🔴,"pH < 6 or > 9, Turbidity > 10 NTU"

