#!/bin/bash

# Function Call History Checker for GoGent
set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}🔍 GoGent Function Call History Checker${NC}"
echo "========================================"

# Source config if available
if [ -f "config.env" ]; then
    source config.env
else
    echo -e "${RED}❌ config.env not found! Please ensure it exists.${NC}"
    exit 1
fi

# Build MySQL connection string
MYSQL_CMD="mysql -h ${DB_HOST:-localhost} -P ${DB_PORT:-3306} -u ${DB_USER} -p${DB_PASSWORD} ${DB_NAME}"

echo -e "${YELLOW}📊 Recent Execution Runs (with function calling status):${NC}"
$MYSQL_CMD -e "
SELECT 
    id,
    name,
    enable_function_calling,
    created_at
FROM execution_runs 
ORDER BY created_at DESC 
LIMIT 10;" 2>/dev/null || echo "❌ Could not connect to database"

echo ""
echo -e "${YELLOW}🔧 Function Calls from Recent Executions:${NC}"
$MYSQL_CMD -e "
SELECT 
    fc.function_name,
    fc.execution_status,
    fc.execution_time_ms,
    ar.prompt,
    er.name as execution_name,
    fc.created_at
FROM function_calls fc
JOIN api_requests ar ON fc.request_id = ar.id
JOIN execution_runs er ON ar.execution_run_id = er.id
ORDER BY fc.created_at DESC
LIMIT 20;" 2>/dev/null || echo "❌ Could not query function calls"

echo ""
echo -e "${YELLOW}🌤️ Weather-related Function Calls:${NC}"
$MYSQL_CMD -e "
SELECT 
    fc.function_name,
    fc.function_arguments,
    fc.function_response,
    fc.execution_status,
    ar.prompt,
    er.name as execution_name,
    fc.created_at
FROM function_calls fc
JOIN api_requests ar ON fc.request_id = ar.id
JOIN execution_runs er ON ar.execution_run_id = er.id
WHERE ar.prompt LIKE '%weather%' 
   OR ar.prompt LIKE '%LA%'
   OR ar.prompt LIKE '%Los Angeles%'
   OR fc.function_name LIKE '%weather%'
ORDER BY fc.created_at DESC;" 2>/dev/null || echo "❌ Could not query weather-related calls"

echo ""
echo -e "${YELLOW}📈 Function Call Statistics:${NC}"
$MYSQL_CMD -e "
SELECT 
    fc.function_name,
    COUNT(*) as total_calls,
    COUNT(CASE WHEN fc.execution_status = 'success' THEN 1 END) as successful_calls,
    COUNT(CASE WHEN fc.execution_status = 'error' THEN 1 END) as failed_calls,
    AVG(fc.execution_time_ms) as avg_execution_time_ms
FROM function_calls fc
GROUP BY fc.function_name
ORDER BY total_calls DESC;" 2>/dev/null || echo "❌ Could not get function call stats"

echo ""
echo -e "${BLUE}💡 Tips:${NC}"
echo "• If no function calls appear above, the AI might not have triggered any tools"
echo "• Check if you have function definitions configured in the database"
echo "• Weather queries typically need a weather API function to be defined"
echo "• Function calls only happen when the AI determines they're needed" 