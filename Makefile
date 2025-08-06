.PHONY: test build clean coverage lint format tdd ci build-frontend build-all test-frontend quickstart

# 測試相關
test:
	go test ./... -v

test-watch:
	@echo "File watching requires 'entr' tool. Install with: apt install entr (Ubuntu) or brew install entr (macOS)"
	@if command -v entr >/dev/null 2>&1; then \
		find . -name "*.go" | entr -c go test ./... -v; \
	else \
		echo "entr not found. Running tests once..."; \
		go test ./... -v; \
	fi

coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# 建構相關
build:
	go build -o run-script-service main.go

clean:
	rm -f run-script-service coverage.out coverage.html

# 代碼品質
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, running go vet instead"; \
		go vet ./...; \
	fi

format:
	go fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	else \
		echo "goimports not installed, skipping import formatting"; \
	fi

# TDD 循環
tdd: test-watch


# 前端建構相關 (Plan 09)
build-frontend:
	@echo "Building Vue.js + TypeScript frontend..."
	@if [ ! -d "web/frontend/node_modules" ]; then \
		echo "Installing frontend dependencies..."; \
		cd web/frontend && pnpm install; \
	fi
	@echo "Running frontend build..."
	cd web/frontend && pnpm build
	@echo "Frontend build completed successfully"

test-frontend:
	@echo "Running frontend unit tests..."
	cd web/frontend && pnpm run test:unit -- --run --reporter=verbose
	@echo "Frontend tests completed"


build-all: build-frontend build
	@echo "Complete build process finished"
	@echo "Backend binary: ./run-script-service"
	@echo "Frontend assets: embedded in binary"

# 快速設置目標
quickstart:
	@echo "🚀 Quick Start Setup for Run Script Service"
	@echo "Installing frontend dependencies..."
	@if [ ! -d "web/frontend/node_modules" ]; then \
		cd web/frontend && pnpm install; \
	else \
		echo "Frontend dependencies already installed"; \
	fi
	@echo "Building complete service..."
	@$(MAKE) build-all
	@echo ""
	@echo "✅ Setup complete! Next steps:"
	@echo "1. Add a script: ./run-script-service add-script --name=test --path=./test.sh --interval=30s"
	@echo "2. Start service: ./run-script-service daemon start"
	@echo "3. Check status: ./run-script-service daemon status"
	@echo "4. Web interface: http://localhost:8080"

# CI 管道
ci: format lint test test-frontend build-all
