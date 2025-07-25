-- name: CreateAPIResponse :exec
INSERT INTO api_responses (
    id, request_id, response_status, response_text, function_call_response,
    usage_metadata, safety_ratings, finish_reason, error_message,
    response_time_ms, response_headers, response_body
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetAPIResponse :one
SELECT * FROM api_responses
WHERE id = ?;

-- name: GetAPIResponseByRequest :one
SELECT * FROM api_responses
WHERE request_id = ?;

-- name: GetAPIResponsesByStatus :many
SELECT * FROM api_responses
WHERE response_status = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: GetAPIResponsesByTimeRange :many
SELECT * FROM api_responses
WHERE created_at BETWEEN ? AND ?
ORDER BY created_at DESC;

-- name: GetAPIResponsesWithRequests :many
SELECT 
    r.*,
    req.prompt,
    req.request_type,
    req.function_name,
    c.variation_name,
    c.model_name
FROM api_responses r
JOIN api_requests req ON r.request_id = req.id
JOIN api_configurations c ON req.configuration_id = c.id
WHERE req.execution_run_id = ?
ORDER BY r.created_at;

-- name: ListAPIResponses :many
SELECT * FROM api_responses
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: UpdateAPIResponse :exec
UPDATE api_responses
SET response_status = ?, response_text = ?, function_call_response = ?,
    usage_metadata = ?, safety_ratings = ?, finish_reason = ?,
    error_message = ?, response_time_ms = ?, response_headers = ?, response_body = ?
WHERE id = ?;

-- name: DeleteAPIResponse :exec
DELETE FROM api_responses
WHERE id = ?; 