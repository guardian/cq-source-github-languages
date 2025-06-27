.PHONY: test
test:
	go test -timeout 10m ./...

.PHONY: lint
lint:
	@golangci-lint run --timeout 10m

.PHONY: build
build:
	go build -o cq-source-github-languages .

.PHONY: build-test-auth
build-test-auth:
	go build -o test-auth ./cmd/test-auth

.PHONY: gen-docs
gen-docs: build
	rm -rf ./docs/tables/*
	mkdir -p ./docs/tables
	# Use cloudquery command from PATH to generate docs
	cloudquery tables docs/spec.yml --output-dir . --format markdown

# All gen targets
.PHONY: gen
gen: gen-docs