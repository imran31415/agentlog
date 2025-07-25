-- Insert Neo4j Graph Lookup Function Definition
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
    'func-neo4j-001',
    'query_graph',
    'Query Neo4j Graph Database',
    'Query a Neo4j graph database to find nodes, relationships, and paths. Supports Cypher queries to retrieve connected data and graph structures.',
    JSON_OBJECT(
        'type', 'object',
        'properties', JSON_OBJECT(
            'query', JSON_OBJECT(
                'type', 'string',
                'description', 'Cypher query to execute against the Neo4j database. Use MATCH clauses to find patterns and RETURN to specify what data to retrieve.',
                'examples', JSON_ARRAY(
                    'MATCH (n:Person) RETURN n.name LIMIT 10',
                    'MATCH (p:Person)-[:WORKS_FOR]->(c:Company) RETURN p.name, c.name',
                    'MATCH (start:Location {name: "New York"})-[:CONNECTED_TO*1..3]-(end:Location) RETURN DISTINCT end.name'
                )
            ),
            'limit', JSON_OBJECT(
                'type', 'integer',
                'description', 'Maximum number of results to return (1-100)',
                'default', 25,
                'minimum', 1,
                'maximum', 100
            )
        ),
        'required', JSON_ARRAY('query')
    ),
    JSON_OBJECT(
        'nodes', JSON_ARRAY(
            JSON_OBJECT(
                'id', 'person1',
                'labels', JSON_ARRAY('Person'),
                'properties', JSON_OBJECT(
                    'name', 'John Doe',
                    'age', 30,
                    'city', 'New York'
                )
            ),
            JSON_OBJECT(
                'id', 'company1',
                'labels', JSON_ARRAY('Company'),
                'properties', JSON_OBJECT(
                    'name', 'Tech Corp',
                    'industry', 'Technology',
                    'founded', 2010
                )
            )
        ),
        'relationships', JSON_ARRAY(
            JSON_OBJECT(
                'id', 'rel1',
                'type', 'WORKS_FOR',
                'startNode', 'person1',
                'endNode', 'company1',
                'properties', JSON_OBJECT(
                    'since', '2020-01-15',
                    'position', 'Software Engineer'
                )
            )
        ),
        'summary', JSON_OBJECT(
            'totalNodes', 2,
            'totalRelationships', 1,
            'executionTime', '15ms',
            'query', 'MATCH (p:Person)-[:WORKS_FOR]->(c:Company) RETURN p, c LIMIT 25'
        )
    ),
    NULL, -- endpoint_url (Neo4j connection handled internally)
    'POST',
    JSON_OBJECT(
        'Content-Type', 'application/json'
    ),
    JSON_OBJECT(
        'requires_auth', TRUE,
        'auth_type', 'basic',
        'description', 'Neo4j database authentication required'
    ),
    TRUE,
    NOW(),
    NOW()
) ON DUPLICATE KEY UPDATE
    display_name = VALUES(display_name),
    description = VALUES(description),
    parameters_schema = VALUES(parameters_schema),
    mock_response = VALUES(mock_response),
    endpoint_url = VALUES(endpoint_url),
    http_method = VALUES(http_method),
    headers = VALUES(headers),
    auth_config = VALUES(auth_config),
    is_active = VALUES(is_active),
    updated_at = NOW(); 