// Orders API - Gestión de pedidos con Go (net/http) + SQLite
package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	_ "modernc.org/sqlite" // driver de SQLite escrito en Go puro
)

// Order representa un pedido. Las etiquetas `json` definen los nombres en el JSON.
type Order struct {
	ID        int64   `json:"id"`
	Customer  string  `json:"customer"`
	Product   string  `json:"product"`
	Quantity  int     `json:"quantity"`
	Total     float64 `json:"total"`
	Status    string  `json:"status"`
	CreatedAt string  `json:"created_at"`
}

// Estados permitidos para un pedido
var validStatus = map[string]bool{
	"pending": true, "shipped": true, "delivered": true, "cancelled": true,
}

var db *sql.DB

// ---------- Base de datos ----------

func initDB() {
	// Ruta de la base de datos (Docker la pasa con la variable DB_PATH)
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data/orders.db"
	}
	os.MkdirAll(filepath.Dir(dbPath), 0755)

	var err error
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatal(err)
	}

	// Crear la tabla si todavía no existe
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS orders (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			customer   TEXT NOT NULL,
			product    TEXT NOT NULL,
			quantity   INTEGER NOT NULL,
			total      REAL NOT NULL,
			status     TEXT NOT NULL DEFAULT 'pending',
			created_at TEXT DEFAULT CURRENT_TIMESTAMP
		)`)
	if err != nil {
		log.Fatal(err)
	}
}

// findOrder busca un pedido por ID. Devuelve nil si no existe.
func findOrder(id string) *Order {
	var o Order
	err := db.QueryRow("SELECT id, customer, product, quantity, total, status, created_at FROM orders WHERE id = ?", id).
		Scan(&o.ID, &o.Customer, &o.Product, &o.Quantity, &o.Total, &o.Status, &o.CreatedAt)
	if err != nil {
		return nil
	}
	return &o
}

// writeJSON envía cualquier dato como respuesta JSON
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func errorJSON(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"detail": message})
}

// ---------- Endpoints ----------

func health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]string{"service": "orders-api", "status": "ok"})
}

// Crear pedido
func createOrder(w http.ResponseWriter, r *http.Request) {
	var o Order
	if err := json.NewDecoder(r.Body).Decode(&o); err != nil || o.Customer == "" || o.Product == "" || o.Quantity <= 0 {
		errorJSON(w, 400, "Los campos customer, product y quantity (> 0) son obligatorios")
		return
	}
	result, err := db.Exec("INSERT INTO orders (customer, product, quantity, total) VALUES (?, ?, ?, ?)",
		o.Customer, o.Product, o.Quantity, o.Total)
	if err != nil {
		errorJSON(w, 500, err.Error())
		return
	}
	id, _ := result.LastInsertId()
	writeJSON(w, 201, findOrder(itoa(id)))
}

// Listar pedidos
func listOrders(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, customer, product, quantity, total, status, created_at FROM orders")
	if err != nil {
		errorJSON(w, 500, err.Error())
		return
	}
	defer rows.Close()

	orders := []Order{} // lista vacía en vez de null si no hay pedidos
	for rows.Next() {
		var o Order
		rows.Scan(&o.ID, &o.Customer, &o.Product, &o.Quantity, &o.Total, &o.Status, &o.CreatedAt)
		orders = append(orders, o)
	}
	writeJSON(w, 200, orders)
}

// Consultar pedido por ID
func getOrder(w http.ResponseWriter, r *http.Request) {
	o := findOrder(r.PathValue("id"))
	if o == nil {
		errorJSON(w, 404, "Pedido no encontrado")
		return
	}
	writeJSON(w, 200, o)
}

// Actualizar el estado de un pedido
func updateStatus(w http.ResponseWriter, r *http.Request) {
	o := findOrder(r.PathValue("id"))
	if o == nil {
		errorJSON(w, 404, "Pedido no encontrado")
		return
	}
	var body struct {
		Status string `json:"status"`
	}
	json.NewDecoder(r.Body).Decode(&body)
	if !validStatus[body.Status] {
		errorJSON(w, 400, "Estado inválido. Usa: pending, shipped, delivered o cancelled")
		return
	}
	db.Exec("UPDATE orders SET status = ? WHERE id = ?", body.Status, o.ID)
	writeJSON(w, 200, findOrder(itoa(o.ID)))
}

// Eliminar pedido
func deleteOrder(w http.ResponseWriter, r *http.Request) {
	o := findOrder(r.PathValue("id"))
	if o == nil {
		errorJSON(w, 404, "Pedido no encontrado")
		return
	}
	db.Exec("DELETE FROM orders WHERE id = ?", o.ID)
	writeJSON(w, 200, map[string]string{"message": "Pedido " + itoa(o.ID) + " eliminado"})
}

// itoa convierte un número a texto (ej: 5 -> "5")
func itoa(n int64) string {
	return strconv.FormatInt(n, 10)
}

func main() {
	initDB()

	// Rutas: "MÉTODO /ruta/{parametro}" (disponible desde Go 1.22)
	http.HandleFunc("GET /{$}", health) // {$} = solo la ruta "/" exacta
	http.HandleFunc("POST /orders", createOrder)
	http.HandleFunc("GET /orders", listOrders)
	http.HandleFunc("GET /orders/{id}", getOrder)
	http.HandleFunc("PUT /orders/{id}/status", updateStatus)
	http.HandleFunc("DELETE /orders/{id}", deleteOrder)

	log.Println("Orders API escuchando en el puerto 8002")
	log.Fatal(http.ListenAndServe(":8002", nil))
}
