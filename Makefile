.PHONY: all
all:

# Tools
sync-vendor:
	@echo Download go.mod dependencies
	@go mod download && go mod tidy && go mod vendor

.PHONY: tools
tools:
	@echo Installing tools from tools.go
	@cat tools/tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % go install %

.PHONY: mocks
mocks:
	@echo Generating mocks using 'mockery'
	@mockery --name=Cache --recursive --keeptree --dir algorithms

gen-large-file:
	@echo Gen large file
	@base64 /dev/urandom | head -c 100000000000 > file.txt

test:
	@echo Running test
	@go test ./...

benchmark:
	@echo Running benchmark
	@go test -bench=. ./...