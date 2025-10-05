#!/bin/bash
set -e

# Create multiple databases
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE DATABASE users_db;
    CREATE DATABASE products_db;
    CREATE DATABASE orders_db;
EOSQL

echo "Multiple databases created successfully!"