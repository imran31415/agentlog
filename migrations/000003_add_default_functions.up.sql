-- Add default function definitions for Neo4j and OpenWeather API

-- Insert Neo4j node lookup function
INSERT INTO function_definitions (
    id,
    user_id,
    name,
    display_name,
    description,
    parameters_schema,
    endpoint_url,
    http_method,
    headers,
    is_active,
    is_system_resource,
    created_at,
    updated_at
) VALUES (
    'func-neo4j-lookup',
    'system',
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
                'description', 'Key-value pairs to match against node properties'
            ),
            'limit', JSON_OBJECT(
                'type', 'integer',
                'description', 'Maximum number of nodes to return',
                'minimum', 1,
                'maximum', 100
            )
        ),
        'required', JSON_ARRAY('label')
    ),
    'http://localhost:7474/db/neo4j/tx/commit',
    'POST',
    JSON_OBJECT(
        'Content-Type', 'application/json',
        'Accept', 'application/json'
    ),
    true,
    true,
    NOW(),
    NOW()
);

-- Insert OpenWeather API function
INSERT INTO function_definitions (
    id,
    user_id,
    name,
    display_name,
    description,
    parameters_schema,
    endpoint_url,
    http_method,
    headers,
    is_active,
    is_system_resource,
    created_at,
    updated_at
) VALUES (
    'func-openweather-current',
    'system',
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
                'description', 'Units of measurement. metric: Celsius, imperial: Fahrenheit, kelvin: Kelvin'
            ),
            'lang', JSON_OBJECT(
                'type', 'string',
                'description', 'Language code for weather description (e.g., en, es, fr)'
            )
        ),
        'required', JSON_ARRAY('location')
    ),
    'https://api.openweathermap.org/data/2.5/weather',
    'GET',
    JSON_OBJECT(
        'Accept', 'application/json'
    ),
    true,
    true,
    NOW(),
    NOW()
); 