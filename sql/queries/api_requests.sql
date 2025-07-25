-- name: CreateAPIRequest :exec
INSERT INTO api_requests (
    id, execution_run_id, configuration_id, request_type, prompt,
    context, function_name, function_parameters, request_headers, request_body
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetAPIRequest :one
SELECT * FROM api_requests
WHERE id = ?;

-- name: GetAPIRequestsByRun :many
SELECT * FROM api_requests
WHERE execution_run_id = ?
ORDER BY created_at;

-- name: GetAPIRequestsByConfiguration :many
SELECT * FROM api_requests
WHERE configuration_id = ?
ORDER BY created_at;

-- name: GetAPIRequestsByType :many
SELECT * FROM api_requests
WHERE request_type = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: ListAPIRequests :many
SELECT * FROM api_requests
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: UpdateAPIRequest :exec
UPDATE api_requests
SET prompt = ?, context = ?, function_name = ?, function_parameters = ?,
    request_headers = ?, request_body = ?
WHERE id = ?;

-- name: DeleteAPIRequest :exec
DELETE FROM api_requests
WHERE id = ?; 