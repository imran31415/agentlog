-- name: CreateAPIRequest :exec
INSERT INTO api_requests (
    id, user_id, execution_run_id, configuration_id, request_type, prompt,
    context, function_name, function_parameters, request_headers, request_body
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetAPIRequest :one
SELECT * FROM api_requests
WHERE id = ? AND user_id = ?;

-- name: GetAPIRequestsByRun :many
SELECT * FROM api_requests
WHERE execution_run_id = ? AND user_id = ?
ORDER BY created_at;

-- name: GetAPIRequestsByConfiguration :many
SELECT * FROM api_requests
WHERE configuration_id = ? AND user_id = ?
ORDER BY created_at;

-- name: GetAPIRequestsByType :many
SELECT * FROM api_requests
WHERE request_type = ? AND user_id = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: ListAPIRequests :many
SELECT * FROM api_requests
WHERE user_id = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: UpdateAPIRequest :exec
UPDATE api_requests
SET prompt = ?, context = ?, function_name = ?, function_parameters = ?,
    request_headers = ?, request_body = ?
WHERE id = ? AND user_id = ?;

-- name: DeleteAPIRequest :exec
DELETE FROM api_requests
WHERE id = ? AND user_id = ?;

-- name: CountAPIRequestsByUser :one
SELECT COUNT(*) FROM api_requests WHERE user_id = ?; 