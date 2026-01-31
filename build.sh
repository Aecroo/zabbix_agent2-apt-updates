#!/bin/bash
# Zabbix Agent 2 APT Updates Plugin - Build Script
#
# This script provides convenient commands for building and deploying the plugin.

set -e

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Display help
show_help() {
    echo -e "${BLUE}Zabbix Agent 2 APT Updates Plugin - Build Script${NC}"
    echo ""
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  build          Build the plugin for all platforms (native)"
    echo "  build-docker   Build using Docker (cross-platform)"
    echo "  test           Run tests"
    echo "  clean          Clean build artifacts"
    echo "  deploy         Deploy with Docker Compose"
    echo "  start          Start the Zabbix Agent container"
    echo "  stop           Stop the Zabbix Agent container"
    echo "  logs           View container logs"
    echo "  shell          Get a shell in the dev container"
    echo "  help           Show this help message"
    echo ""
}

# Build natively using Makefile
build_native() {
    echo -e "${BLUE}Building plugin natively...${NC}"
    cd "$PROJECT_DIR"

    # Create dist directory if it doesn't exist
    mkdir -p dist

    # Build for multiple platforms
    echo -e "${GREEN}Building for Linux AMD64...${NC}"
    make GOOS=linux GOARCH=amd64 build

    echo -e "${GREEN}Building for Linux ARM64...${NC}"
    make GOOS=linux GOARCH=arm64 build

    echo -e "${GREEN}Building for Linux ARMv7...${NC}"
    make GOOS=linux GOARCH=arm GOARM=7 build

    # Rename files for clarity
    mv dist/zabbix-apt-updates-linux-amd64 dist/zabbix-apt-updates-linux-amd64 2>/dev/null || true
    mv dist/zabbix-apt-updates-linux-arm64 dist/zabbix-apt-updates-linux-arm64 2>/dev/null || true
    mv dist/zabbix-apt-updates-linux-armv7 dist/zabbix-apt-updates-linux-armv7 2>/dev/null || true

    echo -e "${GREEN}Build complete!${NC}"
    ls -lh dist/
}

# Build using Docker
build_docker() {
    echo -e "${BLUE}Building plugin using Docker...${NC}"
    cd "$PROJECT_DIR"

    # Create dist directory if it doesn't exist
    mkdir -p dist

    # Run docker compose build for builder service
    docker compose up --build builder

    echo -e "${GREEN}Docker build complete!${NC}"
    ls -lh dist/
}

# Run tests
test() {
    echo -e "${BLUE}Running tests...${NC}"
    cd "$PROJECT_DIR"
    go test -v ./...
}

# Clean build artifacts
clean() {
    echo -e "${BLUE}Cleaning build artifacts...${NC}"
    cd "$PROJECT_DIR"
    rm -rf dist/
    rm -f zabbix-apt-updates*
    echo -e "${GREEN}Clean complete!${NC}"
}

# Deploy with Docker Compose
deploy() {
    echo -e "${BLUE}Deploying with Docker Compose...${NC}"
    cd "$PROJECT_DIR"

    # Build first if dist doesn't exist or is empty
    if [ ! -d "dist" ] || [ -z "$(ls -A dist 2>/dev/null)" ]; then
        echo -e "${YELLOW}No build artifacts found. Building first...${NC}"
        build_docker
    fi

    docker compose up -d agent
    echo -e "${GREEN}Deployment complete!${NC}"
    echo -e "${BLUE}You can check the status with: docker-compose ps${NC}"
}

# Start the Zabbix Agent container
start() {
    echo -e "${BLUE}Starting Zabbix Agent container...${NC}"
    cd "$PROJECT_DIR"
    docker compose start agent
    echo -e "${GREEN}Container started!${NC}"
}

# Stop the Zabbix Agent container
stop() {
    echo -e "${BLUE}Stopping Zabbix Agent container...${NC}"
    cd "$PROJECT_DIR"
    docker compose stop agent
    echo -e "${GREEN}Container stopped!${NC}"
}

# View logs
logs() {
    echo -e "${BLUE}Viewing container logs...${NC}"
    cd "$PROJECT_DIR"
    docker compose logs -f agent
}

# Get shell in dev container
shell() {
    echo -e "${BLUE}Starting development container...${NC}"
    cd "$PROJECT_DIR"
    docker compose run --service-ports dev sh
}

# Main execution
main() {
    case "$1" in
        build)
            build_native
            ;;
        build-docker)
            build_docker
            ;;
        test)
            test
            ;;
        clean)
            clean
            ;;
        deploy)
            deploy
            ;;
        start)
            start
            ;;
        stop)
            stop
            ;;
        logs)
            logs
            ;;
        shell)
            shell
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            echo -e "${RED}Unknown command: $1${NC}"
            echo -e "${BLUE}Use '$0 help' for usage information.${NC}"
            exit 1
            ;;
    esac
}

main "$@"
