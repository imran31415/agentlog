-- Function Calls queries

-- name: CreateFunctionCall :exec
INSERT INTO function_calls (
    id, request_id, function_name, function_arguments, function_response,
    execution_status, execution_time_ms, error_details
) VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetFunctionCall :one
SELECT * FROM function_calls WHERE id = ?;

-- name: ListFunctionCallsByRequest :many
SELECT * FROM function_calls 
WHERE request_id = ?
ORDER BY created_at ASC;

-- name: ListFunctionCallsByExecution :many
SELECT fc.*, ar.prompt, ar.created_at as request_created_at
FROM function_calls fc
JOIN api_requests ar ON fc.request_id = ar.id
WHERE ar.execution_run_id = ?
ORDER BY fc.created_at DESC;

-- name: UpdateFunctionCall :exec
UPDATE function_calls 
SET function_response = ?, execution_status = ?, 
    execution_time_ms = ?, error_details = ?
WHERE id = ?;

-- name: GetFunctionCallStats :one
SELECT 
    COUNT(*) as total_calls,
    COUNT(CASE WHEN execution_status = 'success' THEN 1 END) as successful_calls,
    COUNT(CASE WHEN execution_status = 'error' THEN 1 END) as failed_calls,
    AVG(execution_time_ms) as avg_execution_time,
    MAX(execution_time_ms) as max_execution_time,
    MIN(execution_time_ms) as min_execution_time
FROM function_calls 
WHERE request_id IN (
    SELECT id FROM api_requests WHERE execution_run_id = ?
);

-- name: GetFunctionCallsByName :many
SELECT fc.*, ar.execution_run_id, ar.prompt
FROM function_calls fc
JOIN api_requests ar ON fc.request_id = ar.id
WHERE fc.function_name = ?
ORDER BY fc.created_at DESC
LIMIT ?;

-- name: GetRecentFunctionCalls :many
SELECT fc.*, ar.execution_run_id, ar.prompt, er.name as execution_name
FROM function_calls fc
JOIN api_requests ar ON fc.request_id = ar.id
JOIN execution_runs er ON ar.execution_run_id = er.id
ORDER BY fc.created_at DESC
LIMIT ?;

-- name: DeleteFunctionCallsByRequest :exec
DELETE FROM function_calls WHERE request_id = ?; 