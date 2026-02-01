# Zabbix Agent 2 APT Updates Plugin - Builder Image
#
# This image builds the plugin binary for multiple platforms.
# Based on official Go image with necessary build tools.

FROM golang:1.24-alpine AS builder

WORKDIR /build

# Copy source files
COPY . .
# Copy zabbix_example for SDK replacement
COPY zabbix_example /build/zabbix_example

# Install build dependencies
RUN apk add --no-cache \
    make \
    git \
    ca-certificates

# Initialize Go module and download dependencies
RUN sed -i '/replace golang.zabbix.com\/sdk/d' go.mod && \
    # Download SDK first, then tidy
    GO111MODULE=on GOPRIVATE=golang.zabbix.com go mod download -x && \
    go mod tidy

# Build for multiple platforms
RUN for GOOS in linux; do \
    for GOARCH in amd64 arm64 arm; do \
        if [ "$GOARCH" = "arm" ]; then \
            for GOARM in 7; do \
                export GOARM=$GOARM; \
                echo "Building for ${GOOS}-${GOARCH}v${GOARM}..." && \
                make GOOS=$GOOS GOARCH=$GOARCH build || \
                (echo "Build failed for armv${GOARM}"; exit 1); \
            done; \
        else \
            echo "Building for ${GOOS}-${GOARCH}..." && \
            make GOOS=$GOOS GOARCH=$GOARCH build || \
            (echo "Build failed for ${GOOS}-${GOARCH}"; exit 1); \
        fi; \
    done; \
done

# Create distribution package
RUN mkdir -p /dist && \
    cp -r dist/* /dist/

# Final artifact location
FROM alpine:latest
WORKDIR /output
COPY --from=builder /dist .
# Rebuild marker So 1. Feb 03:26:01 UTC 2026
