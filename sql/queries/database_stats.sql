-- Database Statistics queries (user-specific)

-- name: GetUserExecutionRunsCount :one
SELECT COUNT(*) FROM execution_runs WHERE user_id = ?;

-- name: GetUserAPIRequestsCount :one  
SELECT COUNT(*) FROM api_requests WHERE user_id = ?;

-- name: GetUserAPIResponsesCount :one
SELECT COUNT(*) FROM api_responses WHERE user_id = ?;

-- name: GetUserFunctionCallsCount :one
SELECT COUNT(*) FROM function_calls WHERE user_id = ?;

-- name: GetUserAvgResponseTime :one
SELECT COALESCE(AVG(response_time_ms), 0) FROM api_responses WHERE user_id = ?;

-- name: GetUserSuccessRate :one
SELECT COALESCE(
    (COUNT(CASE WHEN response_status = 'success' THEN 1 END) * 100.0 / NULLIF(COUNT(*), 0)), 
    0
) as success_rate FROM api_responses WHERE user_id = ?;

-- name: GetUserExecutionStats :one
SELECT 
    COUNT(*) as total_executions,
    COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_executions,
    COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_executions,
    COUNT(CASE WHEN status = 'running' THEN 1 END) as running_executions,
    COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending_executions
FROM execution_runs 
WHERE user_id = ?;

-- name: GetUserAPIRequestStats :one
SELECT 
    COUNT(*) as total_requests,
    COUNT(CASE WHEN request_type = 'generate' THEN 1 END) as generate_requests,
    COUNT(CASE WHEN request_type = 'chat' THEN 1 END) as chat_requests,
    COUNT(CASE WHEN request_type = 'function_call' THEN 1 END) as function_call_requests
FROM api_requests 
WHERE user_id = ?;

-- name: GetUserFunctionCallStats :one
SELECT 
    COUNT(*) as total_function_calls,
    COUNT(CASE WHEN execution_status = 'success' THEN 1 END) as successful_calls,
    COUNT(CASE WHEN execution_status = 'error' THEN 1 END) as failed_calls,
    COUNT(CASE WHEN execution_status = 'pending' THEN 1 END) as pending_calls,
    COALESCE(AVG(execution_time_ms), 0) as avg_execution_time
FROM function_calls 
WHERE user_id = ?;

-- name: GetUserActivityByDay :many
SELECT 
    DATE(created_at) as activity_date,
    COUNT(*) as execution_count
FROM execution_runs 
WHERE user_id = ? 
    AND created_at >= DATE_SUB(CURRENT_DATE, INTERVAL 30 DAY)
GROUP BY DATE(created_at)
ORDER BY activity_date DESC; 