-- migrate:up
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- GIN index for full-text search on user profiles
CREATE INDEX IF NOT EXISTS idx_users_search ON users USING GIN(to_tsvector('english', full_name || ' ' || username || ' ' || email));

-- migrate:down
DROP TABLE IF EXISTS users;