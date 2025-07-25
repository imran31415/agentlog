# Database Migrations - Quick Reference

## Daily Usage Commands

```bash
# Run all pending migrations
make migrate

# Check migration status
make migrate-status

# Initialize database (old method - deprecated)
make init-db
```

## Creating New Migrations

1. Create a new file in `sql/migrations/` with format: `XXX_description.sql`
2. Include migration metadata at the top:
   ```sql
   -- Migration: XXX_description.sql
   -- Description: What this migration does
   -- Created: YYYY-MM-DD
   ```
3. Write your SQL statements
4. Run `make migrate` to apply

## Migration File Naming Convention

- `000_create_migrations_table.sql` - Migration tracking table
- `001_initial_schema.sql` - Initial database schema
- `002_add_user_preferences.sql` - Example future migration
- `003_update_function_definitions.sql` - Example future migration

## Automatic Migration on Server Start

The gogent server automatically runs migrations when it starts up. No manual intervention needed for production deployments.

## Status Meanings

- **applied**: Migration successfully executed
- **failed**: Migration encountered an error
- **rolled_back**: Migration was reverted

## Important Notes

- Migrations run in transactions for safety
- Each migration is checksummed to detect changes
- Never modify existing migration files once applied
- Always test migrations on development environment first 