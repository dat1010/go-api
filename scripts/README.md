# Database Scripts

## Turso to Aurora Data Migration

Use the one-off data migration helper to copy data from Turso into Aurora Postgres.

Location: `scripts/turso_migrate`

### What it does

- Reads all rows from the `posts` table in Turso
- Upserts them into Aurora Postgres
- Safe to re-run (idempotent by `id`)

### Environment Variables

- `TURSO_DATABASE_URL`: Your Turso database URL
- `TURSO_AUTH_TOKEN`: Your Turso auth token
- `PG_DSN`: Postgres DSN (e.g., `postgres://user:pass@host:port/app?sslmode=require`)
- `DB_SECRET_ARN`: Secrets Manager ARN for Aurora (used if `PG_DSN` is not set)

If the Turso variables are not set, the script logs a skip message and exits without failing.

### Run it

```bash
cd scripts/turso_migrate
go run .
```

### Notes

- Make sure migrations have already run on Aurora so the `posts` table exists.
- For large datasets, run from a trusted host with stable network access.
- If you cannot reach Aurora from your laptop, run the GitHub Actions workflow `Turso Data Migration` to execute the migration as a one-off ECS task inside the VPC.
