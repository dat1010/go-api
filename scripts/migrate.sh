#!/bin/bash

# Database migration script for Turso
# This script applies pending migrations to the Turso database

set -e  # Exit on any error

# Turso connection parameters
TURSO_DATABASE_URL="${TURSO_DATABASE_URL}"

if [ -z "$TURSO_DATABASE_URL" ]; then
    echo "Error: TURSO_DATABASE_URL must be set"
    exit 1
fi

# Extract database name from URL
DATABASE_NAME=$(echo "$TURSO_DATABASE_URL" | sed 's|libsql://||' | sed 's|-dat1010\..*||')
echo "Using database: $DATABASE_NAME"

echo "Starting Turso database migrations..."

# Function to run a migration
run_migration() {
    local migration_file=$1
    local migration_name=$(basename "$migration_file" .sql)
    
    echo "Checking if migration $migration_name has been applied..."
    
    # Check if migration has already been applied
    if turso db shell "$DATABASE_NAME" -c "SELECT 1 FROM migrations WHERE name = '$migration_name' LIMIT 1;" 2>/dev/null | grep -q 1; then
        echo "Migration $migration_name already applied, skipping..."
        return 0
    fi
    
    echo "Applying migration: $migration_name"
    
    # Apply the migration
    turso db shell "$DATABASE_NAME" < "$migration_file"
    
    # Record the migration as applied
    turso db shell "$DATABASE_NAME" -c "INSERT INTO migrations (name, applied_at) VALUES ('$migration_name', datetime('now'));"
    
    echo "Migration $migration_name applied successfully"
}

# Create migrations table if it doesn't exist
echo "Ensuring migrations table exists..."
turso db shell "$DATABASE_NAME" -c "
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