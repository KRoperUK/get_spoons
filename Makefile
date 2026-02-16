.PHONY: build run clean all test lint fmt vet

BINARY_NAME=get_spoons
CLI_PATH=./cmd/get_spoons
CSV_OUTPUT=latest_list.csv

all: build

VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || cat .release-please-manifest.json 2>/dev/null || echo "v0.0.0")

build:
	go build -ldflags="-X main.Version=$(VERSION)" -o $(BINARY_NAME) $(CLI_PATH)/main.go

run:
	@if [ -z "$(JDW_TOKEN)" ]; then \
		echo "Error: JDW_TOKEN environment variable is not set."; \
		echo "Use: JDW_TOKEN=your_token make run"; \
		exit 1; \
	fi
	go run $(CLI_PATH)/main.go --output $(CSV_OUTPUT)

test:
	go test ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

test-live:
	JDW_LIVE_TESTS=true JDW_TOKEN="$(JDW_TOKEN)" go test -v ./jdw/...

lint:
	golangci-lint run

fmt:
	go fmt ./...

vet:
	go vet ./...

clean:
	rm -f $(BINARY_NAME) $(CSV_OUTPUT) coverage.out
