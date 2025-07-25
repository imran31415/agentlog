package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"gogent/internal/auth"
	"gogent/internal/gogent"
	"gogent/internal/types"

	"github.com/joho/godotenv"
)

// BusinessLogic contains the core business logic separated from gRPC concerns
type BusinessLogic struct {
	client         *gogent.Client
	config         *types.GeminiClientConfig
	executions     map[string]*ExecutionStatus
	executionMutex sync.RWMutex
}

// NewBusinessLogic creates a new business logic instance
func NewBusinessLogic() (*BusinessLogic, error) {
	// Load environment variables
	if err := godotenv.Load("config.env"); err != nil {
		log.Printf("Warning: could not load config.env file: %v", err)
	}

	// Get configuration from environment
	apiKey := os.Getenv("GEMINI_API_KEY")
	dbURL := os.Getenv("DB_URL")

	if dbURL == "" {
		return nil, fmt.Errorf("DB_URL environment variable is required")
	}

	// Create Gemini client configuration
	config := &types.GeminiClientConfig{
		APIKey:      apiKey,
		MaxRetries:  3,
		TimeoutSecs: 30,
	}

	// Create gogent client
	client, err := gogent.NewClient(dbURL, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create gogent client: %w", err)
	}

	return &BusinessLogic{
		client:     client,
		config:     config,
		executions: make(map[string]*ExecutionStatus),
	}, nil
}

// Close closes the business logic resources
func (bl *BusinessLogic) Close() error {
	if bl.client != nil {
		return bl.client.Close()
	}
	return nil
}

// GetDB returns the database connection for direct queries
func (bl *BusinessLogic) GetDB() *sql.DB {
	return bl.client.GetDB()
}

// =============================================================================
// AUTHENTICATION & USER MANAGEMENT
// =============================================================================

func (bl *BusinessLogic) LoginUser(username, password string) (*auth.User, string, time.Time, error) {
	// TODO: Implement actual authentication logic
	log.Printf("🔐 Login attempt for user: %s", username)

	now := time.Now()
	loginTime := now // Create separate variable for address
	user := &auth.User{
		ID:            "user-1",
		Username:      username,
		Email:         &[]string{username + "@example.com"}[0],
		EmailVerified: true,
		IsTemporary:   false,
		CreatedAt:     now,
		UpdatedAt:     now,
		LastLoginAt:   &loginTime,
	}

	token := "mock-jwt-token"
	expiresAt := time.Now().Add(24 * time.Hour)

	return user, token, expiresAt, nil
}

func (bl *BusinessLogic) RegisterUser(username, email, password string) (*auth.User, string, error) {
	// TODO: Implement actual registration logic
	log.Printf("📝 Registration attempt for user: %s", username)

	now := time.Now()
	user := &auth.User{
		ID:            fmt.Sprintf("user-%d", now.Unix()),
		Username:      username,
		Email:         &email,
		EmailVerified: false,
		IsTemporary:   false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	token := "mock-jwt-token"
	return user, token, nil
}

func (bl *BusinessLogic) CreateTemporaryUser(sessionID string) (*auth.User, string, string, error) {
	log.Printf("👤 Creating temporary user with session ID: %s", sessionID)

	now := time.Now()
	tempID := fmt.Sprintf("temp-%d", now.Unix())
	user := &auth.User{
		ID:          tempID,
		Username:    "temp-user-" + tempID,
		Email:       nil,
		IsTemporary: true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	tempPassword := "temp-password-123"
	token := "temp-jwt-token"

	return user, tempPassword, token, nil
}

func (bl *BusinessLogic) SaveTemporaryAccount(email string) (*auth.User, bool, error) {
	log.Printf("💾 Saving temporary account with email: %s", email)

	now := time.Now()
	user := &auth.User{
		ID:            "saved-user-1",
		Username:      strings.Split(email, "@")[0],
		Email:         &email,
		EmailVerified: false,
		IsTemporary:   false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	emailSent := true
	return user, emailSent, nil
}

func (bl *BusinessLogic) VerifyEmail(token string) (*auth.User, bool, error) {
	log.Printf("✅ Verifying email with token: %s", token)

	email := "user@example.com"
	now := time.Now()
	user := &auth.User{
		ID:            "verified-user-1",
		Username:      "verified-user",
		Email:         &email,
		EmailVerified: true,
		IsTemporary:   false,
		UpdatedAt:     now,
	}

	verified := true
	return user, verified, nil
}

func (bl *BusinessLogic) GetCurrentUser() (*auth.User, error) {
	log.Printf("👤 Getting current user")

	// TODO: Extract user from JWT token in context
	email := "current@example.com"
	now := time.Now()
	lastLogin := now // Create a separate variable for the address
	user := &auth.User{
		ID:            "current-user-1",
		Username:      "current-user",
		Email:         &email,
		EmailVerified: true,
		IsTemporary:   false,
		CreatedAt:     now,
		UpdatedAt:     now,
		LastLoginAt:   &lastLogin,
	}

	return user, nil
}

// =============================================================================
// EXECUTION MANAGEMENT
// =============================================================================

func (bl *BusinessLogic) StartExecution(request *types.MultiExecutionRequest, useMock bool, additionalConfig *types.GeminiClientConfig) (string, *types.ExecutionRun, error) {
	log.Printf("🚀 Starting execution: %s", request.ExecutionRunName)

	// Generate execution run ID
	executionID := fmt.Sprintf("exec-%d", time.Now().UnixNano()/1000000)

	// Track execution status
	bl.executionMutex.Lock()
	bl.executions[executionID] = &ExecutionStatus{
		ID:        executionID,
		Status:    "pending",
		StartTime: time.Now(),
	}
	bl.executionMutex.Unlock()

	// Create execution run for response
	executionRun := &types.ExecutionRun{
		ID:                    executionID,
		Name:                  request.ExecutionRunName,
		Description:           request.Description,
		EnableFunctionCalling: request.EnableFunctionCalling,
		Status:                "pending",
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	// Start async execution
	go bl.runAsyncExecution(executionID, request, useMock, additionalConfig)

	return executionID, executionRun, nil
}

func (bl *BusinessLogic) GetExecutionStatus(ctx context.Context, executionID string) (string, time.Time, *time.Time, string, *types.ExecutionResult, error) {
	log.Printf("📊 Getting execution status for: %s", executionID)

	bl.executionMutex.RLock()
	execStatus, exists := bl.executions[executionID]
	bl.executionMutex.RUnlock()

	if !exists {
		// Check if this is a real execution ID from database
		realResult, err := bl.client.GetExecutionResult(ctx, executionID)
		if err != nil {
			return "", time.Time{}, nil, "", nil, fmt.Errorf("execution not found: %s", executionID)
		}

		now := time.Now()
		return "completed", now, &now, "", realResult, nil
	}

	var result *types.ExecutionResult
	if execStatus.Status == "completed" && execStatus.RealExecutionRunID != "" {
		realResult, err := bl.client.GetExecutionResult(ctx, execStatus.RealExecutionRunID)
		if err == nil {
			result = realResult
		}

		// Clean up completed execution from map
		bl.executionMutex.Lock()
		delete(bl.executions, executionID)
		bl.executionMutex.Unlock()
	}

	return execStatus.Status, execStatus.StartTime, execStatus.EndTime, execStatus.ErrorMessage, result, nil
}

func (bl *BusinessLogic) GetExecutionResult(ctx context.Context, executionRunID string) (*types.ExecutionResult, error) {
	log.Printf("📊 Getting execution result for: %s", executionRunID)

	return bl.client.GetExecutionResult(ctx, executionRunID)
}

func (bl *BusinessLogic) ListExecutionRuns(ctx context.Context, limit, offset int32) ([]*types.ExecutionRun, error) {
	log.Printf("📋 Listing execution runs (limit: %d, offset: %d)", limit, offset)

	if limit == 0 {
		limit = 10
	}

	return bl.client.ListExecutionRuns(ctx, limit, offset)
}

func (bl *BusinessLogic) DeleteExecutionRun(ctx context.Context, executionRunID string) error {
	log.Printf("🗑️ Deleting execution run: %s", executionRunID)

	// TODO: Implement actual deletion logic
	return nil
}

// =============================================================================
// CONFIGURATION MANAGEMENT
// =============================================================================

func (bl *BusinessLogic) GetDefaultConfigurations() []types.APIConfiguration {
	temp1 := float32(0.2)
	temp2 := float32(0.5)
	temp3 := float32(0.8)
	maxTokens := int32(500)
	topP1 := float32(0.8)
	topP2 := float32(0.9)
	topP3 := float32(0.95)
	topK1 := int32(10)
	topK2 := int32(20)
	topK3 := int32(40)

	return []types.APIConfiguration{
		{
			ID:            "config-conservative",
			VariationName: "Conservative",
			ModelName:     "gemini-1.5-flash",
			SystemPrompt:  "You are a helpful assistant. Provide balanced, informative responses.",
			Temperature:   &temp1,
			MaxTokens:     &maxTokens,
			TopP:          &topP1,
			TopK:          &topK1,
			CreatedAt:     time.Now(),
		},
		{
			ID:            "config-balanced",
			VariationName: "Balanced",
			ModelName:     "gemini-1.5-flash",
			SystemPrompt:  "You are a helpful assistant. Provide balanced, informative responses.",
			Temperature:   &temp2,
			MaxTokens:     &maxTokens,
			TopP:          &topP2,
			TopK:          &topK2,
			CreatedAt:     time.Now(),
		},
		{
			ID:            "config-creative",
			VariationName: "Creative",
			ModelName:     "gemini-1.5-flash",
			SystemPrompt:  "You are a creative assistant. Provide imaginative and engaging responses.",
			Temperature:   &temp3,
			MaxTokens:     &maxTokens,
			TopP:          &topP3,
			TopK:          &topK3,
			CreatedAt:     time.Now(),
		},
	}
}

func (bl *BusinessLogic) CreateConfiguration(config *types.APIConfiguration) (*types.APIConfiguration, error) {
	log.Printf("➕ Creating configuration: %s", config.VariationName)

	// TODO: Implement actual creation logic
	config.ID = fmt.Sprintf("config-%d", time.Now().Unix())
	config.CreatedAt = time.Now()

	return config, nil
}

func (bl *BusinessLogic) UpdateConfiguration(id string, config *types.APIConfiguration) (*types.APIConfiguration, error) {
	log.Printf("✏️ Updating configuration: %s", id)

	// TODO: Implement actual update logic
	config.ID = id

	return config, nil
}

func (bl *BusinessLogic) DeleteConfiguration(id string) error {
	log.Printf("🗑️ Deleting configuration: %s", id)

	// TODO: Implement actual deletion logic
	return nil
}

// =============================================================================
// FUNCTION MANAGEMENT
// =============================================================================

func (bl *BusinessLogic) ListFunctions(ctx context.Context) ([]*types.FunctionDefinition, error) {
	log.Printf("📋 Listing functions")

	// Query the database directly for function definitions
	query := `
		SELECT id, name, display_name, description, parameters_schema,
		       mock_response, endpoint_url, http_method, headers, auth_config,
		       is_active, created_at, updated_at
		FROM function_definitions
		WHERE is_active = true
		ORDER BY display_name ASC
	`

	rows, err := bl.client.GetDB().QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query functions: %v", err)
	}
	defer rows.Close()

	var functions []*types.FunctionDefinition

	for rows.Next() {
		var function types.FunctionDefinition
		var parametersSchemaJSON string
		var mockResponseJSON, headersJSON, authConfigJSON sql.NullString
		var endpointURL sql.NullString

		err := rows.Scan(
			&function.ID,
			&function.Name,
			&function.DisplayName,
			&function.Description,
			&parametersSchemaJSON,
			&mockResponseJSON,
			&endpointURL,
			&function.HttpMethod,
			&headersJSON,
			&authConfigJSON,
			&function.IsActive,
			&function.CreatedAt,
			&function.UpdatedAt,
		)
		if err != nil {
			log.Printf("❌ Failed to scan function row: %v", err)
			continue
		}

		// Set endpoint URL
		if endpointURL.Valid {
			function.EndpointURL = endpointURL.String
		}

		// TODO: Parse JSON fields
		functions = append(functions, &function)
	}

	return functions, nil
}

func (bl *BusinessLogic) GetFunction(id string) (*types.FunctionDefinition, error) {
	log.Printf("🔍 Getting function: %s", id)

	// TODO: Implement actual database lookup
	if id == "func-1" {
		function := &types.FunctionDefinition{
			ID:          "func-1",
			Name:        "get_weather",
			DisplayName: "Get Weather",
			Description: "Get current weather information for a location",
			EndpointURL: "https://api.weather.com/v1/current",
			HttpMethod:  "GET",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		return function, nil
	}

	return nil, fmt.Errorf("function not found: %s", id)
}

func (bl *BusinessLogic) CreateFunction(function *types.FunctionDefinition) (*types.FunctionDefinition, error) {
	log.Printf("➕ Creating function: %s", function.DisplayName)

	// TODO: Implement actual database insertion
	function.ID = fmt.Sprintf("func-%d", time.Now().Unix())
	function.CreatedAt = time.Now()
	function.UpdatedAt = time.Now()
	function.IsActive = true

	return function, nil
}

func (bl *BusinessLogic) UpdateFunction(id string, function *types.FunctionDefinition) (*types.FunctionDefinition, error) {
	log.Printf("✏️ Updating function: %s", id)

	// TODO: Implement actual database update
	function.ID = id
	function.UpdatedAt = time.Now()

	return function, nil
}

func (bl *BusinessLogic) DeleteFunction(id string) error {
	log.Printf("🗑️ Deleting function: %s", id)

	// TODO: Implement actual database deletion
	return nil
}

func (bl *BusinessLogic) TestFunction(functionID string, useMockData bool) (bool, bool, int32, map[string]interface{}, string, error) {
	log.Printf("🧪 Testing function: %s", functionID)

	startTime := time.Now()

	// For now, simulate function execution
	if useMockData {
		mockResponse := map[string]interface{}{
			"status": "mock_success",
			"data":   "Mock response generated",
		}

		return true, true, int32(time.Since(startTime).Milliseconds()), mockResponse, "", nil
	}

	// TODO: Implement real function execution
	realResponse := map[string]interface{}{
		"status":  "not_implemented",
		"message": "Real function testing not implemented yet",
	}

	return false, false, int32(time.Since(startTime).Milliseconds()), realResponse, "Real function testing not implemented", nil
}

// =============================================================================
// DATABASE MANAGEMENT
// =============================================================================

func (bl *BusinessLogic) GetDatabaseStats() (int32, int32, int32, int32, float64, float64) {
	log.Printf("📊 Getting database stats")

	// TODO: Implement actual database statistics
	return 25, 156, 156, 8, 450.5, 0.94
}

func (bl *BusinessLogic) ListDatabaseTables() []string {
	log.Printf("📋 Listing database tables")

	return []string{
		"execution_runs",
		"comparison_results",
		"function_calls",
		"api_configurations",
		"api_requests",
		"api_responses",
		"execution_logs",
		"function_definitions",
		"execution_function_configs",
	}
}

func (bl *BusinessLogic) GetTableData(tableName string) ([]string, [][]interface{}, int32, error) {
	log.Printf("📊 Getting table data for: %s", tableName)

	// TODO: Implement actual table data retrieval
	// For now, return mock data
	columns := []string{"id", "name", "created_at"}
	rows := [][]interface{}{
		{"1", "Sample data", time.Now().Format(time.RFC3339)},
	}

	return columns, rows, 1, nil
}

// =============================================================================
// HEALTH & SYSTEM
// =============================================================================

func (bl *BusinessLogic) GetHealthStatus() (string, string, bool, bool) {
	log.Printf("🏥 Health check")

	status := "ok"
	version := "1.0.0"
	database := bl.client != nil
	geminiAPI := bl.config.APIKey != ""

	return status, version, database, geminiAPI
}

// =============================================================================
// HELPER METHODS
// =============================================================================

// runAsyncExecution runs the execution in a goroutine
func (bl *BusinessLogic) runAsyncExecution(executionID string, request *types.MultiExecutionRequest, useMock bool, additionalConfig *types.GeminiClientConfig) {
	// Update status to running
	bl.executionMutex.Lock()
	if status, exists := bl.executions[executionID]; exists {
		status.Status = "running"
	}
	bl.executionMutex.Unlock()

	log.Printf("🚀 Starting async execution: %s", executionID)

	// Create temporary client configuration with additional APIs
	tempConfig := &types.GeminiClientConfig{
		APIKey:      bl.config.APIKey,
		MaxRetries:  bl.config.MaxRetries,
		TimeoutSecs: bl.config.TimeoutSecs,
	}

	// Merge additional configuration if provided
	if additionalConfig != nil {
		if additionalConfig.OpenWeatherAPIKey != "" {
			tempConfig.OpenWeatherAPIKey = additionalConfig.OpenWeatherAPIKey
		}
		if additionalConfig.Neo4jURL != "" {
			tempConfig.Neo4jURL = additionalConfig.Neo4jURL
			tempConfig.Neo4jUsername = additionalConfig.Neo4jUsername
			tempConfig.Neo4jPassword = additionalConfig.Neo4jPassword
			tempConfig.Neo4jDatabase = additionalConfig.Neo4jDatabase
		}
	}

	if useMock || tempConfig.APIKey == "" {
		tempConfig.APIKey = "" // Force mock mode
		log.Printf("Using mock mode for execution")
	}

	// Create temporary client
	dbURL := os.Getenv("DB_URL")
	tempClient, err := gogent.NewClient(dbURL, tempConfig)
	if err != nil {
		bl.markExecutionFailed(executionID, fmt.Sprintf("Failed to create client: %v", err))
		return
	}
	defer tempClient.Close()

	// Execute
	ctx := context.Background()
	result, err := tempClient.ExecuteMultiVariation(ctx, request)
	if err != nil {
		bl.markExecutionFailed(executionID, fmt.Sprintf("Execution failed: %v", err))
		return
	}

	// Mark execution as completed
	bl.executionMutex.Lock()
	if status, exists := bl.executions[executionID]; exists {
		status.Status = "completed"
		status.RealExecutionRunID = result.ExecutionRun.ID
		endTime := time.Now()
		status.EndTime = &endTime
	}
	bl.executionMutex.Unlock()

	log.Printf("✅ Async execution completed: %s", executionID)
}

// markExecutionFailed marks an execution as failed
func (bl *BusinessLogic) markExecutionFailed(executionID, errorMessage string) {
	bl.executionMutex.Lock()
	if execStatus, exists := bl.executions[executionID]; exists {
		execStatus.Status = "failed"
		execStatus.ErrorMessage = errorMessage
		endTime := time.Now()
		execStatus.EndTime = &endTime
	}
	bl.executionMutex.Unlock()
	log.Printf("❌ Async execution failed: %s - %s", executionID, errorMessage)
}
