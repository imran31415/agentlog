package main

import (
	"fmt"
	"time"

	"gogent/internal/types"
)

// SimpleDemoClient demonstrates the core gogent concepts without database complexity
type SimpleDemoClient struct {
	config *types.GeminiClientConfig
}

// NewSimpleDemoClient creates a basic demo client
func NewSimpleDemoClient(config *types.GeminiClientConfig) *SimpleDemoClient {
	return &SimpleDemoClient{
		config: config,
	}
}

// ExecuteMultiVariationDemo demonstrates multi-variation execution with mock responses
func (c *SimpleDemoClient) ExecuteMultiVariationDemo(request *types.MultiExecutionRequest) *types.ExecutionResult {
	fmt.Printf("🚀 Starting Multi-Variation Execution: %s\n", request.ExecutionRunName)
	fmt.Printf("📝 Description: %s\n", request.Description)
	fmt.Printf("💭 Base Prompt: %s\n", request.BasePrompt)
	fmt.Printf("🔧 Number of Variations: %d\n", len(request.Configurations))
	fmt.Println()

	executionRun := types.ExecutionRun{
		ID:          generateID(),
		Name:        request.ExecutionRunName,
		Description: request.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	startTime := time.Now()
	results := make([]types.VariationResult, 0, len(request.Configurations))

	for i, config := range request.Configurations {
		fmt.Printf("⚙️  Executing Variation %d: %s\n", i+1, config.VariationName)
		fmt.Printf("   Model: %s\n", config.ModelName)
		if config.Temperature != nil {
			fmt.Printf("   Temperature: %.1f\n", *config.Temperature)
		}
		if config.SystemPrompt != "" {
			fmt.Printf("   System Prompt: %s\n", config.SystemPrompt)
		}

		// Simulate API call with realistic delay
		time.Sleep(time.Duration(200+i*50) * time.Millisecond)

		// Create mock response with variation-specific content
		responseText := generateMockResponse(request.BasePrompt, config)
		responseTime := int32(200 + i*50)

		apiRequest := types.APIRequest{
			ID:              generateID(),
			ExecutionRunID:  executionRun.ID,
			ConfigurationID: config.ID,
			RequestType:     types.RequestTypeGenerate,
			Prompt:          request.BasePrompt,
			Context:         request.Context,
			CreatedAt:       time.Now(),
		}

		apiResponse := types.APIResponse{
			ID:             generateID(),
			RequestID:      apiRequest.ID,
			ResponseStatus: types.ResponseStatusSuccess,
			ResponseText:   responseText,
			FinishReason:   "stop",
			ResponseTimeMs: responseTime,
			CreatedAt:      time.Now(),
		}

		variationResult := types.VariationResult{
			Configuration: config,
			Request:       apiRequest,
			Response:      apiResponse,
			ExecutionTime: int64(responseTime), // Already in milliseconds
		}

		results = append(results, variationResult)

		fmt.Printf("   ✅ Response (%dms): %s\n", responseTime, truncateString(responseText, 100))
		fmt.Println()
	}

	totalTime := time.Since(startTime)

	result := &types.ExecutionResult{
		ExecutionRun: executionRun,
		Results:      results,
		TotalTime:    totalTime.Milliseconds(),
		SuccessCount: len(results),
		ErrorCount:   0,
	}

	// Simple comparison - find fastest response
	if request.ComparisonConfig != nil && request.ComparisonConfig.Enabled {
		var fastest *types.VariationResult
		for i := range results {
			if fastest == nil || results[i].Response.ResponseTimeMs < fastest.Response.ResponseTimeMs {
				fastest = &results[i]
			}
		}

		if fastest != nil {
			result.Comparison = &types.ComparisonResult{
				ID:                  generateID(),
				ExecutionRunID:      executionRun.ID,
				ComparisonType:      "performance",
				MetricName:          "response_time",
				BestConfigurationID: fastest.Configuration.ID,
				AnalysisNotes:       fmt.Sprintf("Fastest response: %dms with variation '%s'", fastest.Response.ResponseTimeMs, fastest.Configuration.VariationName),
				CreatedAt:           time.Now(),
			}
		}
	}

	return result
}

// Helper functions
func generateID() string {
	return fmt.Sprintf("demo-%d", time.Now().UnixNano()%1000000)
}

func generateMockResponse(prompt string, config types.APIConfiguration) string {
	responses := map[string]string{
		"creative":   "🎨 [Creative Response] " + prompt + " - This response emphasizes creativity and artistic expression, with vivid imagery and imaginative elements.",
		"analytical": "🔍 [Analytical Response] " + prompt + " - This response provides a structured, logical analysis with clear reasoning and factual information.",
		"balanced":   "⚖️ [Balanced Response] " + prompt + " - This response offers a well-rounded perspective, combining creativity with analytical thinking.",
	}

	// Use variation name to determine response style
	for key, response := range responses {
		if containsString(config.VariationName, key) {
			return response
		}
	}

	// Default response based on temperature
	if config.Temperature != nil {
		if *config.Temperature < 0.4 {
			return responses["analytical"]
		} else if *config.Temperature > 0.7 {
			return responses["creative"]
		}
	}

	return responses["balanced"]
}

func containsString(text, substr string) bool {
	return len(text) >= len(substr) &&
		text[:len(substr)] == substr[:len(substr)] ||
		len(text) > len(substr) &&
			text[len(text)-len(substr):] == substr
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func runSimpleDemo() {
	fmt.Println("🎯 GoGent Simple Demo - Multi-Variation AI Execution")
	fmt.Println("=====================================================")
	fmt.Println()

	// Create demo client
	config := &types.GeminiClientConfig{
		APIKey:      "demo-api-key",
		MaxRetries:  3,
		TimeoutSecs: 30,
	}

	client := NewSimpleDemoClient(config)

	// Create test configurations
	temp1 := float32(0.2)
	temp2 := float32(0.7)
	temp3 := float32(1.0)
	maxTokens := int32(150)

	request := &types.MultiExecutionRequest{
		ExecutionRunName: "creative-writing-comparison",
		Description:      "Comparing different temperature settings for creative story generation",
		BasePrompt:       "Write a short story about a robot learning to paint masterpieces",
		Context:          "This is a test to see how temperature affects creativity in storytelling",
		Configurations: []types.APIConfiguration{
			{
				ID:            generateID(),
				VariationName: "analytical-precise",
				ModelName:     "gemini-1.5-flash",
				SystemPrompt:  "You are a precise, analytical storyteller who focuses on logical narrative structure.",
				Temperature:   &temp1,
				MaxTokens:     &maxTokens,
			},
			{
				ID:            generateID(),
				VariationName: "creative-expressive",
				ModelName:     "gemini-1.5-flash",
				SystemPrompt:  "You are a highly creative storyteller who uses vivid imagery and emotional depth.",
				Temperature:   &temp2,
				MaxTokens:     &maxTokens,
			},
			{
				ID:            generateID(),
				VariationName: "experimental-wild",
				ModelName:     "gemini-1.5-flash",
				SystemPrompt:  "You are an experimental storyteller who takes bold creative risks.",
				Temperature:   &temp3,
				MaxTokens:     &maxTokens,
			},
		},
		ComparisonConfig: &types.ComparisonConfig{
			Enabled: true,
			Metrics: []string{"response_time", "creativity"},
		},
	}

	// Execute multi-variation request
	result := client.ExecuteMultiVariationDemo(request)

	// Display results
	fmt.Println("📊 EXECUTION RESULTS")
	fmt.Println("====================")
	fmt.Printf("✅ Execution Run: %s\n", result.ExecutionRun.Name)
	fmt.Printf("⏱️  Total Time: %v\n", result.TotalTime)
	fmt.Printf("✅ Success Count: %d\n", result.SuccessCount)
	fmt.Printf("❌ Error Count: %d\n", result.ErrorCount)
	fmt.Println()

	fmt.Println("🔍 VARIATION RESULTS")
	fmt.Println("====================")
	for i, variation := range result.Results {
		fmt.Printf("\n%d. %s (Model: %s)\n", i+1, variation.Configuration.VariationName, variation.Configuration.ModelName)
		fmt.Printf("   📝 System Prompt: %s\n", variation.Configuration.SystemPrompt)
		if variation.Configuration.Temperature != nil {
			fmt.Printf("   🌡️  Temperature: %.1f\n", *variation.Configuration.Temperature)
		}
		fmt.Printf("   ⏱️  Response Time: %dms\n", variation.Response.ResponseTimeMs)
		fmt.Printf("   📄 Response: %s\n", variation.Response.ResponseText)
	}

	if result.Comparison != nil {
		fmt.Println("\n🏆 COMPARISON RESULTS")
		fmt.Println("====================")
		fmt.Printf("🥇 Best Configuration: %s\n", result.Comparison.BestConfigurationID)
		fmt.Printf("📈 Analysis: %s\n", result.Comparison.AnalysisNotes)
	}

	fmt.Println("\n✨ Demo completed successfully!")
	fmt.Println("\n🔧 NEXT STEPS:")
	fmt.Println("1. Set up your database with: make init-db")
	fmt.Println("2. Add your real Gemini API key to config.env")
	fmt.Println("3. Run the full version with database logging")
	fmt.Println("4. Integrate real Gemini API calls (currently using mock responses)")
}

// This function can be called from main.go for a simple demo
