-- Migration: 004_add_default_configurations.sql
-- Description: Add default AI configurations to the database
-- Created: 2025-07-25

-- Insert Conservative configuration
INSERT INTO api_configurations (
    id,
    user_id,
    execution_run_id,
    variation_name,
    model_name,
    system_prompt,
    temperature,
    max_tokens,
    top_p,
    top_k,
    safety_settings,
    generation_config,
    tools,
    tool_config,
    created_at
) VALUES (
    'config-conservative',
    'system', -- System-wide configuration
    'system-default', -- Global configuration identifier
    'Conservative',
    'gemini-1.5-flash',
    'You are a helpful assistant. Provide balanced, informative responses with careful attention to accuracy and safety.',
    0.2,
    500,
    0.8,
    10,
    JSON_OBJECT(
        'HARM_CATEGORY_HARASSMENT', 'BLOCK_MEDIUM_AND_ABOVE',
        'HARM_CATEGORY_HATE_SPEECH', 'BLOCK_MEDIUM_AND_ABOVE',
        'HARM_CATEGORY_SEXUALLY_EXPLICIT', 'BLOCK_MEDIUM_AND_ABOVE',
        'HARM_CATEGORY_DANGEROUS_CONTENT', 'BLOCK_MEDIUM_AND_ABOVE'
    ),
    JSON_OBJECT(
        'temperature', 0.2,
        'maxOutputTokens', 500,
        'topP', 0.8,
        'topK', 10,
        'candidateCount', 1
    ),
    JSON_ARRAY(),
    JSON_OBJECT(),
    NOW()
);

-- Insert Balanced configuration
INSERT INTO api_configurations (
    id,
    user_id,
    execution_run_id,
    variation_name,
    model_name,
    system_prompt,
    temperature,
    max_tokens,
    top_p,
    top_k,
    safety_settings,
    generation_config,
    tools,
    tool_config,
    created_at
) VALUES (
    'config-balanced',
    'system', -- System-wide configuration
    'system-default', -- Global configuration identifier
    'Balanced',
    'gemini-1.5-flash',
    'You are a helpful assistant. Provide balanced, informative responses that are both accurate and engaging.',
    0.5,
    500,
    0.9,
    20,
    JSON_OBJECT(
        'HARM_CATEGORY_HARASSMENT', 'BLOCK_MEDIUM_AND_ABOVE',
        'HARM_CATEGORY_HATE_SPEECH', 'BLOCK_MEDIUM_AND_ABOVE',
        'HARM_CATEGORY_SEXUALLY_EXPLICIT', 'BLOCK_MEDIUM_AND_ABOVE',
        'HARM_CATEGORY_DANGEROUS_CONTENT', 'BLOCK_MEDIUM_AND_ABOVE'
    ),
    JSON_OBJECT(
        'temperature', 0.5,
        'maxOutputTokens', 500,
        'topP', 0.9,
        'topK', 20,
        'candidateCount', 1
    ),
    JSON_ARRAY(),
    JSON_OBJECT(),
    NOW()
);

-- Insert Creative configuration
INSERT INTO api_configurations (
    id,
    user_id,
    execution_run_id,
    variation_name,
    model_name,
    system_prompt,
    temperature,
    max_tokens,
    top_p,
    top_k,
    safety_settings,
    generation_config,
    tools,
    tool_config,
    created_at
) VALUES (
    'config-creative',
    'system', -- System-wide configuration
    'system-default', -- Global configuration identifier
    'Creative',
    'gemini-1.5-flash',
    'You are a creative assistant. Provide imaginative and engaging responses with vivid imagery and artistic flair.',
    0.8,
    500,
    0.95,
    40,
    JSON_OBJECT(
        'HARM_CATEGORY_HARASSMENT', 'BLOCK_MEDIUM_AND_ABOVE',
        'HARM_CATEGORY_HATE_SPEECH', 'BLOCK_MEDIUM_AND_ABOVE',
        'HARM_CATEGORY_SEXUALLY_EXPLICIT', 'BLOCK_MEDIUM_AND_ABOVE',
        'HARM_CATEGORY_DANGEROUS_CONTENT', 'BLOCK_MEDIUM_AND_ABOVE'
    ),
    JSON_OBJECT(
        'temperature', 0.8,
        'maxOutputTokens', 500,
        'topP', 0.95,
        'topK', 40,
        'candidateCount', 1
    ),
    JSON_ARRAY(),
    JSON_OBJECT(),
    NOW()
);

-- Insert Function Calling configuration
INSERT INTO api_configurations (
    id,
    user_id,
    execution_run_id,
    variation_name,
    model_name,
    system_prompt,
    temperature,
    max_tokens,
    top_p,
    top_k,
    safety_settings,
    generation_config,
    tools,
    tool_config,
    created_at
) VALUES (
    'config-function-calling',
    'system', -- System-wide configuration
    'system-default', -- Global configuration identifier
    'Function Calling',
    'gemini-1.5-flash',
    'You are a helpful assistant with access to external tools. Use the available functions when they can help answer the user''s question more accurately.',
    0.3,
    500,
    0.85,
    15,
    JSON_OBJECT(
        'HARM_CATEGORY_HARASSMENT', 'BLOCK_MEDIUM_AND_ABOVE',
        'HARM_CATEGORY_HATE_SPEECH', 'BLOCK_MEDIUM_AND_ABOVE',
        'HARM_CATEGORY_SEXUALLY_EXPLICIT', 'BLOCK_MEDIUM_AND_ABOVE',
        'HARM_CATEGORY_DANGEROUS_CONTENT', 'BLOCK_MEDIUM_AND_ABOVE'
    ),
    JSON_OBJECT(
        'temperature', 0.3,
        'maxOutputTokens', 500,
        'topP', 0.85,
        'topK', 15,
        'candidateCount', 1
    ),
    JSON_ARRAY(
        JSON_OBJECT(
            'name', 'get_current_weather',
            'description', 'Get current weather information for a specific location',
            'parameters', JSON_OBJECT(
                'type', 'object',
                'properties', JSON_OBJECT(
                    'location', JSON_OBJECT(
                        'type', 'string',
                        'description', 'City name, state code, and country code'
                    ),
                    'units', JSON_OBJECT(
                        'type', 'string',
                        'enum', JSON_ARRAY('metric', 'imperial', 'kelvin'),
                        'default', 'metric'
                    )
                ),
                'required', JSON_ARRAY('location')
            )
        ),
        JSON_OBJECT(
            'name', 'neo4j_node_lookup',
            'description', 'Look up nodes in a Neo4j graph database',
            'parameters', JSON_OBJECT(
                'type', 'object',
                'properties', JSON_OBJECT(
                    'label', JSON_OBJECT(
                        'type', 'string',
                        'description', 'The node label to search for'
                    ),
                    'properties', JSON_OBJECT(
                        'type', 'object',
                        'description', 'Key-value pairs to match against node properties'
                    ),
                    'limit', JSON_OBJECT(
                        'type', 'integer',
                        'default', 10,
                        'minimum', 1,
                        'maximum', 100
                    )
                ),
                'required', JSON_ARRAY('label')
            )
        )
    ),
    JSON_OBJECT(
        'function_calling_config', JSON_OBJECT(
            'mode', 'AUTO'
        )
    ),
    NOW()
); 