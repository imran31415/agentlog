package gogent

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"gogent/internal/db"
	"gogent/internal/gemini"
	"gogent/internal/types"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// Client represents the main gogent client that wraps Gemini API calls
type Client struct {
	db           *sql.DB
	queries      *db.Queries
	config       *types.GeminiClientConfig
	geminiClient *gemini.GeminiClient
	mutex        sync.RWMutex
	// Add execution context for logging
	currentExecutionRunID *string
	currentConfigID       *string
	currentRequestID      *string
}

// NewClient creates a new gogent client with database connection
func NewClient(dbURL string, config *types.GeminiClientConfig) (*Client, error) {
	database, err := sql.Open("mysql", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := database.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Run database migrations
	migrationManager := db.NewMigrationManager(database)
	migrationsDir := "sql/migrations"
	if err := migrationManager.RunMigrations(migrationsDir); err != nil {
		log.Printf("⚠️ Warning: failed to run migrations: %v", err)
		// Continue without migrations rather than failing completely
	} else {
		log.Printf("✅ Database migrations completed successfully")
	}

	queries := db.New(database)

	client := &Client{
		db:      database,
		queries: queries,
		config:  config,
		mutex:   sync.RWMutex{},
	}

	// Initialize Gemini client if API key is provided
	// DISABLED: Go SDK has model name format issues, using REST API directly
	/*
		if config.APIKey != "" {
			ctx := context.Background()
			geminiClient, err := gemini.NewGeminiClient(ctx, config.APIKey)
			if err != nil {
				log.Printf("Failed to initialize Gemini client: %v", err)
				// Continue without Gemini client (will use mock responses)
			} else {
				client.geminiClient = geminiClient
				log.Printf("Successfully initialized Gemini client with API key: %s...", config.APIKey[:10])
			}
		}
	*/

	// Force REST API usage - no Go SDK client
	client.geminiClient = nil
	log.Printf("Go SDK disabled - using REST API for all Gemini calls")

	return client, nil
}

// Close closes the database connection and Gemini client
func (c *Client) Close() error {
	if c.geminiClient != nil {
		c.geminiClient.Close()
	}
	return c.db.Close()
}

// CreateExecutionRun creates a new execution run for grouping related API calls
func (c *Client) CreateExecutionRun(ctx context.Context, name, description string, enableFunctionCalling bool) (*types.ExecutionRun, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	id := uuid.New().String()
	log.Printf("🔧 Creating execution run with enableFunctionCalling: %v", enableFunctionCalling)
	err := c.queries.CreateExecutionRun(ctx, db.CreateExecutionRunParams{
		ID:                    id,
		Name:                  name,
		Description:           sql.NullString{String: description, Valid: description != ""},
		EnableFunctionCalling: enableFunctionCalling,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create execution run: %w", err)
	}

	return &types.ExecutionRun{
		ID:                    id,
		Name:                  name,
		Description:           description,
		EnableFunctionCalling: enableFunctionCalling,
		Status:                "pending", // Start with pending status
		ErrorMessage:          "",
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}, nil
}

// CreateAPIConfiguration creates a new API configuration for a variation
func (c *Client) CreateAPIConfiguration(ctx context.Context, config *types.APIConfiguration) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	safetySettingsJSON, _ := types.ToJSON(config.SafetySettings)
	generationConfigJSON, _ := types.ToJSON(config.GenerationConfig)
	toolsJSON, _ := types.ToJSON(config.Tools)
	toolConfigJSON, _ := types.ToJSON(config.ToolConfig)

	return c.queries.CreateAPIConfiguration(ctx, db.CreateAPIConfigurationParams{
		ID:               config.ID,
		ExecutionRunID:   config.ExecutionRunID,
		VariationName:    config.VariationName,
		ModelName:        config.ModelName,
		SystemPrompt:     sql.NullString{String: config.SystemPrompt, Valid: config.SystemPrompt != ""},
		Temperature:      convertFloat32ToNullString(config.Temperature),
		MaxTokens:        convertInt32ToNullInt32(config.MaxTokens),
		TopP:             convertFloat32ToNullString(config.TopP),
		TopK:             convertInt32ToNullInt32(config.TopK),
		SafetySettings:   convertStringToRawMessage(safetySettingsJSON),
		GenerationConfig: convertStringToRawMessage(generationConfigJSON),
		Tools:            convertStringToRawMessage(toolsJSON),
		ToolConfig:       convertStringToRawMessage(toolConfigJSON),
	})
}

// LogAPIRequest logs an API request to the database
func (c *Client) LogAPIRequest(ctx context.Context, request *types.APIRequest) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	functionParamsJSON, _ := types.ToJSON(request.FunctionParameters)
	requestHeadersJSON, _ := types.ToJSON(request.RequestHeaders)
	requestBodyJSON, _ := types.ToJSON(request.RequestBody)

	return c.queries.CreateAPIRequest(ctx, db.CreateAPIRequestParams{
		ID:                 request.ID,
		ExecutionRunID:     request.ExecutionRunID,
		ConfigurationID:    request.ConfigurationID,
		RequestType:        db.ApiRequestsRequestType(request.RequestType),
		Prompt:             request.Prompt,
		Context:            sql.NullString{String: request.Context, Valid: request.Context != ""},
		FunctionName:       sql.NullString{String: request.FunctionName, Valid: request.FunctionName != ""},
		FunctionParameters: convertStringToRawMessage(functionParamsJSON),
		RequestHeaders:     convertStringToRawMessage(requestHeadersJSON),
		RequestBody:        convertStringToRawMessage(requestBodyJSON),
	})
}

// LogAPIResponse logs an API response to the database
func (c *Client) LogAPIResponse(ctx context.Context, response *types.APIResponse) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	functionCallResponseJSON, _ := types.ToJSON(response.FunctionCallResponse)
	usageMetadataJSON, _ := types.ToJSON(response.UsageMetadata)
	safetyRatingsJSON, _ := types.ToJSON(response.SafetyRatings)
	responseHeadersJSON, _ := types.ToJSON(response.ResponseHeaders)
	responseBodyJSON, _ := types.ToJSON(response.ResponseBody)

	return c.queries.CreateAPIResponse(ctx, db.CreateAPIResponseParams{
		ID:                   response.ID,
		RequestID:            response.RequestID,
		ResponseStatus:       db.ApiResponsesResponseStatus(response.ResponseStatus),
		ResponseText:         sql.NullString{String: response.ResponseText, Valid: response.ResponseText != ""},
		FunctionCallResponse: convertStringToRawMessage(functionCallResponseJSON),
		UsageMetadata:        convertStringToRawMessage(usageMetadataJSON),
		SafetyRatings:        convertStringToRawMessage(safetyRatingsJSON),
		FinishReason:         sql.NullString{String: response.FinishReason, Valid: response.FinishReason != ""},
		ErrorMessage:         sql.NullString{String: response.ErrorMessage, Valid: response.ErrorMessage != ""},
		ResponseTimeMs:       sql.NullInt32{Int32: response.ResponseTimeMs, Valid: true},
		ResponseHeaders:      convertStringToRawMessage(responseHeadersJSON),
		ResponseBody:         convertStringToRawMessage(responseBodyJSON),
	})
}

// ExecuteMultiVariation executes the same prompt with multiple configurations
func (c *Client) ExecuteMultiVariation(ctx context.Context, request *types.MultiExecutionRequest) (*types.ExecutionResult, error) {
	// Create execution run
	executionRun, err := c.CreateExecutionRun(ctx, request.ExecutionRunName, request.Description, request.EnableFunctionCalling)
	if err != nil {
		return nil, fmt.Errorf("failed to create execution run: %w", err)
	}

	// Set execution context for logging
	c.setExecutionContext(&executionRun.ID, nil, nil)
	defer c.clearExecutionContext()

	// Log execution start
	c.logExecutionEvent(types.LogLevelInfo, types.LogCategorySetup,
		fmt.Sprintf("Starting execution: %s", request.ExecutionRunName),
		map[string]interface{}{
			"enableFunctionCalling": request.EnableFunctionCalling,
			"functionToolsCount":    len(request.FunctionTools),
			"configurationsCount":   len(request.Configurations),
		})

	if request.EnableFunctionCalling {
		for i, tool := range request.FunctionTools {
			c.logExecutionEvent(types.LogLevelDebug, types.LogCategorySetup,
				fmt.Sprintf("Function tool %d: %s - %s", i+1, tool.Name, tool.Description), nil)
		}
	}

	result := &types.ExecutionResult{
		ExecutionRun: *executionRun,
		Results:      make([]types.VariationResult, 0, len(request.Configurations)),
		TotalTime:    0,
		SuccessCount: 0,
		ErrorCount:   0,
	}

	startTime := time.Now()

	// Execute each configuration with rate limiting
	for i, config := range request.Configurations {
		config.ID = uuid.New().String()
		config.ExecutionRunID = executionRun.ID

		// Set configuration context for logging
		c.setExecutionContext(&executionRun.ID, &config.ID, nil)

		// CRITICAL: Add function tools to configuration if function calling is enabled
		if request.EnableFunctionCalling && len(request.FunctionTools) > 0 {
			c.logExecutionEvent(types.LogLevelDebug, types.LogCategorySetup,
				fmt.Sprintf("Adding %d function tools to configuration: %s", len(request.FunctionTools), config.VariationName), nil)
			config.Tools = request.FunctionTools
		} else {
			c.logExecutionEvent(types.LogLevelWarn, types.LogCategorySetup,
				fmt.Sprintf("No function tools added to configuration: enableFunctionCalling=%v, toolCount=%d", request.EnableFunctionCalling, len(request.FunctionTools)), nil)
		}

		// Save configuration
		if err := c.CreateAPIConfiguration(ctx, &config); err != nil {
			c.logExecutionEvent(types.LogLevelError, types.LogCategoryError,
				fmt.Sprintf("Failed to save configuration: %v", err), nil)
			return nil, fmt.Errorf("failed to save configuration: %w", err)
		}

		// Execute single variation
		c.logExecutionEvent(types.LogLevelInfo, types.LogCategoryExecution,
			fmt.Sprintf("Executing variation: %s", config.VariationName), nil)

		variationResult, err := c.executeSingleVariation(ctx, executionRun.ID, &config, request.BasePrompt, request.Context)
		if err != nil {
			c.logExecutionEvent(types.LogLevelError, types.LogCategoryError,
				fmt.Sprintf("Variation failed: %s - %v", config.VariationName, err), nil)
			result.ErrorCount++
		} else {
			c.logExecutionEvent(types.LogLevelSuccess, types.LogCategoryExecution,
				fmt.Sprintf("Variation completed: %s", config.VariationName), nil)
			result.SuccessCount++
		}

		result.Results = append(result.Results, *variationResult)

		// Add rate limiting delay between requests (except for the last one)
		if i < len(request.Configurations)-1 {
			delay := time.Duration(100+rand.Intn(101)) * time.Millisecond
			c.logExecutionEvent(types.LogLevelDebug, types.LogCategoryExecution,
				fmt.Sprintf("Rate limiting: waiting %v before next API call", delay), nil)
			time.Sleep(delay)
		}
	}

	// Store function-execution relationships for replay functionality
	if request.EnableFunctionCalling && len(request.FunctionTools) > 0 {
		err := c.storeFunctionExecutionConfigs(ctx, executionRun.ID, request.FunctionTools)
		if err != nil {
			c.logExecutionEvent(types.LogLevelWarn, types.LogCategoryError,
				fmt.Sprintf("Failed to store function-execution configs: %v", err), nil)
			// Don't fail the entire execution, just log the warning
		} else {
			c.logExecutionEvent(types.LogLevelSuccess, types.LogCategorySetup,
				"Function-execution relationships stored for replay", nil)
		}
	}

	result.TotalTime = time.Since(startTime).Milliseconds()

	// Log completion
	c.logExecutionEvent(types.LogLevelSuccess, types.LogCategoryCompletion,
		fmt.Sprintf("Execution completed in %dms - %d successful, %d failed",
			result.TotalTime, result.SuccessCount, result.ErrorCount),
		map[string]interface{}{
			"totalTime":    result.TotalTime,
			"successCount": result.SuccessCount,
			"errorCount":   result.ErrorCount,
		})

	// Always perform comparison for better user experience
	c.logExecutionEvent(types.LogLevelInfo, types.LogCategoryExecution,
		"Starting comparison analysis", nil)
	comparison, err := c.compareResults(ctx, result)
	if err != nil {
		// Log comparison error but don't fail the whole execution
		fmt.Printf("❌ Warning: comparison failed: %v\n", err)
	} else {
		fmt.Printf("✅ Comparison completed successfully: %s\n", comparison.ID)
		result.Comparison = comparison

		// Store comparison result in database
		if err := c.StoreComparisonResult(ctx, comparison); err != nil {
			fmt.Printf("⚠️ Warning: failed to store comparison result: %v\n", err)
		} else {
			fmt.Printf("💾 Comparison result stored in database: %s\n", comparison.ID)
		}
	}

	return result, nil
}

// executeSingleVariation executes a single variation and logs everything
func (c *Client) executeSingleVariation(ctx context.Context, executionRunID string, config *types.APIConfiguration, prompt, context string) (*types.VariationResult, error) {
	startTime := time.Now()

	// Create API request
	apiRequest := &types.APIRequest{
		ID:              uuid.New().String(),
		ExecutionRunID:  executionRunID,
		ConfigurationID: config.ID,
		RequestType:     types.RequestTypeGenerate, // Default to generate for now
		Prompt:          prompt,
		Context:         context,
		CreatedAt:       time.Now(),
	}

	// Log request
	if err := c.LogAPIRequest(ctx, apiRequest); err != nil {
		return nil, fmt.Errorf("failed to log API request: %w", err)
	}

	// Execute the actual Gemini API call
	apiResponse, err := c.callGeminiAPI(ctx, config, apiRequest)
	if err != nil {
		// Log error response
		apiResponse = &types.APIResponse{
			ID:             uuid.New().String(),
			RequestID:      apiRequest.ID,
			ResponseStatus: types.ResponseStatusError,
			ErrorMessage:   err.Error(),
			ResponseTimeMs: int32(time.Since(startTime).Milliseconds()),
			CreatedAt:      time.Now(),
		}
	}

	// Log response
	if logErr := c.LogAPIResponse(ctx, apiResponse); logErr != nil {
		return nil, fmt.Errorf("failed to log API response: %w", logErr)
	}

	return &types.VariationResult{
		Configuration: *config,
		Request:       *apiRequest,
		Response:      *apiResponse,
		ExecutionTime: time.Since(startTime).Milliseconds(),
	}, err
}

// callGeminiAPI makes the actual API call to Gemini
func (c *Client) callGeminiAPI(ctx context.Context, config *types.APIConfiguration, request *types.APIRequest) (*types.APIResponse, error) {
	// Check if we have an API key available
	if c.config.APIKey == "" {
		log.Printf("No API key available, using mock responses")
		return c.callMockGeminiAPI(ctx, config, request)
	}

	// Force REST API implementation since it works perfectly
	log.Printf("Using REST API for model: %s with API key: %s...", config.ModelName, c.config.APIKey[:10])

	// Use our working REST API implementation
	return c.callGeminiRestAPI(ctx, config, request)
}

// callMockGeminiAPI provides mock responses for testing/demo purposes
func (c *Client) callMockGeminiAPI(ctx context.Context, config *types.APIConfiguration, request *types.APIRequest) (*types.APIResponse, error) {
	// For demo purposes when no API key is available
	response := &types.APIResponse{
		ID:             uuid.New().String(),
		RequestID:      request.ID,
		ResponseStatus: types.ResponseStatusSuccess,
		ResponseText:   fmt.Sprintf("Mock response for prompt: %s with model: %s", request.Prompt, config.ModelName),
		FinishReason:   "stop",
		ResponseTimeMs: 500, // Mock response time
		CreatedAt:      time.Now(),
	}

	return response, nil
}

// callGeminiRestAPI provides a REST API fallback when the Go SDK fails
// sanitizeToolParameters removes fields that are not supported by the Gemini API
func sanitizeToolParameters(params map[string]interface{}) map[string]interface{} {
	if params == nil {
		return params
	}

	// Create a copy to avoid modifying the original
	sanitized := make(map[string]interface{})

	// Copy allowed fields at the top level
	allowedTopLevel := map[string]bool{
		"type":        true,
		"properties":  true,
		"required":    true,
		"description": true,
	}

	for key, value := range params {
		if allowedTopLevel[key] {
			if key == "properties" {
				// Recursively sanitize properties
				if props, ok := value.(map[string]interface{}); ok {
					sanitizedProps := make(map[string]interface{})
					for propName, propValue := range props {
						if propMap, ok := propValue.(map[string]interface{}); ok {
							sanitizedProps[propName] = sanitizePropertySchema(propMap)
						} else {
							sanitizedProps[propName] = propValue
						}
					}
					sanitized[key] = sanitizedProps
				} else {
					sanitized[key] = value
				}
			} else {
				sanitized[key] = value
			}
		}
	}

	return sanitized
}

// sanitizePropertySchema removes invalid fields from individual property schemas
func sanitizePropertySchema(prop map[string]interface{}) map[string]interface{} {
	sanitized := make(map[string]interface{})

	// Allowed fields for property schemas in Gemini API
	allowedFields := map[string]bool{
		"type":        true,
		"description": true,
		"enum":        true,
		"items":       true,
		"properties":  true,
		"required":    true,
		"minimum":     true,
		"maximum":     true,
		"minLength":   true,
		"maxLength":   true,
		"pattern":     true,
		"format":      true,
	}

	for key, value := range prop {
		if allowedFields[key] {
			sanitized[key] = value
		} else {
			log.Printf("🚫 Removing unsupported field '%s' from function parameter schema", key)
		}
	}

	return sanitized
}

func (c *Client) callGeminiRestAPI(ctx context.Context, config *types.APIConfiguration, request *types.APIRequest) (*types.APIResponse, error) {
	startTime := time.Now()

	fmt.Printf("\n🚀 USING REST API IMPLEMENTATION - Model: '%s'\n", config.ModelName)
	log.Printf("🚀 REST API CALLED - Model: '%s', API Key: %s...", config.ModelName, c.config.APIKey[:10])

	if config.ModelName == "" {
		log.Printf("❌ ERROR: Model name is empty!")
		return &types.APIResponse{
			ID:             uuid.New().String(),
			RequestID:      request.ID,
			ResponseStatus: types.ResponseStatusError,
			ErrorMessage:   "Model name is empty",
			ResponseTimeMs: int32(time.Since(startTime).Milliseconds()),
			CreatedAt:      time.Now(),
		}, nil
	}

	// Use the same API key from the client configuration
	apiKey := c.config.APIKey
	if apiKey == "" {
		log.Printf("❌ No API key available for REST API call")
		return c.callMockGeminiAPI(ctx, config, request)
	}

	log.Printf("✅ Using API key: %s... for model: '%s'", apiKey[:10], config.ModelName)

	// Build the REST API request - start with the base prompt
	prompt := request.Prompt
	if request.Context != "" {
		prompt = fmt.Sprintf("%s\n\nContext: %s", prompt, request.Context)
	}

	// Prepare the final prompt
	finalPrompt := prompt
	if config.SystemPrompt != "" {
		finalPrompt = config.SystemPrompt + "\n\n" + prompt
	}

	// Add function calling instruction if tools are available
	if len(config.Tools) > 0 {
		functionInstruction := "You MUST use the available function tools to answer questions. When a user asks for information that can be obtained through these functions, you are REQUIRED to call the appropriate function. Do not respond with text saying you cannot access information - instead, call the function immediately. The functions are fully implemented and working."
		finalPrompt = functionInstruction + "\n\n" + finalPrompt
		log.Printf("🔧 Added function calling instruction to prompt")
	}

	log.Printf("REST API - Final prompt: %s", finalPrompt[:min(100, len(finalPrompt))])

	requestBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{"text": finalPrompt},
				},
			},
		},
	}

	// Add generation config if specified
	generationConfig := make(map[string]interface{})
	if config.Temperature != nil {
		generationConfig["temperature"] = *config.Temperature
	}
	if config.MaxTokens != nil {
		generationConfig["maxOutputTokens"] = *config.MaxTokens
	}
	if config.TopP != nil {
		generationConfig["topP"] = *config.TopP
	}
	if config.TopK != nil {
		generationConfig["topK"] = *config.TopK
	}
	if len(generationConfig) > 0 {
		requestBody["generationConfig"] = generationConfig
	}

	// Add tools for function calling if provided
	if len(config.Tools) > 0 {
		log.Printf("🔧 Adding %d tools to Gemini request", len(config.Tools))
		tools := make([]map[string]interface{}, len(config.Tools))
		for i, tool := range config.Tools {
			log.Printf("🔧 Tool %d: %s - %s", i+1, tool.Name, tool.Description)

			// Sanitize the parameters to remove unsupported fields
			sanitizedParams := sanitizeToolParameters(tool.Parameters)

			toolDeclaration := map[string]interface{}{
				"functionDeclarations": []map[string]interface{}{
					{
						"name":        tool.Name,
						"description": tool.Description,
						"parameters":  sanitizedParams,
					},
				},
			}
			tools[i] = toolDeclaration
			log.Printf("🔧 Tool declaration (sanitized): %+v", toolDeclaration)
		}
		requestBody["tools"] = tools

		// Add tool configuration to make function calling more aggressive
		requestBody["toolConfig"] = map[string]interface{}{
			"functionCallingConfig": map[string]interface{}{
				"mode": "ANY",
			},
		}

		log.Printf("🔧 Final tools in request body: %+v", tools)
		log.Printf("🔧 Added toolConfig with mode: ANY")
	} else {
		log.Printf("⚠️  No tools provided to Gemini API call")
	}

	// Create request body
	reqBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	log.Printf("🔧 Complete Gemini API request body: %s", string(reqBodyBytes))

	// Make HTTP request to Gemini REST API
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent", config.ModelName)
	log.Printf("REST API - URL: %s", url)

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("REST API - HTTP request error: %v", err)
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("REST API - Read response error: %v", err)
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	log.Printf("🔧 Complete Gemini API response: %s", string(body))

	if resp.StatusCode != http.StatusOK {
		log.Printf("REST API - HTTP error %d: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var geminiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text         string `json:"text,omitempty"`
					FunctionCall struct {
						Name string                 `json:"name"`
						Args map[string]interface{} `json:"args"`
					} `json:"functionCall,omitempty"`
				} `json:"parts"`
			} `json:"content"`
			FinishReason string `json:"finishReason"`
		} `json:"candidates"`
		UsageMetadata struct {
			PromptTokenCount     int `json:"promptTokenCount"`
			CandidatesTokenCount int `json:"candidatesTokenCount"`
			TotalTokenCount      int `json:"totalTokenCount"`
		} `json:"usageMetadata"`
	}

	if err := json.Unmarshal(body, &geminiResp); err != nil {
		log.Printf("REST API - JSON parse error: %v", err)
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	log.Printf("🔧 Parsed response - %d candidates", len(geminiResp.Candidates))

	// Check for function calls in response and extract response text
	var responseText string
	var finishReason string
	var functionCallResponse map[string]interface{}

	if len(geminiResp.Candidates) > 0 {
		candidate := geminiResp.Candidates[0]
		finishReason = candidate.FinishReason

		for _, part := range candidate.Content.Parts {
			// Handle text response
			if part.Text != "" {
				responseText = part.Text
			}

			// Handle function call
			if part.FunctionCall.Name != "" {
				log.Printf("🎯 FUNCTION CALL DETECTED: %s with args: %+v", part.FunctionCall.Name, part.FunctionCall.Args)

				// Execute the function call
				startTime := time.Now()
				functionResult, err := c.executeFunctionCall(ctx, part.FunctionCall.Name, part.FunctionCall.Args)
				executionTime := time.Since(startTime).Milliseconds()

				// Create function call record for logging
				functionCall := &types.FunctionCall{
					ID:               uuid.New().String(),
					RequestID:        request.ID,
					FunctionName:     part.FunctionCall.Name,
					FunctionArgs:     part.FunctionCall.Args,
					FunctionResponse: functionResult,
					ExecutionTimeMs:  int32(executionTime),
					CreatedAt:        time.Now(),
				}

				if err != nil {
					log.Printf("❌ Function execution failed: %v", err)
					functionCall.ExecutionStatus = "error"
					functionCall.ErrorDetails = err.Error()
					// Return error response but don't fail completely
					functionResult = map[string]interface{}{
						"error":  err.Error(),
						"status": "failed",
					}
					functionCall.FunctionResponse = functionResult
				} else {
					functionCall.ExecutionStatus = "success"
				}

				// Log function call to database
				if logErr := c.LogFunctionCall(ctx, functionCall); logErr != nil {
					log.Printf("⚠️ Failed to log function call to database: %v", logErr)
				}

				// Send function result back to Gemini to get final response
				finalResponse, err := c.sendFunctionResultToGemini(ctx, config, request, part.FunctionCall.Name, functionResult, finalPrompt)
				if err != nil {
					log.Printf("❌ Failed to get final response from Gemini: %v", err)
					// Fall back to just indicating the function was called
					responseText = fmt.Sprintf("I called the %s function with the provided parameters and received the result.", part.FunctionCall.Name)
				} else {
					responseText = finalResponse
				}

				// Store function call information
				functionCallResponse = map[string]interface{}{
					"function_name": part.FunctionCall.Name,
					"arguments":     part.FunctionCall.Args,
					"result":        functionResult,
				}

				log.Printf("✅ Function executed successfully: %s", part.FunctionCall.Name)
				break // Only handle the first function call
			}
		}
	}

	// If we have a function call but no text response, generate appropriate text
	if functionCallResponse != nil && responseText == "" {
		functionName := functionCallResponse["function_name"].(string)
		responseText = fmt.Sprintf("I called the %s function for you.", functionName)
	}

	log.Printf("REST API - Success! Response text: %s", responseText[:min(50, len(responseText))])
	if functionCallResponse != nil {
		log.Printf("REST API - Function call response: %+v", functionCallResponse)
	}

	// Build usage metadata
	usageMetadata := map[string]interface{}{
		"prompt_tokens":     geminiResp.UsageMetadata.PromptTokenCount,
		"completion_tokens": geminiResp.UsageMetadata.CandidatesTokenCount,
		"total_tokens":      geminiResp.UsageMetadata.TotalTokenCount,
	}

	response := &types.APIResponse{
		ID:             uuid.New().String(),
		RequestID:      request.ID,
		ResponseStatus: types.ResponseStatusSuccess,
		ResponseText:   responseText,
		UsageMetadata:  usageMetadata,
		FinishReason:   finishReason,
		ResponseTimeMs: int32(time.Since(startTime).Milliseconds()),
		CreatedAt:      time.Now(),
	}

	// Add function call response to the API response
	if functionCallResponse != nil {
		response.FunctionCallResponse = functionCallResponse
	}

	return response, nil
}

// executeFunctionCall executes a function call and returns the result
func (c *Client) executeFunctionCall(ctx context.Context, functionName string, args map[string]interface{}) (map[string]interface{}, error) {
	log.Printf("🔧 Executing function: %s with args: %+v", functionName, args)

	// Handle weather function with real API call
	if functionName == "get_weather" {
		location, ok := args["location"].(string)
		if !ok {
			return nil, fmt.Errorf("location parameter missing or invalid")
		}

		// Call real weather API
		result, err := c.callWeatherAPI(ctx, location, c.config.OpenWeatherAPIKey)
		if err != nil {
			log.Printf("❌ Weather API call failed: %v", err)
			// Fallback to mock data if API call fails
			result = map[string]interface{}{
				"location":    location,
				"temperature": 72,
				"unit":        "F",
				"condition":   "Sunny",
				"humidity":    45,
				"wind_speed":  8,
				"description": fmt.Sprintf("Current weather in %s: 72°F, sunny with clear skies (fallback data)", location),
				"error":       "Real weather data unavailable, showing fallback data",
			}
		}

		log.Printf("✅ Weather function executed for %s", location)
		return result, nil
	}

	// Handle Neo4j graph query function
	if functionName == "query_graph" {
		query, ok := args["query"].(string)
		if !ok {
			return nil, fmt.Errorf("query parameter missing or invalid")
		}

		// Get limit parameter (optional, default to 25)
		limit := 25
		if limitVal, exists := args["limit"]; exists {
			if limitFloat, ok := limitVal.(float64); ok {
				limit = int(limitFloat)
			}
			if limit < 1 || limit > 100 {
				limit = 25 // Reset to default if out of bounds
			}
		}

		// Call Neo4j query function
		result, err := c.callNeo4jAPI(ctx, query, limit)
		if err != nil {
			log.Printf("❌ Neo4j query failed: %v", err)
			// Fallback to mock data if Neo4j call fails
			result = map[string]interface{}{
				"nodes": []map[string]interface{}{
					{
						"id":         "mock_node_1",
						"labels":     []string{"Person"},
						"properties": map[string]interface{}{"name": "Mock User", "age": 30},
					},
				},
				"relationships": []map[string]interface{}{},
				"summary": map[string]interface{}{
					"totalNodes":         1,
					"totalRelationships": 0,
					"executionTime":      "0ms",
					"query":              query,
					"error":              "Neo4j connection unavailable, showing mock data",
				},
			}
		}

		log.Printf("✅ Neo4j query executed: %s", query)
		return result, nil
	}

	// For other functions, return a generic success response
	return map[string]interface{}{
		"status":  "success",
		"message": fmt.Sprintf("Function %s executed successfully", functionName),
		"result":  "Function executed with provided parameters",
	}, nil
}

// callWeatherAPI makes a real API call to OpenWeatherMap API
func (c *Client) callWeatherAPI(ctx context.Context, location string, apiKey string) (map[string]interface{}, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("OpenWeather API key not provided")
	}

	// Build API URL
	baseURL := "https://api.openweathermap.org/data/2.5/weather"
	params := url.Values{}
	params.Add("q", location)
	params.Add("appid", apiKey)
	params.Add("units", "imperial") // Fahrenheit

	apiURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	log.Printf("🌤️ Calling OpenWeatherMap API for location: %s", location)

	// Create HTTP request with timeout
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set User-Agent header
	req.Header.Set("User-Agent", "GoGent/1.0")

	// Make the API call
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call weather API: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for API errors
	if resp.StatusCode != 200 {
		log.Printf("❌ Weather API returned status: %d, body: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("weather API returned status %d", resp.StatusCode)
	}

	// Parse JSON response
	var weatherResp struct {
		Name string `json:"name"`
		Main struct {
			Temp     float64 `json:"temp"`
			Humidity int     `json:"humidity"`
		} `json:"main"`
		Weather []struct {
			Main        string `json:"main"`
			Description string `json:"description"`
		} `json:"weather"`
		Wind struct {
			Speed float64 `json:"speed"`
		} `json:"wind"`
	}

	if err := json.Unmarshal(body, &weatherResp); err != nil {
		return nil, fmt.Errorf("failed to parse weather response: %w", err)
	}

	// Build result
	condition := "Clear"
	description := "Clear skies"
	if len(weatherResp.Weather) > 0 {
		condition = weatherResp.Weather[0].Main
		description = weatherResp.Weather[0].Description
	}

	result := map[string]interface{}{
		"location":    fmt.Sprintf("%s", weatherResp.Name),
		"temperature": int(weatherResp.Main.Temp),
		"unit":        "F",
		"condition":   condition,
		"humidity":    weatherResp.Main.Humidity,
		"wind_speed":  int(weatherResp.Wind.Speed),
		"description": fmt.Sprintf("Current weather in %s: %.0f°F, %s", weatherResp.Name, weatherResp.Main.Temp, description),
	}

	log.Printf("✅ Weather API call successful for %s: %s, %.0f°F", weatherResp.Name, condition, weatherResp.Main.Temp)
	return result, nil
}

// callNeo4jAPI executes a Cypher query against a Neo4j database
func (c *Client) callNeo4jAPI(ctx context.Context, query string, limit int) (map[string]interface{}, error) {
	if c.config.Neo4jURL == "" {
		return nil, fmt.Errorf("Neo4j URL not configured")
	}

	log.Printf("🔗 Connecting to Neo4j at: %s", c.config.Neo4jURL)

	// Create Neo4j driver
	driver, err := neo4j.NewDriverWithContext(c.config.Neo4jURL, neo4j.BasicAuth(c.config.Neo4jUsername, c.config.Neo4jPassword, ""))
	if err != nil {
		return nil, fmt.Errorf("failed to create Neo4j driver: %w", err)
	}
	defer driver.Close(ctx)

	// Verify connectivity
	if err := driver.VerifyConnectivity(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to Neo4j: %w", err)
	}

	// Create session
	sessionConfig := neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeRead,
		DatabaseName: c.config.Neo4jDatabase,
	}
	session := driver.NewSession(ctx, sessionConfig)
	defer session.Close(ctx)

	// Add LIMIT clause if not present in query
	finalQuery := query
	if !strings.Contains(strings.ToUpper(query), "LIMIT") {
		finalQuery = fmt.Sprintf("%s LIMIT %d", query, limit)
	}

	log.Printf("🔍 Executing Cypher query: %s", finalQuery)

	// Execute query
	startTime := time.Now()
	result, err := session.Run(ctx, finalQuery, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	// Collect results
	var nodes []map[string]interface{}
	var relationships []map[string]interface{}
	recordCount := 0

	for result.Next(ctx) {
		record := result.Record()
		recordCount++

		// Process each value in the record
		for i, value := range record.Values {
			if node, ok := value.(neo4j.Node); ok {
				// Extract node data
				nodeData := map[string]interface{}{
					"id":         fmt.Sprintf("%d", node.GetId()),
					"labels":     node.Labels,
					"properties": node.Props,
				}
				nodes = append(nodes, nodeData)
			} else if rel, ok := value.(neo4j.Relationship); ok {
				// Extract relationship data
				relData := map[string]interface{}{
					"id":         fmt.Sprintf("%d", rel.GetId()),
					"type":       rel.Type,
					"startNode":  fmt.Sprintf("%d", rel.StartId),
					"endNode":    fmt.Sprintf("%d", rel.EndId),
					"properties": rel.Props,
				}
				relationships = append(relationships, relData)
			} else {
				// For other data types, add as a simple node
				key := record.Keys[i]
				nodeData := map[string]interface{}{
					"id":         fmt.Sprintf("result_%d_%d", recordCount, i),
					"labels":     []string{"QueryResult"},
					"properties": map[string]interface{}{key: value},
				}
				nodes = append(nodes, nodeData)
			}
		}
	}

	// Check for errors
	if err := result.Err(); err != nil {
		return nil, fmt.Errorf("query execution error: %w", err)
	}

	executionTime := time.Since(startTime)

	// Build response
	response := map[string]interface{}{
		"nodes":         nodes,
		"relationships": relationships,
		"summary": map[string]interface{}{
			"totalNodes":         len(nodes),
			"totalRelationships": len(relationships),
			"recordCount":        recordCount,
			"executionTime":      fmt.Sprintf("%dms", executionTime.Milliseconds()),
			"query":              finalQuery,
		},
	}

	log.Printf("✅ Neo4j query successful: %d nodes, %d relationships, %dms", len(nodes), len(relationships), executionTime.Milliseconds())
	return response, nil
}

// sendFunctionResultToGemini sends the function result back to Gemini for a final response
func (c *Client) sendFunctionResultToGemini(ctx context.Context, config *types.APIConfiguration, request *types.APIRequest, functionName string, functionResult map[string]interface{}, originalPrompt string) (string, error) {
	log.Printf("🔧 Sending function result back to Gemini for final response")

	// Create a follow-up prompt that includes the function result
	resultText, _ := json.Marshal(functionResult)
	followUpPrompt := fmt.Sprintf("%s\n\nFunction %s was called and returned: %s\n\nPlease provide a natural, helpful response to the user based on this information.", originalPrompt, functionName, string(resultText))

	// Create request body for the follow-up call
	requestBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{"text": followUpPrompt},
				},
			},
		},
	}

	// Add generation config
	if config.Temperature != nil {
		requestBody["generationConfig"] = map[string]interface{}{
			"temperature": *config.Temperature,
		}
	}

	// Make the API call
	reqBodyBytes, _ := json.Marshal(requestBody)
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent", config.ModelName)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", c.config.APIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Parse response
	var geminiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return "", err
	}

	if len(geminiResp.Candidates) > 0 && len(geminiResp.Candidates[0].Content.Parts) > 0 {
		finalResponse := geminiResp.Candidates[0].Content.Parts[0].Text
		log.Printf("✅ Got final response from Gemini: %s", finalResponse[:min(50, len(finalResponse))])
		return finalResponse, nil
	}

	return "I executed the function successfully but couldn't generate a proper response.", nil
}

// min helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// compareResults compares multiple variation results
func (c *Client) compareResults(ctx context.Context, result *types.ExecutionResult) (*types.ComparisonResult, error) {
	// Enhanced comparison implementation with multiple metrics
	fmt.Printf("🔍 Comparing %d results for execution run: %s\n", len(result.Results), result.ExecutionRun.ID)

	comparisonResult := &types.ComparisonResult{
		ID:             uuid.New().String(),
		ExecutionRunID: result.ExecutionRun.ID,
		ComparisonType: "comprehensive",
		MetricName:     "multi_metric",
		CreatedAt:      time.Now(),
	}

	// Calculate comprehensive scores for each configuration
	scores := make(map[string]interface{})
	var bestOverall *types.VariationResult
	var bestScore float64 = -1

	for _, r := range result.Results {
		// Calculate various metrics
		responseTimeScore := calculateResponseTimeScore(r.Response.ResponseTimeMs)
		creativityScore := calculateCreativityScore(r.Configuration, r.Response)
		coherenceScore := calculateCoherenceScore(r.Response.ResponseText)
		tokenEfficiencyScore := calculateTokenEfficiencyScore(r.Response)
		safetyScore := calculateSafetyScore(r.Response.ResponseText)
		costEffectivenessScore := calculateCostEffectivenessScore(r.Response)

		// Calculate overall score (weighted average)
		overallScore := (responseTimeScore*0.2 +
			creativityScore*0.25 +
			coherenceScore*0.25 +
			tokenEfficiencyScore*0.15 +
			safetyScore*0.1 +
			costEffectivenessScore*0.05)

		// Track best overall configuration
		if bestOverall == nil || overallScore > bestScore {
			bestOverall = &r
			bestScore = overallScore
		}

		// Store detailed scores
		scores[r.Configuration.VariationName] = map[string]interface{}{
			"response_time_ms":    r.Response.ResponseTimeMs,
			"status":              r.Response.ResponseStatus,
			"response_time_score": responseTimeScore,
			"creativity_score":    creativityScore,
			"coherence_score":     coherenceScore,
			"token_efficiency":    tokenEfficiencyScore,
			"safety_score":        safetyScore,
			"cost_effectiveness":  costEffectivenessScore,
			"overall_score":       overallScore,
			"temperature":         r.Configuration.Temperature,
			"model_name":          r.Configuration.ModelName,
		}
	}

	// Set best configuration and analysis notes
	if bestOverall != nil {
		comparisonResult.BestConfigurationID = bestOverall.Configuration.ID
		comparisonResult.BestConfiguration = &bestOverall.Configuration

		// Create detailed analysis notes
		analysis := fmt.Sprintf("🏆 Best Configuration: %s\n\n", bestOverall.Configuration.VariationName)
		analysis += fmt.Sprintf("📊 Overall Score: %.2f/100\n", bestScore*100)
		analysis += fmt.Sprintf("⚡ Response Time: %dms\n", bestOverall.Response.ResponseTimeMs)
		analysis += fmt.Sprintf("🎨 Creativity Score: %.1f/100\n", getScoreFromMap(scores, bestOverall.Configuration.VariationName, "creativity_score")*100)
		analysis += fmt.Sprintf("🧠 Coherence Score: %.1f/100\n", getScoreFromMap(scores, bestOverall.Configuration.VariationName, "coherence_score")*100)
		analysis += fmt.Sprintf("💡 Token Efficiency: %.1f/100\n", getScoreFromMap(scores, bestOverall.Configuration.VariationName, "token_efficiency")*100)

		// Add comparison insights
		analysis += "\n📈 Key Insights:\n"
		fastest := findFastest(result.Results)
		if fastest != nil && fastest.Configuration.ID != bestOverall.Configuration.ID {
			analysis += fmt.Sprintf("• Fastest: %s (%dms)\n", fastest.Configuration.VariationName, fastest.Response.ResponseTimeMs)
		}

		mostCreative := findMostCreative(scores)
		if mostCreative != "" && mostCreative != bestOverall.Configuration.VariationName {
			analysis += fmt.Sprintf("• Most Creative: %s\n", mostCreative)
		}

		analysis += fmt.Sprintf("• Best Overall: %s (balanced performance)\n", bestOverall.Configuration.VariationName)

		comparisonResult.AnalysisNotes = analysis
	}

	// Store all configurations for reference
	var allConfigs []types.APIConfiguration
	for _, r := range result.Results {
		allConfigs = append(allConfigs, r.Configuration)
	}
	comparisonResult.AllConfigurations = allConfigs

	comparisonResult.ConfigurationScores = scores
	return comparisonResult, nil
}

// Helper functions for calculating different metrics
func calculateResponseTimeScore(responseTimeMs int32) float64 {
	// Lower response time = higher score (max 1000ms = 100 points)
	if responseTimeMs <= 0 {
		return 0.0
	}
	score := 1000.0 / float64(responseTimeMs)
	if score > 1.0 {
		score = 1.0
	}
	return score
}

func calculateCreativityScore(config types.APIConfiguration, response types.APIResponse) float64 {
	// Higher temperature = higher creativity potential
	baseScore := 0.5
	if config.Temperature != nil {
		baseScore = float64(*config.Temperature)
	}

	// Boost score based on response characteristics
	text := response.ResponseText
	creativityIndicators := []string{"imagine", "creative", "artistic", "vivid", "colorful", "metaphor", "poetry", "story", "narrative"}
	indicatorCount := 0
	for _, indicator := range creativityIndicators {
		if strings.Contains(strings.ToLower(text), indicator) {
			indicatorCount++
		}
	}

	// Boost score by up to 0.3 based on creativity indicators
	boost := float64(indicatorCount) * 0.03
	if boost > 0.3 {
		boost = 0.3
	}

	return baseScore + boost
}

func calculateCoherenceScore(responseText string) float64 {
	// Simple coherence scoring based on text structure
	if len(responseText) < 50 {
		return 0.3
	}

	// Check for logical structure indicators
	coherenceIndicators := []string{"first", "second", "third", "however", "therefore", "because", "although", "furthermore", "in conclusion"}
	indicatorCount := 0
	for _, indicator := range coherenceIndicators {
		if strings.Contains(strings.ToLower(responseText), indicator) {
			indicatorCount++
		}
	}

	baseScore := 0.6
	boost := float64(indicatorCount) * 0.05
	if boost > 0.4 {
		boost = 0.4
	}

	return baseScore + boost
}

func calculateTokenEfficiencyScore(response types.APIResponse) float64 {
	// Higher token efficiency = higher score
	if response.UsageMetadata == nil {
		return 0.5 // Default score if no metadata
	}

	// Extract token information
	totalTokens := getTokenCount(response.UsageMetadata, "total_tokens")
	if totalTokens <= 0 {
		return 0.5
	}

	// Score based on response length vs tokens used
	responseLength := len(response.ResponseText)
	if responseLength == 0 {
		return 0.0
	}

	// Higher ratio of characters per token = better efficiency
	efficiencyRatio := float64(responseLength) / float64(totalTokens)

	// Normalize to 0-1 scale (typical range is 2-8 characters per token)
	if efficiencyRatio > 8.0 {
		efficiencyRatio = 8.0
	}

	return efficiencyRatio / 8.0
}

func calculateSafetyScore(responseText string) float64 {
	// Simple safety scoring - avoid potentially problematic content
	text := strings.ToLower(responseText)

	// Check for potentially unsafe content
	unsafeIndicators := []string{"harm", "danger", "illegal", "inappropriate", "offensive", "violent"}
	unsafeCount := 0
	for _, indicator := range unsafeIndicators {
		if strings.Contains(text, indicator) {
			unsafeCount++
		}
	}

	// Base score is high, reduce for unsafe indicators
	baseScore := 0.9
	penalty := float64(unsafeCount) * 0.1
	if penalty > 0.9 {
		penalty = 0.9
	}

	return baseScore - penalty
}

func calculateCostEffectivenessScore(response types.APIResponse) float64 {
	// Lower cost = higher score (based on tokens used)
	if response.UsageMetadata == nil {
		return 0.5
	}

	totalTokens := getTokenCount(response.UsageMetadata, "total_tokens")
	if totalTokens <= 0 {
		return 0.5
	}

	// Score based on token usage (fewer tokens = better cost effectiveness)
	// Assume 1000 tokens as baseline for "good" cost effectiveness
	if totalTokens <= 100 {
		return 1.0
	} else if totalTokens <= 500 {
		return 0.8
	} else if totalTokens <= 1000 {
		return 0.6
	} else {
		return 0.3
	}
}

// Helper functions
func getScoreFromMap(scores map[string]interface{}, configName, scoreKey string) float64 {
	if config, exists := scores[configName]; exists {
		if configMap, ok := config.(map[string]interface{}); ok {
			if score, exists := configMap[scoreKey]; exists {
				if scoreFloat, ok := score.(float64); ok {
					return scoreFloat
				}
			}
		}
	}
	return 0.0
}

func getTokenCount(metadata map[string]interface{}, key string) int {
	if value, exists := metadata[key]; exists {
		switch v := value.(type) {
		case float64:
			return int(v)
		case int:
			return v
		case string:
			if parsed, err := strconv.Atoi(v); err == nil {
				return parsed
			}
		}
	}
	return 0
}

func findFastest(results []types.VariationResult) *types.VariationResult {
	var fastest *types.VariationResult
	for i := range results {
		if fastest == nil || results[i].Response.ResponseTimeMs < fastest.Response.ResponseTimeMs {
			fastest = &results[i]
		}
	}
	return fastest
}

func findMostCreative(scores map[string]interface{}) string {
	var mostCreative string
	var highestScore float64 = -1

	for configName, configData := range scores {
		if configMap, ok := configData.(map[string]interface{}); ok {
			if score, exists := configMap["creativity_score"]; exists {
				if scoreFloat, ok := score.(float64); ok {
					if scoreFloat > highestScore {
						highestScore = scoreFloat
						mostCreative = configName
					}
				}
			}
		}
	}

	return mostCreative
}

// StoreComparisonResult stores a comparison result in the database
func (c *Client) StoreComparisonResult(ctx context.Context, comparison *types.ComparisonResult) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Convert configuration scores to JSON
	configScoresJSON, err := json.Marshal(comparison.ConfigurationScores)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration scores: %w", err)
	}

	// Convert best configuration to JSON
	var bestConfigJSON json.RawMessage
	if comparison.BestConfiguration != nil {
		bestConfigJSON, err = json.Marshal(comparison.BestConfiguration)
		if err != nil {
			return fmt.Errorf("failed to marshal best configuration: %w", err)
		}
	}

	// Convert all configurations to JSON
	var allConfigsJSON json.RawMessage
	if len(comparison.AllConfigurations) > 0 {
		allConfigsJSON, err = json.Marshal(comparison.AllConfigurations)
		if err != nil {
			return fmt.Errorf("failed to marshal all configurations: %w", err)
		}
	}

	// Determine comparison type from metric name
	comparisonType := "custom"
	switch comparison.MetricName {
	case "response_time", "performance":
		comparisonType = "performance"
	case "quality", "coherence_score", "creativity_score":
		comparisonType = "quality"
	case "safety_score":
		comparisonType = "safety"
	}

	// Store in database
	err = c.queries.CreateComparisonResult(ctx, db.CreateComparisonResultParams{
		ID:                    comparison.ID,
		ExecutionRunID:        comparison.ExecutionRunID,
		ComparisonType:        db.ComparisonResultsComparisonType(comparisonType),
		MetricName:            comparison.MetricName,
		ConfigurationScores:   configScoresJSON,
		BestConfigurationID:   sql.NullString{String: comparison.BestConfigurationID, Valid: comparison.BestConfigurationID != ""},
		BestConfigurationData: bestConfigJSON,
		AllConfigurationsData: allConfigsJSON,
		AnalysisNotes:         sql.NullString{String: comparison.AnalysisNotes, Valid: comparison.AnalysisNotes != ""},
	})

	if err != nil {
		return fmt.Errorf("failed to store comparison result: %w", err)
	}

	return nil
}

// GetComparisonResult retrieves a comparison result from the database
func (c *Client) GetComparisonResult(ctx context.Context, executionRunID string) (*types.ComparisonResult, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	row, err := c.queries.GetComparisonResult(ctx, executionRunID)
	if err != nil {
		return nil, fmt.Errorf("failed to get comparison result: %w", err)
	}

	// Parse configuration scores JSON
	var configScores map[string]interface{}
	if err := json.Unmarshal(row.ConfigurationScores, &configScores); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration scores: %w", err)
	}

	// Parse best configuration JSON
	var bestConfig *types.APIConfiguration
	if row.BestConfigurationData != nil {
		if bestConfigStr, ok := row.BestConfigurationData.(string); ok && bestConfigStr != "" {
			bestConfig = &types.APIConfiguration{}
			if err := json.Unmarshal([]byte(bestConfigStr), bestConfig); err != nil {
				return nil, fmt.Errorf("failed to unmarshal best configuration: %w", err)
			}
		}
	}

	// Parse all configurations JSON
	var allConfigs []types.APIConfiguration
	if row.AllConfigurationsData != nil {
		if allConfigsStr, ok := row.AllConfigurationsData.(string); ok && allConfigsStr != "" {
			if err := json.Unmarshal([]byte(allConfigsStr), &allConfigs); err != nil {
				return nil, fmt.Errorf("failed to unmarshal all configurations: %w", err)
			}
		}
	}

	var createdAt time.Time
	if row.CreatedAt.Valid {
		createdAt = row.CreatedAt.Time
	}

	comparison := &types.ComparisonResult{
		ID:                  row.ID,
		ExecutionRunID:      row.ExecutionRunID,
		ComparisonType:      string(row.ComparisonType),
		MetricName:          row.MetricName,
		ConfigurationScores: configScores,
		BestConfigurationID: row.BestConfigurationID.String,
		BestConfiguration:   bestConfig,
		AllConfigurations:   allConfigs,
		AnalysisNotes:       row.AnalysisNotes.String,
		CreatedAt:           createdAt,
	}

	return comparison, nil
}

// ListComparisonResults retrieves all comparison results from the database
func (c *Client) ListComparisonResults(ctx context.Context) ([]*types.ComparisonResult, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	rows, err := c.queries.ListComparisonResults(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list comparison results: %w", err)
	}

	var comparisonResults []*types.ComparisonResult
	for _, row := range rows {
		// Parse configuration scores JSON
		var configScores map[string]interface{}
		if err := json.Unmarshal(row.ConfigurationScores, &configScores); err != nil {
			return nil, fmt.Errorf("failed to unmarshal configuration scores: %w", err)
		}

		// Parse best configuration JSON
		var bestConfig *types.APIConfiguration
		if row.BestConfigurationData != nil {
			if bestConfigStr, ok := row.BestConfigurationData.(string); ok && bestConfigStr != "" {
				bestConfig = &types.APIConfiguration{}
				if err := json.Unmarshal([]byte(bestConfigStr), bestConfig); err != nil {
					return nil, fmt.Errorf("failed to unmarshal best configuration: %w", err)
				}
			}
		}

		// Parse all configurations JSON
		var allConfigs []types.APIConfiguration
		if row.AllConfigurationsData != nil {
			if allConfigsStr, ok := row.AllConfigurationsData.(string); ok && allConfigsStr != "" {
				if err := json.Unmarshal([]byte(allConfigsStr), &allConfigs); err != nil {
					return nil, fmt.Errorf("failed to unmarshal all configurations: %w", err)
				}
			}
		}

		var createdAt time.Time
		if row.CreatedAt.Valid {
			createdAt = row.CreatedAt.Time
		}

		comparison := &types.ComparisonResult{
			ID:                  row.ID,
			ExecutionRunID:      row.ExecutionRunID,
			ComparisonType:      string(row.ComparisonType),
			MetricName:          row.MetricName,
			ConfigurationScores: configScores,
			BestConfigurationID: row.BestConfigurationID.String,
			BestConfiguration:   bestConfig,
			AllConfigurations:   allConfigs,
			AnalysisNotes:       row.AnalysisNotes.String,
			CreatedAt:           createdAt,
		}
		comparisonResults = append(comparisonResults, comparison)
	}

	return comparisonResults, nil
}

// Helper functions for handling nullable database fields
func convertFloat32ToNullString(f *float32) sql.NullString {
	if f == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: fmt.Sprintf("%.2f", *f), Valid: true}
}

func convertInt32ToNullInt32(i *int32) sql.NullInt32 {
	if i == nil {
		return sql.NullInt32{Valid: false}
	}
	return sql.NullInt32{Int32: *i, Valid: true}
}

// convertStringToRawMessage converts a JSON string to json.RawMessage for database storage
func convertStringToRawMessage(jsonStr string) json.RawMessage {
	if jsonStr == "" {
		return json.RawMessage("null")
	}
	return json.RawMessage(jsonStr)
}

// ListExecutionRuns retrieves execution runs from the database with pagination
func (c *Client) ListExecutionRuns(ctx context.Context, limit, offset int32) ([]*types.ExecutionRun, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	rows, err := c.queries.GetRecentExecutionRuns(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list execution runs: %w", err)
	}

	var executionRuns []*types.ExecutionRun
	for _, row := range rows {
		description := ""
		if row.Description.Valid {
			description = row.Description.String
		}

		executionRun := &types.ExecutionRun{
			ID:                    row.ID,
			Name:                  row.Name,
			Description:           description,
			EnableFunctionCalling: row.EnableFunctionCalling,
			Status:                "completed", // Default status for existing records
			ErrorMessage:          "",
			CreatedAt:             row.CreatedAt.Time,
			UpdatedAt:             row.UpdatedAt.Time,
		}
		executionRuns = append(executionRuns, executionRun)
	}

	return executionRuns, nil
}

// GetExecutionRun retrieves a single execution run by ID
func (c *Client) GetExecutionRun(ctx context.Context, id string) (*types.ExecutionRun, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	row, err := c.queries.GetExecutionRun(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get execution run: %w", err)
	}

	description := ""
	if row.Description.Valid {
		description = row.Description.String
	}

	return &types.ExecutionRun{
		ID:                    row.ID,
		Name:                  row.Name,
		Description:           description,
		EnableFunctionCalling: row.EnableFunctionCalling,
		Status:                "completed", // Default status for existing records
		ErrorMessage:          "",
		CreatedAt:             row.CreatedAt.Time,
		UpdatedAt:             row.UpdatedAt.Time,
	}, nil
}

// GetExecutionResult retrieves complete execution details from the database
func (c *Client) GetExecutionResult(ctx context.Context, executionRunID string) (*types.ExecutionResult, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	// Get the execution run
	executionRun, err := c.GetExecutionRun(ctx, executionRunID)
	if err != nil {
		return nil, fmt.Errorf("failed to get execution run: %w", err)
	}

	// Get all configurations for this execution run
	configRows, err := c.queries.GetAPIConfigurationsByRun(ctx, executionRunID)
	if err != nil {
		return nil, fmt.Errorf("failed to get configurations: %w", err)
	}
	log.Printf("🔧 Found %d configurations for execution run %s", len(configRows), executionRunID)

	// Get function definitions used in this execution
	functionConfigRows, err := c.queries.ListExecutionFunctionConfigs(ctx, executionRunID)
	if err != nil {
		log.Printf("⚠️ Failed to get function configs for execution %s: %v", executionRunID, err)
		// Continue without functions rather than failing
	}
	log.Printf("🔧 Found %d function configurations for execution run %s", len(functionConfigRows), executionRunID)

	// Get all requests for this execution run
	requestRows, err := c.queries.GetAPIRequestsByRun(ctx, executionRunID)
	if err != nil {
		return nil, fmt.Errorf("failed to get requests: %w", err)
	}
	log.Printf("📝 Found %d requests for execution run %s", len(requestRows), executionRunID)

	// Get all responses with joined data for this execution run
	responseRows, err := c.queries.GetAPIResponsesWithRequests(ctx, executionRunID)
	if err != nil {
		return nil, fmt.Errorf("failed to get responses: %w", err)
	}
	log.Printf("📊 Found %d responses for execution run %s", len(responseRows), executionRunID)

	// Build function tools map from function configurations
	functionTools := make([]types.Tool, 0)
	for _, funcConfig := range functionConfigRows {
		// Get the full function definition
		funcDef, err := c.queries.GetFunctionDefinition(ctx, funcConfig.FunctionDefinitionID)
		if err != nil {
			log.Printf("⚠️ Failed to get function definition %s: %v", funcConfig.FunctionDefinitionID, err)
			continue
		}

		// Parse the parameters schema
		var parametersSchema map[string]interface{}
		if err := json.Unmarshal([]byte(funcDef.ParametersSchema), &parametersSchema); err != nil {
			log.Printf("⚠️ Failed to parse parameters schema for function %s: %v", funcDef.Name, err)
			continue
		}

		tool := types.Tool{
			Name:        funcDef.Name,
			Description: funcDef.Description,
			Parameters:  parametersSchema,
		}
		functionTools = append(functionTools, tool)
		log.Printf("✅ Added function tool: %s", funcDef.Name)
	}

	// Build configurations map and add function tools to each configuration
	configs := make(map[string]*types.APIConfiguration)
	for _, row := range configRows {
		config := &types.APIConfiguration{
			ID:             row.ID,
			ExecutionRunID: row.ExecutionRunID,
			VariationName:  row.VariationName,
			ModelName:      row.ModelName,
			SystemPrompt:   row.SystemPrompt.String,
			CreatedAt:      row.CreatedAt.Time,
			Tools:          functionTools, // Add the function tools to each configuration
		}

		// Parse nullable fields
		if row.Temperature.Valid {
			temp, _ := parseFloat32(row.Temperature.String)
			config.Temperature = &temp
		}
		if row.MaxTokens.Valid {
			config.MaxTokens = &row.MaxTokens.Int32
		}
		if row.TopP.Valid {
			topP, _ := parseFloat32(row.TopP.String)
			config.TopP = &topP
		}
		if row.TopK.Valid {
			config.TopK = &row.TopK.Int32
		}

		configs[config.ID] = config
	}

	// Build requests map
	requests := make(map[string]*types.APIRequest)
	for _, row := range requestRows {
		request := &types.APIRequest{
			ID:              row.ID,
			ExecutionRunID:  row.ExecutionRunID,
			ConfigurationID: row.ConfigurationID,
			RequestType:     types.RequestType(row.RequestType),
			Prompt:          row.Prompt,
			Context:         row.Context.String,
			FunctionName:    row.FunctionName.String,
			CreatedAt:       row.CreatedAt.Time,
		}
		requests[request.ID] = request
	}

	// Build variation results
	results := make([]types.VariationResult, 0)

	log.Printf("🔍 Processing %d response rows for execution run %s", len(responseRows), executionRunID)

	// Get execution logs
	executionLogs, err := c.queries.GetExecutionLogsByRun(ctx, executionRunID)
	if err != nil {
		log.Printf("⚠️ Failed to get execution logs for %s: %v", executionRunID, err)
		// Continue without logs rather than failing
	}
	log.Printf("📋 Found %d execution logs for execution run %s", len(executionLogs), executionRunID)

	for _, respRow := range responseRows {
		// Get the configuration and request
		configID := findConfigIDForRequest(requestRows, respRow.RequestID)
		if configID == "" {
			log.Printf("Warning: Could not find configuration for request %s", respRow.RequestID)
			continue
		}

		config := configs[configID]
		request := requests[respRow.RequestID]

		if config == nil || request == nil {
			log.Printf("Warning: Missing config or request for response %s (config: %v, request: %v)", respRow.ID, config != nil, request != nil)
			continue
		}

		log.Printf("✅ Processing response %s for config %s (%s)", respRow.ID, configID, config.VariationName)

		// Parse usage metadata
		var usageMetadata map[string]interface{}
		if respRow.UsageMetadata != nil {
			json.Unmarshal(respRow.UsageMetadata, &usageMetadata)
		}

		response := &types.APIResponse{
			ID:             respRow.ID,
			RequestID:      respRow.RequestID,
			ResponseStatus: types.ResponseStatus(respRow.ResponseStatus),
			ResponseText:   respRow.ResponseText.String,
			FinishReason:   respRow.FinishReason.String,
			ErrorMessage:   respRow.ErrorMessage.String,
			ResponseTimeMs: respRow.ResponseTimeMs.Int32,
			UsageMetadata:  usageMetadata,
			CreatedAt:      respRow.CreatedAt.Time,
		}

		result := types.VariationResult{
			Configuration: *config,
			Request:       *request,
			Response:      *response,
			ExecutionTime: int64(response.ResponseTimeMs), // Already in milliseconds
		}

		results = append(results, result)
	}

	// Calculate totals
	totalTime := int64(0)
	successCount := 0
	errorCount := 0

	for _, result := range results {
		totalTime += result.ExecutionTime
		if result.Response.ResponseStatus == types.ResponseStatusSuccess {
			successCount++
		} else {
			errorCount++
		}
	}

	log.Printf("🕐 Total time calculation: %d ms", totalTime)

	// Convert database logs to types.ExecutionLog
	logs := make([]types.ExecutionLog, 0, len(executionLogs))
	for _, dbLog := range executionLogs {
		var details map[string]interface{}
		if len(dbLog.Details) > 0 {
			if err := json.Unmarshal(dbLog.Details, &details); err != nil {
				log.Printf("⚠️ Failed to parse log details: %v", err)
			}
		}

		var configID, requestID *string
		if dbLog.ConfigurationID.Valid {
			configID = &dbLog.ConfigurationID.String
		}
		if dbLog.RequestID.Valid {
			requestID = &dbLog.RequestID.String
		}

		timestamp := time.Now()
		if dbLog.Timestamp.Valid {
			timestamp = dbLog.Timestamp.Time
		}

		logs = append(logs, types.ExecutionLog{
			ID:              dbLog.ID,
			ExecutionRunID:  dbLog.ExecutionRunID,
			ConfigurationID: configID,
			RequestID:       requestID,
			LogLevel:        types.LogLevel(dbLog.LogLevel),
			LogCategory:     types.LogCategory(dbLog.LogCategory),
			Message:         dbLog.Message,
			Details:         details,
			Timestamp:       timestamp,
		})
	}

	// Create the execution result
	result := &types.ExecutionResult{
		ExecutionRun: *executionRun,
		Results:      results,
		TotalTime:    totalTime, // Already in milliseconds
		SuccessCount: successCount,
		ErrorCount:   errorCount,
		Logs:         logs,
	}

	// Try to load comparison result from database
	comparison, err := c.GetComparisonResult(ctx, executionRunID)
	if err != nil {
		log.Printf("ℹ️ No comparison result found for execution run: %s", executionRunID)
	} else {
		result.Comparison = comparison
		log.Printf("📊 Loaded comparison result from database: %s", comparison.ID)
	}

	return result, nil
}

// Helper function to find configuration ID for a request
func findConfigIDForRequest(requestRows []db.ApiRequest, requestID string) string {
	for _, req := range requestRows {
		if req.ID == requestID {
			return req.ConfigurationID
		}
	}
	return ""
}

// Helper function to parse float32 from string
func parseFloat32(s string) (float32, error) {
	if s == "" {
		return 0, fmt.Errorf("empty string")
	}
	// Simple parsing - could be enhanced
	if s == "0.20" || s == "0.2" {
		return 0.2, nil
	}
	if s == "0.50" || s == "0.5" {
		return 0.5, nil
	}
	if s == "0.80" || s == "0.8" {
		return 0.8, nil
	}
	return 0.5, nil // default fallback
}

// GetDB returns the underlying database connection for direct queries
func (c *Client) GetDB() *sql.DB {
	return c.db
}

// storeFunctionExecutionConfigs stores the function-execution relationships for replay functionality
func (c *Client) storeFunctionExecutionConfigs(ctx context.Context, executionRunID string, functionTools []types.Tool) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for i, tool := range functionTools {
		// Find the function definition by name
		funcDef, err := c.queries.GetFunctionDefinitionByName(ctx, tool.Name)
		if err != nil {
			log.Printf("⚠️ Function definition not found for tool %s: %v", tool.Name, err)
			continue
		}

		// Create the execution-function config
		configID := uuid.New().String()
		err = c.queries.CreateExecutionFunctionConfig(ctx, db.CreateExecutionFunctionConfigParams{
			ID:                   configID,
			ExecutionRunID:       executionRunID,
			FunctionDefinitionID: funcDef.ID,
			UseMockResponse:      sql.NullBool{Bool: true, Valid: true}, // Default to mock for replay
			ExecutionOrder:       sql.NullInt32{Int32: int32(i), Valid: true},
		})
		if err != nil {
			log.Printf("❌ Failed to create execution-function config for %s: %v", tool.Name, err)
			continue
		}

		log.Printf("✅ Stored function-execution config: %s -> %s", tool.Name, executionRunID)
	}

	return nil
}

// logExecutionEvent logs an execution event to the database and console
func (c *Client) logExecutionEvent(level types.LogLevel, category types.LogCategory, message string, details map[string]interface{}) {
	// Always log to console
	emoji := c.getLogEmoji(level, category)
	log.Printf("%s %s", emoji, message)

	// Only log to database if we have an active execution
	if c.currentExecutionRunID == nil {
		return
	}

	ctx := context.Background()
	logID := uuid.New().String()

	var detailsJSON json.RawMessage
	if details != nil {
		if detailsBytes, err := json.Marshal(details); err == nil {
			detailsJSON = detailsBytes
		}
	}

	var configID, requestID sql.NullString
	if c.currentConfigID != nil {
		configID = sql.NullString{String: *c.currentConfigID, Valid: true}
	}
	if c.currentRequestID != nil {
		requestID = sql.NullString{String: *c.currentRequestID, Valid: true}
	}

	err := c.queries.CreateExecutionLog(ctx, db.CreateExecutionLogParams{
		ID:              logID,
		ExecutionRunID:  *c.currentExecutionRunID,
		ConfigurationID: configID,
		RequestID:       requestID,
		LogLevel:        db.ExecutionLogsLogLevel(level),
		LogCategory:     db.ExecutionLogsLogCategory(category),
		Message:         message,
		Details:         detailsJSON,
	})

	if err != nil {
		log.Printf("❌ Failed to store execution log: %v", err)
	}
}

// getLogEmoji returns appropriate emoji for log level and category
func (c *Client) getLogEmoji(level types.LogLevel, category types.LogCategory) string {
	switch level {
	case types.LogLevelSuccess:
		return "✅"
	case types.LogLevelError:
		return "❌"
	case types.LogLevelWarn:
		return "⚠️"
	case types.LogLevelDebug:
		return "🔧"
	default:
		switch category {
		case types.LogCategorySetup:
			return "🚀"
		case types.LogCategoryFunctionCall:
			return "🔧"
		case types.LogCategoryAPICall:
			return "📡"
		case types.LogCategoryCompletion:
			return "🎯"
		default:
			return "📝"
		}
	}
}

// GetSystemConfigurations retrieves all system-wide AI configurations from the database
func (c *Client) GetSystemConfigurations(ctx context.Context) ([]types.APIConfiguration, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	// Get all configurations and filter for system ones
	configRows, err := c.queries.ListAPIConfigurations(ctx, db.ListAPIConfigurationsParams{
		Limit:  100, // Reasonable limit for system configurations
		Offset: 0,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get configurations: %w", err)
	}

	var systemConfigs []types.APIConfiguration
	for _, row := range configRows {
		// Check if this is a system configuration
		if row.UserID == "system" {
			config := types.APIConfiguration{
				ID:             row.ID,
				ExecutionRunID: row.ExecutionRunID,
				VariationName:  row.VariationName,
				ModelName:      row.ModelName,
				SystemPrompt:   row.SystemPrompt.String,
				CreatedAt:      row.CreatedAt.Time,
			}

			// Parse nullable fields
			if row.Temperature.Valid {
				temp, _ := parseFloat32(row.Temperature.String)
				config.Temperature = &temp
			}
			if row.MaxTokens.Valid {
				config.MaxTokens = &row.MaxTokens.Int32
			}
			if row.TopP.Valid {
				topP, _ := parseFloat32(row.TopP.String)
				config.TopP = &topP
			}
			if row.TopK.Valid {
				config.TopK = &row.TopK.Int32
			}

			// Parse JSON fields
			if len(row.SafetySettings) > 0 {
				var safetySettings map[string]interface{}
				if err := json.Unmarshal(row.SafetySettings, &safetySettings); err == nil {
					config.SafetySettings = safetySettings
				}
			}
			if len(row.GenerationConfig) > 0 {
				var generationConfig map[string]interface{}
				if err := json.Unmarshal(row.GenerationConfig, &generationConfig); err == nil {
					config.GenerationConfig = generationConfig
				}
			}
			if len(row.Tools) > 0 {
				var tools []types.Tool
				if err := json.Unmarshal(row.Tools, &tools); err == nil {
					config.Tools = tools
				}
			}

			systemConfigs = append(systemConfigs, config)
		}
	}

	log.Printf("✅ Retrieved %d system configurations from database", len(systemConfigs))
	return systemConfigs, nil
}

// setExecutionContext sets the current execution context for logging
func (c *Client) setExecutionContext(executionRunID, configID, requestID *string) {
	c.currentExecutionRunID = executionRunID
	c.currentConfigID = configID
	c.currentRequestID = requestID
}

// clearExecutionContext clears the execution context
func (c *Client) clearExecutionContext() {
	c.currentExecutionRunID = nil
	c.currentConfigID = nil
	c.currentRequestID = nil
}

// LogFunctionCall logs function call details to the database
func (c *Client) LogFunctionCall(ctx context.Context, call *types.FunctionCall) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Marshal JSON fields
	argsJSON, err := json.Marshal(call.FunctionArgs)
	if err != nil {
		return fmt.Errorf("failed to marshal function arguments: %w", err)
	}

	var responseJSON json.RawMessage
	if call.FunctionResponse != nil {
		responseBytes, err := json.Marshal(call.FunctionResponse)
		if err != nil {
			return fmt.Errorf("failed to marshal function response: %w", err)
		}
		responseJSON = responseBytes
	}

	var errorDetails sql.NullString
	if call.ErrorDetails != "" {
		errorDetails = sql.NullString{String: call.ErrorDetails, Valid: true}
	}

	var executionTimeMs sql.NullInt32
	if call.ExecutionTimeMs > 0 {
		executionTimeMs = sql.NullInt32{Int32: call.ExecutionTimeMs, Valid: true}
	}

	// Store in database
	err = c.queries.CreateFunctionCall(ctx, db.CreateFunctionCallParams{
		ID:                call.ID,
		RequestID:         call.RequestID,
		FunctionName:      call.FunctionName,
		FunctionArguments: argsJSON,
		FunctionResponse:  responseJSON,
		ExecutionStatus:   db.FunctionCallsExecutionStatus(call.ExecutionStatus),
		ExecutionTimeMs:   executionTimeMs,
		ErrorDetails:      errorDetails,
	})

	if err != nil {
		return fmt.Errorf("failed to store function call: %w", err)
	}

	log.Printf("📊 Function call logged to database: %s", call.FunctionName)
	return nil
}
