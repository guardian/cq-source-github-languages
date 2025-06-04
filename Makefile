.PHONY: test
test:
	go test -timeout 10m ./...

.PHONY: lint
lint:
	@golangci-lint run --timeout 10m

.PHONY: build
build:
	go build -o cq-source-github-languages .

.PHONY: gen-docs
gen-docs: build
	rm -rf ./docs/tables/*
	mkdir -p ./docs/tables
	# Use cloudquery tables command with a simple spec file - markdown is no longer supported
	cloudquery tables docs/spec.yml --output-dir . --format markdown

# All gen targets
.PHONY: gen
gen: gen-docs