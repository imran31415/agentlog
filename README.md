# agentlog - AI Multi-Variation Execution Platform with Interface Architecture

live demo:  https://agentlog.scalebase.io

agentlog is a Go platform that wraps AI APIs (starting with Google Gemini) with multi-variation execution, database logging, and use case-specific implementations. It enables you to run the same AI prompt with different configurations, compare results, and implement domain-specific AI solutions like procurement management, legal analysis, and more.

<img width="627" height="717" alt="image" src="https://github.com/user-attachments/assets/93715c59-6cc6-4aaa-98d9-467dcb5c8647" />


https://github.com/user-attachments/assets/43184a26-4a6f-4f93-a655-f58f7badaddc




## 🚀 Quick Start

Get AgentLog running in 2 steps:

### 1. Start the Backend Server 

Update the config.env file with your local db and gemini api keys and then:

```bash
make run-server
```
This starts the agentlog backend on `localhost:8080` with REST API endpoints for multi-variation AI execution, function calling, and database logging.

```
2025/07/24 20:36:54 Go SDK disabled - using REST API for all Gemini calls
🚀 agentlog HTTP Server starting on port 8080
📡 Health check: http://localhost:8080/health
🔧 API endpoints:
   POST /api/execute - Multi-variation execution
   GET  /api/execution-runs - Execution history
   GET  /api/configurations - List API configurations
   GET  /api/functions - List function definitions
   POST /api/functions - Create function definition
   GET  /api/functions/{id} - Get function by ID
   PUT  /api/functions/{id} - Update function
   DELETE /api/functions/{id} - Delete function
   POST /api/functions/test/{id} - Test function execution
   GET  /api/database/stats - Database statistics
   GET  /api/database/tables - Database tables
💡 Use X-Use-Mock: true header for mock responses
🔑 Set GEMINI_API_KEY in config.env for real API calls

2025/07/24 20:36:57 📋 Listing function definitions from database
2025/07/24 20:36:57 ✅ Successfully loaded 2 function definitions from database


```

### 2. Start the Frontend App
```bash
make frontend-start
```
This launches the React Native development server with the mobile interface for configuring AI models, executing prompts, and viewing results.

<img width="490" height="672" alt="image" src="https://github.com/user-attachments/assets/4a31e3d8-21e6-46e3-b041-1b272cee5f76" />


### That's it! 
You now have an AI experimentation platform running locally. The frontend will connect to the backend automatically.

<img width="482" height="276" alt="image" src="https://github.com/user-attachments/assets/285a93b3-ddae-4831-8583-0dd53bdff699" />

<img width="480" height="437" alt="image" src="https://github.com/user-attachments/assets/7c5cac7d-9365-499a-a655-680206bf7f20" />


## 📋 Overview

### The Problem
When building AI agents with Gemini (or any LLM), you need **visibility and control** for debugging and optimization. Most implementations lack:

1. **Traceability & Monitoring** - No logging of AI interactions
2. **Configuration Flexibility** - Can't easily adjust temperature, tokens, system prompts
3. **Parallel Testing** - No way to run multiple model variations simultaneously  
4. **Centralized Management** - No unified platform to track and compare executions

### The Solution: AgentLog Platform

AgentLog is an **AI experimentation platform** that gives you control over your Gemini agents:

#### 🔧 **Configuration Control**
- Configure Gemini API keys and any custom function API keys
- Adjust model parameters: temperature, max tokens, top-P, top-K
- Customize system prompts and context for different use cases
- Set up parallel executions with variation testing

#### 📊 **Execution & Tracking**  
- Run multiple AI model configurations simultaneously
- Compare results side-by-side with analysis
- Database logging of every API call and response
- Track function calls, execution times, and model performance

#### 🛠️ **Function System**
- Add custom functions for external API integrations
- Built-in support for weather APIs, Neo4j graph databases
- Create domain-specific AI workflows (procurement, legal analysis, etc.)
- Function call tracing and debugging capabilities

#### 📱 **Frontend Interface**
- React Native mobile app for platform management
- Real-time execution monitoring with loading states
- Historical analysis with searchable execution logs
- Database inspection tools for debugging

#### 🏢 **Features**
- MySQL database with audit trails
- RESTful API architecture for integration
- Multi-variation execution engine
- Production deployment capabilities

**Result**: Instead of blind AI development, you get a platform with the tools needed to build, test, and optimize agents with visibility into their behavior.

## 🌟 Key Features

### Core Platform
- **📊 Multi-Variation Execution**: Run the same prompt with different configurations simultaneously
- **🗄️ Database Logging**: Every API call and response logged to MySQL database
- **🔍 Result Comparison**: Analyze and compare results across variations
- **⚙️ Configuration Support**: Support for different models, temperatures, system prompts, and more
- **🛡️ Type-Safe Operations**: Uses sqlc for generated type-safe SQL queries
- **🧩 Interface Architecture**: Clean, extensible interfaces for different use cases

### Use Case Implementations
- **🏢 AI Procurement Manager**: Solution for vendor evaluation, contract analysis, negotiation strategies
- **📋 Framework**: Easy to implement new domains (legal, content, risk assessment, etc.)
- **🏭 Factory Pattern**: Simple instantiation of different implementations
- **🔌 Plugin System**: Extensible architecture for custom functionality

### Features
- **📈 Analytics & Insights**: Performance metrics, cost analysis, model comparison
- **🔄 Multi-Provider Support**: Extensible to support different AI providers
- **🧪 A/B Testing**: Built-in experimentation framework for AI prompts
- **📝 Audit Trail**: Compliance and audit logging

## 🏗️ Architecture

```
agentlog Platform
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
│   ├── agentlogFactory
│   ├── ClientAdapter
│   └── MockFactory (testing)
├── 🗄️ Database Layer (MySQL + sqlc)
├── 🔧 Core Client (AI API Wrapper)
└── 📊 Analytics & Comparison Engine
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
- **Real API Integration**: Uses real Gemini API when API key is configured
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

