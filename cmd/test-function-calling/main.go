package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"gogent/internal/gogent"
	"gogent/internal/types"

	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("🧪 Function Calling Test")
	fmt.Println("========================")
	fmt.Println()

	// Load environment variables
	if err := godotenv.Load("config.env"); err != nil {
		log.Printf("Warning: could not load config.env file: %v", err)
		return
	}

	dbURL := os.Getenv("DB_URL")
	apiKey := os.Getenv("GEMINI_API_KEY")

	if dbURL == "" {
		log.Fatal("DB_URL environment variable is required")
	}

	config := &types.GeminiClientConfig{
		APIKey:      apiKey,
		MaxRetries:  3,
		TimeoutSecs: 30,
	}

	client, err := gogent.NewClient(dbURL, config)
	if err != nil {
		log.Fatalf("Failed to create gogent client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Check if function definitions exist
	fmt.Println("🔍 Checking Function Definitions...")
	// Note: We would need to implement ListFunctionDefinitions method
	fmt.Println("💡 Please ensure you have saved your weather function!")
	fmt.Println()

	// Test execution with function calling enabled
	fmt.Println("🧪 Testing Function Calling with Weather Query...")

	temp := float32(0.7)
	maxTokens := int32(150)

	request := &types.MultiExecutionRequest{
		ExecutionRunName:      "function-calling-test",
		Description:           "Testing weather function calling",
		BasePrompt:            "What's the weather like in Los Angeles?",
		EnableFunctionCalling: true, // Key: Enable function calling!
		Configurations: []types.APIConfiguration{
			{
				VariationName: "function-enabled",
				ModelName:     "gemini-1.5-flash",
				SystemPrompt:  "You are a helpful assistant that can call functions to get real-time information.",
				Temperature:   &temp,
				MaxTokens:     &maxTokens,
				// Note: In real usage, Tools would be populated from function definitions
			},
		},
		// Include function tools (this would normally come from database)
		FunctionTools: []types.Tool{
			{
				Name:        "get_weather",
				Description: "Get current weather information for a location",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"location": map[string]interface{}{
							"type":        "string",
							"description": "The location to get weather for",
						},
					},
					"required": []string{"location"},
				},
			},
		},
	}

	result, err := client.ExecuteMultiVariation(ctx, request)
	if err != nil {
		log.Printf("❌ Execution failed: %v", err)
		return
	}

	fmt.Printf("✅ Execution completed!\n")
	fmt.Printf("📊 Results: %d successful, %d errors\n", result.SuccessCount, result.ErrorCount)
	fmt.Printf("⏱️ Total time: %d ms\n", result.TotalTime)
	fmt.Printf("🔧 Function calling enabled: %v\n", result.ExecutionRun.EnableFunctionCalling)

	// Check if any function calls were made
	functionCallsDetected := false
	for i, variation := range result.Results {
		fmt.Printf("\n📝 Variation %d: %s\n", i+1, variation.Configuration.VariationName)

		if variation.Response.FunctionCallResponse != nil && len(variation.Response.FunctionCallResponse) > 0 {
			functionCallsDetected = true
			fmt.Printf("  🔧 Function Call Detected: %+v\n", variation.Response.FunctionCallResponse)
		}

		responseText := variation.Response.ResponseText
		if len(responseText) > 150 {
			responseText = responseText[:150] + "..."
		}
		fmt.Printf("  💬 Response: %s\n", responseText)
	}

	fmt.Println()
	if functionCallsDetected {
		fmt.Println("🎉 SUCCESS: Function calls were detected!")
		fmt.Println("🔍 Check the function_calls table in your database for detailed logs.")
	} else {
		fmt.Println("❌ No function calls detected.")
		fmt.Println("💡 This might happen if:")
		fmt.Println("   • Function definitions aren't properly saved")
		fmt.Println("   • The AI doesn't determine functions are needed")
		fmt.Println("   • Function calling integration needs debugging")
	}

	fmt.Println()
	fmt.Println("🔧 Next steps:")
	fmt.Println("1. Save your function definition in the frontend")
	fmt.Println("2. Enable 'Function Calling' toggle in Execute screen")
	fmt.Println("3. Select your weather function")
	fmt.Println("4. Try prompt: 'What's the current weather in Los Angeles?'")
}
