package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// Print usage info
	if len(os.Args) > 1 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		printUsage()
		return
	}

	// Check command line arguments for demo mode
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--server":
			runServer()
		case "--real-api":
			runRealApiDemo()
		case "--simple-api":
			runSimpleRealApiDemo()
		case "--simple":
			runSimpleDemo()
		case "--database":
			runFullDatabaseDemo()
		default:
			fmt.Printf("Unknown option: %s\n", os.Args[1])
			printUsage()
		}
	} else {
		// Default behavior - try simple API, fallback to mock demo
		runAutoDemo()
	}
}

func runAutoDemo() {
	fmt.Println("🎯 GoGent Auto Demo - Detecting Configuration")
	fmt.Println("===========================================")
	fmt.Println()

	// Load environment variables
	if err := godotenv.Load("config.env"); err != nil {
		fmt.Println("📝 No config.env found, running simple demo...")
		runSimpleDemo()
		return
	}

	// Check if we have a valid API key
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey != "" && apiKey != "your_gemini_api_key_here" {
		fmt.Println("🔑 API key detected, running simple real API demo...")
		runSimpleRealApiDemo()
	} else {
		fmt.Println("🎭 No API key configured, running simple demo...")
		runSimpleDemo()
	}
}

func runFullDatabaseDemo() {
	// Load environment variables
	if err := godotenv.Load("config.env"); err != nil {
		log.Printf("Warning: could not load config.env file: %v", err)
	}

	// Get database URL from environment
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL environment variable is required")
	}

	// Get Gemini API key from environment
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY environment variable is required")
	}

	fmt.Println("🚧 Full database demo is not yet implemented due to type compatibility issues.")
	fmt.Println("📝 TODO: Fix database integration with generated sqlc types")
	fmt.Println("🎯 Running real API demo instead...")
	fmt.Println()

	runRealApiDemo()
}

func printUsage() {
	fmt.Println("🎯 GoGent - Multi-Variation AI Execution Engine")
	fmt.Println("===============================================")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  go run cmd/gogent/*.go [option]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  (no args)      Auto-detect: Use real API if configured, otherwise mock demo")
	fmt.Println("  --server       Start HTTP server for frontend integration")
	fmt.Println("  --simple       Run simple demo with mock responses")
	fmt.Println("  --simple-api   Run simple demo with real Gemini API (no database)")
	fmt.Println("  --real-api     Run demo with real Gemini API + database logging")
	fmt.Println("  --database     Run with full database integration (requires DB setup)")
	fmt.Println("  --help, -h     Show this help message")
	fmt.Println()
	fmt.Println("Setup:")
	fmt.Println("  1. Copy config.example.env to config.env")
	fmt.Println("  2. Add your GEMINI_API_KEY to config.env")
	fmt.Println("  3. For database features: set up MySQL and run 'make init-db'")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  make run                         # Auto-detect demo mode")
	fmt.Println("  go run cmd/gogent/*.go --simple         # Mock responses")
	fmt.Println("  go run cmd/gogent/*.go --simple-api     # Real API, no database")
	fmt.Println("  go run cmd/gogent/*.go --real-api       # Real API + database")
	fmt.Println()
}
