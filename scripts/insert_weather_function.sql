-- Insert weather function definition into the database
INSERT INTO function_definitions (
    id, 
    name, 
    display_name, 
    description, 
    parameters_schema, 
    mock_response, 
    endpoint_url, 
    http_method, 
    headers, 
    auth_config, 
    is_active,
    created_at,
    updated_at
) VALUES (
    'func-weather-001',
    'get_weather',
    'Get Weather Information',
    'Get current weather information for a specific location',
    JSON_OBJECT(
        'type', 'object',
        'properties', JSON_OBJECT(
            'location', JSON_OBJECT(
                'type', 'string',
                'description', 'City name or location to get weather for (e.g., Los Angeles, London, Tokyo)'
            )
        ),
        'required', JSON_ARRAY('location')
    ),
    JSON_OBJECT(
        'location', 'Los Angeles, CA',
        'temperature', 72,
        'condition', 'Sunny',
        'humidity', 65,
        'wind_speed', 5,
        'wind_direction', 'W',
        'feels_like', 75,
        'description', 'Clear sky with light winds'
    ),
    'https://api.openweathermap.org/data/2.5/weather',
    'GET',
    JSON_OBJECT('Content-Type', 'application/json'),
    JSON_OBJECT('api_key_required', true),
    true,
    NOW(),
    NOW()
);

-- Verify the insertion
SELECT id, name, display_name, description, is_active FROM function_definitions WHERE name = 'get_weather'; 