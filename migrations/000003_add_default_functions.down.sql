-- Remove default function definitions
DELETE FROM function_definitions WHERE id IN ('func-neo4j-lookup', 'func-openweather-current'); 