#!/bin/bash

# Database setup script for GoGent
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}GoGent Database Setup Script${NC}"
echo "================================="

# Check if config.env exists
if [ ! -f "config.env" ]; then
    echo -e "${RED}Error: config.env file not found!${NC}"
    echo "Please copy config.example.env to config.env and edit it with your settings."
    exit 1
fi

# Source the config file
source config.env

# Check required variables
if [ -z "$DB_HOST" ] || [ -z "$DB_USER" ] || [ -z "$DB_NAME" ]; then
    echo -e "${RED}Error: Missing required database configuration!${NC}"
    echo "Please ensure DB_HOST, DB_USER, and DB_NAME are set in config.env"
    exit 1
fi

# Check if MySQL is available
if ! command -v mysql &> /dev/null; then
    echo -e "${RED}Error: MySQL client not found!${NC}"
    echo "Please install MySQL client to continue."
    exit 1
fi

echo -e "${YELLOW}Database configuration:${NC}"
echo "Host: $DB_HOST"
echo "Port: ${DB_PORT:-3306}"
echo "User: $DB_USER"
echo "Database: $DB_NAME"
echo ""

# Prompt for password if not set
if [ -z "$DB_PASSWORD" ]; then
    echo -n "Please enter MySQL password for user $DB_USER: "
    read -s DB_PASSWORD
    echo ""
fi

# Test database connection
echo -e "${YELLOW}Testing database connection...${NC}"
if mysql -h "$DB_HOST" -P "${DB_PORT:-3306}" -u "$DB_USER" -p"$DB_PASSWORD" -e "SELECT 1;" &> /dev/null; then
    echo -e "${GREEN}✓ Database connection successful!${NC}"
else
    echo -e "${RED}✗ Database connection failed!${NC}"
    echo "Please check your database credentials and ensure MySQL is running."
    exit 1
fi

# Create database if it doesn't exist
echo -e "${YELLOW}Creating database if it doesn't exist...${NC}"
mysql -h "$DB_HOST" -P "${DB_PORT:-3306}" -u "$DB_USER" -p"$DB_PASSWORD" -e "CREATE DATABASE IF NOT EXISTS \`$DB_NAME\`;"
echo -e "${GREEN}✓ Database '$DB_NAME' is ready!${NC}"

# Run schema migration
echo -e "${YELLOW}Running database schema migration...${NC}"
mysql -h "$DB_HOST" -P "${DB_PORT:-3306}" -u "$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" < sql/schema.sql
echo -e "${GREEN}✓ Schema migration completed!${NC}"

# Verify tables were created
echo -e "${YELLOW}Verifying table creation...${NC}"
TABLES=$(mysql -h "$DB_HOST" -P "${DB_PORT:-3306}" -u "$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" -e "SHOW TABLES;" | tail -n +2)
if [ -n "$TABLES" ]; then
    echo -e "${GREEN}✓ Tables created successfully:${NC}"
    echo "$TABLES" | sed 's/^/  - /'
else
    echo -e "${RED}✗ No tables found! Schema migration may have failed.${NC}"
    exit 1
fi

echo ""
echo -e "${GREEN}🎉 Database setup completed successfully!${NC}"
echo ""
echo "You can now:"
echo "  - Run 'make generate-db' to generate Go database code"
echo "  - Run 'make run' to start the GoGent application"
echo "  - Run 'make run-tests' to execute tests"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo "1. Ensure your GEMINI_API_KEY is set in config.env"
echo "2. Run: make setup"
echo "3. Run: make run" 