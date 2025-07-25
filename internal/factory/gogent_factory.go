package factory

import (
	"fmt"

	"gogent/examples/procurement"
	"gogent/internal/adapters"
	"gogent/internal/gogent"
	"gogent/internal/interfaces"
	"gogent/internal/types"
)

// DefaultGoGentFactory implements the GoGentFactory interface
type DefaultGoGentFactory struct{}

// NewGoGentFactory creates a new factory instance
func NewGoGentFactory() interfaces.GoGentFactory {
	return &DefaultGoGentFactory{}
}

// CreateClient creates a standard GoGent client
func (f *DefaultGoGentFactory) CreateClient(config *types.GeminiClientConfig, dbURL string) (interfaces.GoGentClient, error) {
	// Create the underlying gogent client
	client, err := gogent.NewClient(dbURL, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create gogent client: %w", err)
	}

	// Wrap it with our adapter to implement the interfaces
	adapter := adapters.NewGoGentClientAdapter(client)

	return adapter, nil
}

// CreateProcurementManager creates a procurement-specific implementation
func (f *DefaultGoGentFactory) CreateProcurementManager(config *types.GeminiClientConfig, dbURL string) (interfaces.ProcurementManager, error) {
	// Create the base client
	baseClient, err := f.CreateClient(config, dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create base client for procurement manager: %w", err)
	}

	// Create the procurement manager with the required interfaces
	procurementManager := procurement.NewProcurementAIManager(
		baseClient, // MultiVariationExecutor
		baseClient, // ExecutionLogger
		baseClient, // ResultComparator
	)

	return procurementManager, nil
}

// CreateCustomExecutor creates a custom use-case executor
func (f *DefaultGoGentFactory) CreateCustomExecutor(useCaseName string, config *types.GeminiClientConfig, dbURL string) (interfaces.UseCaseSpecificExecutor, error) {
	switch useCaseName {
	case "procurement", "ai-procurement-manager":
		return f.CreateProcurementManager(config, dbURL)

	case "legal-analysis":
		return f.createLegalAnalysisExecutor(config, dbURL)

	case "content-generation":
		return f.createContentGenerationExecutor(config, dbURL)

	case "risk-assessment":
		return f.createRiskAssessmentExecutor(config, dbURL)

	default:
		return nil, fmt.Errorf("unknown use case: %s", useCaseName)
	}
}

// CreateAnalyticsProvider creates an analytics provider
func (f *DefaultGoGentFactory) CreateAnalyticsProvider(dbURL string) (interfaces.AnalyticsProvider, error) {
	// TODO: Implement analytics provider
	return nil, fmt.Errorf("analytics provider not yet implemented")
}

// Helper methods for creating specific use case executors

func (f *DefaultGoGentFactory) createLegalAnalysisExecutor(config *types.GeminiClientConfig, dbURL string) (interfaces.UseCaseSpecificExecutor, error) {
	// TODO: Implement legal analysis use case
	return nil, fmt.Errorf("legal analysis executor not yet implemented")
}

func (f *DefaultGoGentFactory) createContentGenerationExecutor(config *types.GeminiClientConfig, dbURL string) (interfaces.UseCaseSpecificExecutor, error) {
	// TODO: Implement content generation use case
	return nil, fmt.Errorf("content generation executor not yet implemented")
}

func (f *DefaultGoGentFactory) createRiskAssessmentExecutor(config *types.GeminiClientConfig, dbURL string) (interfaces.UseCaseSpecificExecutor, error) {
	// TODO: Implement risk assessment use case
	return nil, fmt.Errorf("risk assessment executor not yet implemented")
}

// Convenience functions for common usage patterns

// QuickCreateProcurementManager creates a procurement manager with default configuration
func QuickCreateProcurementManager(apiKey, dbURL string) (interfaces.ProcurementManager, error) {
	factory := NewGoGentFactory()

	config := &types.GeminiClientConfig{
		APIKey:      apiKey,
		MaxRetries:  3,
		TimeoutSecs: 30,
	}

	return factory.CreateProcurementManager(config, dbURL)
}

// QuickCreateClient creates a standard client with default configuration
func QuickCreateClient(apiKey, dbURL string) (interfaces.GoGentClient, error) {
	factory := NewGoGentFactory()

	config := &types.GeminiClientConfig{
		APIKey:      apiKey,
		MaxRetries:  3,
		TimeoutSecs: 30,
	}

	return factory.CreateClient(config, dbURL)
}

// CreateMockFactory creates a factory that returns mock implementations for testing
func CreateMockFactory() interfaces.GoGentFactory {
	return &MockGoGentFactory{}
}

// MockGoGentFactory for testing purposes
type MockGoGentFactory struct{}

func (f *MockGoGentFactory) CreateClient(config *types.GeminiClientConfig, dbURL string) (interfaces.GoGentClient, error) {
	// Return a mock implementation for testing
	return nil, fmt.Errorf("mock client not yet implemented")
}

func (f *MockGoGentFactory) CreateProcurementManager(config *types.GeminiClientConfig, dbURL string) (interfaces.ProcurementManager, error) {
	// Return a mock procurement manager for testing
	return nil, fmt.Errorf("mock procurement manager not yet implemented")
}

func (f *MockGoGentFactory) CreateCustomExecutor(useCaseName string, config *types.GeminiClientConfig, dbURL string) (interfaces.UseCaseSpecificExecutor, error) {
	// Return a mock custom executor for testing
	return nil, fmt.Errorf("mock custom executor not yet implemented")
}

func (f *MockGoGentFactory) CreateAnalyticsProvider(dbURL string) (interfaces.AnalyticsProvider, error) {
	// Return a mock analytics provider for testing
	return nil, fmt.Errorf("mock analytics provider not yet implemented")
}
