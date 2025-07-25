package gogent

import (
	"fmt"
	"testing"
	"time"

	"gogent/internal/types"

	"github.com/google/uuid"
)

func TestConfigurationValidation(t *testing.T) {
	tests := []struct {
		name           string
		config         types.APIConfiguration
		expectError    bool
		expectedErrMsg string
	}{
		{
			name: "valid_configuration",
			config: types.APIConfiguration{
				ID:            uuid.New().String(),
				VariationName: "test-variation",
				ModelName:     "gemini-1.5-flash",
				SystemPrompt:  "You are a helpful assistant",
				Temperature:   &[]float32{0.7}[0],
				MaxTokens:     &[]int32{150}[0],
			},
			expectError: false,
		},
		{
			name: "missing_variation_name",
			config: types.APIConfiguration{
				ID:        uuid.New().String(),
				ModelName: "gemini-1.5-flash",
			},
			expectError:    true,
			expectedErrMsg: "variation name is required",
		},
		{
			name: "missing_model_name",
			config: types.APIConfiguration{
				ID:            uuid.New().String(),
				VariationName: "test-variation",
			},
			expectError:    true,
			expectedErrMsg: "model name is required",
		},
		{
			name: "invalid_temperature_too_low",
			config: types.APIConfiguration{
				ID:            uuid.New().String(),
				VariationName: "test-variation",
				ModelName:     "gemini-1.5-flash",
				Temperature:   &[]float32{-0.1}[0],
			},
			expectError:    true,
			expectedErrMsg: "temperature must be between 0.0 and 2.0",
		},
		{
			name: "invalid_temperature_too_high",
			config: types.APIConfiguration{
				ID:            uuid.New().String(),
				VariationName: "test-variation",
				ModelName:     "gemini-1.5-flash",
				Temperature:   &[]float32{2.1}[0],
			},
			expectError:    true,
			expectedErrMsg: "temperature must be between 0.0 and 2.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfiguration(&tt.config)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if !contains(err.Error(), tt.expectedErrMsg) {
					t.Errorf("expected error to contain %q, got %q", tt.expectedErrMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestMultiExecutionRequestValidation(t *testing.T) {
	tests := []struct {
		name           string
		request        types.MultiExecutionRequest
		expectError    bool
		expectedErrMsg string
	}{
		{
			name: "valid_request",
			request: types.MultiExecutionRequest{
				ExecutionRunName: "test-run",
				Description:      "Test execution",
				BasePrompt:       "Test prompt",
				Configurations: []types.APIConfiguration{
					{
						ID:            uuid.New().String(),
						VariationName: "variation-1",
						ModelName:     "gemini-1.5-flash",
					},
				},
			},
			expectError: false,
		},
		{
			name: "missing_execution_run_name",
			request: types.MultiExecutionRequest{
				BasePrompt: "Test prompt",
				Configurations: []types.APIConfiguration{
					{
						ID:            uuid.New().String(),
						VariationName: "variation-1",
						ModelName:     "gemini-1.5-flash",
					},
				},
			},
			expectError:    true,
			expectedErrMsg: "execution run name is required",
		},
		{
			name: "missing_base_prompt",
			request: types.MultiExecutionRequest{
				ExecutionRunName: "test-run",
				Configurations: []types.APIConfiguration{
					{
						ID:            uuid.New().String(),
						VariationName: "variation-1",
						ModelName:     "gemini-1.5-flash",
					},
				},
			},
			expectError:    true,
			expectedErrMsg: "base prompt is required",
		},
		{
			name: "no_configurations",
			request: types.MultiExecutionRequest{
				ExecutionRunName: "test-run",
				BasePrompt:       "Test prompt",
				Configurations:   []types.APIConfiguration{},
			},
			expectError:    true,
			expectedErrMsg: "at least one configuration is required",
		},
		{
			name: "too_many_configurations",
			request: types.MultiExecutionRequest{
				ExecutionRunName: "test-run",
				BasePrompt:       "Test prompt",
				Configurations:   make([]types.APIConfiguration, 11), // Assuming max is 10
			},
			expectError:    true,
			expectedErrMsg: "maximum 10 configurations allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMultiExecutionRequest(&tt.request)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if !contains(err.Error(), tt.expectedErrMsg) {
					t.Errorf("expected error to contain %q, got %q", tt.expectedErrMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestExecutionResult(t *testing.T) {
	tests := []struct {
		name               string
		results            []types.VariationResult
		expectedSuccess    int
		expectedError      int
		expectedFastest    string
		expectedComparison bool
	}{
		{
			name: "all_successful_results",
			results: []types.VariationResult{
				{
					Configuration: types.APIConfiguration{
						ID:            "config-1",
						VariationName: "variation-1",
						ModelName:     "gemini-1.5-flash",
					},
					Response: types.APIResponse{
						ID:             "response-1",
						ResponseStatus: types.ResponseStatusSuccess,
						ResponseText:   "Response 1",
						ResponseTimeMs: 200,
					},
					ExecutionTime: 200,
				},
				{
					Configuration: types.APIConfiguration{
						ID:            "config-2",
						VariationName: "variation-2",
						ModelName:     "gemini-1.5-flash",
					},
					Response: types.APIResponse{
						ID:             "response-2",
						ResponseStatus: types.ResponseStatusSuccess,
						ResponseText:   "Response 2",
						ResponseTimeMs: 150,
					},
					ExecutionTime: 150,
				},
			},
			expectedSuccess:    2,
			expectedError:      0,
			expectedFastest:    "config-2",
			expectedComparison: true,
		},
		{
			name: "mixed_results",
			results: []types.VariationResult{
				{
					Configuration: types.APIConfiguration{
						ID:            "config-1",
						VariationName: "variation-1",
						ModelName:     "gemini-1.5-flash",
					},
					Response: types.APIResponse{
						ID:             "response-1",
						ResponseStatus: types.ResponseStatusSuccess,
						ResponseText:   "Response 1",
						ResponseTimeMs: 200,
					},
					ExecutionTime: 200,
				},
				{
					Configuration: types.APIConfiguration{
						ID:            "config-2",
						VariationName: "variation-2",
						ModelName:     "gemini-1.5-flash",
					},
					Response: types.APIResponse{
						ID:             "response-2",
						ResponseStatus: types.ResponseStatusError,
						ErrorMessage:   "API Error",
						ResponseTimeMs: 0,
					},
					ExecutionTime: 0,
				},
			},
			expectedSuccess:    1,
			expectedError:      1,
			expectedFastest:    "config-1",
			expectedComparison: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzeExecutionResults(tt.results)

			if result.SuccessCount != tt.expectedSuccess {
				t.Errorf("expected success count %d, got %d", tt.expectedSuccess, result.SuccessCount)
			}
			if result.ErrorCount != tt.expectedError {
				t.Errorf("expected error count %d, got %d", tt.expectedError, result.ErrorCount)
			}

			if tt.expectedComparison && len(tt.results) > 0 {
				fastest := findFastestResult(tt.results)
				if fastest != nil && fastest.Configuration.ID != tt.expectedFastest {
					t.Errorf("expected fastest %s, got %s", tt.expectedFastest, fastest.Configuration.ID)
				}
			}
		})
	}
}

func TestComparisonMetrics(t *testing.T) {
	tests := []struct {
		name        string
		results     []types.VariationResult
		metrics     []string
		expectBest  string
		expectNotes string
	}{
		{
			name: "performance_comparison",
			results: []types.VariationResult{
				{
					Configuration: types.APIConfiguration{
						ID:            "fast-config",
						VariationName: "fast-variation",
					},
					Response: types.APIResponse{
						ResponseTimeMs: 100,
						ResponseStatus: types.ResponseStatusSuccess,
					},
				},
				{
					Configuration: types.APIConfiguration{
						ID:            "slow-config",
						VariationName: "slow-variation",
					},
					Response: types.APIResponse{
						ResponseTimeMs: 300,
						ResponseStatus: types.ResponseStatusSuccess,
					},
				},
			},
			metrics:     []string{"response_time"},
			expectBest:  "fast-config",
			expectNotes: "fastest response time",
		},
		{
			name: "quality_comparison",
			results: []types.VariationResult{
				{
					Configuration: types.APIConfiguration{
						ID:            "detailed-config",
						VariationName: "detailed-variation",
					},
					Response: types.APIResponse{
						ResponseText:   "This is a very detailed and comprehensive response with lots of information.",
						ResponseStatus: types.ResponseStatusSuccess,
						ResponseTimeMs: 200,
					},
				},
				{
					Configuration: types.APIConfiguration{
						ID:            "brief-config",
						VariationName: "brief-variation",
					},
					Response: types.APIResponse{
						ResponseText:   "Brief response.",
						ResponseStatus: types.ResponseStatusSuccess,
						ResponseTimeMs: 150,
					},
				},
			},
			metrics:     []string{"response_length"},
			expectBest:  "detailed-config",
			expectNotes: "longest response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comparison := compareResultsByMetrics(tt.results, tt.metrics)

			if comparison.BestConfigurationID != tt.expectBest {
				t.Errorf("expected best config %s, got %s", tt.expectBest, comparison.BestConfigurationID)
			}
			if !contains(comparison.AnalysisNotes, tt.expectNotes) {
				t.Errorf("expected notes to contain %q, got %q", tt.expectNotes, comparison.AnalysisNotes)
			}
		})
	}
}

// Helper functions for the tests
func validateConfiguration(config *types.APIConfiguration) error {
	if config.VariationName == "" {
		return fmt.Errorf("variation name is required")
	}
	if config.ModelName == "" {
		return fmt.Errorf("model name is required")
	}
	if config.Temperature != nil && (*config.Temperature < 0.0 || *config.Temperature > 2.0) {
		return fmt.Errorf("temperature must be between 0.0 and 2.0")
	}
	return nil
}

func validateMultiExecutionRequest(request *types.MultiExecutionRequest) error {
	if request.ExecutionRunName == "" {
		return fmt.Errorf("execution run name is required")
	}
	if request.BasePrompt == "" {
		return fmt.Errorf("base prompt is required")
	}
	if len(request.Configurations) == 0 {
		return fmt.Errorf("at least one configuration is required")
	}
	if len(request.Configurations) > 10 {
		return fmt.Errorf("maximum 10 configurations allowed")
	}
	return nil
}

func analyzeExecutionResults(results []types.VariationResult) types.ExecutionResult {
	successCount := 0
	errorCount := 0
	totalTime := int64(0)

	for _, result := range results {
		if result.Response.ResponseStatus == types.ResponseStatusSuccess {
			successCount++
		} else {
			errorCount++
		}
		totalTime += result.ExecutionTime
	}

	return types.ExecutionResult{
		Results:      results,
		SuccessCount: successCount,
		ErrorCount:   errorCount,
		TotalTime:    totalTime,
	}
}

func findFastestResult(results []types.VariationResult) *types.VariationResult {
	var fastest *types.VariationResult

	for i := range results {
		if results[i].Response.ResponseStatus == types.ResponseStatusSuccess {
			if fastest == nil || results[i].Response.ResponseTimeMs < fastest.Response.ResponseTimeMs {
				fastest = &results[i]
			}
		}
	}

	return fastest
}

func compareResultsByMetrics(results []types.VariationResult, metrics []string) types.ComparisonResult {
	if len(results) == 0 {
		return types.ComparisonResult{}
	}

	comparison := types.ComparisonResult{
		ID:             uuid.New().String(),
		ComparisonType: "multi-metric",
		CreatedAt:      time.Now(),
	}

	for _, metric := range metrics {
		switch metric {
		case "response_time":
			fastest := findFastestResult(results)
			if fastest != nil {
				comparison.BestConfigurationID = fastest.Configuration.ID
				comparison.MetricName = "response_time"
				comparison.AnalysisNotes = fmt.Sprintf("fastest response time: %dms", fastest.Response.ResponseTimeMs)
			}
		case "response_length":
			var longest *types.VariationResult
			for i := range results {
				if results[i].Response.ResponseStatus == types.ResponseStatusSuccess {
					if longest == nil || len(results[i].Response.ResponseText) > len(longest.Response.ResponseText) {
						longest = &results[i]
					}
				}
			}
			if longest != nil {
				comparison.BestConfigurationID = longest.Configuration.ID
				comparison.MetricName = "response_length"
				comparison.AnalysisNotes = fmt.Sprintf("longest response: %d characters", len(longest.Response.ResponseText))
			}
		}
	}

	return comparison
}

// Helper function for string contains check
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
