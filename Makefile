.PHONY: setup install-deps generate-db init-db run-tests clean frontend-setup frontend-install frontend-start frontend-ios frontend-android frontend-web frontend-build frontend-clean

# Setup the entire project (backend + frontend)
setup: install-deps generate-db frontend-setup

# Backend Setup Commands
# ======================

# Install Go dependencies
install-deps:
	go mod tidy
	go mod download

# Install sqlc for code generation
install-sqlc:
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Generate database code from SQL schema
generate-db: install-sqlc
	sqlc generate

# Initialize the database with schema
init-db:
	mysql -h $(DB_HOST) -u $(DB_USER) -p$(DB_PASSWORD) < sql/schema.sql

# Run database migrations
migrate:
	DB_URL=$(DB_URL) go run cmd/migrate/main.go

# Show migration status
migrate-status:
	DB_URL=$(DB_URL) go run cmd/migrate/main.go -status

# Build migration tool
build-migrate:
	go build -o bin/migrate ./cmd/migrate

# Generate protobuf Go code
generate-proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/gogent.proto

# Install protobuf tools
install-proto-tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Run backend tests
run-tests:
	go test ./...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Build the backend project
build:
	go build -o bin/gogent ./cmd/gogent

# Backend Demo Commands
# ====================

# Run auto-demo (detects configuration)
run:
	go run cmd/gogent/*.go

# Run simple demo with mock responses
run-simple:
	go run cmd/gogent/*.go --simple

# Run simple demo with real Gemini API (no database)
run-simple-api:
	go run cmd/gogent/*.go --simple-api

# Start HTTP server for frontend integration (alias for run-server)
run-api: run-server

# Start HTTP server for frontend integration
run-server:
	@echo "🧹 Cleaning up any existing processes..."
	@pkill -9 -f gogent 2>/dev/null || true
	@pkill -9 -f "go run.*gogent" 2>/dev/null || true
	@lsof -ti:8080 | xargs kill -9 2>/dev/null || true
	@sleep 2
	@echo "✅ Port 8080 is now available"
	@echo "🚀 Starting GoGent HTTP Server..."
	go run cmd/gogent/*.go --server

# Run real API demo with database logging (one-time execution)
run-api-demo:
	go run cmd/gogent/*.go --real-api

# Run the database version (requires DB setup)
run-db:
	go run cmd/gogent/*.go --database

# Show help
help:
	go run cmd/gogent/*.go --help

# Kill all server processes and free port 8080
kill-server:
	@echo "🧹 Stopping all GoGent processes..."
	@pkill -9 -f gogent 2>/dev/null || true
	@pkill -9 -f "go run.*gogent" 2>/dev/null || true
	@lsof -ti:8080 | xargs kill -9 2>/dev/null || true
	@echo "✅ All processes stopped and port 8080 freed"

# Frontend Setup Commands
# =======================

# Setup frontend project
frontend-setup: frontend-install
	@echo "🎯 Frontend Setup Complete!"
	@echo ""
	@echo "📱 Next steps:"
	@echo "1. Make sure the backend is running: 'make run-api'"
	@echo "2. Start the frontend: 'make frontend-start'"
	@echo "3. Run on iOS: 'make frontend-ios'"
	@echo "4. Run on Android: 'make frontend-android'"
	@echo ""

# Install frontend dependencies
frontend-install:
	@echo "📦 Installing frontend dependencies..."
	cd frontend && yarn install

# Frontend Development Commands
# ============================

# Start Expo development server
frontend-start:
	@echo "🚀 Starting Expo development server..."
	cd frontend && yarn start

# Run on iOS simulator
frontend-ios:
	@echo "📱 Starting iOS app..."
	cd frontend && yarn ios

# Run on Android simulator
frontend-android:
	@echo "🤖 Starting Android app..."
	cd frontend && yarn android

# Run on web browser
frontend-web:
	@echo "🌐 Starting web app..."
	cd frontend && yarn web

# Build frontend for production
frontend-build:
	@echo "🔨 Building frontend for production..."
	cd frontend && yarn build

# Frontend Maintenance Commands
# =============================

# Clean frontend dependencies and cache
frontend-clean:
	@echo "🧹 Cleaning frontend..."
	cd frontend && rm -rf node_modules yarn.lock
	cd frontend && rm -rf .expo

# Reinstall frontend dependencies
frontend-reinstall: frontend-clean frontend-install

# Type check frontend
frontend-typecheck:
	@echo "🔍 Type checking frontend..."
	cd frontend && yarn type-check

# Lint frontend code
frontend-lint:
	@echo "🔍 Linting frontend..."
	cd frontend && yarn lint

# Fix frontend linting issues
frontend-lint-fix:
	@echo "🔧 Fixing frontend lint issues..."
	cd frontend && yarn lint --fix

# Backend Maintenance Commands
# ============================

# Clean generated files
clean:
	rm -rf internal/db
	rm -f coverage.out coverage.html

# Format backend code
fmt:
	go fmt ./...

# Lint backend code
lint:
	golangci-lint run

# Full Project Commands
# ====================

# Clean everything (backend + frontend)
clean-all: clean frontend-clean
	@echo "🧹 Cleaned backend and frontend"

# Install all dependencies (backend + frontend)
install-all: install-deps frontend-install
	@echo "📦 Installed all dependencies"

# Development setup (first time setup)
dev-setup: setup
	cp config.example.env config.env
	@echo "🎯 GoGent Full Stack Development Setup Complete!"
	@echo ""
	@echo "📝 Backend Setup:"
	@echo "1. Edit config.env and add your GEMINI_API_KEY"
	@echo "2. Get your API key from: https://aistudio.google.com/app/apikey"
	@echo "3. For database features: set up MySQL and run 'make init-db'"
	@echo ""
	@echo "📱 Frontend Setup:"
	@echo "4. Backend must be running for full functionality"
	@echo "5. Start frontend with: 'make frontend-start'"
	@echo ""
	@echo "🚀 Quick start commands:"
	@echo "  make run-api                # Start backend with real API + database"
	@echo "  make frontend-start         # Start mobile app development server"
	@echo "  make frontend-ios           # Run on iOS simulator"
	@echo "  make frontend-android       # Run on Android simulator"
	@echo ""
	@echo "🔍 Development commands:"
	@echo "  make help                   # Backend help"
	@echo "  make frontend-typecheck     # Check TypeScript types"
	@echo "  make frontend-lint          # Lint frontend code"
	@echo "  make clean-all              # Clean everything"

# Status check - verify everything is set up correctly
status:
	@echo "🔍 GoGent Project Status Check"
	@echo "================================"
	@echo ""
	@echo "📊 Backend Status:"
	@go version 2>/dev/null && echo "✅ Go installed" || echo "❌ Go not found"
	@test -f config.env && echo "✅ Config file exists" || echo "❌ Config file missing (run 'make dev-setup')"
	@test -d internal/db && echo "✅ Database code generated" || echo "❌ Database code missing (run 'make generate-db')"
	@echo ""
	@echo "📱 Frontend Status:"
	@cd frontend && yarn --version 2>/dev/null && echo "✅ Yarn installed" || echo "❌ Yarn not found"
	@cd frontend && test -d node_modules && echo "✅ Frontend dependencies installed" || echo "❌ Frontend dependencies missing (run 'make frontend-install')"
	@cd frontend && test -f yarn.lock && echo "✅ Yarn lockfile exists" || echo "❌ No yarn lockfile"
	@echo ""
	@echo "🚀 Ready to start:"
	@echo "  Backend:  make run-api"
	@echo "  Frontend: make frontend-start"

# Show all available commands
commands:
	@echo "🛠️  GoGent Available Commands"
	@echo "============================"
	@echo ""
	@echo "📦 Setup & Installation:"
	@echo "  dev-setup              # First-time setup (backend + frontend)"
	@echo "  setup                  # Setup backend only"
	@echo "  frontend-setup         # Setup frontend only"
	@echo "  install-all            # Install all dependencies"
	@echo "  status                 # Check project status"
	@echo ""
	@echo "🔧 Backend Commands:"
	@echo "  run                    # Auto-detect demo mode"
	@echo "  run-simple             # Mock responses demo"
	@echo "  run-simple-api         # Real API demo (no DB)"
	@echo "  run-api                # Real API + database demo"
	@echo "  run-db                 # Database demo"
	@echo "  build                  # Build backend binary"
	@echo "  run-tests              # Run backend tests"
	@echo ""
	@echo "📱 Frontend Commands:"
	@echo "  frontend-start         # Start Expo dev server"
	@echo "  frontend-ios           # Run on iOS simulator"
	@echo "  frontend-android       # Run on Android simulator"
	@echo "  frontend-web           # Run in web browser"
	@echo "  frontend-build         # Build for production"
	@echo ""
	@echo "🧹 Maintenance:"
	@echo "  clean-all              # Clean everything"
	@echo "  frontend-clean         # Clean frontend only"
	@echo "  frontend-reinstall     # Reinstall frontend deps"
	@echo "  frontend-lint          # Lint frontend code"
	@echo "  frontend-typecheck     # Check TypeScript types" 