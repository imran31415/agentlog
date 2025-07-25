-- Migration: 001_initial_schema.sql
-- Description: Initial database schema for gogent application
-- Created: 2025-07-25
-- Status: Applied

-- Create database if not exists
CREATE DATABASE IF NOT EXISTS gogent;
USE gogent;

-- Function definitions for reusable tools
CREATE TABLE function_definitions (
    id VARCHAR(36) NOT NULL,
    name VARCHAR(255) NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    parameters_schema JSON NOT NULL,
    mock_response JSON DEFAULT NULL,
    endpoint_url VARCHAR(500) DEFAULT NULL,
    http_method ENUM('GET','POST','PUT','DELETE','PATCH') DEFAULT 'POST',
    headers JSON DEFAULT NULL,
    auth_config JSON DEFAULT NULL,
    is_active TINYINT(1) DEFAULT 1,
    created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY name (name),
    KEY idx_name (name),
    KEY idx_active (is_active),
    KEY idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Execution runs for grouping related API calls and variations
CREATE TABLE execution_runs (
    id VARCHAR(36) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    base_prompt TEXT,
    context_prompt TEXT,
    created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    enable_function_calling TINYINT(1) NOT NULL DEFAULT 0,
    status ENUM('pending','running','completed','failed') DEFAULT 'pending',
    error_message TEXT,
    PRIMARY KEY (id),
    KEY idx_created_at (created_at),
    KEY idx_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- API call configurations for multi-variation execution
CREATE TABLE api_configurations (
    id VARCHAR(36) NOT NULL,
    execution_run_id VARCHAR(36) NOT NULL,
    variation_name VARCHAR(255) NOT NULL,
    model_name VARCHAR(255) NOT NULL,
    system_prompt TEXT,
    temperature DECIMAL(3,2) DEFAULT NULL,
    max_tokens INT DEFAULT NULL,
    top_p DECIMAL(3,2) DEFAULT NULL,
    top_k INT DEFAULT NULL,
    safety_settings JSON DEFAULT NULL,
    generation_config JSON DEFAULT NULL,
    tools JSON DEFAULT NULL,
    tool_config JSON DEFAULT NULL,
    created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_execution_run (execution_run_id),
    KEY idx_variation (variation_name),
    KEY idx_model (model_name),
    CONSTRAINT api_configurations_ibfk_1 FOREIGN KEY (execution_run_id) REFERENCES execution_runs (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- API requests
CREATE TABLE api_requests (
    id VARCHAR(36) NOT NULL,
    execution_run_id VARCHAR(36) NOT NULL,
    configuration_id VARCHAR(36) NOT NULL,
    request_type ENUM('generate','chat','function_call') NOT NULL,
    prompt TEXT NOT NULL,
    context TEXT,
    function_name VARCHAR(255) DEFAULT NULL,
    function_parameters JSON DEFAULT NULL,
    request_headers JSON DEFAULT NULL,
    request_body JSON DEFAULT NULL,
    created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_execution_run (execution_run_id),
    KEY idx_configuration (configuration_id),
    KEY idx_request_type (request_type),
    KEY idx_function_name (function_name),
    KEY idx_created_at (created_at),
    CONSTRAINT api_requests_ibfk_1 FOREIGN KEY (execution_run_id) REFERENCES execution_runs (id) ON DELETE CASCADE,
    CONSTRAINT api_requests_ibfk_2 FOREIGN KEY (configuration_id) REFERENCES api_configurations (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- API responses
CREATE TABLE api_responses (
    id VARCHAR(36) NOT NULL,
    request_id VARCHAR(36) NOT NULL,
    response_status ENUM('success','error','timeout') NOT NULL,
    response_text TEXT,
    function_call_response JSON DEFAULT NULL,
    usage_metadata JSON DEFAULT NULL,
    safety_ratings JSON DEFAULT NULL,
    finish_reason VARCHAR(100) DEFAULT NULL,
    error_message TEXT,
    response_time_ms INT DEFAULT NULL,
    response_headers JSON DEFAULT NULL,
    response_body JSON DEFAULT NULL,
    created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_request (request_id),
    KEY idx_status (response_status),
    KEY idx_response_time (response_time_ms),
    KEY idx_created_at (created_at),
    CONSTRAINT api_responses_ibfk_1 FOREIGN KEY (request_id) REFERENCES api_requests (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Function call details for tracking tool usage
CREATE TABLE function_calls (
    id VARCHAR(36) NOT NULL,
    request_id VARCHAR(36) NOT NULL,
    function_name VARCHAR(255) NOT NULL,
    function_arguments JSON DEFAULT NULL,
    function_response JSON DEFAULT NULL,
    execution_status ENUM('pending','success','error') NOT NULL,
    execution_time_ms INT DEFAULT NULL,
    error_details TEXT,
    created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_request (request_id),
    KEY idx_function_name (function_name),
    KEY idx_status (execution_status),
    KEY idx_created_at (created_at),
    CONSTRAINT function_calls_ibfk_1 FOREIGN KEY (request_id) REFERENCES api_requests (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Link execution runs to enabled function definitions
CREATE TABLE execution_function_configs (
    id VARCHAR(36) NOT NULL,
    execution_run_id VARCHAR(36) NOT NULL,
    function_definition_id VARCHAR(36) NOT NULL,
    use_mock_response TINYINT(1) DEFAULT 1,
    execution_order INT DEFAULT 0,
    created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY unique_execution_function (execution_run_id,function_definition_id),
    KEY idx_execution_run (execution_run_id),
    KEY idx_function_def (function_definition_id),
    CONSTRAINT execution_function_configs_ibfk_1 FOREIGN KEY (execution_run_id) REFERENCES execution_runs (id) ON DELETE CASCADE,
    CONSTRAINT execution_function_configs_ibfk_2 FOREIGN KEY (function_definition_id) REFERENCES function_definitions (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Execution logs for storing detailed execution information
CREATE TABLE execution_logs (
    id VARCHAR(36) NOT NULL,
    execution_run_id VARCHAR(36) NOT NULL,
    configuration_id VARCHAR(36) DEFAULT NULL,
    request_id VARCHAR(36) DEFAULT NULL,
    log_level ENUM('INFO','DEBUG','WARN','ERROR','SUCCESS') NOT NULL DEFAULT 'INFO',
    log_category ENUM('SETUP','EXECUTION','FUNCTION_CALL','API_CALL','COMPLETION','ERROR') NOT NULL DEFAULT 'EXECUTION',
    message TEXT NOT NULL,
    details JSON DEFAULT NULL,
    timestamp TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_execution_run (execution_run_id),
    KEY idx_configuration (configuration_id),
    KEY idx_request (request_id),
    KEY idx_log_level (log_level),
    KEY idx_log_category (log_category),
    KEY idx_timestamp (timestamp),
    CONSTRAINT execution_logs_ibfk_1 FOREIGN KEY (execution_run_id) REFERENCES execution_runs (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Comparison results for analyzing variations
CREATE TABLE comparison_results (
    id VARCHAR(36) NOT NULL,
    execution_run_id VARCHAR(36) NOT NULL,
    comparison_type ENUM('quality','performance','safety','custom') NOT NULL,
    metric_name VARCHAR(255) NOT NULL,
    configuration_scores JSON DEFAULT NULL,
    best_configuration_id VARCHAR(36) DEFAULT NULL,
    best_configuration_data JSON DEFAULT NULL,
    all_configurations_data JSON DEFAULT NULL,
    analysis_notes TEXT,
    created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY best_configuration_id (best_configuration_id),
    KEY idx_execution_run (execution_run_id),
    KEY idx_comparison_type (comparison_type),
    KEY idx_metric (metric_name),
    CONSTRAINT comparison_results_ibfk_1 FOREIGN KEY (execution_run_id) REFERENCES execution_runs (id) ON DELETE CASCADE,
    CONSTRAINT comparison_results_ibfk_2 FOREIGN KEY (best_configuration_id) REFERENCES api_configurations (id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci; 