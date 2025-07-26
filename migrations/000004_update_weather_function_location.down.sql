-- Rollback get_current_weather function to original strict location format

UPDATE function_definitions 
SET parameters_schema = JSON_OBJECT(
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
updated_at = NOW()
WHERE name = 'get_current_weather' AND user_id = 'system'; 