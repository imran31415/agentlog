# agentlog - AI Multi-Variation Execution Platform with Interface Architecture

live demo:  https://agentlog.scalebase.io

agentlog is a Go platform that wraps AI APIs (starting with Google Gemini) with multi-variation execution, database logging, and use case-specific implementations. It enables you to run the same AI prompt with different configurations, compare results, and implement domain-specific AI solutions like procurement management, legal analysis, and more.



## Key Features

* **Multi-Variation Execution:** Run the same prompt with different configurations simultaneously
* **Database Logging:** Every API call and response logged to MySQL database
* **Result Comparison:** Analyze and compare results across variations
* **Configuration Support:** Support for different models, temperatures, system prompts, and more
* **Type-Safe Operations:** Uses sqlc for generated type-safe SQL queries
* **Interface Architecture:** Clean, extensible interfaces for different use cases

## Architecture

```
agentlog Platform
├── Interface Layer
│   ├── MultiVariationExecutor
│   ├── ExecutionLogger  
│   ├── ConfigurationManager
│   ├── ResultComparator
│   └── Use Case Interfaces
├── Domain Implementations
│   ├── ProcurementManager
│   ├── LegalAnalyzer (extensible)
│   ├── ContentGenerator (extensible)
│   └── RiskAssessor (extensible)
├── Factory & Adapters
│   ├── agentlogFactory
│   ├── ClientAdapter
│   └── MockFactory (testing)
├── Database Layer (MySQL + sqlc)
├── Core Client (AI API Wrapper)
└── Analytics & Comparison Engine
```

## Usage

See the [Quick Start](path/to/quickstart) for details on setting up and running agentlog.

## Contributing

Contributions are welcome! See the [Contributing Guide](path/to/contributing) for details.

## License

[License information here]
