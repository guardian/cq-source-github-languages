.PHONY: test lint gen-docs build clean

# Run tests
test:
	@go test -timeout 10m ./...

# Run linter
lint:
	@golangci-lint run --timeout 10m

# Generate documentation
gen-docs:
	@go run main.go doc docs

# Build the plugin
build:
	@go build -o cq-source-github-languages .

# Clean build artifacts
clean:
	@rm -f cq-source-github-languages