-- Migration: 000_create_migrations_table.sql
-- Description: Create migrations tracking table
-- Created: 2025-07-25
-- Status: Applied

-- Create database if not exists
CREATE DATABASE IF NOT EXISTS gogent;
USE gogent;

-- Migration tracking table
CREATE TABLE IF NOT EXISTS schema_migrations (
    id INT AUTO_INCREMENT PRIMARY KEY,
    migration_name VARCHAR(255) NOT NULL UNIQUE,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    checksum VARCHAR(64) NOT NULL,
    execution_time_ms INT DEFAULT 0,
    status ENUM('applied', 'failed', 'rolled_back') DEFAULT 'applied',
    error_message TEXT,
    INDEX idx_migration_name (migration_name),
    INDEX idx_applied_at (applied_at),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci; 