# Database Migrations

This directory contains database migration files for the gogent application. Migrations are automatically applied when the application starts, ensuring the database schema is always up to date.

## Migration Files

- `000_create_migrations_table.sql` - Creates the migration tracking table
- `001_initial_schema.sql` - Initial database schema with all tables

## How It Works

1. **Automatic Migration**: When the gogent client starts, it automatically runs all pending migrations
2. **Migration Tracking**: The `schema_migrations` table tracks which migrations have been applied
3. **Checksum Validation**: Each migration is checksummed to detect changes
4. **Transaction Safety**: Migrations run in transactions for rollback safety

## Running Migrations Manually

### Using Make Commands

```bash
# Run all pending migrations
make migrate

# Show migration status
make migrate-status

# Build migration tool
make build-migrate
```

### Using the Migration Tool Directly

```bash
# Run migrations
go run cmd/migrate/main.go

# Show status
go run cmd/migrate/main.go -status

# Use custom database URL
go run cmd/migrate/main.go -db-url="user:pass@tcp(localhost:3306)/dbname"

# Show help
go run cmd/migrate/main.go -help
```

## Creating New Migrations

1. Create a new SQL file in this directory with a descriptive name
2. Use the format: `XXX_description.sql` where XXX is a sequential number
3. Include a header comment with description and date
4. Test the migration before committing

Example migration file:

```sql
-- Migration: 002_add_user_table.sql
-- Description: Add user management table
-- Created: 2025-07-25
-- Status: Applied

CREATE TABLE users (
    id VARCHAR(36) PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## Migration Best Practices

1. **Always use transactions** - The migration system handles this automatically
2. **Test migrations** - Test on a copy of production data
3. **Keep migrations small** - One logical change per migration
4. **Use descriptive names** - Make it clear what each migration does
5. **Include rollback considerations** - Think about how to undo changes if needed

## Troubleshooting

### Migration Fails

If a migration fails:

1. Check the error message in the `schema_migrations` table
2. Fix the migration file
3. Re-run migrations - the system will reapply failed migrations

### Database Connection Issues

Ensure your database connection is working:

```bash
# Test connection
mysql -h localhost -u root -p -e "SELECT 1;"
```

### Migration Status

Check which migrations have been applied:

```bash
make migrate-status
```

## Schema Changes

When making schema changes:

1. Create a new migration file
2. Test the migration
3. Commit the migration file
4. Deploy - migrations will run automatically on startup

The migration system ensures that all environments (development, staging, production) stay in sync with the latest schema. 