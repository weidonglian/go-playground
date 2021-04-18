.PHONY: all
all: start

# Tools
sync-vendor:
	@echo Download go.mod dependencies
	@go mod download && go mod tidy && go mod vendor

gen-large-file:
	@echo Gen large file
	@base64 /dev/urandom | head -c 100000000000 > file.txt

test:
	@echo Running test
	@go test ./...
