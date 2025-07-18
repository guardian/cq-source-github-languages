.PHONY: test test-clean lint build gen-docs clean

# Run tests
test:
	@go test -timeout 10m ./...

# Run tests with cleared cache
test-clean:
	@go clean -testcache
	@go test -timeout 10m ./...

# Run linter
lint:
	@golangci-lint run --timeout 10m

# Build the plugin
build:
	@go build -o cq-source-github-languages .

# Generate documentation
gen-docs: build
	rm -rf ./docs/tables/*
	mkdir -p ./docs/tables
	# Use cloudquery command from PATH to generate docs
	cloudquery tables docs/spec.yml --output-dir . --format markdown

# Clean build artifacts
clean:
	@rm -f cq-source-github-languages