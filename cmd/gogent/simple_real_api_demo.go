package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"gogent/internal/gemini"
	"gogent/internal/types"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func runSimpleRealApiDemo() {
	fmt.Println("🚀 GoGent Simple Real API Demo")
	fmt.Println("===============================")
	fmt.Println("📡 Using real Gemini API without database logging")
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
		return
	}

	fmt.Printf("🔑 Using API Key: %s...%s\n", apiKey[:8], apiKey[len(apiKey)-4:])
	fmt.Println()

	// Create Gemini client
	ctx := context.Background()
	geminiClient, err := gemini.NewGeminiClient(ctx, apiKey)
	if err != nil {
		log.Fatalf("Failed to create Gemini client: %v", err)
	}
	defer geminiClient.Close()

	// Create test configurations with different temperatures
	temp1 := float32(0.2)
	temp2 := float32(0.7)
	temp3 := float32(1.1)
	maxTokens := int32(150)

	configurations := []types.APIConfiguration{
		{
			ID:            uuid.New().String(),
			VariationName: "conservative-precise",
			ModelName:     "gemini-1.5-flash",
			SystemPrompt:  "You are a precise, analytical storyteller. Write concise, well-structured narratives.",
			Temperature:   &temp1,
			MaxTokens:     &maxTokens,
		},
		{
			ID:            uuid.New().String(),
			VariationName: "balanced-creative",
			ModelName:     "gemini-1.5-flash",
			SystemPrompt:  "You are a creative storyteller who balances imagination with clarity.",
			Temperature:   &temp2,
			MaxTokens:     &maxTokens,
		},
		{
			ID:            uuid.New().String(),
			VariationName: "experimental-wild",
			ModelName:     "gemini-1.5-flash",
			SystemPrompt:  "You are a wildly imaginative storyteller who uses unexpected metaphors and bold creative risks.",
			Temperature:   &temp3,
			MaxTokens:     &maxTokens,
		},
	}

	prompt := "Write a 2-sentence story about a robot who discovers emotions while painting sunsets."
	context := "This is a creative writing test to explore how temperature affects storytelling creativity."

	fmt.Println("📝 Testing Prompt:")
	fmt.Printf("   %s\n", prompt)
	fmt.Println()
	fmt.Println("🎯 Executing 3 variations with different temperature settings...")
	fmt.Println()

	var results []VariationResult
	totalStartTime := time.Now()

	for i, config := range configurations {
		fmt.Printf("⚙️  Variation %d: %s\n", i+1, config.VariationName)
		fmt.Printf("   🌡️  Temperature: %.1f\n", *config.Temperature)
		fmt.Printf("   🤖 Model: %s\n", config.ModelName)
		fmt.Printf("   📋 System: %s\n", config.SystemPrompt)

		startTime := time.Now()

		// Make real API call
		response, err := geminiClient.GenerateContent(ctx, &config, prompt, context)

		duration := time.Since(startTime)

		if err != nil {
			fmt.Printf("   ❌ Error: %v\n", err)
			results = append(results, VariationResult{
				Config:   config,
				Error:    err,
				Duration: duration,
				Success:  false,
			})
		} else {
			fmt.Printf("   ✅ Success (%dms)\n", response.ResponseTimeMs)
			fmt.Printf("   📄 Response: %s\n", response.ResponseText)
			if response.UsageMetadata != nil {
				fmt.Printf("   📊 Tokens: %v\n", response.UsageMetadata)
			}

			results = append(results, VariationResult{
				Config:   config,
				Response: response,
				Duration: duration,
				Success:  true,
			})
		}
		fmt.Println()
	}

	totalDuration := time.Since(totalStartTime)

	// Display summary
	fmt.Println("📊 EXECUTION SUMMARY")
	fmt.Println("====================")
	fmt.Printf("⏱️  Total Time: %v\n", totalDuration)

	successCount := 0
	var fastestResult *VariationResult

	for i := range results {
		if results[i].Success {
			successCount++
			if fastestResult == nil || results[i].Response.ResponseTimeMs < fastestResult.Response.ResponseTimeMs {
				fastestResult = &results[i]
			}
		}
	}

	fmt.Printf("✅ Successful: %d/%d\n", successCount, len(results))
	fmt.Printf("❌ Failed: %d/%d\n", len(results)-successCount, len(results))

	if fastestResult != nil {
		fmt.Println()
		fmt.Println("🏆 PERFORMANCE WINNER")
		fmt.Println("======================")
		fmt.Printf("🥇 Fastest: %s (%dms)\n", fastestResult.Config.VariationName, fastestResult.Response.ResponseTimeMs)
		fmt.Printf("🌡️  Temperature: %.1f\n", *fastestResult.Config.Temperature)
	}

	fmt.Println()
	fmt.Println("🎯 CREATIVITY ANALYSIS")
	fmt.Println("======================")
	for _, result := range results {
		if result.Success {
			creativity := estimateCreativity(result.Response.ResponseText)
			fmt.Printf("• %s: %s creativity\n", result.Config.VariationName, creativity)
		}
	}

	fmt.Println()
	fmt.Println("✨ Real API demo completed!")
	fmt.Println("💡 Try different prompts or temperatures to see how responses vary")
}

// Helper types for simple demo
type VariationResult struct {
	Config   types.APIConfiguration
	Response *types.APIResponse
	Error    error
	Duration time.Duration
	Success  bool
}

func estimateCreativity(text string) string {
	// Simple heuristic based on text characteristics
	if len(text) < 50 {
		return "Low"
	}

	creativityWords := []string{"magical", "mysterious", "vivid", "imagination", "dreams", "wonder", "ethereal", "whispered", "danced", "shimmered"}
	count := 0

	for _, word := range creativityWords {
		if containsWord(text, word) {
			count++
		}
	}

	if count >= 3 {
		return "High"
	} else if count >= 1 {
		return "Medium"
	}
	return "Low"
}

func containsWord(text, word string) bool {
	// Simple contains check (case-insensitive would be better)
	return len(text) >= len(word) &&
		(text == word ||
			(len(text) > len(word) &&
				(text[:len(word)] == word ||
					text[len(text)-len(word):] == word)))
}
