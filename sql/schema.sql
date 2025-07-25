-- gogent database schema for logging Gemini API interactions

CREATE DATABASE IF NOT EXISTS gogent;
USE gogent;

-- Function definitions for reusable tools
CREATE TABLE function_definitions (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    display_name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    parameters_schema JSON NOT NULL, -- JSON schema for function parameters
    mock_response JSON, -- Mock response for testing
    endpoint_url VARCHAR(500), -- For real API integration
    http_method ENUM('GET', 'POST', 'PUT', 'DELETE', 'PATCH') DEFAULT 'POST',
    headers JSON, -- HTTP headers for real API calls
    auth_config JSON, -- Authentication configuration
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_name (name),
    INDEX idx_active (is_active),
    INDEX idx_created_at (created_at)
);

-- Link execution runs to enabled function definitions
CREATE TABLE execution_function_configs (
    id VARCHAR(36) PRIMARY KEY,
    execution_run_id VARCHAR(36) NOT NULL,
    function_definition_id VARCHAR(36) NOT NULL,
    use_mock_response BOOLEAN DEFAULT TRUE,
    execution_order INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (execution_run_id) REFERENCES execution_runs(id) ON DELETE CASCADE,
    FOREIGN KEY (function_definition_id) REFERENCES function_definitions(id) ON DELETE CASCADE,
    UNIQUE KEY unique_execution_function (execution_run_id, function_definition_id),
    INDEX idx_execution_run (execution_run_id),
    INDEX idx_function_def (function_definition_id)
);

-- Execution runs for grouping related API calls and variations
CREATE TABLE execution_runs (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    enable_function_calling BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_created_at (created_at),
    INDEX idx_name (name),
    INDEX idx_function_calling (enable_function_calling)
);

-- Execution logs for storing detailed execution information
CREATE TABLE execution_logs (
    id VARCHAR(36) PRIMARY KEY,
    execution_run_id VARCHAR(36) NOT NULL,
    configuration_id VARCHAR(36) NULL, -- NULL for run-level logs
    request_id VARCHAR(36) NULL, -- NULL for run/config-level logs
    log_level ENUM('INFO', 'DEBUG', 'WARN', 'ERROR', 'SUCCESS') NOT NULL DEFAULT 'INFO',
    log_category ENUM('SETUP', 'EXECUTION', 'FUNCTION_CALL', 'API_CALL', 'COMPLETION', 'ERROR') NOT NULL DEFAULT 'EXECUTION',
    message TEXT NOT NULL,
    details JSON NULL, -- Additional structured data
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (execution_run_id) REFERENCES execution_runs(id) ON DELETE CASCADE,
    FOREIGN KEY (configuration_id) REFERENCES api_configurations(id) ON DELETE CASCADE,
    FOREIGN KEY (request_id) REFERENCES api_requests(id) ON DELETE CASCADE,
    INDEX idx_execution_run (execution_run_id),
    INDEX idx_configuration (configuration_id),
    INDEX idx_request (request_id),
    INDEX idx_log_level (log_level),
    INDEX idx_log_category (log_category),
    INDEX idx_timestamp (timestamp)
);

-- API call configurations for multi-variation execution
CREATE TABLE api_configurations (
    id VARCHAR(36) PRIMARY KEY,
    execution_run_id VARCHAR(36) NOT NULL,
    variation_name VARCHAR(255) NOT NULL,
    model_name VARCHAR(255) NOT NULL,
    system_prompt TEXT,
    temperature DECIMAL(3,2),
    max_tokens INT,
    top_p DECIMAL(3,2),
    top_k INT,
    safety_settings JSON,
    generation_config JSON,
    tools JSON,
    tool_config JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (execution_run_id) REFERENCES execution_runs(id) ON DELETE CASCADE,
    INDEX idx_execution_run (execution_run_id),
    INDEX idx_variation (variation_name),
    INDEX idx_model (model_name)
);

-- API requests
CREATE TABLE api_requests (
    id VARCHAR(36) PRIMARY KEY,
    execution_run_id VARCHAR(36) NOT NULL,
    configuration_id VARCHAR(36) NOT NULL,
    request_type ENUM('generate', 'chat', 'function_call') NOT NULL,
    prompt TEXT NOT NULL,
    context TEXT,
    function_name VARCHAR(255),
    function_parameters JSON,
    request_headers JSON,
    request_body JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (execution_run_id) REFERENCES execution_runs(id) ON DELETE CASCADE,
    FOREIGN KEY (configuration_id) REFERENCES api_configurations(id) ON DELETE CASCADE,
    INDEX idx_execution_run (execution_run_id),
    INDEX idx_configuration (configuration_id),
    INDEX idx_request_type (request_type),
    INDEX idx_function_name (function_name),
    INDEX idx_created_at (created_at)
);

-- API responses
CREATE TABLE api_responses (
    id VARCHAR(36) PRIMARY KEY,
    request_id VARCHAR(36) NOT NULL,
    response_status ENUM('success', 'error', 'timeout') NOT NULL,
    response_text TEXT,
    function_call_response JSON,
    usage_metadata JSON,
    safety_ratings JSON,
    finish_reason VARCHAR(100),
    error_message TEXT,
    response_time_ms INT,
    response_headers JSON,
    response_body JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (request_id) REFERENCES api_requests(id) ON DELETE CASCADE,
    INDEX idx_request (request_id),
    INDEX idx_status (response_status),
    INDEX idx_response_time (response_time_ms),
    INDEX idx_created_at (created_at)
);

-- Function call details for tracking tool usage
CREATE TABLE function_calls (
    id VARCHAR(36) PRIMARY KEY,
    request_id VARCHAR(36) NOT NULL,
    function_name VARCHAR(255) NOT NULL,
    function_arguments JSON,
    function_response JSON,
    execution_status ENUM('pending', 'success', 'error') NOT NULL,
    execution_time_ms INT,
    error_details TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (request_id) REFERENCES api_requests(id) ON DELETE CASCADE,
    INDEX idx_request (request_id),
    INDEX idx_function_name (function_name),
    INDEX idx_status (execution_status),
    INDEX idx_created_at (created_at)
);

-- Comparison results for analyzing variations
CREATE TABLE comparison_results (
    id VARCHAR(36) PRIMARY KEY,
    execution_run_id VARCHAR(36) NOT NULL,
    comparison_type ENUM('quality', 'performance', 'safety', 'custom') NOT NULL,
    metric_name VARCHAR(255) NOT NULL,
    configuration_scores JSON,
    best_configuration_id VARCHAR(36),
    best_configuration_data JSON,
    all_configurations_data JSON,
    analysis_notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (execution_run_id) REFERENCES execution_runs(id) ON DELETE CASCADE,
    FOREIGN KEY (best_configuration_id) REFERENCES api_configurations(id) ON DELETE SET NULL,
    INDEX idx_execution_run (execution_run_id),
    INDEX idx_comparison_type (comparison_type),
    INDEX idx_metric (metric_name)
); 