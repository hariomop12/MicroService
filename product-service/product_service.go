package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/lib/pq"
)

var db *sql.DB

type Product struct {
	ID            int       `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Price         float64   `json:"price"`
	StockQuantity int       `json:"stock_quantity"`
	Category      string    `json:"category"`
	Tags          []string  `json:"tags"`
	CreatedAt     time.Time `json:"created_at"`
}

type CreateProductRequest struct {
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Price         float64  `json:"price"`
	StockQuantity int      `json:"stock_quantity"`
	Category      string   `json:"category"`
	Tags          []string `json:"tags"`
}

func initDB() {
	connStr := "host=localhost port=5432 user=postgres password=postgres dbname=products_db sslmode=disable"
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to products database")
}

func createProductHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var productID int
	err := db.QueryRow(`
		INSERT INTO products (name, description, price, stock_quantity, category, tags)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id
	`, req.Name, req.Description, req.Price, req.StockQuantity, req.Category, pq.Array(req.Tags)).Scan(&productID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":    "Product created successfully",
		"product_id": productID,
	})
}

func getProductHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID := vars["id"]

	var product Product
	var tags pq.StringArray

	err := db.QueryRow(
		`SELECT id, name, description, price, stock_quantity, category, tags, created_at FROM products WHERE id = $1
		`, productID).Scan(
		&product.ID, &product.Name, &product.Description, &product.Price, &product.StockQuantity, &product.Category, &tags, &product.CreatedAt,
	)

	if err != nil {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	product.Tags = tags

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}

func searchProductsHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	// Using GIN index for full-text search
	rows, err := db.Query(`
		SELECT id, name, description, price, stock_quantity, category, tags, created_at 
		FROM products 
		WHERE to_tsvector('english', name || ' ' || COALESCE(description, '')) @@ plainto_tsquery('english', $1)
		ORDER BY ts_rank(to_tsvector('english', name || ' ' || COALESCE(description, '')), plainto_tsquery('english', $1)) DESC
		LIMIT 50
	`, query)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var product Product
		var tags pq.StringArray
		if err := rows.Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.StockQuantity, &product.Category, &tags, &product.CreatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		product.Tags = tags
		products = append(products, product)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)

}

func searchByTagsHandler(w http.ResponseWriter, r *http.Request) {
	tag := r.URL.Query().Get("tag")
	if tag == "" {
		http.Error(w, "Tags parameter required", http.StatusBadRequest)
		return
	}

	// Using GIN index for array Search
	rows, err := db.Query(`
		SELECT id, name, description, price, stock_quantity, category, tags, created_at 
		FROM products 
		WHERE tags @> ARRAY[$1]::text[]
		LIMIT 50
	`, tag)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var product Product
		var tags pq.StringArray
		if err := rows.Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.StockQuantity, &product.Category, &tags, &product.CreatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		product.Tags = tags
		products = append(products, product)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

func updateStockHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID := vars["id"]

	var req struct {
		Quantity int `json:"quantity"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := db.Exec(`
	UPDATE products 
		SET stock_quantity = stock_quantity + $1, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $2
	`, req.Quantity, productID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()

	if rowsAffected == 0 {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Stock updated successfully"})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func main() {
	initDB()
	defer db.Close()

	r := mux.NewRouter()
	r.HandleFunc("/health", healthHandler).Methods("GET")
	r.HandleFunc("/api/products", createProductHandler).Methods("POST")
	r.HandleFunc("/api/products/{id}", getProductHandler).Methods("GET")
	r.HandleFunc("/api/products/search", searchProductsHandler).Methods("GET")
	r.HandleFunc("/api/products/tags", searchByTagsHandler).Methods("GET")
	r.HandleFunc("/api/products/{id}/stock", updateStockHandler).Methods("PATCH")

	fmt.Println("Product Service running on :8002")
	log.Fatal(http.ListenAndServe(":8002", r))
}