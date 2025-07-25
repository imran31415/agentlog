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

func runRealApiDemo() {
	fmt.Println("🔥 GoGent Real API Demo - With Gemini Integration")
	fmt.Println("================================================")
	fmt.Println()

	// Load environment variables
	if err := godotenv.Load("config.env"); err != nil {
		log.Printf("Warning: could not load config.env file: %v", err)
		fmt.Println("⚠️  Please ensure you have a config.env file with your GEMINI_API_KEY")
		return
	}

	// Get Gemini API key from environment
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" || apiKey == "your_gemini_api_key_here" {
		fmt.Println("⚠️  GEMINI_API_KEY not found or still using example value")
		fmt.Println("📝 Please edit config.env and add your real Gemini API key")
		fmt.Println("🔗 Get your API key from: https://aistudio.google.com/app/apikey")
		fmt.Println()
		fmt.Println("🎯 Running simple demo with mock responses instead...")
		runSimpleDemo()
		return
	}

	// Get database URL from environment
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		fmt.Println("⚠️  DB_URL not found, using SQLite for demo")
		// For demo purposes, we'll use the simple demo instead
		fmt.Println("🎯 Running simple demo without database...")
		runSimpleDemo()
		return
	}

	fmt.Printf("🔑 Using API Key: %s...%s\n", apiKey[:8], apiKey[len(apiKey)-4:])
	fmt.Println("🚀 Connecting to real Gemini API...")
	fmt.Println()

	// Create Gemini client configuration
	config := &types.GeminiClientConfig{
		APIKey:      apiKey,
		MaxRetries:  3,
		TimeoutSecs: 30,
	}

	// Create gogent client
	client, err := gogent.NewClient(dbURL, config)
	if err != nil {
		log.Printf("Failed to create gogent client: %v", err)
		fmt.Println("🎯 Falling back to simple demo...")
		runSimpleDemo()
		return
	}
	defer client.Close()

	// Example: Execute multiple variations of the same prompt
	ctx := context.Background()

	// Create different configurations for comparison
	temp1 := float32(0.1)
	temp2 := float32(0.7)
	temp3 := float32(1.2)

	maxTokens := int32(200)

	request := &types.MultiExecutionRequest{
		ExecutionRunName: "real-gemini-temperature-test",
		Description:      "Testing real Gemini API with different temperature settings",
		BasePrompt:       "Write a creative 3-sentence story about a time-traveling chef who discovers the secret ingredient to happiness.",
		Context:          "This is a real API test to compare how temperature affects Gemini's creativity",
		Configurations: []types.APIConfiguration{
			{
				VariationName: "precise-conservative",
				ModelName:     "gemini-1.5-flash",
				SystemPrompt:  "You are a precise, structured storyteller. Focus on clear narrative and logical flow.",
				Temperature:   &temp1,
				MaxTokens:     &maxTokens,
			},
			{
				VariationName: "balanced-creative",
				ModelName:     "gemini-1.5-flash",
				SystemPrompt:  "You are a creative storyteller who balances imagination with coherence.",
				Temperature:   &temp2,
				MaxTokens:     &maxTokens,
			},
			{
				VariationName: "highly-experimental",
				ModelName:     "gemini-1.5-flash",
				SystemPrompt:  "You are a wildly creative storyteller who takes bold narrative risks and uses unexpected metaphors.",
				Temperature:   &temp3,
				MaxTokens:     &maxTokens,
			},
		},
		ComparisonConfig: &types.ComparisonConfig{
			Enabled: true,
			Metrics: []string{"response_time", "creativity", "coherence"},
		},
	}

	fmt.Println("📡 Executing multi-variation request with real Gemini API...")
	fmt.Println("⏳ This may take a few seconds...")
	fmt.Println()

	result, err := client.ExecuteMultiVariation(ctx, request)
	if err != nil {
		log.Fatalf("Failed to execute multi-variation: %v", err)
	}

	// Display results
	fmt.Println("🎉 REAL API EXECUTION RESULTS")
	fmt.Println("==============================")
	fmt.Printf("✅ Execution Run: %s\n", result.ExecutionRun.Name)
	fmt.Printf("⏱️  Total Time: %v\n", result.TotalTime)
	fmt.Printf("✅ Success Count: %d\n", result.SuccessCount)
	fmt.Printf("❌ Error Count: %d\n", result.ErrorCount)
	fmt.Println()

	fmt.Println("📊 REAL GEMINI RESPONSES")
	fmt.Println("========================")
	for i, variation := range result.Results {
		fmt.Printf("\n%d. %s (Model: %s)\n", i+1, variation.Configuration.VariationName, variation.Configuration.ModelName)
		fmt.Printf("   🌡️  Temperature: %.1f\n", *variation.Configuration.Temperature)
		fmt.Printf("   📝 System Prompt: %s\n", variation.Configuration.SystemPrompt)
		fmt.Printf("   ⏱️  Response Time: %dms\n", variation.Response.ResponseTimeMs)
		fmt.Printf("   🏁 Finish Reason: %s\n", variation.Response.FinishReason)

		if variation.Response.UsageMetadata != nil {
			fmt.Printf("   📊 Token Usage: %v\n", variation.Response.UsageMetadata)
		}

		fmt.Printf("   📄 Response:\n")
		fmt.Printf("      %s\n", variation.Response.ResponseText)

		if variation.Response.ErrorMessage != "" {
			fmt.Printf("   ❌ Error: %s\n", variation.Response.ErrorMessage)
		}
	}

	if result.Comparison != nil {
		fmt.Println("\n🏆 COMPARISON ANALYSIS")
		fmt.Println("======================")
		fmt.Printf("🥇 Best Configuration: %s\n", result.Comparison.BestConfigurationID)
		fmt.Printf("📈 Analysis: %s\n", result.Comparison.AnalysisNotes)
	}

	fmt.Println("\n✨ Real API demo completed successfully!")
	fmt.Println()
	fmt.Println("💾 DATABASE LOGGING")
	fmt.Println("===================")
	fmt.Println("All API calls, responses, and metadata have been logged to your database:")
	fmt.Println("• execution_runs - Run metadata and timing")
	fmt.Println("• api_configurations - Each variation's parameters")
	fmt.Println("• api_requests - Complete request details")
	fmt.Println("• api_responses - Full responses with usage stats")
	fmt.Println()
	fmt.Println("🔍 Query your database to analyze the data:")
	fmt.Println("SELECT * FROM api_responses ORDER BY created_at DESC LIMIT 3;")
}
