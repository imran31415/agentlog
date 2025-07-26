-- name: CreateAPIResponse :exec
INSERT INTO api_responses (
    id, user_id, request_id, response_status, response_text, function_call_response,
    usage_metadata, safety_ratings, finish_reason, error_message,
    response_time_ms, response_headers, response_body
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetAPIResponse :one
SELECT * FROM api_responses
WHERE id = ? AND user_id = ?;

-- name: GetAPIResponseByRequest :one
SELECT * FROM api_responses
WHERE request_id = ? AND user_id = ?;

-- name: GetAPIResponsesByStatus :many
SELECT * FROM api_responses
WHERE response_status = ? AND user_id = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: GetAPIResponsesByTimeRange :many
SELECT * FROM api_responses
WHERE created_at BETWEEN ? AND ?
ORDER BY created_at DESC;

-- name: GetAPIResponsesWithRequests :many
SELECT 
    r.id, r.user_id, r.request_id, r.response_status, r.response_text,
    r.function_call_response, r.usage_metadata, r.safety_ratings,
    r.finish_reason, r.error_message, r.response_time_ms,
    r.response_headers, r.response_body, r.created_at
FROM api_responses r
JOIN api_requests req ON r.request_id = req.id
WHERE req.execution_run_id = ? AND r.user_id = ?
ORDER BY r.created_at;

-- name: ListAPIResponses :many
SELECT * FROM api_responses
WHERE user_id = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: UpdateAPIResponse :exec
UPDATE api_responses
SET response_status = ?, response_text = ?, function_call_response = ?,
    usage_metadata = ?, safety_ratings = ?, finish_reason = ?, error_message = ?,
    response_time_ms = ?, response_headers = ?, response_body = ?
WHERE id = ? AND user_id = ?;

-- name: DeleteAPIResponse :exec
DELETE FROM api_responses
WHERE id = ? AND user_id = ?; 

-- name: CountAPIResponsesByUser :one
SELECT COUNT(*) FROM api_responses WHERE user_id = ?; 