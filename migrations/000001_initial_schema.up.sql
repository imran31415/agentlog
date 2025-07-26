-- Initial schema with user authentication
-- Based on actual SQL queries in sql/queries/ directory

-- Users table (from user authentication migration)
CREATE TABLE users (
    id VARCHAR(255) PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    email VARCHAR(255) UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    email_verified BOOLEAN DEFAULT FALSE,
    is_temporary BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP NULL
);

-- User sessions table
CREATE TABLE user_sessions (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    token VARCHAR(500) NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Execution runs table (must come before api_configurations due to foreign key)
CREATE TABLE execution_runs (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    base_prompt TEXT,
    context_prompt TEXT,
    enable_function_calling BOOLEAN NOT NULL DEFAULT FALSE,
    status ENUM('pending','running','completed','failed') DEFAULT 'pending',
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Function definitions table
CREATE TABLE function_definitions (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    name VARCHAR(100) NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    parameters_schema JSON,
    mock_response JSON DEFAULT NULL,
    endpoint_url VARCHAR(500),
    http_method VARCHAR(10) DEFAULT 'POST',
    headers JSON,
    auth_config JSON DEFAULT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    is_system_resource BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY unique_user_function (user_id, name),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- API configurations table
CREATE TABLE api_configurations (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    execution_run_id VARCHAR(255) NOT NULL,
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
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (execution_run_id) REFERENCES execution_runs(id) ON DELETE CASCADE
);

-- API requests table (based on sql/queries/api_requests.sql)
CREATE TABLE api_requests (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    execution_run_id VARCHAR(255) NOT NULL,
    configuration_id VARCHAR(255) NOT NULL,
    request_type VARCHAR(100),
    prompt TEXT,
    context TEXT,
    function_name VARCHAR(100),
    function_parameters JSON,
    request_headers JSON,
    request_body JSON NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (execution_run_id) REFERENCES execution_runs(id) ON DELETE CASCADE,
    FOREIGN KEY (configuration_id) REFERENCES api_configurations(id) ON DELETE CASCADE
);

-- API responses table (based on sql/queries/api_responses.sql)
CREATE TABLE api_responses (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    request_id VARCHAR(255) NOT NULL,
    response_status VARCHAR(50),
    response_text TEXT,
    function_call_response JSON,
    usage_metadata JSON,
    safety_ratings JSON,
    finish_reason VARCHAR(50),
    error_message TEXT,
    response_time_ms INT,
    response_headers JSON,
    response_body JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (request_id) REFERENCES api_requests(id) ON DELETE CASCADE
);

-- Function calls table (based on sql/queries/function_calls.sql)
CREATE TABLE function_calls (
    id VARCHAR(255) PRIMARY KEY,
    request_id VARCHAR(255) NOT NULL,
    function_name VARCHAR(100) NOT NULL,
    function_arguments JSON,
    function_response JSON,
    execution_status VARCHAR(50) DEFAULT 'pending',
    execution_time_ms INT,
    error_details TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (request_id) REFERENCES api_requests(id) ON DELETE CASCADE
);

-- Comparison results table (based on sql/queries/comparison_results.sql)
CREATE TABLE comparison_results (
    id VARCHAR(255) PRIMARY KEY,
    execution_run_id VARCHAR(255) NOT NULL,
    comparison_type VARCHAR(100),
    metric_name VARCHAR(100),
    configuration_scores JSON,
    best_configuration_id VARCHAR(255),
    best_configuration_data JSON,
    all_configurations_data JSON,
    analysis_notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (execution_run_id) REFERENCES execution_runs(id) ON DELETE CASCADE,
    FOREIGN KEY (best_configuration_id) REFERENCES api_configurations(id) ON DELETE SET NULL
);

-- Execution logs table (based on sql/queries/execution_logs.sql)
CREATE TABLE execution_logs (
    id VARCHAR(255) PRIMARY KEY,
    execution_run_id VARCHAR(255) NOT NULL,
    configuration_id VARCHAR(255),
    request_id VARCHAR(255),
    log_level VARCHAR(20) DEFAULT 'INFO',
    log_category VARCHAR(50),
    message TEXT NOT NULL,
    details JSON,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (execution_run_id) REFERENCES execution_runs(id) ON DELETE CASCADE,
    FOREIGN KEY (configuration_id) REFERENCES api_configurations(id) ON DELETE CASCADE,
    FOREIGN KEY (request_id) REFERENCES api_requests(id) ON DELETE CASCADE
);

-- Execution function configs table
CREATE TABLE execution_function_configs (
    id VARCHAR(255) PRIMARY KEY,
    execution_run_id VARCHAR(255) NOT NULL,
    function_definition_id VARCHAR(255) NOT NULL,
    use_mock_response BOOLEAN DEFAULT FALSE,
    execution_order INT DEFAULT 0,
    config JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (execution_run_id) REFERENCES execution_runs(id) ON DELETE CASCADE,
    FOREIGN KEY (function_definition_id) REFERENCES function_definitions(id) ON DELETE CASCADE
);

-- Create indexes for better performance
CREATE INDEX idx_execution_runs_user_id ON execution_runs(user_id);
CREATE INDEX idx_execution_runs_created_at ON execution_runs(created_at);
CREATE INDEX idx_api_requests_execution_run_id ON api_requests(execution_run_id);
CREATE INDEX idx_api_requests_configuration_id ON api_requests(configuration_id);
CREATE INDEX idx_api_responses_request_id ON api_responses(request_id);
CREATE INDEX idx_function_calls_request_id ON function_calls(request_id);
CREATE INDEX idx_execution_logs_execution_run_id ON execution_logs(execution_run_id);
CREATE INDEX idx_execution_logs_configuration_id ON execution_logs(configuration_id);
CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_token ON user_sessions(token);
CREATE INDEX idx_function_definitions_user_id ON function_definitions(user_id);
CREATE INDEX idx_api_configurations_user_id ON api_configurations(user_id);

-- Create system user immediately to avoid foreign key issues
INSERT INTO users (id, username, email, password_hash, email_verified, is_temporary, created_at, updated_at)
VALUES ('system', 'system', NULL, '', TRUE, FALSE, NOW(), NOW()); 