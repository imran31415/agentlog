-- name: CreateAPIConfiguration :exec
INSERT INTO api_configurations (
    id, user_id, execution_run_id, variation_name, model_name, system_prompt,
    temperature, max_tokens, top_p, top_k, safety_settings,
    generation_config, tools, tool_config
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetAPIConfiguration :one
SELECT id, user_id, execution_run_id, variation_name, model_name, system_prompt, temperature, max_tokens, top_p, top_k, safety_settings, generation_config, tools, tool_config, created_at FROM api_configurations
WHERE id = ? AND user_id = ?;

-- name: GetAPIConfigurationsByRun :many
SELECT id, user_id, execution_run_id, variation_name, model_name, system_prompt, temperature, max_tokens, top_p, top_k, safety_settings, generation_config, tools, tool_config, created_at FROM api_configurations
WHERE execution_run_id = ? AND user_id = ?
ORDER BY variation_name;

-- name: GetAPIConfigurationByVariation :one
SELECT id, user_id, execution_run_id, variation_name, model_name, system_prompt, temperature, max_tokens, top_p, top_k, safety_settings, generation_config, tools, tool_config, created_at FROM api_configurations
WHERE execution_run_id = ? AND variation_name = ? AND user_id = ?;

-- name: ListAPIConfigurations :many
SELECT id, user_id, execution_run_id, variation_name, model_name, system_prompt, temperature, max_tokens, top_p, top_k, safety_settings, generation_config, tools, tool_config, created_at FROM api_configurations
WHERE user_id = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: ListAPIConfigurationsByUser :many
SELECT id, user_id, execution_run_id, variation_name, model_name, system_prompt, temperature, max_tokens, top_p, top_k, safety_settings, generation_config, tools, tool_config, created_at FROM api_configurations
WHERE user_id = ?
ORDER BY created_at DESC;

-- name: UpdateAPIConfiguration :exec
UPDATE api_configurations
SET variation_name = ?, model_name = ?, system_prompt = ?,
    temperature = ?, max_tokens = ?, top_p = ?, top_k = ?,
    safety_settings = ?, generation_config = ?, tools = ?, tool_config = ?
WHERE id = ? AND user_id = ?;

-- name: DeleteAPIConfiguration :exec
DELETE FROM api_configurations
WHERE id = ? AND user_id = ?;

-- name: CountAPIConfigurationsByUser :one
SELECT COUNT(*) FROM api_configurations WHERE user_id = ?; 