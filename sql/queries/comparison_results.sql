-- name: CreateComparisonResult :exec
INSERT INTO comparison_results (
    id, execution_run_id, comparison_type, metric_name, 
    configuration_scores, best_configuration_id, best_configuration_data, 
    all_configurations_data, analysis_notes
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?
);

-- name: GetComparisonResult :one
SELECT 
    id, execution_run_id, comparison_type, metric_name,
    configuration_scores, best_configuration_id, 
    CAST(best_configuration_data AS CHAR) as best_configuration_data,
    CAST(all_configurations_data AS CHAR) as all_configurations_data, 
    analysis_notes, created_at
FROM comparison_results 
WHERE execution_run_id = ? 
LIMIT 1;

-- name: ListComparisonResults :many
SELECT 
    id, execution_run_id, comparison_type, metric_name,
    configuration_scores, best_configuration_id, 
    CAST(best_configuration_data AS CHAR) as best_configuration_data,
    CAST(all_configurations_data AS CHAR) as all_configurations_data,
    analysis_notes, created_at
FROM comparison_results 
ORDER BY created_at DESC;

-- name: GetComparisonResultsByExecutionRun :many
SELECT 
    id, execution_run_id, comparison_type, metric_name,
    configuration_scores, best_configuration_id, 
    CAST(best_configuration_data AS CHAR) as best_configuration_data,
    CAST(all_configurations_data AS CHAR) as all_configurations_data,
    analysis_notes, created_at
FROM comparison_results 
WHERE execution_run_id = ?
ORDER BY created_at DESC; 