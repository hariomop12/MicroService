# 🚀 Microservice E-commerce Platform

A complete microservices architecture built with Go, featuring user management, product catalog, order processing, and API gateway with PostgreSQL databases and Docker containerization.

## 📋 Table of Contents

- [Architecture Overview](#architecture-overview)
- [Tech Stack](#tech-stack)
- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Database Migration with dbmate](#database-migration-with-dbmate)
- [API Documentation](#api-documentation)
- [Testing with cURL](#testing-with-curl)
- [Project Structure](#project-structure)
- [Services](#services)
- [Troubleshooting](#troubleshooting)

## 🏗️ Architecture Overview

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│                 │    │                  │    │                 │
│   NGINX         │────│   API Gateway    │────│   User Service  │
│   (Port 80)     │    │   (Port 8000)    │    │   (Port 8001)   │
│                 │    │                  │    │                 │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                │
                                ├────────────────────────────────────┐
                                │                                    │
                       ┌─────────────────┐              ┌─────────────────┐
                       │                 │              │                 │
                       │ Product Service │              │  Order Service  │
                       │   (Port 8002)   │              │   (Port 8003)   │
                       │                 │              │                 │
                       └─────────────────┘              └─────────────────┘
                                │                                    │
                                └────────────────┬───────────────────┘
                                                 │
                                    ┌─────────────────────┐
                                    │                     │
                                    │   PostgreSQL        │
                                    │   (Port 5432)       │
                                    │ ┌─────────────────┐ │
                                    │ │   users_db      │ │
                                    │ │   products_db   │ │
                                    │ │   orders_db     │ │
                                    │ └─────────────────┘ │
                                    └─────────────────────┘
```

## 🛠️ Tech Stack

- **Backend**: Go 1.23
- **Database**: PostgreSQL 15
- **Migration Tool**: dbmate
- **Containerization**: Docker & Docker Compose
- **Reverse Proxy**: NGINX
- **Libraries**:
  - Gorilla Mux (HTTP routing)
  - PostgreSQL Driver (lib/pq)
  - JWT Authentication (golang-jwt/jwt)
  - Bcrypt (password hashing)

## 📋 Prerequisites

Before running this project, make sure you have:

- [Docker](https://docs.docker.com/get-docker/) installed
- [Docker Compose](https://docs.docker.com/compose/install/) installed
- [Go 1.23+](https://golang.org/dl/) (for local development)
- [dbmate](https://github.com/amacneil/dbmate) (for database migrations)

### Install dbmate:
```bash
# Using go install
go install github.com/amacneil/dbmate/v2@latest

# Or using curl (Linux/macOS)
sudo curl -fsSL -o /usr/local/bin/dbmate https://github.com/amacneil/dbmate/releases/latest/download/dbmate-linux-amd64
sudo chmod +x /usr/local/bin/dbmate
```

## 🚀 Quick Start

### 1. Clone the Repository
```bash
git clone https://github.com/hariomop12/MicroService.git
cd MicroService
```

### 2. Start All Services
```bash
# Build and start all containers
docker-compose up --build -d

# Check service status
docker-compose ps
```

### 3. Database Migration with dbmate

**Set up all databases:**
```bash
# Users Database
export DATABASE_URL="postgresql://postgres:postgres@localhost:5432/users_db?sslmode=disable"
dbmate up

# Products Database  
export DATABASE_URL="postgresql://postgres:postgres@localhost:5432/products_db?sslmode=disable"
dbmate up

# Orders Database
export DATABASE_URL="postgresql://postgres:postgres@localhost:5432/orders_db?sslmode=disable"
dbmate up
```

**Alternative - Manual Database Setup:**
```bash
# Create tables manually if dbmate migration fails
docker exec -it postgres_main psql -U postgres -d users_db -c "
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);"

docker exec -it postgres_main psql -U postgres -d products_db -c "
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL,
    stock_quantity INTEGER NOT NULL DEFAULT 0,
    category VARCHAR(100),
    tags TEXT[],
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);"

docker exec -it postgres_main psql -U postgres -d orders_db -c "
CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    total_amount DECIMAL(10, 2) NOT NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS order_items (
    id SERIAL PRIMARY KEY,
    order_id INTEGER REFERENCES orders(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL,
    quantity INTEGER NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);"
```

### 4. Insert Sample Data
```bash
# Add sample products
docker exec -it postgres_main psql -U postgres -d products_db -c "
INSERT INTO products (name, description, price, stock_quantity, category, tags) VALUES
('Laptop Pro 15', 'High-performance laptop with 16GB RAM', 1299.99, 50, 'Electronics', ARRAY['laptop', 'computer', 'electronics']),
('Wireless Mouse', 'Ergonomic wireless mouse with USB receiver', 29.99, 200, 'Accessories', ARRAY['mouse', 'wireless', 'accessories']),
('Mechanical Keyboard', 'RGB mechanical keyboard with blue switches', 89.99, 100, 'Accessories', ARRAY['keyboard', 'mechanical', 'rgb']),
('USB-C Hub', '7-in-1 USB-C hub with HDMI and ethernet', 49.99, 150, 'Accessories', ARRAY['hub', 'usb-c', 'adapter']),
('Monitor 27', '4K IPS monitor with HDR support', 399.99, 75, 'Electronics', ARRAY['monitor', 'display', '4k'])
ON CONFLICT DO NOTHING;"
```

## 🌐 Service Endpoints

| Service | Port | Endpoint |
|---------|------|----------|
| API Gateway | 8000 | http://localhost:8000 |
| User Service | 8001 | http://localhost:8001 |
| Product Service | 8002 | http://localhost:8002 |
| Order Service | 8003 | http://localhost:8003 |
| NGINX | 80 | http://localhost |
| PostgreSQL | 5432 | localhost:5432 |

## 📡 API Documentation

### Health Checks
```bash
curl http://localhost:8001/health  # User Service
curl http://localhost:8002/health  # Product Service  
curl http://localhost:8003/health  # Order Service
```

## 🧪 Testing with cURL

### 1. User Service APIs

#### Register a New User
```bash
curl -X POST http://localhost:8001/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john.doe@example.com",
    "username": "johndoe", 
    "password": "securepassword123",
    "full_name": "John Doe"
  }'
```

#### Login User
```bash
curl -X POST http://localhost:8001/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john.doe@example.com",
    "password": "securepassword123"
  }'
```

#### Get User by ID
```bash
curl -X GET http://localhost:8001/users/1
```

#### Search Users
```bash
curl -X GET "http://localhost:8001/users/search?q=john"
```

### 2. Product Service APIs

#### Create a Product
```bash
curl -X POST http://localhost:8002/api/products \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Gaming Chair",
    "description": "Ergonomic gaming chair with RGB lighting",
    "price": 299.99,
    "stock_quantity": 25,
    "category": "Furniture",
    "tags": ["gaming", "chair", "rgb", "ergonomic"]
  }'
```

#### Get Product by ID
```bash
curl -X GET http://localhost:8002/api/products/1
```

#### Search Products
```bash
curl -X GET "http://localhost:8002/api/products/search?q=laptop"
```

#### Search Products by Tags
```bash
curl -X GET "http://localhost:8002/api/products/tags?tags=laptop,computer"
```

#### Update Product Stock
```bash
curl -X PATCH http://localhost:8002/api/products/1/stock \
  -H "Content-Type: application/json" \
  -d '{
    "quantity": 45
  }'
```

### 3. Order Service APIs

#### Create an Order
```bash
curl -X POST http://localhost:8003/api/orders \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "items": [
      {
        "product_id": 1,
        "quantity": 2
      },
      {
        "product_id": 2,
        "quantity": 1
      }
    ]
  }'
```

#### Get Order by ID
```bash
curl -X GET http://localhost:8003/api/orders/1
```

#### Get Orders by User ID
```bash
curl -X GET http://localhost:8003/api/orders/user/1
```

#### Update Order Status
```bash
curl -X PATCH http://localhost:8003/api/orders/1/status \
  -H "Content-Type: application/json" \
  -d '{
    "status": "shipped"
  }'
```

### 4. API Gateway (Proxy Routes)

#### User Registration via Gateway
```bash
curl -X POST http://localhost:8000/api/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "jane.smith@example.com",
    "username": "janesmith",
    "password": "mypassword456", 
    "full_name": "Jane Smith"
  }'
```

#### Product Creation via Gateway
```bash
curl -X POST http://localhost:8000/api/products \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Wireless Headphones",
    "description": "Noise-cancelling wireless headphones",
    "price": 199.99,
    "stock_quantity": 80,
    "category": "Electronics",
    "tags": ["headphones", "wireless", "audio"]
  }'
```

#### Order Creation via Gateway
```bash
curl -X POST http://localhost:8000/api/orders \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "items": [
      {
        "product_id": 3,
        "quantity": 1
      }
    ]
  }'
```

## 🧪 Complete Test Flow

Run this sequence to test the entire system:

```bash
# 1. Register a user
curl -X POST http://localhost:8001/register \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "username": "testuser", "password": "password123", "full_name": "Test User"}'

# 2. Create a product  
curl -X POST http://localhost:8002/api/products \
  -H "Content-Type: application/json" \
  -d '{"name": "Test Product", "description": "A test product", "price": 99.99, "stock_quantity": 10, "category": "Test", "tags": ["test"]}'

# 3. Create an order
curl -X POST http://localhost:8003/api/orders \
  -H "Content-Type: application/json" \
  -d '{"user_id": 1, "items": [{"product_id": 1, "quantity": 2}]}'

# 4. Check the order
curl -X GET http://localhost:8003/api/orders/1

# 5. Update order status
curl -X PATCH http://localhost:8003/api/orders/1/status \
  -H "Content-Type: application/json" \
  -d '{"status": "processing"}'
```

## 📁 Project Structure

```
MicroService/
├── api-getway/                 # API Gateway service
│   ├── api_getway.go
│   ├── Dockerfile
│   ├── go.mod
│   └── go.sum
├── db/
│   └── migrations/             # Database migrations
│       ├── 20251005121033_m1.sql
│       └── 20251005121034_users.sql
├── nginx/                      # NGINX configuration
│   ├── nginx.conf
│   └── logs/
├── order-service/              # Order management service
│   ├── order_service.go
│   ├── Dockerfile
│   ├── go.mod
│   └── go.sum
├── product-service/            # Product catalog service
│   ├── product_service.go
│   ├── Dockerfile
│   ├── go.mod
│   └── go.sum
├── user-service/              # User management service
│   ├── main.go
│   ├── Dockerfile
│   ├── go.mod
│   └── go.sum
├── docker-compose.yml         # Docker composition
├── init-databases.sh          # Database initialization script
└── README.md                  # This file
```

## 🔧 Services

### User Service (Port 8001)
- User registration and authentication
- JWT token generation
- User profile management
- Password hashing with bcrypt

### Product Service (Port 8002)  
- Product catalog management
- Full-text search capabilities
- Tag-based filtering
- Stock management

### Order Service (Port 8003)
- Order creation and management
- Order status tracking
- Integration with Product Service for pricing
- Order history by user

### API Gateway (Port 8000)
- Request routing and load balancing
- Centralized logging
- Rate limiting capabilities
- Service discovery

## 🛠️ Development

### Local Development Setup
```bash
# Install dependencies for each service
cd user-service && go mod tidy
cd ../product-service && go mod tidy  
cd ../order-service && go mod tidy
cd ../api-getway && go mod tidy

# Run services locally (requires PostgreSQL running)
cd user-service && go run main.go      # Port 8001
cd product-service && go run *.go      # Port 8002
cd order-service && go run *.go        # Port 8003
cd api-getway && go run *.go           # Port 8000
```

### Database Management
```bash
# Connect to PostgreSQL
docker exec -it postgres_main psql -U postgres

# View databases
\l

# Connect to specific database
\c users_db

# View tables
\dt

# View table structure
\d users
```

## 🐛 Troubleshooting

### Common Issues

#### 1. Containers Not Starting
```bash
# Check container logs
docker-compose logs -f

# Rebuild containers
docker-compose down
docker-compose up --build -d
```

#### 2. Database Connection Issues
```bash
# Check PostgreSQL health
docker exec postgres_main pg_isready -U postgres

# Restart database
docker-compose restart postgres
```

#### 3. Port Conflicts
```bash
# Check what's using ports
sudo netstat -tulpn | grep :8001
sudo netstat -tulpn | grep :5432

# Stop conflicting services
sudo systemctl stop postgresql  # If local PostgreSQL is running
```

#### 4. Migration Issues
```bash
# Reset migrations
dbmate rollback
dbmate up

# Check migration status
dbmate status
```

### Service Health Checks
```bash
# Check all service health
curl http://localhost:8001/health && echo ""
curl http://localhost:8002/health && echo ""  
curl http://localhost:8003/health && echo ""
```

## 📝 Environment Variables

Create a `.env` file for custom configuration:
```bash
# Database
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_MULTIPLE_DATABASES=users_db,products_db,orders_db

# Services
USER_SERVICE_PORT=8001
PRODUCT_SERVICE_PORT=8002
ORDER_SERVICE_PORT=8003
API_GATEWAY_PORT=8000

# JWT Secret
JWT_SECRET=your-super-secret-jwt-key-change-in-production
```

## 🚀 Deployment

### Production Deployment
1. Update environment variables
2. Configure proper secrets management
3. Set up SSL certificates
4. Configure production database
5. Set up monitoring and logging

### Docker Hub Deployment
```bash
# Build and tag images
docker build -t yourusername/user-service ./user-service
docker build -t yourusername/product-service ./product-service
docker build -t yourusername/order-service ./order-service
docker build -t yourusername/api-gateway ./api-getway

# Push to Docker Hub
docker push yourusername/user-service
docker push yourusername/product-service
docker push yourusername/order-service
docker push yourusername/api-gateway
```

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 👨‍💻 Author

**Hariom Prajapati**
- GitHub: [@hariomop12](https://github.com/hariomop12)
- Repository: [MicroService](https://github.com/hariomop12/MicroService)

---

## 🎯 Next Steps / Roadmap

- [ ] Add authentication middleware to API Gateway
- [ ] Implement Redis caching for frequently accessed data
- [ ] Add Prometheus metrics and Grafana dashboards  
- [ ] Implement event-driven architecture with message queues
- [ ] Add comprehensive unit and integration tests
- [ ] Set up CI/CD pipeline with GitHub Actions
- [ ] Add API versioning support
- [ ] Implement distributed tracing with Jaeger
- [ ] Add rate limiting and circuit breaker patterns
- [ ] Create Kubernetes deployment manifests

---

🎉 **Happy Coding!** If you find this project helpful, please give it a ⭐️ on GitHub!