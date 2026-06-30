.PHONY: all build clean test lint release install dev

APP_NAME    := issgo
VERSION     := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME  := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS     := -s -w -X github.com/issgo/issgo/cmd.Version=$(VERSION) -X github.com/issgo/issgo/cmd.BuildTime=$(BUILD_TIME)
BIN_DIR     := bin
GO          := go
GOFMT       := gofmt

all: tidy build

build:
	@mkdir -p $(BIN_DIR)
	$(GO) build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(APP_NAME) .

tidy:
	$(GO) mod tidy

fmt:
	$(GOFMT) -s -w .

lint:
	@which golangci-lint >/dev/null || $(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run ./...

test:
	$(GO) test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

test-short:
	$(GO) test -short -v ./...

clean:
	rm -rf $(BIN_DIR) release/

install: build
	cp $(BIN_DIR)/$(APP_NAME) /usr/local/bin/

release:
	@bash scripts/release.sh $(VERSION)

# Linux
release-linux-amd64:
	GOOS=linux GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(APP_NAME) .
	tar czf $(BIN_DIR)/$(APP_NAME)_linux_amd64.tar.gz -C $(BIN_DIR) $(APP_NAME)

release-linux-arm64:
	GOOS=linux GOARCH=arm64 $(GO) build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(APP_NAME) .
	tar czf $(BIN_DIR)/$(APP_NAME)_linux_arm64.tar.gz -C $(BIN_DIR) $(APP_NAME)

# macOS
release-darwin-amd64:
	GOOS=darwin GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(APP_NAME) .
	tar czf $(BIN_DIR)/$(APP_NAME)_darwin_amd64.tar.gz -C $(BIN_DIR) $(APP_NAME)

release-darwin-arm64:
	GOOS=darwin GOARCH=arm64 $(GO) build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(APP_NAME) .
	tar czf $(BIN_DIR)/$(APP_NAME)_darwin_arm64.tar.gz -C $(BIN_DIR) $(APP_NAME)

# Windows
release-windows-amd64:
	GOOS=windows GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(APP_NAME).exe .
	cd $(BIN_DIR) && zip $(APP_NAME)_windows_amd64.zip $(APP_NAME).exe

release-all: release-linux-amd64 release-linux-arm64 release-darwin-amd64 release-darwin-arm64 release-windows-amd64
	@cd $(BIN_DIR) && sha256sum *.tar.gz *.zip > checksums.txt
	@echo "All release artifacts in $(BIN_DIR)/"
	@ls -lh $(BIN_DIR)/*.tar.gz $(BIN_DIR)/*.zip

dev:
	$(GO) run .
