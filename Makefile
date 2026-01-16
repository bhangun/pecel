.PHONY: build clean install test lint release demo cross-compile help

BINARY_NAME=pecel
BINARY_DIR=bin
VERSION=$(shell git describe --tags --always --dirty)
GO_FILES=$(shell find . -name "*.go" -type f)

# Colors
GREEN=\033[0;32m
CYAN=\033[0;36m
YELLOW=\033[1;33m
NC=\033[0m # No Color

help:
	@echo "$(CYAN)Pecel CLI - Build Commands$(NC)"
	@echo ""
	@echo "$(YELLOW)Usage:$(NC)"
	@echo "  make [target]"
	@echo ""
	@echo "$(YELLOW)Targets:$(NC)"
	@echo "  $(GREEN)build$(NC)        - Build for current platform"
	@echo "  $(GREEN)build-all$(NC)    - Build for all major platforms"
	@echo "  $(GREEN)install$(NC)      - Install to system"
	@echo "  $(GREEN)test$(NC)         - Run tests"
	@echo "  $(GREEN)test-coverage$(NC)- Run tests with coverage"
	@echo "  $(GREEN)lint$(NC)         - Run linter"
	@echo "  $(GREEN)clean$(NC)        - Clean build artifacts"
	@echo "  $(GREEN)demo$(NC)         - Run demo with sample options"
	@echo "  $(GREEN)cross-compile$(NC)- Build for all platforms (extended)"
	@echo "  $(GREEN)release$(NC)      - Prepare release binaries"
	@echo "  $(GREEN)tag$(NC)          - Create and push git tag"
	@echo "  $(GREEN)checksums$(NC)    - Generate SHA256 checksums"

build:
	@echo "$(CYAN)Building for current platform...$(NC)"
	@mkdir -p $(BINARY_DIR)
	go build -ldflags="-s -w -X main.version=$(VERSION)" -o $(BINARY_DIR)/$(BINARY_NAME) ./cmd/main

build-all:
	@echo "$(CYAN)Building for all platforms...$(NC)"
	@mkdir -p dist/linux-amd64 dist/linux-arm64 dist/darwin-amd64 dist/darwin-arm64 dist/windows-amd64 dist/windows-arm64
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.version=$(VERSION)" -o dist/linux-amd64/$(BINARY_NAME) ./cmd/main
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w -X main.version=$(VERSION)" -o dist/linux-arm64/$(BINARY_NAME) ./cmd/main
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w -X main.version=$(VERSION)" -o dist/darwin-amd64/$(BINARY_NAME) ./cmd/main
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w -X main.version=$(VERSION)" -o dist/darwin-arm64/$(BINARY_NAME) ./cmd/main
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -X main.version=$(VERSION)" -o dist/windows-amd64/$(BINARY_NAME).exe ./cmd/main
	GOOS=windows GOARCH=arm64 go build -ldflags="-s -w -X main.version=$(VERSION)" -o dist/windows-arm64/$(BINARY_NAME).exe ./cmd/main

clean:
	@echo "$(CYAN)Cleaning build artifacts...$(NC)"
	rm -rf dist/ $(BINARY_DIR)/$(BINARY_NAME) $(BINARY_DIR)/$(BINARY_NAME).exe coverage.out combined.txt output.json output.xml $(BINARY_DIR)

test:
	@echo "$(CYAN)Running tests...$(NC)"
	go test -v ./...

test-coverage:
	@echo "$(CYAN)Running tests with coverage...$(NC)"
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

lint:
	@echo "$(CYAN)Running linter...$(NC)"
	golangci-lint run

install: build
	@echo "$(CYAN)Installing to system...$(NC)"
	sudo cp $(BINARY_NAME) /usr/local/bin/
	@echo "$(GREEN)✓ Installed successfully!$(NC)"

release: clean build-all checksums
	@echo "$(CYAN)Release v$(VERSION) ready in dist/$(NC)"

tag:
	@echo "$(CYAN)Creating tag v$(VERSION)...$(NC)"
	git tag -a v$(VERSION) -m "Release v$(VERSION)"
	git push origin v$(VERSION)

demo:
	@echo "$(CYAN)Running demo...$(NC)"
	./$(BINARY_NAME) -i . -o demo.txt -ext .go,.md -format markdown -verbose

demo-all:
	@echo "$(CYAN)Running all format demos...$(NC)"
	./$(BINARY_NAME) -i . -o demo-text.txt -ext .go,.md -verbose
	./$(BINARY_NAME) -i . -o demo.json -ext .go -format json -verbose
	./$(BINARY_NAME) -i . -o demo.xml -ext .go -format xml -verbose
	./$(BINARY_NAME) -i . -o demo.md -ext .go -format markdown -verbose
	@echo "$(GREEN)✓ All demos completed!$(NC)"

cross-compile:
	@echo "$(CYAN)Cross-compiling for all platforms...$(NC)"
	@mkdir -p dist
	@echo "$(YELLOW)Linux...$(NC)"
	@mkdir -p dist
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dist/pecel-linux-amd64 ./cmd/main
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o dist/pecel-linux-arm64 ./cmd/main
	GOOS=linux GOARCH=386 go build -ldflags="-s -w" -o dist/pecel-linux-386 ./cmd/main

	@echo "$(YELLOW)macOS...$(NC)"
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o dist/pecel-darwin-amd64 ./cmd/main
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o dist/pecel-darwin-arm64 ./cmd/main

	@echo "$(YELLOW)Windows...$(NC)"
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/pecel-windows-amd64.exe ./cmd/main
	GOOS=windows GOARCH=arm64 go build -ldflags="-s -w" -o dist/pecel-windows-arm64.exe ./cmd/main
	GOOS=windows GOARCH=386 go build -ldflags="-s -w" -o dist/pecel-windows-386.exe ./cmd/main
	
	@echo "$(GREEN)✓ Cross-compilation complete!$(NC)"

checksums:
	@echo "$(CYAN)Generating SHA256 checksums...$(NC)"
	@if [ -d "dist" ]; then cd dist && find . -type f -name "pecel*" -exec sha256sum {} \; > ../checksums.txt; fi
	@echo "$(GREEN)✓ Checksums generated in checksums.txt$(NC)"

benchmark:
	@echo "$(CYAN)Running benchmarks...$(NC)"
	go test -bench=. -benchmem ./...

size:
	@echo "$(CYAN)Binary sizes:$(NC)"
	@ls -lh $(BINARY_NAME) 2>/dev/null || echo "Binary not built yet"
	@ls -lh dist/*/$(BINARY_NAME) 2>/dev/null || echo "No dist binaries"

.PHONY: docker-build docker-run

docker-build:
	@echo "$(CYAN)Building Docker image...$(NC)"
	docker build -t pecel:latest -f Dockerfile .

docker-run:
	@echo "$(CYAN)Running in Docker...$(NC)"
	docker run --rm -v $(PWD):/data pecel:latest -i /data -o /data/docker-output.txt