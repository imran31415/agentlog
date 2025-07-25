package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"gogent/internal/gogent"
	"gogent/internal/types"

	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("🔍 GoGent Function Call History Checker")
	fmt.Println("=====================================")
	fmt.Println()

	// Load environment variables
	if err := godotenv.Load("config.env"); err != nil {
		log.Printf("Warning: could not load config.env file: %v", err)
		return
	}

	// Get database URL and API key from environment
	dbURL := os.Getenv("DB_URL")
	apiKey := os.Getenv("GEMINI_API_KEY")

	if dbURL == "" {
		log.Fatal("DB_URL environment variable is required")
	}

	// Create client configuration
	config := &types.GeminiClientConfig{
		APIKey:      apiKey,
		MaxRetries:  3,
		TimeoutSecs: 30,
	}

	// Create gogent client
	client, err := gogent.NewClient(dbURL, config)
	if err != nil {
		log.Fatalf("Failed to create gogent client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Check recent execution runs
	fmt.Println("📊 Recent Execution Runs:")
	executionRuns, err := client.ListExecutionRuns(ctx, 10, 0)
	if err != nil {
		log.Printf("Failed to get execution runs: %v", err)
	} else {
		for _, run := range executionRuns {
			fmt.Printf("  • %s - %s (Function Calling: %v) - %s\n",
				run.Name, run.Description, run.EnableFunctionCalling, run.CreatedAt.Format(time.RFC3339))
		}
	}

	fmt.Println()

	// If there are recent execution runs, let's examine the most recent one
	if len(executionRuns) > 0 {
		mostRecentRun := executionRuns[0]
		fmt.Printf("🔬 Examining Most Recent Execution: %s\n", mostRecentRun.Name)

		// Get detailed execution result
		result, err := client.GetExecutionResult(ctx, mostRecentRun.ID)
		if err != nil {
			log.Printf("Failed to get execution result: %v", err)
		} else {
			fmt.Printf("  • Total Variations: %d\n", len(result.Results))
			fmt.Printf("  • Success Count: %d\n", result.SuccessCount)
			fmt.Printf("  • Error Count: %d\n", result.ErrorCount)
			fmt.Printf("  • Total Time: %d ms\n", result.TotalTime)

			// Check each variation for function call activity
			functionCallsFound := false
			for i, variation := range result.Results {
				fmt.Printf("\n  📝 Variation %d: %s\n", i+1, variation.Configuration.VariationName)
				fmt.Printf("    Model: %s\n", variation.Configuration.ModelName)
				fmt.Printf("    Response Time: %d ms\n", variation.Response.ResponseTimeMs)
				fmt.Printf("    Status: %s\n", variation.Response.ResponseStatus)

				// Check if this response has function call data
				if variation.Response.FunctionCallResponse != nil && len(variation.Response.FunctionCallResponse) > 0 {
					functionCallsFound = true
					fmt.Printf("    🔧 Function Call Response: %+v\n", variation.Response.FunctionCallResponse)
				}

				// Show first 100 characters of the response
				responseText := variation.Response.ResponseText
				if len(responseText) > 100 {
					responseText = responseText[:100] + "..."
				}
				fmt.Printf("    Response: %s\n", responseText)
			}

			if !functionCallsFound {
				fmt.Println("\n❌ No function calls detected in the most recent execution")
				fmt.Println("💡 Possible reasons:")
				fmt.Println("   • No function definitions are configured")
				fmt.Println("   • The AI didn't determine functions were needed for this prompt")
				fmt.Println("   • Functions may be configured but not working properly")
			}
		}
	}

	fmt.Println()
	fmt.Println("💡 To see function calls:")
	fmt.Println("1. Ensure you have function definitions in the database")
	fmt.Println("2. Make sure function calling is enabled in your execution runs")
	fmt.Println("3. Use prompts that clearly need external data (like weather)")
	fmt.Println("4. Check if your function definitions are properly configured")

	fmt.Println()
	fmt.Println("🔧 Quick function call test ideas:")
	fmt.Println("   • 'What's the current time in Tokyo?'")
	fmt.Println("   • 'Get me the latest stock price for AAPL'")
	fmt.Println("   • 'What's the weather forecast for San Francisco?'")
}
