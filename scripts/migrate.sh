#!/bin/bash

# Database migration script for Turso
# This script applies pending migrations to the Turso database

set -e  # Exit on any error

# Turso connection parameters
TURSO_DATABASE_URL="${TURSO_DATABASE_URL}"
TURSO_AUTH_TOKEN="${TURSO_AUTH_TOKEN}"

if [ -z "$TURSO_DATABASE_URL" ] || [ -z "$TURSO_AUTH_TOKEN" ]; then
    echo "Error: TURSO_DATABASE_URL and TURSO_AUTH_TOKEN must be set"
    exit 1
fi

echo "Starting Turso database migrations..."

# Function to run a migration
run_migration() {
    local migration_file=$1
    local migration_name=$(basename "$migration_file" .sql)
    
    echo "Checking if migration $migration_name has been applied..."
    
    # Check if migration has already been applied
    if turso db shell --url "$TURSO_DATABASE_URL" --auth-token "$TURSO_AUTH_TOKEN" -c "SELECT 1 FROM migrations WHERE name = '$migration_name' LIMIT 1;" 2>/dev/null | grep -q 1; then
        echo "Migration $migration_name already applied, skipping..."
        return 0
    fi
    
    echo "Applying migration: $migration_name"
    
    # Apply the migration
    turso db shell --url "$TURSO_DATABASE_URL" --auth-token "$TURSO_AUTH_TOKEN" < "$migration_file"
    
    # Record the migration as applied
    turso db shell --url "$TURSO_DATABASE_URL" --auth-token "$TURSO_AUTH_TOKEN" -c "INSERT INTO migrations (name, applied_at) VALUES ('$migration_name', datetime('now'));"
    
    echo "Migration $migration_name applied successfully"
}

# Create migrations table if it doesn't exist
echo "Ensuring migrations table exists..."
turso db shell --url "$TURSO_DATABASE_URL" --auth-token "$TURSO_AUTH_TOKEN" -c "
CREATE TABLE IF NOT EXISTS migrations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
);" 2>/dev/null || echo "Migrations table already exists or could not be created"

# Apply migrations in order
echo "Applying pending migrations..."

# List of migrations in order (add new migrations here)
MIGRATIONS=(
    "migrations/002_remove_published_field.sql"
)

for migration in "${MIGRATIONS[@]}"; do
    if [ -f "$migration" ]; then
        run_migration "$migration"
    else
        echo "Warning: Migration file $migration not found, skipping..."
    fi
done

echo "All migrations completed successfully!" 