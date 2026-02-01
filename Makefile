# Zabbix Agent 2 APT Updates Plugin - Makefile
#
# This Makefile provides standard build targets for the plugin.

.PHONY: all build clean test dist install uninstall version help

# Project settings
NAME = zabbix-agent2-plugin-apt-updates
VERSION = 0.2.0
BUILD_DIR = dist
GOOS ?= linux
GOARCH ?= amd64

all: build

build: $(NAME)

$(NAME): main.go go.mod go.sum
	@echo "Building $(NAME) v$(VERSION) for $(GOOS)/$(GOARCH)..."
	@if [ -n "$(GOARM)" ]; then \
		GOOS=$(GOOS) GOARCH=$(GOARCH) GOARM=$(GOARM) go build -o $(BUILD_DIR)/$(NAME)-$(GOOS)-$(GOARCH)v$(GOARM); \
	else \
		GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(BUILD_DIR)/$(NAME)-$(GOOS)-$(GOARCH); \
	fi
	@echo "Build complete: $(BUILD_DIR)/$(NAME)-$(GOOS)-$(GOARCH)$(if $(GOARM),v$(GOARM),)"

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f $(NAME) $(NAME)-*
	@echo "Clean complete."

dist: build
	@echo "Creating distribution package..."
	@mkdir -p $(BUILD_DIR)/package
	@cp $(BUILD_DIR)/$(NAME)-$(GOOS)-$(GOARCH) $(BUILD_DIR)/package/
	@cp README.md $(BUILD_DIR)/package/
	@tar -czf $(BUILD_DIR)/$(NAME)-$(VERSION)-$(GOOS)-$(GOARCH).tar.gz -C $(BUILD_DIR)/package .
	@echo "Distribution package created: $(BUILD_DIR)/$(NAME)-$(VERSION)-$(GOOS)-$(GOARCH).tar.gz"

install: build
	@echo "Installing to /usr/local/bin..."
	@sudo install -m 755 $(BUILD_DIR)/$(NAME)-$(GOOS)-$(GOARCH) /usr/local/bin/$(NAME)
	@echo "Installation complete."

uninstall:
	@echo "Removing from /usr/local/bin..."
	@sudo rm -f /usr/local/bin/$(NAME)
	@echo "Uninstallation complete."

test:
	@echo "Running tests..."
	@go test -v ./...
	@echo "Tests complete."

version:
	@echo "$(NAME) v$(VERSION)"

help:
	@echo "Available targets:"
	@echo "  all        - Build the project (default)"
	@echo "  build      - Build the binary"
	@echo "  clean      - Remove build artifacts"
	@echo "  test       - Run tests"
	@echo "  dist       - Create distribution package"
	@echo "  install    - Install to /usr/local/bin"
	@echo "  uninstall  - Remove from /usr/local/bin"
	@echo "  version    - Show version information"
	@echo "  help       - Show this help message"

.PHONY: build-linux-amd64 build-linux-arm64 build-linux-arm7

# Cross-compilation targets
build-linux-amd64:
	make GOOS=linux GOARCH=amd64 build

build-linux-arm64:
	make GOOS=linux GOARCH=arm64 build

build-linux-arm7:
	make GOOS=linux GOARCH=arm GOARM=7 build
