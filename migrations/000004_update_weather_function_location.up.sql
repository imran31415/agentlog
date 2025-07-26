-- Update get_current_weather function to accept more flexible location formats

UPDATE function_definitions 
SET parameters_schema = JSON_OBJECT(
    'type', 'object',
    'properties', JSON_OBJECT(
        'location', JSON_OBJECT(
            'type', 'string',
            'description', 'Location to get weather for. Can be: city name (e.g., "Los Angeles"), city and state (e.g., "Los Angeles, CA"), city and country (e.g., "London, UK"), or full format (e.g., "Los Angeles, CA, US")'
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
updated_at = NOW()
WHERE name = 'get_current_weather' AND user_id = 'system'; 