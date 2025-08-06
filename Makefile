.PHONY: test build clean coverage lint format tdd ci build-frontend embed-frontend build-all test-frontend

# 測試相關
test:
	go test ./... -v

test-watch:
	find . -name "*.go" | entr -c go test ./... -v

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

embed-frontend: build-frontend
	@echo "Frontend assets embedded via Go embed.FS"
	@echo "Re-run 'make build' to include updated frontend assets"

build-all: build-frontend build
	@echo "Complete build process finished"
	@echo "Backend binary: ./run-script-service"
	@echo "Frontend assets: embedded in binary"

# CI 管道
ci: format lint test test-frontend build-all
