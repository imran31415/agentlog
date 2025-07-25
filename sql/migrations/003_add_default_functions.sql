-- Migration: 003_add_default_functions.sql
-- Description: Add default function definitions for Neo4j and OpenWeather API
-- Created: 2025-07-25

-- Insert Neo4j node lookup function
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
    'func-neo4j-lookup',
    'neo4j_node_lookup',
    'Neo4j Node Lookup',
    'Look up nodes in a Neo4j graph database by label and properties',
    JSON_OBJECT(
        'type', 'object',
        'properties', JSON_OBJECT(
            'label', JSON_OBJECT(
                'type', 'string',
                'description', 'The node label to search for (e.g., Person, Company, Product)'
            ),
            'properties', JSON_OBJECT(
                'type', 'object',
                'description', 'Key-value pairs to match against node properties',
                'additionalProperties', true
            ),
            'limit', JSON_OBJECT(
                'type', 'integer',
                'description', 'Maximum number of nodes to return',
                'default', 10,
                'minimum', 1,
                'maximum', 100
            )
        ),
        'required', JSON_ARRAY('label')
    ),
    JSON_OBJECT(
        'success', true,
        'nodes', JSON_ARRAY(
            JSON_OBJECT(
                'id', 123,
                'labels', JSON_ARRAY('Person'),
                'properties', JSON_OBJECT(
                    'name', 'John Doe',
                    'age', 30,
                    'city', 'New York'
                )
            ),
            JSON_OBJECT(
                'id', 456,
                'labels', JSON_ARRAY('Person'),
                'properties', JSON_OBJECT(
                    'name', 'Jane Smith',
                    'age', 25,
                    'city', 'San Francisco'
                )
            )
        ),
        'count', 2,
        'query_time_ms', 45
    ),
    'http://localhost:7474/db/neo4j/tx/commit',
    'POST',
    JSON_OBJECT(
        'Content-Type', 'application/json',
        'Accept', 'application/json'
    ),
    JSON_OBJECT(
        'type', 'basic',
        'username_field', 'username',
        'password_field', 'password'
    ),
    true,
    NOW(),
    NOW()
);

-- Insert OpenWeather API function
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
    'func-openweather-current',
    'get_current_weather',
    'Get Current Weather',
    'Get current weather information for a specific location using OpenWeather API',
    JSON_OBJECT(
        'type', 'object',
        'properties', JSON_OBJECT(
            'location', JSON_OBJECT(
                'type', 'string',
                'description', 'City name, state code (US only), and country code (ISO 3166) separated by comma. Format: "City,State,Country" or "City,Country"'
            ),
            'units', JSON_OBJECT(
                'type', 'string',
                'enum', JSON_ARRAY('metric', 'imperial', 'kelvin'),
                'description', 'Units of measurement. metric: Celsius, imperial: Fahrenheit, kelvin: Kelvin',
                'default', 'metric'
            ),
            'lang', JSON_OBJECT(
                'type', 'string',
                'description', 'Language code for weather description (e.g., en, es, fr)',
                'default', 'en'
            )
        ),
        'required', JSON_ARRAY('location')
    ),
    JSON_OBJECT(
        'coord', JSON_OBJECT('lon', -122.08, 'lat', 37.39),
        'weather', JSON_ARRAY(
            JSON_OBJECT(
                'id', 800,
                'main', 'Clear',
                'description', 'clear sky',
                'icon', '01d'
            )
        ),
        'base', 'stations',
        'main', JSON_OBJECT(
            'temp', 22.5,
            'feels_like', 21.8,
            'temp_min', 20.1,
            'temp_max', 25.3,
            'pressure', 1013,
            'humidity', 65
        ),
        'visibility', 10000,
        'wind', JSON_OBJECT(
            'speed', 3.6,
            'deg', 240
        ),
        'clouds', JSON_OBJECT('all', 20),
        'dt', 1640995200,
        'sys', JSON_OBJECT(
            'type', 2,
            'id', 2000,
            'country', 'US',
            'sunrise', 1640966400,
            'sunset', 1641002400
        ),
        'timezone', -28800,
        'id', 5375480,
        'name', 'Mountain View',
        'cod', 200
    ),
    'https://api.openweathermap.org/data/2.5/weather',
    'GET',
    JSON_OBJECT(
        'Accept', 'application/json'
    ),
    JSON_OBJECT(
        'type', 'api_key',
        'api_key_param', 'appid',
        'api_key_location', 'query'
    ),
    true,
    NOW(),
    NOW()
); 