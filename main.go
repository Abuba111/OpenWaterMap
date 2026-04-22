package main

import (
    "database/sql"
    "encoding/json"
    "net/http"
    "github.com/gorilla/mux"
    "github.com/rs/cors"
    _ "github.com/mattn/go-sqlite3"
)

type WaterPoint struct {
    ID          int     `json:"id"`
    Location    string  `json:"location"`
    Depth       float64 `json:"depth"`
}

var db *sql.DB

// Initialize database
func initDB() {
    var err error
    db, err = sql.Open("sqlite3", "./water_points.db")
    if err != nil {
        panic(err)
    }
    createTable()
    insertTestData()
}

func createTable() {
    createTableSQL := `CREATE TABLE IF NOT EXISTS water_points (
        "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
        "location" TEXT,
        "depth" REAL
    );`
    db.Exec(createTableSQL)
}

func insertTestData() {
    testPoints := []WaterPoint{
        {Location: "Almaty", Depth: 5.0},
        {Location: "Astana", Depth: 3.0},
        {Location: "Shymkent", Depth: 4.0},
        {Location: "Karaganda", Depth: 6.5},
        {Location: "Aktobe", Depth: 7.0},
        {Location: "Pavlodar", Depth: 4.5},
        {Location: "Ust-Kamenogorsk", Depth: 4.0},
        {Location: "Taraz", Depth: 4.2},
        {Location: "Semey", Depth: 5.3},
        {Location: "Atyrau", Depth: 3.7},
    }

    for _, point := range testPoints {
        db.Exec("INSERT INTO water_points (location, depth) VALUES (?, ?)", point.Location, point.Depth)
    }
}

// Health check endpoint
func healthCheck(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode("Healthy")
}

// Get water points
func getWaterPoints(w http.ResponseWriter, r *http.Request) {
    rows, err := db.Query("SELECT id, location, depth FROM water_points")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var points []WaterPoint
    for rows.Next() {
        var point WaterPoint
        if err := rows.Scan(&point.ID, &point.Location, &point.Depth); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        points = append(points, point)
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(points)
}

func main() {
    initDB()
    router := mux.NewRouter()
    router.Use(cors.New(cors.Options{
        AllowedOrigins:   []string{"*"},
        AllowCredentials: true,
    }).Handler)
    
    router.HandleFunc("/health", healthCheck).Methods("GET")
    router.HandleFunc("/waterpoints", getWaterPoints).Methods("GET")

    http.ListenAndServe(":8080", router)
}