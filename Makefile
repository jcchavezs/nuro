.PHONY: test
test:
	@go test ./...

.PHONY: dist
dist:
	@goreleaser release --snapshot --clean

.PHONY: generate
generate: generate-readme

generate-readme:
	@echo "<!-- Generated file by make generate-readme. DO NOT EDIT. -->" > README.md
	@HELP=$$(go run main.go --help) \
	envsubst '$$HELP' < README.md.tmpl >> README.md
