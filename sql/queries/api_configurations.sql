-- name: CreateAPIConfiguration :exec
INSERT INTO api_configurations (
    id, execution_run_id, variation_name, model_name, system_prompt,
    temperature, max_tokens, top_p, top_k, safety_settings,
    generation_config, tools, tool_config
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetAPIConfiguration :one
SELECT * FROM api_configurations
WHERE id = ?;

-- name: GetAPIConfigurationsByRun :many
SELECT * FROM api_configurations
WHERE execution_run_id = ?
ORDER BY variation_name;

-- name: GetAPIConfigurationByVariation :one
SELECT * FROM api_configurations
WHERE execution_run_id = ? AND variation_name = ?;

-- name: ListAPIConfigurations :many
SELECT * FROM api_configurations
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: UpdateAPIConfiguration :exec
UPDATE api_configurations
SET variation_name = ?, model_name = ?, system_prompt = ?,
    temperature = ?, max_tokens = ?, top_p = ?, top_k = ?,
    safety_settings = ?, generation_config = ?, tools = ?, tool_config = ?
WHERE id = ?;

-- name: DeleteAPIConfiguration :exec
DELETE FROM api_configurations
WHERE id = ?; 