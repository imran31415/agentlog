-- Function Definitions queries

-- name: CreateFunctionDefinition :exec
INSERT INTO function_definitions (
    id, name, display_name, description, parameters_schema, 
    mock_response, endpoint_url, http_method, headers, auth_config, is_active
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetFunctionDefinition :one
SELECT * FROM function_definitions WHERE id = ?;

-- name: GetFunctionDefinitionByName :one
SELECT * FROM function_definitions WHERE name = ?;

-- name: ListFunctionDefinitions :many
SELECT * FROM function_definitions 
WHERE is_active = TRUE 
ORDER BY display_name ASC;

-- name: ListAllFunctionDefinitions :many
SELECT * FROM function_definitions 
ORDER BY created_at DESC;

-- name: UpdateFunctionDefinition :exec
UPDATE function_definitions 
SET display_name = ?, description = ?, parameters_schema = ?, 
    mock_response = ?, endpoint_url = ?, http_method = ?, 
    headers = ?, auth_config = ?, is_active = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: DeleteFunctionDefinition :exec
UPDATE function_definitions 
SET is_active = FALSE, updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: PermanentDeleteFunctionDefinition :exec
DELETE FROM function_definitions WHERE id = ?;

-- name: SearchFunctionDefinitions :many
SELECT * FROM function_definitions 
WHERE is_active = TRUE 
AND (display_name LIKE ? OR description LIKE ? OR name LIKE ?)
ORDER BY display_name ASC;

-- name: GetFunctionDefinitionsForExecution :many
SELECT fd.*, efc.use_mock_response, efc.execution_order
FROM function_definitions fd
JOIN execution_function_configs efc ON fd.id = efc.function_definition_id
WHERE efc.execution_run_id = ?
AND fd.is_active = TRUE
ORDER BY efc.execution_order ASC, fd.display_name ASC; 