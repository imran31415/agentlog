package types

import (
	"encoding/json"
	"time"
)

// RequestType represents the type of API request
type RequestType string

const (
	RequestTypeGenerate     RequestType = "generate"
	RequestTypeChat         RequestType = "chat"
	RequestTypeFunctionCall RequestType = "function_call"
)

// ResponseStatus represents the status of an API response
type ResponseStatus string

const (
	ResponseStatusSuccess ResponseStatus = "success"
	ResponseStatusError   ResponseStatus = "error"
	ResponseStatusTimeout ResponseStatus = "timeout"
)

// LogLevel represents the severity level of a log entry
type LogLevel string

const (
	LogLevelInfo    LogLevel = "INFO"
	LogLevelDebug   LogLevel = "DEBUG"
	LogLevelWarn    LogLevel = "WARN"
	LogLevelError   LogLevel = "ERROR"
	LogLevelSuccess LogLevel = "SUCCESS"
)

// LogCategory represents the category/context of a log entry
type LogCategory string

const (
	LogCategorySetup        LogCategory = "SETUP"
	LogCategoryExecution    LogCategory = "EXECUTION"
	LogCategoryFunctionCall LogCategory = "FUNCTION_CALL"
	LogCategoryAPICall      LogCategory = "API_CALL"
	LogCategoryCompletion   LogCategory = "COMPLETION"
	LogCategoryError        LogCategory = "ERROR"
)

// ExecutionLog represents a log entry for an execution
type ExecutionLog struct {
	ID              string                 `json:"id"`
	ExecutionRunID  string                 `json:"executionRunId"`
	ConfigurationID *string                `json:"configurationId,omitempty"`
	RequestID       *string                `json:"requestId,omitempty"`
	LogLevel        LogLevel               `json:"logLevel"`
	LogCategory     LogCategory            `json:"logCategory"`
	Message         string                 `json:"message"`
	Details         map[string]interface{} `json:"details,omitempty"`
	Timestamp       time.Time              `json:"timestamp"`
}

// ExecutionRun represents a group of related API calls with variations
type ExecutionRun struct {
	ID                    string    `json:"id"`
	Name                  string    `json:"name"`
	Description           string    `json:"description,omitempty"`
	EnableFunctionCalling bool      `json:"enableFunctionCalling"`
	Status                string    `json:"status"` // pending, running, completed, failed
	ErrorMessage          string    `json:"errorMessage,omitempty"`
	CreatedAt             time.Time `json:"createdAt"`
	UpdatedAt             time.Time `json:"updatedAt"`
}

// APIConfiguration represents a specific configuration for API calls
type APIConfiguration struct {
	ID               string                 `json:"id"`
	ExecutionRunID   string                 `json:"executionRunId"`
	VariationName    string                 `json:"variationName"`
	ModelName        string                 `json:"modelName"`
	SystemPrompt     string                 `json:"systemPrompt,omitempty"`
	Temperature      *float32               `json:"temperature,omitempty"`
	MaxTokens        *int32                 `json:"maxTokens,omitempty"`
	TopP             *float32               `json:"topP,omitempty"`
	TopK             *int32                 `json:"topK,omitempty"`
	SafetySettings   map[string]interface{} `json:"safetySettings,omitempty"`
	GenerationConfig map[string]interface{} `json:"generationConfig,omitempty"`
	Tools            []Tool                 `json:"tools,omitempty"`
	ToolConfig       map[string]interface{} `json:"toolConfig,omitempty"`
	CreatedAt        time.Time              `json:"createdAt"`
}

// FunctionDefinition represents a reusable function definition
type FunctionDefinition struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`                   // Unique function name for API calls
	DisplayName      string                 `json:"displayName"`            // Human-readable name
	Description      string                 `json:"description"`            // Function description
	ParametersSchema map[string]interface{} `json:"parametersSchema"`       // JSON schema for parameters
	MockResponse     map[string]interface{} `json:"mockResponse,omitempty"` // Mock response for testing
	EndpointURL      string                 `json:"endpointUrl,omitempty"`  // Real API endpoint
	HttpMethod       string                 `json:"httpMethod"`             // HTTP method (GET, POST, etc.)
	Headers          map[string]interface{} `json:"headers,omitempty"`      // HTTP headers
	AuthConfig       map[string]interface{} `json:"authConfig,omitempty"`   // Authentication config
	IsActive         bool                   `json:"isActive"`
	CreatedAt        time.Time              `json:"createdAt"`
	UpdatedAt        time.Time              `json:"updatedAt"`
}

// ExecutionFunctionConfig represents function configuration for a specific execution
type ExecutionFunctionConfig struct {
	ID                   string    `json:"id"`
	ExecutionRunID       string    `json:"executionRunId"`
	FunctionDefinitionID string    `json:"functionDefinitionId"`
	UseMockResponse      bool      `json:"useMockResponse"`
	ExecutionOrder       int       `json:"executionOrder"`
	CreatedAt            time.Time `json:"createdAt"`

	// Populated from JOIN queries
	FunctionName        string `json:"functionName,omitempty"`
	FunctionDisplayName string `json:"functionDisplayName,omitempty"`
	FunctionDescription string `json:"functionDescription,omitempty"`
}

// FunctionCallStats represents statistics for function calls
type FunctionCallStats struct {
	TotalCalls       int     `json:"totalCalls"`
	SuccessfulCalls  int     `json:"successfulCalls"`
	FailedCalls      int     `json:"failedCalls"`
	AvgExecutionTime float64 `json:"avgExecutionTime"`
	MaxExecutionTime int32   `json:"maxExecutionTime"`
	MinExecutionTime int32   `json:"minExecutionTime"`
}

// FunctionCallHistoryItem represents a function call with context
type FunctionCallHistoryItem struct {
	FunctionCall
	ExecutionRunID   string    `json:"executionRunId"`
	ExecutionName    string    `json:"executionName"`
	Prompt           string    `json:"prompt"`
	RequestCreatedAt time.Time `json:"requestCreatedAt"`
}

// FunctionExecutionMode represents how functions should be executed
type FunctionExecutionMode string

const (
	FunctionExecutionMock FunctionExecutionMode = "mock"
	FunctionExecutionReal FunctionExecutionMode = "real"
	FunctionExecutionAuto FunctionExecutionMode = "auto" // Use real if available, mock otherwise
)

// FunctionCallRequest represents a request to execute a function
type FunctionCallRequest struct {
	FunctionName  string                 `json:"functionName"`
	Arguments     map[string]interface{} `json:"arguments"`
	ExecutionMode FunctionExecutionMode  `json:"executionMode"`
	TimeoutMs     int32                  `json:"timeoutMs,omitempty"`
}

// FunctionCallResult represents the result of a function execution
type FunctionCallResult struct {
	FunctionName    string                 `json:"functionName"`
	Arguments       map[string]interface{} `json:"arguments"`
	Response        map[string]interface{} `json:"response"`
	ExecutionStatus string                 `json:"executionStatus"`
	ExecutionTimeMs int32                  `json:"executionTimeMs"`
	ErrorDetails    string                 `json:"errorDetails,omitempty"`
	UsedMockData    bool                   `json:"usedMockData"`
}

// Tool represents a function tool that can be called by the AI
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// APIRequest represents a request to the Gemini API
type APIRequest struct {
	ID                 string                 `json:"id"`
	ExecutionRunID     string                 `json:"executionRunId"`
	ConfigurationID    string                 `json:"configurationId"`
	RequestType        RequestType            `json:"requestType"`
	Prompt             string                 `json:"prompt"`
	Context            string                 `json:"context,omitempty"`
	FunctionName       string                 `json:"functionName,omitempty"`
	FunctionParameters map[string]interface{} `json:"functionParameters,omitempty"`
	RequestHeaders     map[string]interface{} `json:"requestHeaders,omitempty"`
	RequestBody        map[string]interface{} `json:"requestBody,omitempty"`
	CreatedAt          time.Time              `json:"createdAt"`
}

// APIResponse represents a response from the Gemini API
type APIResponse struct {
	ID                   string                 `json:"id"`
	RequestID            string                 `json:"requestId"`
	ResponseStatus       ResponseStatus         `json:"responseStatus"`
	ResponseText         string                 `json:"responseText,omitempty"`
	FunctionCallResponse map[string]interface{} `json:"functionCallResponse,omitempty"`
	UsageMetadata        map[string]interface{} `json:"usageMetadata,omitempty"`
	SafetyRatings        map[string]interface{} `json:"safetyRatings,omitempty"`
	FinishReason         string                 `json:"finishReason,omitempty"`
	ErrorMessage         string                 `json:"errorMessage,omitempty"`
	ResponseTimeMs       int32                  `json:"responseTimeMs"`
	ResponseHeaders      map[string]interface{} `json:"responseHeaders,omitempty"`
	ResponseBody         map[string]interface{} `json:"responseBody,omitempty"`
	CreatedAt            time.Time              `json:"createdAt"`
}

// FunctionCall represents a function call made during AI execution
type FunctionCall struct {
	ID               string                 `json:"id"`
	RequestID        string                 `json:"request_id"`
	FunctionName     string                 `json:"function_name"`
	FunctionArgs     map[string]interface{} `json:"function_arguments"`
	FunctionResponse map[string]interface{} `json:"function_response,omitempty"`
	ExecutionStatus  string                 `json:"execution_status"`
	ExecutionTimeMs  int32                  `json:"execution_time_ms,omitempty"`
	ErrorDetails     string                 `json:"error_details,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
}

// GeminiClientConfig represents the configuration for the Gemini client
type GeminiClientConfig struct {
	APIKey            string `json:"api_key"`
	OpenWeatherAPIKey string `json:"openweather_api_key,omitempty"`
	ProjectID         string `json:"project_id,omitempty"`
	Region            string `json:"region,omitempty"`
	MaxRetries        int    `json:"max_retries"`
	TimeoutSecs       int    `json:"timeout_secs"`
}

// MultiExecutionRequest represents a request to execute multiple variations
type MultiExecutionRequest struct {
	ExecutionRunName      string             `json:"executionRunName"`
	Description           string             `json:"description,omitempty"`
	BasePrompt            string             `json:"basePrompt"`
	Context               string             `json:"context,omitempty"`
	EnableFunctionCalling bool               `json:"enableFunctionCalling,omitempty"`
	Configurations        []APIConfiguration `json:"configurations"`
	FunctionTools         []Tool             `json:"functionTools,omitempty"`
	ComparisonConfig      *ComparisonConfig  `json:"comparisonConfig,omitempty"`
}

// ComparisonConfig represents configuration for comparing execution results
type ComparisonConfig struct {
	Enabled     bool     `json:"enabled"`
	Metrics     []string `json:"metrics"`
	CustomRules []string `json:"customRules,omitempty"`
}

// ExecutionResult represents the result of a multi-execution
type ExecutionResult struct {
	ExecutionRun ExecutionRun      `json:"executionRun"`
	Results      []VariationResult `json:"results"`
	Comparison   *ComparisonResult `json:"comparison,omitempty"`
	TotalTime    int64             `json:"totalTime"` // milliseconds
	SuccessCount int               `json:"successCount"`
	ErrorCount   int               `json:"errorCount"`
	Logs         []ExecutionLog    `json:"logs,omitempty"`
}

// VariationResult represents the result of a single variation execution
type VariationResult struct {
	Configuration APIConfiguration `json:"configuration"`
	Request       APIRequest       `json:"request"`
	Response      APIResponse      `json:"response"`
	FunctionCalls []FunctionCall   `json:"functionCalls,omitempty"`
	ExecutionTime int64            `json:"executionTime"` // milliseconds
}

// ComparisonResult represents the result of comparing multiple variations
type ComparisonResult struct {
	ID                  string                 `json:"id"`
	ExecutionRunID      string                 `json:"executionRunId"`
	ComparisonType      string                 `json:"comparisonType"`
	MetricName          string                 `json:"metricName"`
	ConfigurationScores map[string]interface{} `json:"configurationScores"`
	BestConfigurationID string                 `json:"bestConfigurationId,omitempty"`
	BestConfiguration   *APIConfiguration      `json:"bestConfiguration,omitempty"`
	AllConfigurations   []APIConfiguration     `json:"allConfigurations,omitempty"`
	AnalysisNotes       string                 `json:"analysisNotes,omitempty"`
	CreatedAt           time.Time              `json:"createdAt"`
}

// Additional types for interface support

// ModelInfo represents information about an AI model
type ModelInfo struct {
	Name             string    `json:"name"`
	DisplayName      string    `json:"display_name"`
	Description      string    `json:"description"`
	Version          string    `json:"version"`
	InputTokenLimit  int32     `json:"input_token_limit"`
	OutputTokenLimit int32     `json:"output_token_limit"`
	SupportedMethods []string  `json:"supported_methods"`
	SupportedFormats []string  `json:"supported_formats"`
	CreatedAt        time.Time `json:"created_at"`
}

// TimeRange represents a time range for analytics
type TimeRange struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

// ExecutionAnalytics represents analytics for an execution run
type ExecutionAnalytics struct {
	ExecutionRunID      string             `json:"execution_run_id"`
	TotalRequests       int                `json:"total_requests"`
	SuccessfulRequests  int                `json:"successful_requests"`
	FailedRequests      int                `json:"failed_requests"`
	AverageResponseTime float64            `json:"average_response_time_ms"`
	TotalTokensUsed     int32              `json:"total_tokens_used"`
	TotalCost           float64            `json:"total_cost"`
	ModelUsage          map[string]int     `json:"model_usage"`
	PerformanceMetrics  map[string]float64 `json:"performance_metrics"`
	CreatedAt           time.Time          `json:"created_at"`
}

// PerformanceMetrics represents performance metrics across runs
type PerformanceMetrics struct {
	TimeRange           TimeRange          `json:"time_range"`
	TotalExecutions     int                `json:"total_executions"`
	AverageResponseTime float64            `json:"average_response_time_ms"`
	P95ResponseTime     float64            `json:"p95_response_time_ms"`
	P99ResponseTime     float64            `json:"p99_response_time_ms"`
	SuccessRate         float64            `json:"success_rate"`
	ThroughputPerHour   float64            `json:"throughput_per_hour"`
	ModelPerformance    map[string]float64 `json:"model_performance"`
	CreatedAt           time.Time          `json:"created_at"`
}

// ToJSON converts any struct to JSON string for database storage
func ToJSON(v interface{}) (string, error) {
	bytes, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// FromJSON converts JSON string from database to struct
func FromJSON(jsonStr string, v interface{}) error {
	if jsonStr == "" {
		return nil
	}
	return json.Unmarshal([]byte(jsonStr), v)
}
