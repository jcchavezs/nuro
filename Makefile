.PHONY: test
test:
	@go test ./...

.PHONY: dist
dist:
	@goreleaser release --snapshot --clean
