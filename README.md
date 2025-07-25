# GoGent - AI Multi-Variation Execution Platform with Interface Architecture

GoGent is a comprehensive Go platform that wraps AI APIs (starting with Google Gemini) with advanced multi-variation execution, database logging, and use case-specific implementations. It enables you to run the same AI prompt with different configurations, compare results, and implement domain-specific AI solutions like procurement management, legal analysis, and more.

## 🌟 Key Features

### Core Platform
- **📊 Multi-Variation Execution**: Run the same prompt with different configurations simultaneously
- **🗄️ Comprehensive Logging**: Every API call and response logged to MySQL database
- **🔍 Intelligent Comparison**: Automatically analyze and compare results across variations
- **⚙️ Flexible Configuration**: Support for different models, temperatures, system prompts, and more
- **🛡️ Type-Safe Operations**: Uses sqlc for generated type-safe SQL queries
- **🧩 Interface Architecture**: Clean, extensible interfaces for different use cases

### Use Case Implementations
- **🏢 AI Procurement Manager**: Complete solution for vendor evaluation, contract analysis, negotiation strategies
- **📋 Extensible Framework**: Easy to implement new domains (legal, content, risk assessment, etc.)
- **🏭 Factory Pattern**: Simple instantiation of different implementations
- **🔌 Plugin System**: Extensible architecture for custom functionality

### Advanced Features
- **📈 Analytics & Insights**: Performance metrics, cost analysis, model comparison
- **🔄 Multi-Provider Support**: Extensible to support different AI providers
- **🧪 A/B Testing**: Built-in experimentation framework for AI prompts
- **📝 Audit Trail**: Complete compliance and audit logging

## 🏗️ Architecture

```
GoGent Platform
├── 🎯 Interface Layer
│   ├── MultiVariationExecutor
│   ├── ExecutionLogger  
│   ├── ConfigurationManager
│   ├── ResultComparator
│   └── Use Case Interfaces
├── 🏢 Domain Implementations
│   ├── ProcurementManager
│   ├── LegalAnalyzer (extensible)
│   ├── ContentGenerator (extensible)
│   └── RiskAssessor (extensible)
├── 🏭 Factory & Adapters
│   ├── GoGentFactory
│   ├── ClientAdapter
│   └── MockFactory (testing)
├── 🗄️ Database Layer (MySQL + sqlc)
├── 🔧 Core Client (AI API Wrapper)
└── 📊 Analytics & Comparison Engine
```

## 🚀 Quick Start

### Prerequisites

- Go 1.21+
- MySQL 8.0+
- Google Gemini API key

### Installation

1. Clone and set up:
```bash
git clone <repository-url>
cd gogent
make dev-setup
```

2. Configure environment:
```bash
cp config.example.env config.env
# Edit config.env with your database and API credentials
```

3. Initialize database:
```bash
make init-db
```

4. Test different modes:
```bash
# Start HTTP server for frontend integration (persistent)
make run-api

# One-time demos:
make run-simple      # Mock demo (no external dependencies)
make run-simple-api  # Real API without database
make run-api-demo    # Real API + database demo (one-time)
```

## 🌐 HTTP Server Mode

The `make run-api` command starts a persistent HTTP server that provides REST API endpoints for the frontend mobile app.

### Server Endpoints

- `GET /health` - Health check with status information
- `POST /api/execute` - Multi-variation execution endpoint
- `GET /api/execution-runs` - Get execution history
- `GET /api/database/stats` - Database statistics
- `GET /api/database/tables` - List database tables

### Server Features

- **Mock Mode Support**: Add `X-Use-Mock: true` header for mock responses
- **Real API Integration**: Automatically uses real Gemini API when API key is configured
- **CORS Enabled**: Ready for frontend integration
- **Database Logging**: All executions logged to MySQL when available

### Example Usage

```bash
# Start the server
make run-server

# Test health endpoint
curl http://localhost:8080/health

# Test execution with mock data
curl -X POST http://localhost:8080/api/execute \
  -H "Content-Type: application/json" \
  -H "X-Use-Mock: true" \
  -d '{
    "execution_run_name": "test",
    "base_prompt": "Write a story about AI",
    "configurations": [{
      "id": "test-1",
      "variation_name": "creative",
      "model_name": "gemini-1.5-flash",
      "temperature": 0.8
    }]
  }'
```

### Frontend Integration

The HTTP server is designed to work with the React Native frontend:

1. **Start Backend**: `make run-server` (runs on localhost:8080)
2. **Start Frontend**: `make frontend-start` 
3. **Configure**: Set backend URL in mobile app settings
4. **Use**: Configure AI models and execute multi-variation prompts

## 💼 Procurement Management Usage

### Quick Procurement Manager Setup

```go
package main

import (
    "context"
    "log"
    "gogent/internal/factory"
)

func main() {
    // Create procurement manager with default config
    procurementManager, err := factory.QuickCreateProcurementManager(
        "your-gemini-api-key",
        "root:password@tcp(localhost:3306)/gogent?parseTime=true",
    )
    if err != nil {
        log.Fatal(err)
    }
    defer procurementManager.Close()
    
    // Now use the procurement manager...
}
```

### 1. Vendor Proposal Evaluation

```go
// Create RFP request
rfp := &types.RFPRequest{
    ID:          "rfp-2024-001",
    Title:       "Cloud Infrastructure Services",
    Description: "Seeking cloud infrastructure provider for enterprise workloads",
    Requirements: []string{
        "99.9% uptime SLA",
        "24/7 technical support",
        "Compliance with SOC 2 Type II",
        "Multi-region availability",
    },
    EvaluationCriteria: []types.EvaluationCriterion{
        {Name: "Technical Capability", Weight: 0.3, Type: "quality"},
        {Name: "Cost Effectiveness", Weight: 0.25, Type: "cost"},
        {Name: "Support Quality", Weight: 0.20, Type: "quality"},
    },
    Budget:   1000000.0, // $1M budget
    Timeline: 6 * 30 * 24 * time.Hour, // 6 months
    VendorProposals: []types.VendorProposal{
        {
            VendorID:    "vendor-aws",
            VendorName:  "Amazon Web Services",
            ProposalDoc: "AWS proposal with comprehensive cloud services...",
            Cost:        850000.0,
        },
        // ... more vendor proposals
    },
}

// AI-powered vendor evaluation with multiple analysis perspectives
result, err := procurementManager.EvaluateVendorProposals(ctx, rfp)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("🎯 Recommendation: %s\n", result.Recommendation)
fmt.Printf("📈 Vendor Scores: %d vendors evaluated\n", len(result.VendorScores))
fmt.Printf("⏱️ Analysis Time: %v\n", result.ExecutionResult.TotalTime)
```

### 2. Negotiation Strategy Generation

```go
// Create vendor profile
vendorProfile := &types.VendorProfile{
    ID:       "vendor-techcorp",
    Name:     "TechCorp Solutions",
    Industry: "Technology Services",
    Size:     "Mid-size (500-1000 employees)",
    Strengths: []string{
        "Strong technical expertise",
        "Proven delivery track record",
    },
    Weaknesses: []string{
        "Higher pricing compared to competitors",
        "Limited global presence",
    },
}

// Generate multiple negotiation strategies
strategies, err := procurementManager.GenerateNegotiationStrategies(ctx, vendorProfile)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("📋 Generated %d strategies\n", len(strategies.Strategies))
fmt.Printf("🏆 Recommended: %s\n", strategies.Recommendation)
```

### 3. Contract Terms Analysis

```go
// Define contract for analysis
contract := &types.ContractTerms{
    ContractID:   "contract-2024-sc-001",
    Title:        "Software Development Services Agreement",
    PaymentTerms: "Net 30 days, milestone-based payments",
    Value:        500000.0, // $500K
    Terms: []types.ContractTerm{
        {
            Name:        "Liability Cap",
            Description: "Vendor liability limited to 12 months of contract value",
            Type:        "legal",
            RiskLevel:   "medium",
        },
        // ... more terms
    },
}

// AI-powered contract analysis
analysis, err := procurementManager.AnalyzeContractTerms(ctx, contract)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("🎯 Overall Risk: %s\n", analysis.OverallRisk)
fmt.Printf("💡 Recommendations: %d\n", len(analysis.Recommendations))
```

### 4. Process Optimization

```go
// Define procurement requirements
requirements := &types.ProcurementRequirements{
    Category: "IT Equipment",
    Requirements: []string{
        "Bulk purchase of laptops and workstations",
        "Warranty and support services",
        "Asset management integration",
    },
    Budget:   2000000.0, // $2M
    Timeline: 4 * 30 * 24 * time.Hour, // 4 months
    Priorities: map[string]float64{
        "cost_optimization": 0.4,
        "quality_assurance": 0.3,
        "delivery_speed":    0.2,
    },
}

// AI-powered process optimization
optimization, err := procurementManager.OptimizeProcurementProcess(ctx, requirements)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("💰 Cost Savings: $%.2f\n", optimization.CostSavings)
fmt.Printf("⏰ Time Reduction: %v\n", optimization.TimeReduction)
fmt.Printf("🛡️ Risk Mitigation: %d strategies\n", len(optimization.RiskMitigation))
```

## 🔧 Basic Multi-Variation Usage

```go
package main

import (
    "context"
    "gogent/internal/factory"
    "gogent/internal/types"
)

func main() {
    // Create standard client
    client, err := factory.QuickCreateClient(
        "your-gemini-api-key",
        "your-db-url",
    )
    if err != nil {
        panic(err)
    }
    defer client.Close()

    // Create multi-variation request
    request := &types.MultiExecutionRequest{
        ExecutionRunName: "temperature-comparison",
        Description:      "Compare different temperature settings for creative writing",
        BasePrompt:       "Write a short story about a robot learning to paint",
        Configurations: []types.APIConfiguration{
            {
                VariationName: "analytical",
                ModelName:     "gemini-1.5-flash",
                SystemPrompt:  "You are a precise, analytical storyteller.",
                Temperature:   &[]float32{0.2}[0],
            },
            {
                VariationName: "creative",
                ModelName:     "gemini-1.5-flash", 
                SystemPrompt:  "You are a highly creative storyteller.",
                Temperature:   &[]float32{0.8}[0],
            },
            {
                VariationName: "experimental",
                ModelName:     "gemini-1.5-flash",
                SystemPrompt:  "You are an experimental storyteller who takes bold risks.",
                Temperature:   &[]float32{1.0}[0],
            },
        },
        ComparisonConfig: &types.ComparisonConfig{
            Enabled: true,
            Metrics: []string{"creativity", "coherence", "response_time"},
        },
    }

    // Execute all variations simultaneously
    result, err := client.ExecuteMultiVariation(context.Background(), request)
    if err != nil {
        panic(err)
    }

    // Analyze results
    fmt.Printf("✅ Success: %d/%d variations\n", 
        result.SuccessCount, 
        result.SuccessCount + result.ErrorCount)
    fmt.Printf("⏱️ Total time: %v\n", result.TotalTime)
    
    for _, variation := range result.Results {
        fmt.Printf("\n🎯 %s (temp: %.1f):\n", 
            variation.Configuration.VariationName,
            *variation.Configuration.Temperature)
        fmt.Printf("📝 %s\n", variation.Response.ResponseText[:100] + "...")
        fmt.Printf("⏱️ %dms\n", variation.Response.ResponseTimeMs)
    }
}
```

## 🔌 Extending for Other Use Cases

### Create Custom Implementation

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

func NewMyUseCaseExecutor(executor interfaces.MultiVariationExecutor, logger interfaces.ExecutionLogger, comparator interfaces.ResultComparator) interfaces.UseCaseSpecificExecutor {
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
            SystemPrompt:  "You are a conservative analyst...",
            Temperature:   &[]float32{0.2}[0],
        },
        // ... more configurations
    }
}

// Implement other required interface methods...
```

### Add to Factory

```go
// In internal/factory/gogent_factory.go
func (f *DefaultGoGentFactory) CreateCustomExecutor(useCaseName string, config *types.GeminiClientConfig, dbURL string) (interfaces.UseCaseSpecificExecutor, error) {
    switch useCaseName {
    case "my-custom-use-case":
        return f.createMyUseCaseExecutor(config, dbURL)
    case "legal-analysis":
        return f.createLegalAnalysisExecutor(config, dbURL)
    case "content-generation":
        return f.createContentGenerationExecutor(config, dbURL)
    // ... other cases
    }
}
```

## 📊 Database Analysis & Insights

Query your execution data for insights:

```sql
-- Find best performing configurations
SELECT 
    c.variation_name,
    AVG(r.response_time_ms) as avg_response_time,
    COUNT(*) as execution_count,
    AVG(JSON_EXTRACT(r.usage_metadata, '$.total_tokens')) as avg_tokens
FROM api_configurations c
JOIN api_requests req ON c.id = req.configuration_id  
JOIN api_responses r ON req.id = r.request_id
WHERE r.response_status = 'success'
GROUP BY c.variation_name
ORDER BY avg_response_time;

-- Cost analysis by use case
SELECT 
    er.name as execution_run,
    COUNT(*) as api_calls,
    AVG(r.response_time_ms) as avg_response_time,
    SUM(JSON_EXTRACT(r.usage_metadata, '$.total_tokens')) as total_tokens
FROM execution_runs er
JOIN api_requests req ON er.id = req.execution_run_id
JOIN api_responses r ON req.id = r.request_id
GROUP BY er.name;

-- Compare different models
SELECT 
    c.model_name,
    c.variation_name,
    COUNT(*) as usage_count,
    AVG(r.response_time_ms) as avg_time,
    SUM(CASE WHEN r.response_status = 'success' THEN 1 ELSE 0 END) / COUNT(*) as success_rate
FROM api_configurations c
JOIN api_requests req ON c.id = req.configuration_id
JOIN api_responses r ON req.id = r.request_id
GROUP BY c.model_name, c.variation_name
ORDER BY success_rate DESC, avg_time;
```

## 🗂️ Project Structure

```
gogent/
├── cmd/gogent/                    # Main application
│   ├── main.go                   # Entry point with demo selection
│   ├── simple_demo.go            # Mock demo (no DB/API needed)
│   ├── real_api_demo.go          # Real API + database demo
│   └── simple_real_api_demo.go   # Real API without database
├── internal/
│   ├── interfaces/               # 🎯 Core interface definitions
│   │   └── gogent.go            # All platform interfaces
│   ├── adapters/                # 🔌 Adapter layer
│   │   └── gogent_adapter.go    # Adapts existing client to interfaces
│   ├── factory/                 # 🏭 Factory pattern implementation
│   │   └── gogent_factory.go    # Creates different implementations
│   ├── db/                      # 🗄️ Generated database code (sqlc)
│   ├── gogent/                  # 🔧 Core client implementation
│   │   └── client.go            # Main GoGent client
│   ├── gemini/                  # 🤖 Gemini API integration
│   │   └── client.go            # Real Gemini API client
│   └── types/                   # 📋 Type definitions
│       └── types.go             # All data structures
├── examples/                    # 📚 Usage examples
│   ├── procurement/             # 🏢 Procurement manager implementation
│   │   └── procurement_manager.go
│   └── usage/                   # 💡 Complete usage examples
│       └── procurement_usage_example.go
├── sql/                         # 🗄️ Database layer
│   ├── schema.sql              # Database schema
│   └── queries/                # SQL queries for code generation
├── docs/                       # 📖 Documentation
│   └── interfaces_architecture.md # Complete architecture guide
├── config.example.env          # Example configuration
├── sqlc.yaml                   # sqlc configuration
├── Makefile                    # Build and development tasks
└── README.md
```

## ⚙️ Configuration

### Environment Variables

```bash
# Database Connection
DB_URL=root:password@tcp(localhost:3306)/gogent?parseTime=true
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your-password
DB_NAME=gogent

# AI API Configuration  
GEMINI_API_KEY=your-gemini-api-key
```

### Available Make Commands

```bash
# Development
make dev-setup          # Complete development setup (backend + frontend)
make run                # Auto-detect best demo mode
make run-server         # Start HTTP server for frontend integration
make run-api            # Start HTTP server (alias for run-server)
make run-simple         # Mock demo (no external dependencies)
make run-simple-api     # Real API without database
make run-api-demo       # Real API + database demo (one-time)

# Frontend
make frontend-start     # Start React Native development server
make frontend-ios       # Run on iOS simulator
make frontend-android   # Run on Android simulator
make frontend-install   # Install frontend dependencies

# Database
make init-db            # Initialize database with schema
make generate-db        # Regenerate sqlc code

# Build & Test
make build              # Build the application
make run-tests          # Run all tests
make fmt                # Format code  
make lint               # Run linter

# Help
make help               # Show available commands
make commands           # Show all available commands
```

## 🎯 Use Case Examples

### Procurement Management
- **Vendor Evaluation**: Multi-perspective analysis of RFP responses
- **Contract Analysis**: Risk assessment and compliance checking
- **Negotiation Strategies**: Generate multiple negotiation approaches
- **Process Optimization**: Identify efficiency improvements

### Legal Analysis (Extensible)
- **Contract Review**: Automated legal risk assessment
- **Compliance Checking**: Regulatory compliance validation
- **Document Summarization**: Legal document analysis

### Content Generation (Extensible)
- **Multi-Style Content**: Generate content with different tones/styles
- **A/B Testing**: Compare different content approaches
- **Quality Assessment**: Automated content quality scoring

### Risk Assessment (Extensible)
- **Financial Risk**: Multi-model financial analysis
- **Operational Risk**: Process and operational risk evaluation
- **Scenario Analysis**: Multiple risk scenario modeling

## 🧪 Advanced Features

### Multi-Variation Configuration

```go
// Temperature comparison
temperatures := []float32{0.1, 0.3, 0.5, 0.7, 0.9}
configs := make([]types.APIConfiguration, len(temperatures))
for i, temp := range temperatures {
    configs[i] = types.APIConfiguration{
        VariationName: fmt.Sprintf("temp-%.1f", temp),
        Temperature:   &temp,
        ModelName:     "gemini-1.5-flash",
    }
}

// Model comparison
models := []string{"gemini-1.5-flash", "gemini-1.5-pro"}
for _, model := range models {
    configs = append(configs, types.APIConfiguration{
        VariationName: fmt.Sprintf("model-%s", model),
        ModelName:     model,
        Temperature:   &[]float32{0.5}[0],
    })
}
```

### Custom Metrics and Comparison

```go
// Define custom comparison metrics
request.ComparisonConfig = &types.ComparisonConfig{
    Enabled: true,
    Metrics: []string{
        "response_time",
        "token_efficiency", 
        "creativity_score",
        "factual_accuracy",
        "cost_effectiveness",
    },
    CustomRules: []string{
        "prefer_faster_responses",
        "penalize_high_token_usage",
        "reward_creative_solutions",
    },
}
```

## 🔬 Testing & Development

### Unit Testing with Mocks

```go
package tests

import (
    "testing"
    "gogent/internal/factory" 
)

func TestProcurementManager(t *testing.T) {
    // Use mock factory for testing
    factory := factory.CreateMockFactory()
    
    pm, err := factory.CreateProcurementManager(mockConfig, mockDBURL)
    assert.NoError(t, err)
    
    // Test procurement functionality
    result, err := pm.EvaluateVendorProposals(ctx, sampleRFP)
    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

### Development Workflow

```bash
# 1. Set up development environment
make dev-setup

# 2. Start with simple demo (no external dependencies)
make run-simple

# 3. Test with real API (no database)
make run-simple-api

# 4. Full integration test
make run-api

# 5. Run tests
make run-tests

# 6. Build for production
make build
```

## 🚀 Production Deployment

### Docker Setup (Coming Soon)

```dockerfile
# Dockerfile example
FROM golang:1.21-alpine AS builder
COPY . /app
WORKDIR /app
RUN make build

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/bin/gogent /usr/local/bin/
CMD ["gogent"]
```

### Performance Considerations

- **Database Connection Pooling**: Configure appropriate pool sizes
- **Rate Limiting**: Implement API rate limiting for production use
- **Caching**: Add Redis caching for configuration and results
- **Monitoring**: Integrate with monitoring solutions (Prometheus, etc.)

## 🎯 Next Steps

1. **🏃‍♂️ Quick Start**: Run `make dev-setup && make run-api` to see it in action
2. **🏢 Try Procurement**: Run the procurement examples to see domain-specific AI
3. **🔧 Customize**: Implement your own use case following the interface patterns
4. **📈 Analyze**: Query the database to understand AI performance patterns
5. **🚀 Scale**: Deploy to production with your specific use case requirements

## 📚 Additional Resources

- [Interface Architecture Guide](docs/interfaces_architecture.md) - Complete implementation guide
- [Procurement Usage Examples](examples/usage/procurement_usage_example.go) - Comprehensive examples
- [Database Schema](sql/schema.sql) - Complete database structure
- [API Documentation](docs/api.md) - Detailed API reference (coming soon)

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Implement your changes following the interface patterns
4. Add comprehensive tests
5. Run `make run-tests && make lint`
6. Submit a pull request

## 📄 License

[License information here]

## 💬 Support & Community

- 🐛 **Issues**: [GitHub Issues](link-to-issues)
- 💬 **Discussions**: [GitHub Discussions](link-to-discussions)  
- 📧 **Contact**: [maintainer-email]
- 📖 **Docs**: [Documentation Site](link-to-docs)

---

**GoGent empowers you to build intelligent, data-driven AI applications with the confidence that comes from systematic experimentation and comprehensive logging.** Start with procurement management or implement your own domain-specific AI solution using our proven interface architecture. 