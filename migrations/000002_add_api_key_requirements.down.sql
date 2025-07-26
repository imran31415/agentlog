-- Remove API key requirements from function definitions
ALTER TABLE function_definitions 
DROP COLUMN required_api_keys,
DROP COLUMN api_key_validation; 