-- Execution Logs queries

-- name: CreateExecutionLog :exec
INSERT INTO execution_logs (
    id, execution_run_id, configuration_id, request_id, 
    log_level, log_category, message, details
) VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetExecutionLogsByRun :many
SELECT 
    id, execution_run_id, configuration_id, request_id,
    log_level, log_category, message, 
    COALESCE(details, JSON_OBJECT()) as details,
    timestamp
FROM execution_logs 
WHERE execution_run_id = ?
ORDER BY timestamp ASC;

-- name: GetExecutionLogsByConfiguration :many
SELECT * FROM execution_logs 
WHERE execution_run_id = ? AND configuration_id = ?
ORDER BY timestamp ASC;

-- name: GetExecutionLogsByRequest :many
SELECT * FROM execution_logs 
WHERE execution_run_id = ? AND request_id = ?
ORDER BY timestamp ASC;

-- name: DeleteExecutionLogsByRun :exec
DELETE FROM execution_logs WHERE execution_run_id = ?;

-- name: CountExecutionLogsByLevel :one
SELECT log_level, COUNT(*) as count
FROM execution_logs 
WHERE execution_run_id = ?
GROUP BY log_level; 