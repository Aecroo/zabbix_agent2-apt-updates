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

    # Rotate old build artifacts
    rotate_old_builds

    # Build the image first (this caches the build)
    echo -e "${YELLOW}Building Docker image...${NC}"
    docker compose build builder

    # Run the builder container to copy files through volume mount
    # Note: Files will be owned by root since Docker runs as root.
    # This is expected behavior without sudo access.
    echo -e "${YELLOW}Running builder container to copy artifacts...${NC}"
    docker compose run --rm builder sh -c 'cp -v /build/dist/* /output/ 2>&1 || echo "No files to copy"'

    echo -e "${GREEN}Docker build complete!${NC}"
    echo -e "${YELLOW}Note: Files are owned by root. Use sudo or run as root for different ownership.${NC}"
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

# Rotate old build artifacts (keep only file, file.1, and file.2)
rotate_old_builds() {
    echo -e "${BLUE}Rotating old build artifacts...${NC}"
    cd "$PROJECT_DIR"

    # Check if dist directory exists and has files
    if [ -d "dist" ] && [ -n "$(ls -A dist 2>/dev/null)" ]; then
        for file in dist/*; do
            # Only process the actual binary files, not rotated ones
            base=$(basename "$file")
            if [[ "$base" =~ ^zabbix-agent2-plugin-apt-updates-linux-(amd64|arm64|armv7)$ ]]; then
                # Rotate: file.2 -> remove, file.1 -> file.2, file -> file.1
                # Preserve timestamps during rotation
                if [ -f "dist/${base}.1" ]; then
                    # Copy timestamp from .1 to .2 before moving
                    touch -r "dist/${base}.1" "dist/${base}.2" 2>/dev/null || true
                    mv -f "dist/${base}.1" "dist/${base}.2"
                fi
                if [ -f "dist/${base}" ]; then
                    # Copy timestamp from file to .1 before moving
                    touch -r "dist/${base}" "dist/${base}.1" 2>/dev/null || true
                    mv -f "dist/${base}" "dist/${base}.1"
                fi
            fi
        done
    fi
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
    # If no argument provided, show help
    if [ $# -eq 0 ]; then
        show_help
        return
    fi

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
