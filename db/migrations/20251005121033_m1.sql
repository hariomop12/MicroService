-- migrate:up
CREATE DATABASE users_db;
CREATE DATABASE products_db;
CREATE DATABASE orders_db;

-- ============================================
-- USERS DATABASE
-- ============================================
\c users_db;

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)

-- GIN index for full-text search on user profiles
CREATE INDEX idx_users_search ON users USING GIN(to_tsvector('english', full_name || ' ' || username || ' ' || email));

-- ============================================
-- PRODUCTS DATABASE
-- ============================================
\c products_db;

CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL,
    stock_quantity INTEGER NOT NULL DEFAULT 0,
    category VARCHAR(100),
    tags TEXT[],
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- GIN index for full-text search on products (name + description)
CREATE INDEX idx_products_search ON products USING GIN(to_tsvector('english', name || ' ' || COALESCE(description, '')));

-- GIN index for array search on tags
CREATE INDEX idx_products_tags ON products USING GIN(tags);

-- Regular B-tree index for category filtering
CREATE INDEX idx_products_category ON products(category);

-- ============================================
-- ORDERS DATABASE
-- ============================================
\c orders_db;
CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    total_amount DECIMAL(10, 2) NOT NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE order_items (
    id SERIAL PRIMARY KEY,
    order_id INTEGER REFERENCES orders(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL,
    quantity INTEGER NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- GIN index for JSONB metadata (if we want to add custom fields)
CREATE INDEX idx_orders_metadata ON orders USING GIN(metadata);
CREATE INDEX idx_order_items_order_id ON order_items(order_id);

-- Insert sample data
INSERT INTO products (name, description, price, stock_quantity, category, tags) VALUES
('Laptop Pro 15', 'High-performance laptop with 16GB RAM', 1299.99, 50, 'Electronics', ARRAY['laptop', 'computer', 'electronics']),
('Wireless Mouse', 'Ergonomic wireless mouse with USB receiver', 29.99, 200, 'Accessories', ARRAY['mouse', 'wireless', 'accessories']),
('Mechanical Keyboard', 'RGB mechanical keyboard with blue switches', 89.99, 100, 'Accessories', ARRAY['keyboard', 'mechanical', 'rgb']),
('USB-C Hub', '7-in-1 USB-C hub with HDMI and ethernet', 49.99, 150, 'Accessories', ARRAY['hub', 'usb-c', 'adapter']),
('Monitor 27"', '4K IPS monitor with HDR support', 399.99, 75, 'Electronics', ARRAY['monitor', 'display', '4k']);

-- migrate:down
DROP DATABASE orders_db;
DROP DATABASE products_db;
DROP DATABASE users_db;
