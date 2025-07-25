package adapters

import (
	"context"
	"fmt"

	"gogent/internal/gogent"
	"gogent/internal/interfaces"
	"gogent/internal/types"
)

// GoGentClientAdapter adapts the current gogent.Client to implement our interfaces
type GoGentClientAdapter struct {
	client *gogent.Client
	userID string // Store user ID for all operations
}

// NewGoGentClientAdapter creates a new adapter for the gogent client
func NewGoGentClientAdapter(client *gogent.Client, userID string) *GoGentClientAdapter {
	return &GoGentClientAdapter{
		client: client,
		userID: userID,
	}
}

// Ensure the adapter implements all required interfaces
var (
	_ interfaces.MultiVariationExecutor = (*GoGentClientAdapter)(nil)
	_ interfaces.ExecutionLogger        = (*GoGentClientAdapter)(nil)
	_ interfaces.ConfigurationManager   = (*GoGentClientAdapter)(nil)
	_ interfaces.ResultComparator       = (*GoGentClientAdapter)(nil)
	_ interfaces.GoGentClient           = (*GoGentClientAdapter)(nil)
)

// MultiVariationExecutor interface implementation

func (adapter *GoGentClientAdapter) ExecuteMultiVariation(ctx context.Context, request *types.MultiExecutionRequest) (*types.ExecutionResult, error) {
	return adapter.client.ExecuteMultiVariation(ctx, adapter.userID, request)
}

func (adapter *GoGentClientAdapter) ExecuteSingleVariation(ctx context.Context, config *types.APIConfiguration, prompt, context string) (*types.VariationResult, error) {
	// Create a mini multi-execution with just one configuration
	request := &types.MultiExecutionRequest{
		ExecutionRunName: fmt.Sprintf("single-variation-%s", config.VariationName),
		Description:      "Single variation execution",
		BasePrompt:       prompt,
		Context:          context,
		Configurations:   []types.APIConfiguration{*config},
	}

	result, err := adapter.client.ExecuteMultiVariation(ctx, adapter.userID, request)
	if err != nil {
		return nil, err
	}

	if len(result.Results) == 0 {
		return nil, fmt.Errorf("no results returned from execution")
	}

	return &result.Results[0], nil
}

func (adapter *GoGentClientAdapter) Close() error {
	return adapter.client.Close()
}

// ExecutionLogger interface implementation

func (adapter *GoGentClientAdapter) CreateExecutionRun(ctx context.Context, name, description string, enableFunctionCalling bool) (*types.ExecutionRun, error) {
	return adapter.client.CreateExecutionRun(ctx, adapter.userID, name, description, enableFunctionCalling)
}

func (adapter *GoGentClientAdapter) LogAPIRequest(ctx context.Context, request *types.APIRequest) error {
	return adapter.client.LogAPIRequest(ctx, adapter.userID, request)
}

func (adapter *GoGentClientAdapter) LogAPIResponse(ctx context.Context, response *types.APIResponse) error {
	return adapter.client.LogAPIResponse(ctx, adapter.userID, response)
}

func (adapter *GoGentClientAdapter) LogFunctionCall(ctx context.Context, call *types.FunctionCall) error {
	return adapter.client.LogFunctionCall(ctx, call)
}

func (adapter *GoGentClientAdapter) GetExecutionRun(ctx context.Context, id string) (*types.ExecutionRun, error) {
	return adapter.client.GetExecutionRun(ctx, adapter.userID, id)
}

func (adapter *GoGentClientAdapter) ListExecutionRuns(ctx context.Context, limit, offset int) ([]*types.ExecutionRun, error) {
	// Convert int to int32 for the client method
	return adapter.client.ListExecutionRuns(ctx, adapter.userID, int32(limit), int32(offset))
}

// ConfigurationManager interface implementation

func (adapter *GoGentClientAdapter) CreateConfiguration(ctx context.Context, config *types.APIConfiguration) error {
	return adapter.client.CreateAPIConfiguration(ctx, adapter.userID, config)
}

func (adapter *GoGentClientAdapter) GetConfiguration(ctx context.Context, id string) (*types.APIConfiguration, error) {
	// TODO: Implement in the underlying client
	return nil, fmt.Errorf("GetConfiguration not yet implemented")
}

func (adapter *GoGentClientAdapter) ListConfigurations(ctx context.Context, executionRunID string) ([]*types.APIConfiguration, error) {
	// TODO: Implement in the underlying client
	return nil, fmt.Errorf("ListConfigurations not yet implemented")
}

func (adapter *GoGentClientAdapter) UpdateConfiguration(ctx context.Context, config *types.APIConfiguration) error {
	// TODO: Implement in the underlying client
	return fmt.Errorf("UpdateConfiguration not yet implemented")
}

func (adapter *GoGentClientAdapter) DeleteConfiguration(ctx context.Context, id string) error {
	// TODO: Implement in the underlying client
	return fmt.Errorf("DeleteConfiguration not yet implemented")
}

// ResultComparator interface implementation

func (adapter *GoGentClientAdapter) CompareResults(ctx context.Context, result *types.ExecutionResult, metrics []string) (*types.ComparisonResult, error) {
	// TODO: Implement proper comparison logic
	return nil, fmt.Errorf("CompareResults not yet implemented")
}

func (adapter *GoGentClientAdapter) SaveComparison(ctx context.Context, comparison *types.ComparisonResult) error {
	return adapter.client.StoreComparisonResult(ctx, adapter.userID, comparison)
}

func (adapter *GoGentClientAdapter) GetComparison(ctx context.Context, id string) (*types.ComparisonResult, error) {
	// TODO: Implement in the underlying client
	return nil, fmt.Errorf("GetComparison not yet implemented")
}

func (adapter *GoGentClientAdapter) ListComparisons(ctx context.Context, executionRunID string) ([]*types.ComparisonResult, error) {
	// TODO: Implement in the underlying client
	return nil, fmt.Errorf("ListComparisons not yet implemented")
}

// GetUnderlyingClient returns the underlying gogent client for advanced usage
func (adapter *GoGentClientAdapter) GetUnderlyingClient() *gogent.Client {
	return adapter.client
}
