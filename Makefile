.PHONY: test build clean coverage lint format install-precommit setup-precommit tdd ci

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

# Pre-commit hooks (using pre-commit framework)
install-precommit:
	@echo "Installing pre-commit framework hooks..."
	@if command -v pre-commit >/dev/null 2>&1; then \
		pre-commit install; \
		echo "Pre-commit hooks installed successfully!"; \
	else \
		echo "❌ pre-commit not found. Install it with: pip install pre-commit"; \
		exit 1; \
	fi

setup-precommit: install-precommit
	@echo "Running pre-commit on all files..."
	@pre-commit run --all-files || echo "Some files were fixed. Please review and commit the changes."

# CI 管道
ci: format lint test build
