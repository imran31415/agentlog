-- Execution Function Configurations queries

-- name: CreateExecutionFunctionConfig :exec
INSERT INTO execution_function_configs (
    id, execution_run_id, function_definition_id, use_mock_response, execution_order
) VALUES (?, ?, ?, ?, ?);

-- name: GetExecutionFunctionConfig :one
SELECT * FROM execution_function_configs WHERE id = ?;

-- name: ListExecutionFunctionConfigs :many
SELECT efc.*, fd.name, fd.display_name, fd.description
FROM execution_function_configs efc
JOIN function_definitions fd ON efc.function_definition_id = fd.id
WHERE efc.execution_run_id = ?
ORDER BY efc.execution_order ASC, fd.display_name ASC;

-- name: UpdateExecutionFunctionConfig :exec
UPDATE execution_function_configs 
SET use_mock_response = ?, execution_order = ?
WHERE id = ?;

-- name: DeleteExecutionFunctionConfig :exec
DELETE FROM execution_function_configs 
WHERE execution_run_id = ? AND function_definition_id = ?;

-- name: DeleteAllExecutionFunctionConfigs :exec
DELETE FROM execution_function_configs WHERE execution_run_id = ?;

-- name: CountExecutionFunctions :one
SELECT COUNT(*) FROM execution_function_configs WHERE execution_run_id = ?;

-- name: CheckExecutionFunctionExists :one
SELECT COUNT(*) FROM execution_function_configs 
WHERE execution_run_id = ? AND function_definition_id = ?; 