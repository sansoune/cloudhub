BINARY_NAME=cloudhub
BINARY_PATH=dist/$(BINARY_NAME)
MAIN_PATH=cmd/main.go
GO=go
GOFLAGS=-v

ARGS ?=

.PHONY: all build run clean

all: clean build

build:
	@echo "building $(BINARY_NAME)"
	@mkdir -p dist
	$(GO) build $(GOFLAGS) -o $(BINARY_PATH) $(MAIN_PATH)
	@echo "✅ Build complete: $(BINARY_PATH)"

run: build
	@echo "Running $(BINARY_NAME)..."
	@./$(BINARY_PATH) $(ARGS)

build-all:
	@echo "Building for all platforms..."
	@mkdir -p dist
	GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=linux GOARCH=arm64 $(GO) build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	GOOS=linux GOARCH=arm GOARM=7 $(GO) build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-armv7 $(MAIN_PATH)
	@echo "✅ Cross-compilation complete"
	@ls -lh dist/

clean:
	@echo "Cleaning..."
	@$(RM) $(BINARY_PATH) 2>/dev/null || true
	@$(RM) coverage.out coverage.html 2>/dev/null || true
	$(GO) clean
	@echo "✅ Clean complete"
