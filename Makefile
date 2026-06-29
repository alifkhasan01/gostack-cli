APP      = gostack
VERSION  ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT   ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE     ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS   = -ldflags "-s -w \
  -X github.com/gostack/cli/internal/cli.Version=$(VERSION) \
  -X github.com/gostack/cli/internal/cli.Commit=$(COMMIT) \
  -X github.com/gostack/cli/internal/cli.BuildDate=$(DATE)"
BIN_DIR   = ./bin

.PHONY: build install test test-short lint vet tidy clean release-dry run-help

## build: compile binary to ./bin/gostack
build:
	@mkdir -p $(BIN_DIR)
	go build $(LDFLAGS) -o $(BIN_DIR)/$(APP) ./cmd/gostack
	@echo "✅ Built: $(BIN_DIR)/$(APP)  ($(VERSION))"

## install: install binary to $GOPATH/bin
install:
	go install $(LDFLAGS) ./cmd/gostack
	@echo "✅ Installed: $(shell go env GOPATH)/bin/$(APP)"

## test: run all tests with verbose output
test:
	go test ./... -v -count=1

## test-short: run tests without verbose
test-short:
	go test ./... -count=1

## vet: run go vet
vet:
	go vet ./...

## lint: run golangci-lint (must be installed)
lint:
	golangci-lint run ./...

## tidy: tidy go modules
tidy:
	go mod tidy

## clean: remove build artifacts
clean:
	rm -rf $(BIN_DIR)

## release-dry: dry-run goreleaser (must be installed)
release-dry:
	goreleaser release --snapshot --clean

## run-help: build and show CLI help
run-help: build
	$(BIN_DIR)/$(APP) --help
