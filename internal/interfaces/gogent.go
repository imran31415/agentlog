package interfaces

import (
	"context"

	"gogent/internal/types"
)

// MultiVariationExecutor defines the core interface for executing AI variations
type MultiVariationExecutor interface {
	// ExecuteMultiVariation runs the same prompt with multiple configurations
	ExecuteMultiVariation(ctx context.Context, request *types.MultiExecutionRequest) (*types.ExecutionResult, error)

	// ExecuteSingleVariation runs a single variation (useful for custom implementations)
	ExecuteSingleVariation(ctx context.Context, config *types.APIConfiguration, prompt, context string) (*types.VariationResult, error)

	// Close releases any resources held by the executor
	Close() error
}

// ExecutionLogger defines the interface for logging AI interactions
type ExecutionLogger interface {
	// CreateExecutionRun creates a new execution run for grouping related API calls
	CreateExecutionRun(ctx context.Context, name, description string, enableFunctionCalling bool) (*types.ExecutionRun, error)

	// LogAPIRequest logs an API request to storage
	LogAPIRequest(ctx context.Context, request *types.APIRequest) error

	// LogAPIResponse logs an API response to storage
	LogAPIResponse(ctx context.Context, response *types.APIResponse) error

	// LogFunctionCall logs function call details
	LogFunctionCall(ctx context.Context, call *types.FunctionCall) error

	// GetExecutionRun retrieves an execution run by ID
	GetExecutionRun(ctx context.Context, id string) (*types.ExecutionRun, error)

	// ListExecutionRuns lists execution runs with pagination
	ListExecutionRuns(ctx context.Context, limit, offset int) ([]*types.ExecutionRun, error)
}

// ConfigurationManager defines the interface for managing AI configurations
type ConfigurationManager interface {
	// CreateConfiguration creates and stores a new API configuration
	CreateConfiguration(ctx context.Context, config *types.APIConfiguration) error

	// GetConfiguration retrieves a configuration by ID
	GetConfiguration(ctx context.Context, id string) (*types.APIConfiguration, error)

	// ListConfigurations lists configurations for an execution run
	ListConfigurations(ctx context.Context, executionRunID string) ([]*types.APIConfiguration, error)

	// UpdateConfiguration updates an existing configuration
	UpdateConfiguration(ctx context.Context, config *types.APIConfiguration) error

	// DeleteConfiguration deletes a configuration
	DeleteConfiguration(ctx context.Context, id string) error
}

// ResultComparator defines the interface for comparing AI execution results
type ResultComparator interface {
	// CompareResults analyzes and compares multiple variation results
	CompareResults(ctx context.Context, result *types.ExecutionResult, metrics []string) (*types.ComparisonResult, error)

	// SaveComparison saves a comparison result
	SaveComparison(ctx context.Context, comparison *types.ComparisonResult) error

	// GetComparison retrieves a comparison result by ID
	GetComparison(ctx context.Context, id string) (*types.ComparisonResult, error)

	// ListComparisons lists comparisons for an execution run
	ListComparisons(ctx context.Context, executionRunID string) ([]*types.ComparisonResult, error)
}

// AIProvider defines the interface for different AI service providers
type AIProvider interface {
	// GenerateContent generates content using the AI service
	GenerateContent(ctx context.Context, config *types.APIConfiguration, prompt, context string) (*types.APIResponse, error)

	// CountTokens counts tokens for cost estimation
	CountTokens(ctx context.Context, modelName, text string) (int32, error)

	// GetModelInfo retrieves information about available models
	GetModelInfo(ctx context.Context, modelName string) (*types.ModelInfo, error)

	// ValidateConfiguration validates if a configuration is supported
	ValidateConfiguration(ctx context.Context, config *types.APIConfiguration) error

	// Close releases provider resources
	Close() error
}

// GoGentClient defines the complete interface for a GoGent implementation
type GoGentClient interface {
	MultiVariationExecutor
	ExecutionLogger
	ConfigurationManager
	ResultComparator
}

// AnalyticsProvider defines the interface for analytics and insights
type AnalyticsProvider interface {
	// GetExecutionAnalytics provides analytics for an execution run
	GetExecutionAnalytics(ctx context.Context, executionRunID string) (*types.ExecutionAnalytics, error)

	// GetPerformanceMetrics calculates performance metrics across runs
	GetPerformanceMetrics(ctx context.Context, timeRange *types.TimeRange) (*types.PerformanceMetrics, error)

	// GetCostAnalysis provides cost analysis and projections
	GetCostAnalysis(ctx context.Context, timeRange *types.TimeRange) (map[string]interface{}, error)

	// GetModelComparison compares different models' performance
	GetModelComparison(ctx context.Context, models []string, timeRange *types.TimeRange) (map[string]interface{}, error)
}

// UseCaseSpecificExecutor defines interface for domain-specific implementations
type UseCaseSpecificExecutor interface {
	MultiVariationExecutor

	// GetUseCaseName returns the name of the specific use case
	GetUseCaseName() string

	// GetDefaultConfigurations returns default configurations for this use case
	GetDefaultConfigurations() []types.APIConfiguration

	// ValidateUseCaseRequest validates a request for this specific use case
	ValidateUseCaseRequest(request *types.MultiExecutionRequest) error

	// PostProcessResults performs use-case specific post-processing
	PostProcessResults(result *types.ExecutionResult) (map[string]interface{}, error)
}

// ProcurementManager defines the interface for AI procurement management use case
type ProcurementManager interface {
	UseCaseSpecificExecutor

	// EvaluateVendorProposals compares AI responses for vendor evaluation
	EvaluateVendorProposals(ctx context.Context, rfp map[string]interface{}) (map[string]interface{}, error)

	// GenerateNegotiationStrategies creates negotiation strategies with different approaches
	GenerateNegotiationStrategies(ctx context.Context, vendorProfile map[string]interface{}) (map[string]interface{}, error)

	// AnalyzeContractTerms analyzes contract terms with different risk profiles
	AnalyzeContractTerms(ctx context.Context, contract map[string]interface{}) (map[string]interface{}, error)

	// OptimizeProcurementProcess finds optimal procurement approaches
	OptimizeProcurementProcess(ctx context.Context, requirements map[string]interface{}) (map[string]interface{}, error)
}

// Factory interface for creating different implementations
type GoGentFactory interface {
	// CreateClient creates a standard GoGent client
	CreateClient(config *types.GeminiClientConfig, dbURL string) (GoGentClient, error)

	// CreateProcurementManager creates a procurement-specific implementation
	CreateProcurementManager(config *types.GeminiClientConfig, dbURL string) (ProcurementManager, error)

	// CreateCustomExecutor creates a custom use-case executor
	CreateCustomExecutor(useCaseName string, config *types.GeminiClientConfig, dbURL string) (UseCaseSpecificExecutor, error)

	// CreateAnalyticsProvider creates an analytics provider
	CreateAnalyticsProvider(dbURL string) (AnalyticsProvider, error)
}

// Plugin interface for extending functionality
type GoGentPlugin interface {
	// GetName returns the plugin name
	GetName() string

	// Initialize initializes the plugin with configuration
	Initialize(config map[string]interface{}) error

	// PreProcess allows plugins to modify requests before execution
	PreProcess(ctx context.Context, request *types.MultiExecutionRequest) (*types.MultiExecutionRequest, error)

	// PostProcess allows plugins to modify results after execution
	PostProcess(ctx context.Context, result *types.ExecutionResult) (*types.ExecutionResult, error)

	// Cleanup releases plugin resources
	Cleanup() error
}
