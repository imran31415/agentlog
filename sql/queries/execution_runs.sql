-- Execution Runs queries

-- name: CreateExecutionRun :exec
INSERT INTO execution_runs (id, name, description, enable_function_calling)
VALUES (?, ?, ?, ?);

-- name: GetExecutionRun :one
SELECT * FROM execution_runs WHERE id = ?;

-- name: GetRecentExecutionRuns :many
SELECT * FROM execution_runs
ORDER BY created_at DESC
LIMIT ?;

-- name: UpdateExecutionRunComplete :exec
UPDATE execution_runs
SET updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: DeleteExecutionRun :exec
DELETE FROM execution_runs WHERE id = ?; 