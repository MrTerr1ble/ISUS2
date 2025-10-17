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

type Unit struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
}

type Warehouse struct {
	ID         int     `json:"id"`
	Name       string  `json:"name"`
	Location   string  `json:"location"`
	Capacity   float64 `json:"capacity"`
	Supervisor string  `json:"supervisor"`
}

type OreType struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description"`
}

type OreBatch struct {
	ID             int     `json:"id"`
	BatchCode      string  `json:"batch_code"`
	OreTypeID      int     `json:"ore_type_id"`
	OreTypeName    string  `json:"ore_type_name"`
	WarehouseID    int     `json:"warehouse_id"`
	WarehouseName  string  `json:"warehouse_name"`
	UnitID         int     `json:"unit_id"`
	UnitName       string  `json:"unit_name"`
	UnitSymbol     string  `json:"unit_symbol"`
	Quantity       float64 `json:"quantity"`
	Quality        float64 `json:"quality"`
	Priority       string  `json:"priority"`
	ExtractionDate string  `json:"extraction_date"`
	Status         string  `json:"status"`
	CreatedAt      string  `json:"created_at"`
}

type EquipmentCategory struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Equipment struct {
	ID            int     `json:"id"`
	Name          string  `json:"name"`
	CategoryID    int     `json:"category_id"`
	CategoryName  string  `json:"category_name"`
	WarehouseID   int     `json:"warehouse_id"`
	WarehouseName string  `json:"warehouse_name"`
	UnitID        int     `json:"unit_id"`
	UnitName      string  `json:"unit_name"`
	UnitSymbol    string  `json:"unit_symbol"`
	Quantity      float64 `json:"quantity"`
	SerialNumber  string  `json:"serial_number"`
	ServiceLife   int     `json:"service_life_months"`
	Status        string  `json:"status"`
	PurchaseDate  string  `json:"purchase_date"`
	CreatedAt     string  `json:"created_at"`
}

type Contractor struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	ContactPerson string `json:"contact_person"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
}

type SalesOrderItem struct {
	ID           int     `json:"id"`
	OrderID      int     `json:"order_id"`
	OreBatchID   int     `json:"ore_batch_id"`
	OreBatchName string  `json:"ore_batch_name"`
	UnitID       int     `json:"unit_id"`
	UnitName     string  `json:"unit_name"`
	UnitSymbol   string  `json:"unit_symbol"`
	Quantity     float64 `json:"quantity"`
	PricePerUnit float64 `json:"price_per_unit"`
}

type SalesOrder struct {
	ID             int              `json:"id"`
	OrderNumber    string           `json:"order_number"`
	ContractorID   int              `json:"contractor_id"`
	ContractorName string           `json:"contractor_name"`
	WarehouseID    int              `json:"warehouse_id"`
	WarehouseName  string           `json:"warehouse_name"`
	Status         string           `json:"status"`
	OrderDate      string           `json:"order_date"`
	TotalQuantity  float64          `json:"total_quantity"`
	Items          []SalesOrderItem `json:"items"`
}

type Transport struct {
	ID            int     `json:"id"`
	Name          string  `json:"name"`
	Type          string  `json:"type"`
	VehicleNumber string  `json:"vehicle_number"`
	Capacity      float64 `json:"capacity"`
	UnitID        int     `json:"unit_id"`
	UnitName      string  `json:"unit_name"`
	UnitSymbol    string  `json:"unit_symbol"`
}

type Shipment struct {
	ID            int    `json:"id"`
	OrderID       int    `json:"order_id"`
	OrderNumber   string `json:"order_number"`
	TransportID   int    `json:"transport_id"`
	TransportName string `json:"transport_name"`
	PlannedDate   string `json:"planned_date"`
	ActualDate    string `json:"actual_date"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
}

type LogEntry struct {
	ID        int    `json:"id"`
	EventTime string `json:"event_time"`
	User      string `json:"user"`
	Action    string `json:"action"`
	Entity    string `json:"entity"`
	Details   string `json:"details"`
}

type ReferenceData struct {
	Units               []Unit              `json:"units"`
	Warehouses          []Warehouse         `json:"warehouses"`
	OreTypes            []OreType           `json:"ore_types"`
	EquipmentCategories []EquipmentCategory `json:"equipment_categories"`
	Contractors         []Contractor        `json:"contractors"`
	Transport           []Transport         `json:"transport"`
}

func main() {
	db, err := sql.Open("sqlite3", "./warehouse.db")
	if err != nil {
		log.Fatalf("can't open db %+v", err)
	}
	defer db.Close()

	if err := initializeSchema(db); err != nil {
		log.Fatalf("failed to initialize schema: %v", err)
	}
	if err := seedReferenceData(db); err != nil {
		log.Fatalf("failed to seed reference data: %v", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/api/reference-data", getReferenceData(db)).Methods("GET")
	router.HandleFunc("/api/ore-batches", getOreBatches(db)).Methods("GET")
	router.HandleFunc("/api/ore-batches", addOreBatch(db)).Methods("POST")
	router.HandleFunc("/api/equipment", getEquipment(db)).Methods("GET")
	router.HandleFunc("/api/equipment", addEquipment(db)).Methods("POST")
	router.HandleFunc("/api/orders", getOrders(db)).Methods("GET")
	router.HandleFunc("/api/orders", addOrder(db)).Methods("POST")
	router.HandleFunc("/api/orders/{id}/status", updateOrderStatus(db)).Methods("PUT")
	router.HandleFunc("/api/shipments", getShipments(db)).Methods("GET")
	router.HandleFunc("/api/shipments", addShipment(db)).Methods("POST")
	router.HandleFunc("/api/logs", getLogs(db)).Methods("GET")

	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./static")))

	fmt.Println("Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func initializeSchema(db *sql.DB) error {
	schema := `
    PRAGMA foreign_keys = ON;
    CREATE TABLE IF NOT EXISTS units (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        symbol TEXT NOT NULL,
        created_at TEXT,
        updated_at TEXT
    );
    CREATE TABLE IF NOT EXISTS warehouses (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        location TEXT,
        supervisor TEXT,
        capacity REAL,
        created_at TEXT,
        updated_at TEXT
    );
    CREATE TABLE IF NOT EXISTS ore_types (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        category TEXT,
        description TEXT,
        created_at TEXT,
        updated_at TEXT
    );
    CREATE TABLE IF NOT EXISTS ore_batches (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        ore_type_id INTEGER NOT NULL,
        warehouse_id INTEGER NOT NULL,
        unit_id INTEGER NOT NULL,
        batch_code TEXT,
        quantity REAL NOT NULL,
        quality REAL,
        priority TEXT,
        extraction_date TEXT,
        status TEXT,
        created_at TEXT,
        updated_at TEXT,
        FOREIGN KEY (ore_type_id) REFERENCES ore_types(id),
        FOREIGN KEY (warehouse_id) REFERENCES warehouses(id),
        FOREIGN KEY (unit_id) REFERENCES units(id)
    );
    CREATE TABLE IF NOT EXISTS equipment_categories (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        created_at TEXT,
        updated_at TEXT
    );
    CREATE TABLE IF NOT EXISTS equipment (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        category_id INTEGER NOT NULL,
        warehouse_id INTEGER NOT NULL,
        unit_id INTEGER NOT NULL,
        name TEXT NOT NULL,
        quantity REAL NOT NULL,
        serial_number TEXT,
        service_life_months INTEGER,
        status TEXT,
        purchase_date TEXT,
        created_at TEXT,
        updated_at TEXT,
        FOREIGN KEY (category_id) REFERENCES equipment_categories(id),
        FOREIGN KEY (warehouse_id) REFERENCES warehouses(id),
        FOREIGN KEY (unit_id) REFERENCES units(id)
    );
    CREATE TABLE IF NOT EXISTS contractors (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        type TEXT,
        contact_person TEXT,
        phone TEXT,
        email TEXT,
        created_at TEXT,
        updated_at TEXT
    );
    CREATE TABLE IF NOT EXISTS sales_orders (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        order_number TEXT NOT NULL,
        contractor_id INTEGER NOT NULL,
        warehouse_id INTEGER NOT NULL,
        status TEXT,
        order_date TEXT,
        total_quantity REAL,
        created_at TEXT,
        updated_at TEXT,
        FOREIGN KEY (contractor_id) REFERENCES contractors(id),
        FOREIGN KEY (warehouse_id) REFERENCES warehouses(id)
    );
    CREATE TABLE IF NOT EXISTS sales_order_items (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        order_id INTEGER NOT NULL,
        ore_batch_id INTEGER NOT NULL,
        unit_id INTEGER NOT NULL,
        quantity REAL NOT NULL,
        price_per_unit REAL,
        created_at TEXT,
        updated_at TEXT,
        FOREIGN KEY (order_id) REFERENCES sales_orders(id) ON DELETE CASCADE,
        FOREIGN KEY (ore_batch_id) REFERENCES ore_batches(id),
        FOREIGN KEY (unit_id) REFERENCES units(id)
    );
    CREATE TABLE IF NOT EXISTS transport (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        type TEXT,
        vehicle_number TEXT,
        capacity REAL,
        unit_id INTEGER,
        created_at TEXT,
        updated_at TEXT,
        FOREIGN KEY (unit_id) REFERENCES units(id)
    );
    CREATE TABLE IF NOT EXISTS shipments (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        order_id INTEGER NOT NULL,
        transport_id INTEGER,
        planned_date TEXT,
        actual_date TEXT,
        status TEXT,
        created_at TEXT,
        updated_at TEXT,
        FOREIGN KEY (order_id) REFERENCES sales_orders(id),
        FOREIGN KEY (transport_id) REFERENCES transport(id)
    );
    CREATE TABLE IF NOT EXISTS logs (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        event_time TEXT,
        user TEXT,
        action TEXT,
        entity TEXT,
        details TEXT
    );
    `
	_, err := db.Exec(schema)
	return err
}

func seedReferenceData(db *sql.DB) error {
	now := time.Now().Format(time.RFC3339)
	if err := seedUnits(db, now); err != nil {
		return err
	}
	if err := seedWarehouses(db, now); err != nil {
		return err
	}
	if err := seedOreTypes(db, now); err != nil {
		return err
	}
	if err := seedEquipmentCategories(db, now); err != nil {
		return err
	}
	if err := seedContractors(db, now); err != nil {
		return err
	}
	if err := seedTransport(db, now); err != nil {
		return err
	}
	return nil
}

func seedUnits(db *sql.DB, now string) error {
	count, err := countRecords(db, "units")
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	units := []struct{ name, symbol string }{
		{"Тонны", "т"},
		{"Килограммы", "кг"},
		{"Штуки", "шт"},
		{"Метры", "м"},
	}
	for _, u := range units {
		if _, err := db.Exec("INSERT INTO units (name, symbol, created_at, updated_at) VALUES (?, ?, ?, ?)", u.name, u.symbol, now, now); err != nil {
			return err
		}
	}
	return nil
}

func seedWarehouses(db *sql.DB, now string) error {
	count, err := countRecords(db, "warehouses")
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	warehouses := []struct {
		name, location, supervisor string
		capacity                   float64
	}{
		{"Склад №1", "Карьер Северный", "Иван Петров", 1200},
		{"Склад №2", "Карьер Южный", "Анна Смирнова", 800},
	}
	for _, w := range warehouses {
		if _, err := db.Exec("INSERT INTO warehouses (name, location, supervisor, capacity, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)", w.name, w.location, w.supervisor, w.capacity, now, now); err != nil {
			return err
		}
	}
	return nil
}

func seedOreTypes(db *sql.DB, now string) error {
	count, err := countRecords(db, "ore_types")
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	ores := []struct {
		name, category, description string
	}{
		{"Железная руда 65%", "Железные руды", "Высокое содержание железа"},
		{"Медная руда", "Цветные руды", "Содержание меди до 25%"},
		{"Золотосодержащая руда", "Драгоценные", "Руда с повышенным содержанием золота"},
	}
	for _, o := range ores {
		if _, err := db.Exec("INSERT INTO ore_types (name, category, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?)", o.name, o.category, o.description, now, now); err != nil {
			return err
		}
	}
	return nil
}

func seedEquipmentCategories(db *sql.DB, now string) error {
	count, err := countRecords(db, "equipment_categories")
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	categories := []string{"Погрузочная техника", "Буровое оборудование", "Транспортировка", "Контроль качества"}
	for _, c := range categories {
		if _, err := db.Exec("INSERT INTO equipment_categories (name, created_at, updated_at) VALUES (?, ?, ?)", c, now, now); err != nil {
			return err
		}
	}
	return nil
}

func seedContractors(db *sql.DB, now string) error {
	count, err := countRecords(db, "contractors")
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	contractors := []struct {
		name, ctype, contact, phone, email string
	}{
		{"ООО МеталлИнвест", "Покупатель", "Дмитрий Кузнецов", "+7 (921) 123-45-67", "sales@metallinvest.ru"},
		{"ЗАО РудаТрейд", "Покупатель", "Алексей Морозов", "+7 (812) 555-12-34", "info@rudatrade.ru"},
		{"АО Горные Машины", "Поставщик", "Екатерина Соколова", "+7 (495) 777-77-77", "supply@gormash.ru"},
	}
	for _, c := range contractors {
		if _, err := db.Exec("INSERT INTO contractors (name, type, contact_person, phone, email, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)", c.name, c.ctype, c.contact, c.phone, c.email, now, now); err != nil {
			return err
		}
	}
	return nil
}

func seedTransport(db *sql.DB, now string) error {
	count, err := countRecords(db, "transport")
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	transports := []struct {
		name, ttype, number string
		capacity            float64
		unitID              int
	}{
		{"БелАЗ 7513", "Самосвал", "A123BC", 90, 1},
		{"MAN TGX", "Тягач", "B456DE", 40, 1},
		{"ЖД состав №12", "Железная дорога", "TR-12", 200, 1},
	}
	for _, t := range transports {
		if _, err := db.Exec("INSERT INTO transport (name, type, vehicle_number, capacity, unit_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)", t.name, t.ttype, t.number, t.capacity, t.unitID, now, now); err != nil {
			return err
		}
	}
	return nil
}

func countRecords(db *sql.DB, table string) (int, error) {
	row := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table))
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func getReferenceData(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := ReferenceData{}
		if units, err := fetchUnits(db); err == nil {
			data.Units = units
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if warehouses, err := fetchWarehouses(db); err == nil {
			data.Warehouses = warehouses
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if oreTypes, err := fetchOreTypes(db); err == nil {
			data.OreTypes = oreTypes
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if categories, err := fetchEquipmentCategories(db); err == nil {
			data.EquipmentCategories = categories
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if contractors, err := fetchContractors(db); err == nil {
			data.Contractors = contractors
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if transport, err := fetchTransport(db); err == nil {
			data.Transport = transport
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	}
}

func fetchUnits(db *sql.DB) ([]Unit, error) {
	rows, err := db.Query("SELECT id, name, symbol FROM units ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var units []Unit
	for rows.Next() {
		var u Unit
		if err := rows.Scan(&u.ID, &u.Name, &u.Symbol); err != nil {
			return nil, err
		}
		units = append(units, u)
	}
	return units, nil
}

func fetchWarehouses(db *sql.DB) ([]Warehouse, error) {
	rows, err := db.Query("SELECT id, name, location, capacity, supervisor FROM warehouses ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var warehouses []Warehouse
	for rows.Next() {
		var w Warehouse
		if err := rows.Scan(&w.ID, &w.Name, &w.Location, &w.Capacity, &w.Supervisor); err != nil {
			return nil, err
		}
		warehouses = append(warehouses, w)
	}
	return warehouses, nil
}

func fetchOreTypes(db *sql.DB) ([]OreType, error) {
	rows, err := db.Query("SELECT id, name, category, description FROM ore_types ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ores []OreType
	for rows.Next() {
		var o OreType
		if err := rows.Scan(&o.ID, &o.Name, &o.Category, &o.Description); err != nil {
			return nil, err
		}
		ores = append(ores, o)
	}
	return ores, nil
}

func fetchEquipmentCategories(db *sql.DB) ([]EquipmentCategory, error) {
	rows, err := db.Query("SELECT id, name FROM equipment_categories ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var categories []EquipmentCategory
	for rows.Next() {
		var c EquipmentCategory
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, nil
}

func fetchContractors(db *sql.DB) ([]Contractor, error) {
	rows, err := db.Query("SELECT id, name, type, contact_person, phone, email FROM contractors ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var contractors []Contractor
	for rows.Next() {
		var c Contractor
		if err := rows.Scan(&c.ID, &c.Name, &c.Type, &c.ContactPerson, &c.Phone, &c.Email); err != nil {
			return nil, err
		}
		contractors = append(contractors, c)
	}
	return contractors, nil
}

func fetchTransport(db *sql.DB) ([]Transport, error) {
	rows, err := db.Query(`
        SELECT t.id, t.name, t.type, t.vehicle_number, t.capacity, t.unit_id, IFNULL(u.name, ''), IFNULL(u.symbol, '')
        FROM transport t
        LEFT JOIN units u ON t.unit_id = u.id
        ORDER BY t.name
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var transport []Transport
	for rows.Next() {
		var t Transport
		if err := rows.Scan(&t.ID, &t.Name, &t.Type, &t.VehicleNumber, &t.Capacity, &t.UnitID, &t.UnitName, &t.UnitSymbol); err != nil {
			return nil, err
		}
		transport = append(transport, t)
	}
	return transport, nil
}

func getOreBatches(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(`
            SELECT ob.id, ob.batch_code, ob.ore_type_id, ot.name, ob.warehouse_id, w.name,
                   ob.unit_id, u.name, u.symbol, ob.quantity, IFNULL(ob.quality, 0), IFNULL(ob.priority, ''),
                   IFNULL(ob.extraction_date, ''), IFNULL(ob.status, ''), IFNULL(ob.created_at, '')
            FROM ore_batches ob
            JOIN ore_types ot ON ob.ore_type_id = ot.id
            JOIN warehouses w ON ob.warehouse_id = w.id
            JOIN units u ON ob.unit_id = u.id
            ORDER BY ob.created_at DESC, ob.id DESC
        `)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var batches []OreBatch
		for rows.Next() {
			var ob OreBatch
			if err := rows.Scan(&ob.ID, &ob.BatchCode, &ob.OreTypeID, &ob.OreTypeName, &ob.WarehouseID, &ob.WarehouseName,
				&ob.UnitID, &ob.UnitName, &ob.UnitSymbol, &ob.Quantity, &ob.Quality, &ob.Priority,
				&ob.ExtractionDate, &ob.Status, &ob.CreatedAt); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			batches = append(batches, ob)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(batches)
	}
}

func addOreBatch(db *sql.DB) http.HandlerFunc {
	type request struct {
		OreTypeID      int      `json:"ore_type_id"`
		WarehouseID    int      `json:"warehouse_id"`
		UnitID         int      `json:"unit_id"`
		BatchCode      string   `json:"batch_code"`
		Quantity       float64  `json:"quantity"`
		Quality        *float64 `json:"quality"`
		Priority       string   `json:"priority"`
		ExtractionDate string   `json:"extraction_date"`
		Status         string   `json:"status"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.OreTypeID == 0 || req.WarehouseID == 0 || req.UnitID == 0 || req.Quantity <= 0 {
			http.Error(w, "Отсутствуют обязательные поля", http.StatusBadRequest)
			return
		}
		now := time.Now().Format(time.RFC3339)
		_, err := db.Exec(`
            INSERT INTO ore_batches (ore_type_id, warehouse_id, unit_id, batch_code, quantity, quality, priority, extraction_date, status, created_at, updated_at)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        `, req.OreTypeID, req.WarehouseID, req.UnitID, req.BatchCode, req.Quantity, req.Quality, req.Priority, req.ExtractionDate, req.Status, now, now)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		logAction(db, "system", "Добавление партии руды", "ore_batches", fmt.Sprintf("Партия %s", req.BatchCode))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Партия руды добавлена"})
	}
}

func getEquipment(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(`
            SELECT e.id, e.name, e.category_id, c.name, e.warehouse_id, w.name, e.unit_id, u.name, u.symbol,
                   e.quantity, IFNULL(e.serial_number, ''), IFNULL(e.service_life_months, 0), IFNULL(e.status, ''),
                   IFNULL(e.purchase_date, ''), IFNULL(e.created_at, '')
            FROM equipment e
            JOIN equipment_categories c ON e.category_id = c.id
            JOIN warehouses w ON e.warehouse_id = w.id
            JOIN units u ON e.unit_id = u.id
            ORDER BY e.created_at DESC, e.id DESC
        `)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var items []Equipment
		for rows.Next() {
			var eq Equipment
			if err := rows.Scan(&eq.ID, &eq.Name, &eq.CategoryID, &eq.CategoryName, &eq.WarehouseID, &eq.WarehouseName,
				&eq.UnitID, &eq.UnitName, &eq.UnitSymbol, &eq.Quantity, &eq.SerialNumber, &eq.ServiceLife,
				&eq.Status, &eq.PurchaseDate, &eq.CreatedAt); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			items = append(items, eq)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(items)
	}
}

func addEquipment(db *sql.DB) http.HandlerFunc {
	type request struct {
		Name         string  `json:"name"`
		CategoryID   int     `json:"category_id"`
		WarehouseID  int     `json:"warehouse_id"`
		UnitID       int     `json:"unit_id"`
		Quantity     float64 `json:"quantity"`
		SerialNumber string  `json:"serial_number"`
		ServiceLife  *int    `json:"service_life_months"`
		Status       string  `json:"status"`
		PurchaseDate string  `json:"purchase_date"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.Name == "" || req.CategoryID == 0 || req.WarehouseID == 0 || req.UnitID == 0 || req.Quantity <= 0 {
			http.Error(w, "Отсутствуют обязательные поля", http.StatusBadRequest)
			return
		}
		now := time.Now().Format(time.RFC3339)
		_, err := db.Exec(`
            INSERT INTO equipment (name, category_id, warehouse_id, unit_id, quantity, serial_number, service_life_months, status, purchase_date, created_at, updated_at)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        `, req.Name, req.CategoryID, req.WarehouseID, req.UnitID, req.Quantity, req.SerialNumber, req.ServiceLife, req.Status, req.PurchaseDate, now, now)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		logAction(db, "system", "Добавление оборудования", "equipment", req.Name)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Единица оборудования добавлена"})
	}
}

func getOrders(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ordersRows, err := db.Query(`
            SELECT o.id, o.order_number, o.contractor_id, c.name, o.warehouse_id, w.name, IFNULL(o.status, ''), IFNULL(o.order_date, ''), IFNULL(o.total_quantity, 0)
            FROM sales_orders o
            JOIN contractors c ON o.contractor_id = c.id
            JOIN warehouses w ON o.warehouse_id = w.id
            ORDER BY o.order_date DESC, o.id DESC
        `)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer ordersRows.Close()

		ordersMap := make(map[int]*SalesOrder)
		var orderIDs []int
		for ordersRows.Next() {
			var o SalesOrder
			if err := ordersRows.Scan(&o.ID, &o.OrderNumber, &o.ContractorID, &o.ContractorName, &o.WarehouseID, &o.WarehouseName, &o.Status, &o.OrderDate, &o.TotalQuantity); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			o.Items = []SalesOrderItem{}
			ordersMap[o.ID] = &o
			orderIDs = append(orderIDs, o.ID)
		}

		if len(orderIDs) > 0 {
			query := `
                SELECT i.id, i.order_id, i.ore_batch_id, IFNULL(ob.batch_code, ''), i.unit_id, u.name, u.symbol, i.quantity, IFNULL(i.price_per_unit, 0)
                FROM sales_order_items i
                LEFT JOIN ore_batches ob ON i.ore_batch_id = ob.id
                LEFT JOIN units u ON i.unit_id = u.id
                WHERE i.order_id IN (` + placeholders(len(orderIDs)) + `)
                ORDER BY i.id
            `
			args := make([]interface{}, len(orderIDs))
			for i, id := range orderIDs {
				args[i] = id
			}
			rows, err := db.Query(query, args...)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer rows.Close()
			for rows.Next() {
				var item SalesOrderItem
				if err := rows.Scan(&item.ID, &item.OrderID, &item.OreBatchID, &item.OreBatchName, &item.UnitID, &item.UnitName, &item.UnitSymbol, &item.Quantity, &item.PricePerUnit); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				if order, ok := ordersMap[item.OrderID]; ok {
					order.Items = append(order.Items, item)
				}
			}
		}

		orders := make([]SalesOrder, 0, len(ordersMap))
		for _, order := range ordersMap {
			orders = append(orders, *order)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(orders)
	}
}

func addOrder(db *sql.DB) http.HandlerFunc {
	type itemRequest struct {
		OreBatchID   int     `json:"ore_batch_id"`
		UnitID       int     `json:"unit_id"`
		Quantity     float64 `json:"quantity"`
		PricePerUnit float64 `json:"price_per_unit"`
	}
	type request struct {
		OrderNumber  string        `json:"order_number"`
		ContractorID int           `json:"contractor_id"`
		WarehouseID  int           `json:"warehouse_id"`
		Status       string        `json:"status"`
		OrderDate    string        `json:"order_date"`
		Items        []itemRequest `json:"items"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.OrderNumber == "" || req.ContractorID == 0 || req.WarehouseID == 0 || len(req.Items) == 0 {
			http.Error(w, "Заполните обязательные поля", http.StatusBadRequest)
			return
		}

		tx, err := db.Begin()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		now := time.Now().Format(time.RFC3339)
		res, err := tx.Exec(`
            INSERT INTO sales_orders (order_number, contractor_id, warehouse_id, status, order_date, total_quantity, created_at, updated_at)
            VALUES (?, ?, ?, ?, ?, 0, ?, ?)
        `, req.OrderNumber, req.ContractorID, req.WarehouseID, req.Status, req.OrderDate, now, now)
		if err != nil {
			tx.Rollback()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		orderID, err := res.LastInsertId()
		if err != nil {
			tx.Rollback()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		total := 0.0
		for _, item := range req.Items {
			if item.OreBatchID == 0 || item.UnitID == 0 || item.Quantity <= 0 {
				tx.Rollback()
				http.Error(w, "Некорректные строки заказа", http.StatusBadRequest)
				return
			}
			total += item.Quantity
			if _, err := tx.Exec(`
                INSERT INTO sales_order_items (order_id, ore_batch_id, unit_id, quantity, price_per_unit, created_at, updated_at)
                VALUES (?, ?, ?, ?, ?, ?, ?)
            `, orderID, item.OreBatchID, item.UnitID, item.Quantity, item.PricePerUnit, now, now); err != nil {
				tx.Rollback()
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		if _, err := tx.Exec("UPDATE sales_orders SET total_quantity = ?, updated_at = ? WHERE id = ?", total, now, orderID); err != nil {
			tx.Rollback()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := tx.Commit(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		logAction(db, "system", "Создание заказа", "sales_orders", req.OrderNumber)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Заказ создан"})
	}
}

func updateOrderStatus(db *sql.DB) http.HandlerFunc {
	type request struct {
		Status string `json:"status"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		orderID := vars["id"]
		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.Status == "" {
			http.Error(w, "Укажите новый статус", http.StatusBadRequest)
			return
		}
		now := time.Now().Format(time.RFC3339)
		res, err := db.Exec("UPDATE sales_orders SET status = ?, updated_at = ? WHERE id = ?", req.Status, now, orderID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			http.Error(w, "Заказ не найден", http.StatusNotFound)
			return
		}
		logAction(db, "system", "Обновление статуса заказа", "sales_orders", fmt.Sprintf("ID %s -> %s", orderID, req.Status))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Статус обновлён"})
	}
}

func getShipments(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(`
            SELECT s.id, s.order_id, o.order_number, IFNULL(s.transport_id, 0), IFNULL(t.name, ''),
                   IFNULL(s.planned_date, ''), IFNULL(s.actual_date, ''), IFNULL(s.status, ''), IFNULL(s.created_at, '')
            FROM shipments s
            JOIN sales_orders o ON s.order_id = o.id
            LEFT JOIN transport t ON s.transport_id = t.id
            ORDER BY s.created_at DESC, s.id DESC
        `)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var shipments []Shipment
		for rows.Next() {
			var s Shipment
			if err := rows.Scan(&s.ID, &s.OrderID, &s.OrderNumber, &s.TransportID, &s.TransportName, &s.PlannedDate, &s.ActualDate, &s.Status, &s.CreatedAt); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			shipments = append(shipments, s)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(shipments)
	}
}

func addShipment(db *sql.DB) http.HandlerFunc {
	type request struct {
		OrderID     int    `json:"order_id"`
		TransportID int    `json:"transport_id"`
		PlannedDate string `json:"planned_date"`
		ActualDate  string `json:"actual_date"`
		Status      string `json:"status"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.OrderID == 0 {
			http.Error(w, "Не указан заказ", http.StatusBadRequest)
			return
		}
		now := time.Now().Format(time.RFC3339)
		_, err := db.Exec(`
            INSERT INTO shipments (order_id, transport_id, planned_date, actual_date, status, created_at, updated_at)
            VALUES (?, ?, ?, ?, ?, ?, ?)
        `, req.OrderID, nullableInt(req.TransportID), req.PlannedDate, req.ActualDate, req.Status, now, now)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		logAction(db, "system", "Создание отгрузки", "shipments", fmt.Sprintf("Заказ %d", req.OrderID))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Отгрузка создана"})
	}
}

func nullableInt(value int) interface{} {
	if value == 0 {
		return nil
	}
	return value
}

func getLogs(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, IFNULL(event_time, ''), IFNULL(user, ''), IFNULL(action, ''), IFNULL(entity, ''), IFNULL(details, '') FROM logs ORDER BY event_time DESC, id DESC LIMIT 100")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		var logs []LogEntry
		for rows.Next() {
			var logEntry LogEntry
			if err := rows.Scan(&logEntry.ID, &logEntry.EventTime, &logEntry.User, &logEntry.Action, &logEntry.Entity, &logEntry.Details); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			logs = append(logs, logEntry)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(logs)
	}
}

func logAction(db *sql.DB, user, action, entity, details string) {
	if user == "" {
		user = "system"
	}
	now := time.Now().Format(time.RFC3339)
	if _, err := db.Exec("INSERT INTO logs (event_time, user, action, entity, details) VALUES (?, ?, ?, ?, ?)", now, user, action, entity, details); err != nil {
		log.Printf("failed to insert log: %v", err)
	}
}

func placeholders(n int) string {
	if n <= 0 {
		return ""
	}
	result := "?"
	for i := 1; i < n; i++ {
		result += ",?"
	}
	return result
}
