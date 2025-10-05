package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

var db *sql.DB

type Order struct {
	ID          int         `json:"id"`
	UserID      int         `json:"user_id"`
	Status      string      `json:"status"`
	TotalAmount float64     `json:"total_amount"`
	Items       []OrderItem `json:"items"`
	CreatedAt   time.Time   `json:"created_at"`
}

type OrderItem struct {
	ID        int     `json:"id"`
	ProductID int     `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

type CreateOrderRequest struct {
	UserID int `json:"user_id"`
	Items  []struct {
		ProductID int `json:"product_id"`
		Quantity  int `json:"quantity"`
	} `json:"items"`
}

type Product struct {
	ID            int     `json:"id"`
	Price         float64 `json:"price"`
	StockQuantity int     `json:"stock_quantity"`
}

func initDB() {
	connStr := "host=postgres port=5432 user=postgres password=postgres dbname=orders_db sslmode=disable"
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to orders database")
}

func getProductFromService(productID int) (*Product, error) {
	resp, err := http.Get(fmt.Sprintf("http://localhost:8002/api/products/%d", productID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("product not found")
	}

	var product Product
	if err := json.NewDecoder(resp.Body).Decode(&product); err != nil {
		return nil, err
	}

	return &product, nil
}

func updateProductStock(productID, quantity int) error {
	reqBody, _ := json.Marshal(map[string]int{"quantity": quantity})
	req, err := http.NewRequest("PATCH",
		fmt.Sprintf("http://localhost:8002/api/products/%d/stock", productID),
		io.NopCloser(io.NopCloser(io.NopCloser(jsonReader{reqBody}))))

	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update stock")
	}

	return nil
}

type jsonReader struct {
	data []byte
}

func (jr jsonReader) Read(p []byte) (n int, err error) {
	if len(jr.data) == 0 {
		return 0, io.EOF
	}
	n = copy(p, jr.data)
	jr.data = jr.data[n:]
	return n, nil
}

func createOrderHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate products and calculate total
	var totalAmount float64
	var orderItems []OrderItem

	for _, item := range req.Items {
		product, err := getProductFromService(item.ProductID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Product %d not found", item.ProductID), http.StatusBadRequest)
			return
		}

		if product.StockQuantity < item.Quantity {
			http.Error(w, fmt.Sprintf("Insufficient stock for product %d", item.ProductID), http.StatusBadRequest)
			return
		}

		itemTotal := product.Price * float64(item.Quantity)
		totalAmount += itemTotal

		orderItems = append(orderItems, OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     product.Price,
		})
	}

	// Create order in transaction
	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var orderID int
	err = tx.QueryRow(`
		INSERT INTO orders (user_id, status, total_amount) 
		VALUES ($1, $2, $3) 
		RETURNING id
	`, req.UserID, "pending", totalAmount).Scan(&orderID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Insert order items
	for _, item := range orderItems {
		_, err := tx.Exec(`
			INSERT INTO order_items (order_id, product_id, quantity, price) 
			VALUES ($1, $2, $3, $4)
		`, orderID, item.ProductID, item.Quantity, item.Price)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Update stock in product service
		if err := updateProductStock(item.ProductID, -item.Quantity); err != nil {
			http.Error(w, "Failed to update product stock", http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":      "Order created successfully",
		"order_id":     orderID,
		"total_amount": totalAmount,
	})
}

func getOrderHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["id"]

	var order Order
	err := db.QueryRow(`
		SELECT id, user_id, status, total_amount, created_at 
		FROM orders WHERE id = $1
	`, orderID).Scan(&order.ID, &order.UserID, &order.Status, &order.TotalAmount, &order.CreatedAt)

	if err != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	// Get order items
	rows, err := db.Query(`
		SELECT id, product_id, quantity, price 
		FROM order_items WHERE order_id = $1
	`, orderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var item OrderItem
		if err := rows.Scan(&item.ID, &item.ProductID, &item.Quantity, &item.Price); err != nil {
			continue
		}
		order.Items = append(order.Items, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func getUserOrdersHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]

	rows, err := db.Query(`
		SELECT id, user_id, status, total_amount, created_at 
		FROM orders WHERE user_id = $1 
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var order Order
		if err := rows.Scan(&order.ID, &order.UserID, &order.Status, &order.TotalAmount, &order.CreatedAt); err != nil {
			continue
		}
		orders = append(orders, order)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func updateOrderStatusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["id"]

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := db.Exec(`
		UPDATE orders 
		SET status = $1, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $2
	`, req.Status, orderID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Order status updated successfully"})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy", "service": "order-service"})
}

func main() {
	initDB()
	defer db.Close()

	r := mux.NewRouter()
	r.HandleFunc("/health", healthHandler).Methods("GET")
	r.HandleFunc("/api/orders", createOrderHandler).Methods("POST")
	r.HandleFunc("/api/orders/{id}", getOrderHandler).Methods("GET")
	r.HandleFunc("/api/orders/user/{user_id}", getUserOrdersHandler).Methods("GET")
	r.HandleFunc("/api/orders/{id}/status", updateOrderStatusHandler).Methods("PATCH")

	fmt.Println("Order Service running on :8003")
	log.Fatal(http.ListenAndServe(":8003", r))
}
