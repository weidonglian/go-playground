.PHONY: all
all: start

# Tools
sync-vendor:
	@echo Download go.mod dependencies
	@go mod download && go mod tidy && go mod vendor

test:
	@echo Running test
	@go test ./...
