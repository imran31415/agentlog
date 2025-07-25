# GoGent Interface Architecture

## Overview

GoGent provides a comprehensive interface-based architecture that allows easy implementation of AI-powered applications for different use cases. The system is designed around the principle of **separation of concerns** and **dependency injection**, making it highly extensible and testable.

## Core Interfaces

### 1. MultiVariationExecutor
The heart of GoGent - enables executing the same AI prompt with multiple configurations to compare different approaches.

```go
type MultiVariationExecutor interface {
    ExecuteMultiVariation(ctx context.Context, request *types.MultiExecutionRequest) (*types.ExecutionResult, error)
    ExecuteSingleVariation(ctx context.Context, config *types.APIConfiguration, prompt, context string) (*types.VariationResult, error)
    Close() error
}
```

**Use Cases:**
- A/B testing different AI prompts
- Comparing temperature settings
- Evaluating different models
- Benchmarking prompt engineering approaches

### 2. ExecutionLogger
Handles comprehensive logging of all AI interactions for analysis and compliance.

```go
type ExecutionLogger interface {
    CreateExecutionRun(ctx context.Context, name, description string) (*types.ExecutionRun, error)
    LogAPIRequest(ctx context.Context, request *types.APIRequest) error
    LogAPIResponse(ctx context.Context, response *types.APIResponse) error
    LogFunctionCall(ctx context.Context, call *types.FunctionCall) error
    GetExecutionRun(ctx context.Context, id string) (*types.ExecutionRun, error)
    ListExecutionRuns(ctx context.Context, limit, offset int) ([]*types.ExecutionRun, error)
}
```

**Benefits:**
- Full audit trail of AI interactions
- Cost tracking and analysis
- Performance monitoring
- Compliance documentation

### 3. ConfigurationManager
Manages AI model configurations and their lifecycle.

```go
type ConfigurationManager interface {
    CreateConfiguration(ctx context.Context, config *types.APIConfiguration) error
    GetConfiguration(ctx context.Context, id string) (*types.APIConfiguration, error)
    ListConfigurations(ctx context.Context, executionRunID string) ([]*types.APIConfiguration, error)
    UpdateConfiguration(ctx context.Context, config *types.APIConfiguration) error
    DeleteConfiguration(ctx context.Context, id string) error
}
```

**Features:**
- Version control for AI configurations
- Template management
- Environment-specific settings
- Configuration validation

### 4. ResultComparator
Analyzes and compares results from different AI variations.

```go
type ResultComparator interface {
    CompareResults(ctx context.Context, result *types.ExecutionResult, metrics []string) (*types.ComparisonResult, error)
    SaveComparison(ctx context.Context, comparison *types.ComparisonResult) error
    GetComparison(ctx context.Context, id string) (*types.ComparisonResult, error)
    ListComparisons(ctx context.Context, executionRunID string) ([]*types.ComparisonResult, error)
}
```

**Capabilities:**
- Quality scoring
- Performance analysis
- Cost comparison
- Automatic best configuration selection

## Use Case Specific Interfaces

### UseCaseSpecificExecutor
Base interface for domain-specific implementations.

```go
type UseCaseSpecificExecutor interface {
    MultiVariationExecutor
    
    GetUseCaseName() string
    GetDefaultConfigurations() []types.APIConfiguration
    ValidateUseCaseRequest(request *types.MultiExecutionRequest) error
    PostProcessResults(result *types.ExecutionResult) (*types.UseCaseResult, error)
}
```

### ProcurementManager
Specialized interface for AI procurement management.

```go
type ProcurementManager interface {
    UseCaseSpecificExecutor
    
    EvaluateVendorProposals(ctx context.Context, rfp *types.RFPRequest) (*types.VendorEvaluationResult, error)
    GenerateNegotiationStrategies(ctx context.Context, vendorProfile *types.VendorProfile) (*types.NegotiationStrategies, error)
    AnalyzeContractTerms(ctx context.Context, contract *types.ContractTerms) (*types.ContractAnalysis, error)
    OptimizeProcurementProcess(ctx context.Context, requirements *types.ProcurementRequirements) (*types.ProcessOptimization, error)
}
```

## Implementation Guide

### Step 1: Create Your Use Case Implementation

```go
package myusecase

import (
    "context"
    "gogent/internal/interfaces"
    "gogent/internal/types"
)

type MyUseCaseExecutor struct {
    executor   interfaces.MultiVariationExecutor
    logger     interfaces.ExecutionLogger
    comparator interfaces.ResultComparator
}

func NewMyUseCaseExecutor(executor interfaces.MultiVariationExecutor, logger interfaces.ExecutionLogger, comparator interfaces.ResultComparator) *MyUseCaseExecutor {
    return &MyUseCaseExecutor{
        executor:   executor,
        logger:     logger,
        comparator: comparator,
    }
}

func (m *MyUseCaseExecutor) GetUseCaseName() string {
    return "my-custom-use-case"
}

func (m *MyUseCaseExecutor) GetDefaultConfigurations() []types.APIConfiguration {
    // Return use-case optimized configurations
    return []types.APIConfiguration{
        {
            VariationName: "conservative",
            ModelName:     "gemini-1.5-flash",
            SystemPrompt:  "You are a conservative analyst...",
            Temperature:   &[]float32{0.2}[0],
        },
        // ... more configurations
    }
}

func (m *MyUseCaseExecutor) ExecuteMultiVariation(ctx context.Context, request *types.MultiExecutionRequest) (*types.ExecutionResult, error) {
    if err := m.ValidateUseCaseRequest(request); err != nil {
        return nil, err
    }
    return m.executor.ExecuteMultiVariation(ctx, request)
}

// Implement other required methods...
```

### Step 2: Add to Factory

```go
func (f *DefaultGoGentFactory) CreateCustomExecutor(useCaseName string, config *types.GeminiClientConfig, dbURL string) (interfaces.UseCaseSpecificExecutor, error) {
    switch useCaseName {
    case "my-custom-use-case":
        return f.createMyUseCaseExecutor(config, dbURL)
    // ... other cases
    }
}

func (f *DefaultGoGentFactory) createMyUseCaseExecutor(config *types.GeminiClientConfig, dbURL string) (interfaces.UseCaseSpecificExecutor, error) {
    baseClient, err := f.CreateClient(config, dbURL)
    if err != nil {
        return nil, err
    }
    
    return myusecase.NewMyUseCaseExecutor(baseClient, baseClient, baseClient), nil
}
```

### Step 3: Use Your Implementation

```go
package main

import (
    "gogent/internal/factory"
    "gogent/internal/types"
)

func main() {
    factory := factory.NewGoGentFactory()
    
    config := &types.GeminiClientConfig{
        APIKey: "your-api-key",
        MaxRetries: 3,
        TimeoutSecs: 30,
    }
    
    executor, err := factory.CreateCustomExecutor("my-custom-use-case", config, "your-db-url")
    if err != nil {
        log.Fatal(err)
    }
    defer executor.Close()
    
    // Use your custom executor
    result, err := executor.ExecuteMultiVariation(ctx, &types.MultiExecutionRequest{
        ExecutionRunName: "my-analysis",
        BasePrompt:       "Analyze this data...",
        Configurations:   executor.GetDefaultConfigurations(),
    })
}
```

## Common Use Case Patterns

### 1. Content Generation
```go
type ContentGenerator interface {
    UseCaseSpecificExecutor
    
    GenerateBlogPost(ctx context.Context, topic string, requirements *ContentRequirements) (*ContentResult, error)
    GenerateMarketingCopy(ctx context.Context, product *Product, audience *Audience) (*MarketingResult, error)
    OptimizeContent(ctx context.Context, content string, metrics []string) (*OptimizationResult, error)
}
```

### 2. Risk Assessment
```go
type RiskAssessor interface {
    UseCaseSpecificExecutor
    
    AssessFinancialRisk(ctx context.Context, data *FinancialData) (*RiskAssessment, error)
    EvaluateOperationalRisk(ctx context.Context, process *BusinessProcess) (*OperationalRisk, error)
    GenerateRiskMitigation(ctx context.Context, risks []Risk) (*MitigationPlan, error)
}
```

### 3. Legal Analysis
```go
type LegalAnalyzer interface {
    UseCaseSpecificExecutor
    
    ReviewContract(ctx context.Context, contract *Contract) (*ContractReview, error)
    AnalyzeCompliance(ctx context.Context, document *Document, regulations []Regulation) (*ComplianceReport, error)
    GenerateLegalSummary(ctx context.Context, documents []Document) (*LegalSummary, error)
}
```

## Benefits of This Architecture

### 1. **Separation of Concerns**
Each interface has a single responsibility, making the system easier to understand and maintain.

### 2. **Dependency Injection**
Components depend on interfaces, not concrete implementations, enabling easy testing and swapping of components.

### 3. **Extensibility**
New use cases can be added without modifying existing code, following the Open/Closed Principle.

### 4. **Testability**
Mock implementations can be easily created for unit testing.

### 5. **Reusability**
Core functionality is shared across different use cases through composition.

### 6. **Flexibility**
Different implementations can be mixed and matched based on requirements.

## Testing Strategy

### Unit Testing with Mocks

```go
package tests

import (
    "testing"
    "gogent/internal/interfaces"
    "gogent/internal/types"
)

type MockMultiVariationExecutor struct{}

func (m *MockMultiVariationExecutor) ExecuteMultiVariation(ctx context.Context, request *types.MultiExecutionRequest) (*types.ExecutionResult, error) {
    return &types.ExecutionResult{
        SuccessCount: 3,
        ErrorCount:   0,
        TotalTime:    time.Second,
        Results:      []types.VariationResult{},
    }, nil
}

func TestProcurementManager(t *testing.T) {
    mockExecutor := &MockMultiVariationExecutor{}
    mockLogger := &MockExecutionLogger{}
    mockComparator := &MockResultComparator{}
    
    pm := procurement.NewProcurementAIManager(mockExecutor, mockLogger, mockComparator)
    
    // Test your procurement manager with mocks
    result, err := pm.EvaluateVendorProposals(ctx, sampleRFP)
    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

## Advanced Features

### 1. Plugin System
```go
type GoGentPlugin interface {
    GetName() string
    Initialize(config map[string]interface{}) error
    PreProcess(ctx context.Context, request *types.MultiExecutionRequest) (*types.MultiExecutionRequest, error)
    PostProcess(ctx context.Context, result *types.ExecutionResult) (*types.ExecutionResult, error)
    Cleanup() error
}
```

### 2. Analytics Provider
```go
type AnalyticsProvider interface {
    GetExecutionAnalytics(ctx context.Context, executionRunID string) (*types.ExecutionAnalytics, error)
    GetPerformanceMetrics(ctx context.Context, timeRange *types.TimeRange) (*types.PerformanceMetrics, error)
    GetCostAnalysis(ctx context.Context, timeRange *types.TimeRange) (*types.CostAnalysis, error)
    GetModelComparison(ctx context.Context, models []string, timeRange *types.TimeRange) (*types.ModelComparison, error)
}
```

### 3. Factory Pattern
The factory pattern enables easy creation and configuration of different implementations:

```go
factory := factory.NewGoGentFactory()

// Create standard client
client, err := factory.CreateClient(config, dbURL)

// Create procurement manager
procManager, err := factory.CreateProcurementManager(config, dbURL)

// Create custom executor
customExec, err := factory.CreateCustomExecutor("my-use-case", config, dbURL)

// Create analytics provider
analytics, err := factory.CreateAnalyticsProvider(dbURL)
```

## Best Practices

### 1. **Interface Design**
- Keep interfaces small and focused
- Prefer composition over inheritance
- Use clear, descriptive method names
- Return appropriate error types

### 2. **Implementation**
- Always validate inputs
- Handle errors gracefully
- Log important operations
- Use context for cancellation and timeouts

### 3. **Configuration**
- Provide sensible defaults
- Make configurations immutable where possible
- Validate configurations at creation time
- Support environment-specific overrides

### 4. **Testing**
- Write unit tests for all implementations
- Use mocks for external dependencies
- Test error conditions
- Validate configuration validation logic

## Migration Guide

### From Existing GoGent Code
1. Wrap existing client with adapter
2. Implement required interface methods
3. Update factory to support new interfaces
4. Gradually migrate to interface-based usage

### Adding New Use Cases
1. Define domain-specific interface extending `UseCaseSpecificExecutor`
2. Implement the interface with your business logic
3. Add factory method for creation
4. Write comprehensive tests
5. Document usage patterns

This architecture provides a solid foundation for building AI-powered applications that are maintainable, testable, and extensible while preserving all the powerful multi-variation capabilities that make GoGent unique. 