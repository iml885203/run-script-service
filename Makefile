.PHONY: test build clean coverage lint format

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

# CI 管道
ci: format lint test build