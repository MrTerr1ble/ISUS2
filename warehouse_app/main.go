package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

type Ore struct {
	ID        int     `json:"id"`
	Type      string  `json:"type"`
	Quantity  float64 `json:"quantity"`
	Location  string  `json:"location"`
	Quality   float64 `json:"quality"`
	Priority  string  `json:"priority"`
	CreatedAt string  `json:"created_at"`
}

type Tool struct {
	ID           int    `json:"id"`
	Type         string `json:"type"`
	Quantity     int    `json:"quantity"`
	SerialNumber string `json:"serial_number"`
	ServiceLife  int    `json:"service_life"`
	CreatedAt    string `json:"created_at"`
}

type Sale struct {
	ID        int     `json:"id"`
	OreType   string  `json:"ore_type"`
	Buyer     string  `json:"buyer"`
	Quantity  float64 `json:"quantity"`
	Status    string  `json:"status"`
	CreatedAt string  `json:"created_at"`
}

type Log struct {
	Date   string `json:"date"`
	User   string `json:"user"`
	Action string `json:"action"`
}

func main() {
	// Инициализация базы данных
	db, err := sql.Open("sqlite3", "./warehouse.db")
	if err != nil {
		log.Fatalf("can't open db %+v", err)
	}
	defer db.Close()

	// Создание таблиц
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS ores (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            type TEXT NOT NULL,
            quantity REAL NOT NULL,
            location TEXT,
            quality REAL,
            priority TEXT,
            created_at TEXT
        );
        CREATE TABLE IF NOT EXISTS tools (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            type TEXT NOT NULL,
            quantity INTEGER NOT NULL,
            serial_number TEXT,
            service_life INTEGER,
            created_at TEXT
        );
        CREATE TABLE IF NOT EXISTS sales (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            ore_type TEXT NOT NULL,
            buyer TEXT,
            quantity REAL NOT NULL,
            status TEXT,
            created_at TEXT
        );
    `)
	if err != nil {
		log.Fatal(err)
	}

	// Инициализация маршрутов
	router := mux.NewRouter()
	router.HandleFunc("/api/ores", getOres(db)).Methods("GET")
	router.HandleFunc("/api/ores", addOre(db)).Methods("POST")
	router.HandleFunc("/api/tools", getTools(db)).Methods("GET")
	router.HandleFunc("/api/tools", addTool(db)).Methods("POST")
	router.HandleFunc("/api/sales", getSales(db)).Methods("GET")
	router.HandleFunc("/api/sales", addSale(db)).Methods("POST")
	router.HandleFunc("/api/sales/{id}", updateSale(db)).Methods("PUT")
	router.HandleFunc("/api/logs", getLogs(db)).Methods("GET")

	// Статические файлы (фронтенд)
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./static")))

	// Запуск сервера
	fmt.Println("Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func getOres(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, type, quantity, location, quality, priority, created_at FROM ores")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var ores []Ore
		for rows.Next() {
			var o Ore
			err := rows.Scan(&o.ID, &o.Type, &o.Quantity, &o.Location, &o.Quality, &o.Priority, &o.CreatedAt)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			ores = append(ores, o)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ores)
	}
}

func addOre(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var ore Ore
		if err := json.NewDecoder(r.Body).Decode(&ore); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		ore.CreatedAt = time.Now().Format(time.RFC3339)
		_, err := db.Exec("INSERT INTO ores (type, quantity, location, quality, priority, created_at) VALUES (?, ?, ?, ?, ?, ?)",
			ore.Type, ore.Quantity, ore.Location, ore.Quality, ore.Priority, ore.CreatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Руда добавлена!"})
	}
}

func getTools(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, type, quantity, serial_number, service_life, created_at FROM tools")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var tools []Tool
		for rows.Next() {
			var t Tool
			err := rows.Scan(&t.ID, &t.Type, &t.Quantity, &t.SerialNumber, &t.ServiceLife, &t.CreatedAt)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			tools = append(tools, t)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tools)
	}
}

func addTool(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var tool Tool
		if err := json.NewDecoder(r.Body).Decode(&tool); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		tool.CreatedAt = time.Now().Format(time.RFC3339)
		_, err := db.Exec("INSERT INTO tools (type, quantity, serial_number, service_life, created_at) VALUES (?, ?, ?, ?, ?)",
			tool.Type, tool.Quantity, tool.SerialNumber, tool.ServiceLife, tool.CreatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Инструменты добавлены!"})
	}
}

func getSales(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, ore_type, buyer, quantity, status, created_at FROM sales")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var sales []Sale
		for rows.Next() {
			var s Sale
			err := rows.Scan(&s.ID, &s.OreType, &s.Buyer, &s.Quantity, &s.Status, &s.CreatedAt)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			sales = append(sales, s)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sales)
	}
}

func addSale(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var sale Sale
		if err := json.NewDecoder(r.Body).Decode(&sale); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		sale.CreatedAt = time.Now().Format(time.RFC3339)
		_, err := db.Exec("INSERT INTO sales (ore_type, buyer, quantity, status, created_at) VALUES (?, ?, ?, ?, ?)",
			sale.OreType, sale.Buyer, sale.Quantity, sale.Status, sale.CreatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Продажа добавлена!"})
	}
}

func updateSale(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		var sale Sale
		if err := json.NewDecoder(r.Body).Decode(&sale); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		_, err := db.Exec("UPDATE sales SET status = ? WHERE id = ?", sale.Status, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Статус обновлён!"})
	}
}

func getLogs(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Для простоты возвращаем статические логи, можно расширить
		logs := []Log{
			{Date: "2025-10-11", User: "Оператор1", Action: "Добавлено 100т железной руды"},
			{Date: "2025-10-10", User: "Менеджер1", Action: "Списано 100т железной руды"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(logs)
	}
}
