# Database Migrations

This directory contains scripts for managing database migrations for **Turso** (SQLite-based database).

## Migration Script (`migrate.sh`)

The `migrate.sh` script automatically applies pending database migrations during deployment using the Turso CLI.

### Features

- **Idempotent**: Safe to run multiple times - won't apply the same migration twice
- **Tracked**: Uses a `migrations` table to track which migrations have been applied
- **Ordered**: Applies migrations in the order specified in the script
- **Error handling**: Exits on any error to prevent partial migrations
- **Turso-native**: Uses Turso CLI for database operations

### How it works

1. Creates a `migrations` table if it doesn't exist
2. Checks which migrations have already been applied
3. Applies only pending migrations using Turso CLI
4. Records each migration as applied after successful execution

### Adding new migrations

1. Create your SQL migration file in the `migrations/` directory
2. Add the migration file path to the `MIGRATIONS` array in `migrate.sh`
3. Commit and push - the migration will run automatically on next deployment

### Environment Variables

The script uses these environment variables (set in GitHub Actions secrets):

- `TURSO_DATABASE_URL`: Your Turso database URL
- `TURSO_AUTH_TOKEN`: Your Turso authentication token

### Manual execution

You can run the script manually for testing:

```bash
export TURSO_DATABASE_URL="libsql://your-database-url"
export TURSO_AUTH_TOKEN="your-auth-token"

chmod +x scripts/migrate.sh
./scripts/migrate.sh
```

### Migration files

Migration files should be placed in the `migrations/` directory with descriptive names:

- `001_create_posts_table.sql`
- `002_remove_published_field.sql`
- `003_add_user_profiles.sql`

Each migration should be self-contained and use SQLite syntax (not MySQL).

### Turso CLI Installation

The deployment workflow automatically installs the Turso CLI. For local development:

```bash
curl -sSfL https://get.tur.so/install.sh | bash
export PATH="$HOME/.turso:$PATH"
```

### SQLite vs MySQL Differences

Since Turso uses SQLite, migrations should use SQLite syntax:

- `INTEGER PRIMARY KEY AUTOINCREMENT` instead of `INT AUTO_INCREMENT PRIMARY KEY`
- `TEXT` instead of `VARCHAR`
- `DATETIME('now')` instead of `NOW()`
- `CREATE TABLE IF NOT EXISTS` for idempotent table creation 