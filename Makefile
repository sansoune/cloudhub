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


clean:
	@echo "Cleaning..."
	@$(RM) $(BINARY_PATH) 2>/dev/null || true
	@$(RM) coverage.out coverage.html 2>/dev/null || true
	$(GO) clean
	@echo "✅ Clean complete"
