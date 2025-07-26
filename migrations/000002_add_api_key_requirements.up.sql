-- Add API key requirements to function definitions

-- Add API key requirements to function definitions
ALTER TABLE function_definitions 
ADD COLUMN required_api_keys JSON DEFAULT NULL COMMENT 'Array of required API key names for this function',
ADD COLUMN api_key_validation JSON DEFAULT NULL COMMENT 'Validation rules for each API key (optional patterns, descriptions)';

-- Example data for existing functions (update as needed)
-- Weather function requires OpenWeather API key
UPDATE function_definitions 
SET required_api_keys = JSON_ARRAY('openWeatherApiKey'),
    api_key_validation = JSON_OBJECT(
        'openWeatherApiKey', JSON_OBJECT(
            'description', 'OpenWeather API key for weather data',
            'pattern', '^[a-zA-Z0-9]{32}$',
            'testEndpoint', 'https://api.openweathermap.org/data/2.5/weather?q=London&appid={key}',
            'errorMessage', 'Please enter a valid OpenWeather API key'
        )
    )
WHERE name = 'get_current_weather';

-- Neo4j functions require Neo4j credentials
UPDATE function_definitions 
SET required_api_keys = JSON_ARRAY('neo4jUrl', 'neo4jUsername', 'neo4jPassword', 'neo4jDatabase'),
    api_key_validation = JSON_OBJECT(
        'neo4jUrl', JSON_OBJECT(
            'description', 'Neo4j database URL',
            'pattern', '^(neo4j|bolt)://.*',
            'errorMessage', 'Please enter a valid Neo4j URL (neo4j:// or bolt://)'
        ),
        'neo4jUsername', JSON_OBJECT(
            'description', 'Neo4j username',
            'errorMessage', 'Please enter your Neo4j username'
        ),
        'neo4jPassword', JSON_OBJECT(
            'description', 'Neo4j password',
            'errorMessage', 'Please enter your Neo4j password'
        ),
        'neo4jDatabase', JSON_OBJECT(
            'description', 'Neo4j database name',
            'errorMessage', 'Please enter the Neo4j database name'
        )
    )
WHERE name = 'neo4j_node_lookup'; 