-- Migration: 002_add_user_authentication.sql
-- Description: Add user authentication system and user_id to all tables
-- Created: 2025-07-25
-- Status: Applied

-- Create users table
CREATE TABLE users (
    id VARCHAR(36) PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    email_verified BOOLEAN DEFAULT FALSE,
    is_temporary BOOLEAN DEFAULT FALSE,
    email_verification_token VARCHAR(255),
    email_verification_expires_at TIMESTAMP NULL,
    password_reset_token VARCHAR(255),
    password_reset_expires_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP NULL,
    INDEX idx_username (username),
    INDEX idx_email (email),
    INDEX idx_temporary (is_temporary),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Add user_id column to execution_runs table
ALTER TABLE execution_runs 
ADD COLUMN user_id VARCHAR(36) NOT NULL DEFAULT 'temp-user' AFTER id,
ADD INDEX idx_user_id (user_id),
ADD CONSTRAINT fk_execution_runs_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Add user_id column to function_definitions table
ALTER TABLE function_definitions 
ADD COLUMN user_id VARCHAR(36) NOT NULL DEFAULT 'temp-user' AFTER id,
ADD INDEX idx_user_id (user_id),
ADD CONSTRAINT fk_function_definitions_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Add user_id column to api_configurations table
ALTER TABLE api_configurations 
ADD COLUMN user_id VARCHAR(36) NOT NULL DEFAULT 'temp-user' AFTER id,
ADD INDEX idx_user_id (user_id),
ADD CONSTRAINT fk_api_configurations_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Add user_id column to api_requests table
ALTER TABLE api_requests 
ADD COLUMN user_id VARCHAR(36) NOT NULL DEFAULT 'temp-user' AFTER id,
ADD INDEX idx_user_id (user_id),
ADD CONSTRAINT fk_api_requests_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Add user_id column to api_responses table
ALTER TABLE api_responses 
ADD COLUMN user_id VARCHAR(36) NOT NULL DEFAULT 'temp-user' AFTER id,
ADD INDEX idx_user_id (user_id),
ADD CONSTRAINT fk_api_responses_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Add user_id column to function_calls table
ALTER TABLE function_calls 
ADD COLUMN user_id VARCHAR(36) NOT NULL DEFAULT 'temp-user' AFTER id,
ADD INDEX idx_user_id (user_id),
ADD CONSTRAINT fk_function_calls_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Add user_id column to execution_logs table
ALTER TABLE execution_logs 
ADD COLUMN user_id VARCHAR(36) NOT NULL DEFAULT 'temp-user' AFTER id,
ADD INDEX idx_user_id (user_id),
ADD CONSTRAINT fk_execution_logs_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Add user_id column to comparison_results table
ALTER TABLE comparison_results 
ADD COLUMN user_id VARCHAR(36) NOT NULL DEFAULT 'temp-user' AFTER id,
ADD INDEX idx_user_id (user_id),
ADD CONSTRAINT fk_comparison_results_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Add user_id column to execution_function_configs table
ALTER TABLE execution_function_configs 
ADD COLUMN user_id VARCHAR(36) NOT NULL DEFAULT 'temp-user' AFTER id,
ADD INDEX idx_user_id (user_id),
ADD CONSTRAINT fk_execution_function_configs_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Create a temporary user for existing data
INSERT INTO users (id, username, email, password_hash, is_temporary, created_at, updated_at) 
VALUES ('temp-user', 'temporary_user', 'temp@example.com', 'temp_hash', TRUE, NOW(), NOW());

-- Create user sessions table for JWT token management
CREATE TABLE user_sessions (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_token_hash (token_hash),
    INDEX idx_expires_at (expires_at),
    CONSTRAINT fk_user_sessions_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci; 