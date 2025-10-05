package apigetway

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type ServiceRegistry struct {
	UserService    *url.URL
	ProductService *url.URL
	OrderService   *url.URL
}

var registry ServiceRegistry

func initServiceRegistry() {
	var err error

	registry.UserService, err = url.Parse("http://localhost:8001")
	if err != nil {
		log.Fatal(err)
	}

	registry.ProductService, err = url.Parse("http://localhost:8002")
	if err != nil {
		log.Fatal(err)
	}

	registry.OrderService, err = url.Parse("http://localhost:8003")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Service registry initialized")
}

// Logging middleware
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("[%s] %s %s", r.Method, r.RequestURI, r.RemoteAddr)
		next.ServeHTTP(w, r)
		log.Printf("Completed in %v", time.Since(start))
	})
}

// CORS middleware
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Rate limiting middleware (simple in-memory implementation)
var requestCounts = make(map[string]int)
var lastReset = time.Now()

func rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Reset counts every minute
		if time.Since(lastReset) > time.Minute {
			requestCounts = make(map[string]int)
			lastReset = time.Now()
		}

		clientIP := r.RemoteAddr
		requestCounts[clientIP]++

		// Allow 100 requests per minute per IP
		if requestCounts[clientIP] > 100 {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Health check aggregator
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	services := map[string]string{
		"user-service":    "http://localhost:8001/health",
		"product-service": "http://localhost:8002/health",
		"order-service":   "http://localhost:8003/health",
	}

	results := make(map[string]string)
	allHealthy := true

	for name, endpoint := range services {
		resp, err := http.Get(endpoint)
		if err != nil || resp.StatusCode != http.StatusOK {
			results[name] = "unhealthy"
			allHealthy = false
		} else {
			results[name] = "healthy"
			resp.Body.Close()
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if !allHealthy {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	fmt.Fprintf(w, `{"gateway": "healthy", "services": %v}`, results)
}

// Create reverse proxy handler
func createReverseProxy(target *url.URL) http.HandlerFunc {
	proxy := httputil.NewSingleHostReverseProxy(target)

	// Custom error handler
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Proxy error: %v", err)
		w.WriteHeader(http.StatusBadGateway)
		fmt.Fprintf(w, `{"error": "Service unavailable"}`)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	}
}

// Route requests to appropriate service
func routeHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Route to User Service
	if strings.HasPrefix(path, "/api/users") {
		createReverseProxy(registry.UserService)(w, r)
		return
	}

	// Route to Product Service
	if strings.HasPrefix(path, "/api/products") {
		createReverseProxy(registry.ProductService)(w, r)
		return
	}

	// Route to Order Service
	if strings.HasPrefix(path, "/api/orders") {
		createReverseProxy(registry.OrderService)(w, r)
		return
	}

	http.Error(w, "Not Found", http.StatusNotFound)
}


// Service discovery endpoint
func servicesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{
		"services": {
			"user-service": "http://localhost:8001",
			"product-service": "http://localhost:8002",
			"order-service": "http://localhost:8003"
		}
	}`)
}

func main() {
	initServiceRegistry()
	
	r := mux.NewRouter()
	
	// Gateway-specific endpoints
	r.HandleFunc("/health", healthCheckHandler).Methods("GET")
	r.HandleFunc("/services", servicesHandler).Methods("GET")
	
	// Proxy all API requests
	r.PathPrefix("/api/").HandlerFunc(routeHandler)
	
	// Apply middleware
	handler := loggingMiddleware(corsMiddleware(rateLimitMiddleware(r)))
	
	fmt.Println("API Gateway running on :8000")
	fmt.Println("-----------------------------------")
	fmt.Println("Routes:")
	fmt.Println("  GET  /health           - Health check")
	fmt.Println("  GET  /services         - Service registry")
	fmt.Println("  *    /api/users/*      -> User Service (8001)")
	fmt.Println("  *    /api/products/*   -> Product Service (8002)")
	fmt.Println("  *    /api/orders/*     -> Order Service (8003)")
	fmt.Println("-----------------------------------")
	
	log.Fatal(http.ListenAndServe(":8000", handler))
}