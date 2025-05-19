.PHONY: test
test:
	@go test ./...

.PHONY: dist
dist:
	@goreleaser release --snapshot --clean

.PHONY: generate
generate: generate-readme

.PHONY: generate-readme
generate-readme:
	@echo "<!-- Generated file by make generate-readme. DO NOT EDIT. -->" > README.md
	@HELP=$$(go run main.go --help) \
	envsubst '$$HELP' < README.md.tmpl >> README.md

.PHONY: install-tools
install-tools: ## Install tools
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6

check-tool-%:
	@which $* > /dev/null || (echo "Install $* with 'make install-tools'"; exit 1 )

.PHONY: lint
lint: check-tool-golangci-lint
	@golangci-lint run ./...
