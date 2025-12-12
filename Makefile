# Makefile for the vertex-ingestor project

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=rag-vec-search
BINARY_UNIX=$(BINARY_NAME)

# All target is the default target
all: build

# Build the binary
build:
	$(GOBUILD) -o $(BINARY_NAME) ./main.go

# Run the application
run:
	$(GOBUILD) -o $(BINARY_NAME) ./main.go
	./$(BINARY_NAME) $(ARGS)

# Clean the binary
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f mem.out

# Run tests
test:
	$(GOTEST) -v ./...

# Get dependencies
deps:
	$(GOCMD) mod tidy
	$(GOCMD) install github.com/vektra/mockery/v3@v3.6.1
	$(GOCMD) install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

generate:
	sqlc generate
	mockery

# Run benchmarks
benchmark:
	$(GOTEST) -bench=. -benchmem -memprofile mem.out ./internal/pipeline

# Run benchmarks
memprofile:
	$(GOCMD) tool pprof -http=:8080 mem.out 

.PHONY: all build run clean test deps ingest benchmark
