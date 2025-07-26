-- Execution Runs queries

-- name: CreateExecutionRun :exec
INSERT INTO execution_runs (id, user_id, name, description, enable_function_calling)
VALUES (?, ?, ?, ?, ?);

-- name: GetExecutionRun :one
SELECT * FROM execution_runs WHERE id = ? AND user_id = ?;

-- name: GetRecentExecutionRuns :many
SELECT * FROM execution_runs
WHERE user_id = ?
ORDER BY created_at DESC
LIMIT ?;

-- name: GetExecutionRunsByUser :many
SELECT * FROM execution_runs
WHERE user_id = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: UpdateExecutionRunComplete :exec
UPDATE execution_runs
SET updated_at = CURRENT_TIMESTAMP
WHERE id = ? AND user_id = ?;

-- name: UpdateExecutionRunStatus :exec
UPDATE execution_runs
SET status = ?, error_message = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ? AND user_id = ?;

-- name: DeleteExecutionRun :exec
DELETE FROM execution_runs WHERE id = ? AND user_id = ?;

-- name: CountExecutionRunsByUser :one
SELECT COUNT(*) FROM execution_runs WHERE user_id = ?; 