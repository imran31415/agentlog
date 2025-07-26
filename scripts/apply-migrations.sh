#!/bin/bash

# Apply database migrations manually
# Usage: ./scripts/apply-migrations.sh

set -e

MYSQL_PASSWORD="Password123!"
DATABASE="gogent"

echo "🚀 Applying database migrations manually..."

# Check if database exists
echo "📊 Checking database connection..."
mysql -u root -p$MYSQL_PASSWORD -e "USE $DATABASE; SELECT 1;" > /dev/null 2>&1 || {
    echo "❌ Database '$DATABASE' not accessible. Creating it..."
    mysql -u root -p$MYSQL_PASSWORD -e "CREATE DATABASE IF NOT EXISTS $DATABASE;"
}

echo "✅ Database connection verified"

# Apply migrations in order
MIGRATIONS_DIR="migrations"
echo "📁 Looking for migrations in $MIGRATIONS_DIR/"

if [ ! -d "$MIGRATIONS_DIR" ]; then
    echo "❌ Migrations directory not found: $MIGRATIONS_DIR"
    exit 1
fi

# Apply each migration file in order
for migration in $(ls $MIGRATIONS_DIR/*.up.sql | sort); do
    echo "🔧 Applying migration: $(basename $migration)"
    mysql -u root -p$MYSQL_PASSWORD $DATABASE < "$migration"
    echo "✅ Applied: $(basename $migration)"
done

echo "🎉 All migrations applied successfully!"

# Show final stats
echo "📊 Database summary:"
mysql -u root -p$MYSQL_PASSWORD -e "
USE $DATABASE; 
SELECT 'Functions' as table_name, COUNT(*) as count FROM function_definitions
UNION ALL
SELECT 'Configurations' as table_name, COUNT(*) as count FROM api_configurations
UNION ALL  
SELECT 'Users' as table_name, COUNT(*) as count FROM users;
"

echo "✅ Migration process completed!" 