package main

import (
	"fmt"
	"os"
)

func main() {
	// Print usage info
	if len(os.Args) > 1 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		printUsage()
		return
	}

	// Check command line arguments for production modes
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--server":
			runServer()
		case "--grpc-server":
			runGRPCServer()
		case "--grpc-gateway":
			runGRPCGateway()
		case "--both":
			go runGRPCServer() // Start gRPC server in background
			runGRPCGateway()   // Start HTTP gateway in foreground
		default:
			fmt.Printf("Unknown option: %s\n", os.Args[1])
			printUsage()
		}
	} else {
		// Default behavior - start REST server (mobile-friendly)
		runServer()
	}
}

func printUsage() {
	fmt.Println("🎯 GoGent - Multi-Variation AI Execution Engine")
	fmt.Println("===============================================")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  go run cmd/gogent/*.go [option]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  (no args)      Start REST HTTP server (mobile-friendly, default)")
	fmt.Println("  --server       Start REST HTTP server for frontend integration")
	fmt.Println("  --grpc-server  Start native gRPC server (port 9090)")
	fmt.Println("  --grpc-gateway Start HTTP-to-gRPC gateway (port 8081)")
	fmt.Println("  --both         Start both gRPC server + HTTP gateway")
	fmt.Println("  --help, -h     Show this help message")
	fmt.Println()
	fmt.Println("Setup:")
	fmt.Println("  1. Copy config.example.env to config.env")
	fmt.Println("  2. Add your GEMINI_API_KEY to config.env")
	fmt.Println("  3. Set up MySQL and configure DB_URL in config.env")
	fmt.Println("  4. Run database migrations with 'make init-db'")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  make run                                 # Start REST server (recommended)")
	fmt.Println("  go run cmd/gogent/*.go --server          # Start REST server")
	fmt.Println("  go run cmd/gogent/*.go --grpc-server     # Start gRPC server")
	fmt.Println("  go run cmd/gogent/*.go --grpc-gateway    # Start HTTP-to-gRPC gateway")
	fmt.Println("  go run cmd/gogent/*.go --both            # Start both gRPC + gateway")
	fmt.Println()
}
