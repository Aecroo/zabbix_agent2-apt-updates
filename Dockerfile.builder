# Zabbix Agent 2 APT Updates Plugin - Builder Image
#
# This image builds the plugin binary for multiple platforms.
# Based on official Go image with necessary build tools.

FROM golang:1.21-alpine AS builder

WORKDIR /build

# Copy source files
COPY . .

# Install build dependencies
RUN apk add --no-cache \
    make \
    git \
    ca-certificates

# Initialize Go module and download dependencies
RUN go mod download

# Build for multiple platforms
RUN for GOOS in linux; do \
    for GOARCH in amd64 arm64 arm; do \
        if [ "$GOARCH" = "arm" ]; then \
            for GOARM in 7; do \
                export GOARM=$GOARM \
                make build-linux-arm${GOARM} \
                && mv dist/zabbix-apt-updates-linux-arm${GOARM} dist/zabbix-apt-updates-linux-armv${GOARM} \
                || echo "Build failed for armv${GOARM}"; \
            done; \
        else \
            make GOOS=$GOOS GOARCH=$GOARCH build \
            && mv dist/zabbix-apt-updates-linux-${GOARCH} dist/zabbix-apt-updates-linux-${GOARCH} \
            || echo "Build failed for ${GOOS}-${GOARCH}"; \
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
