# Docker Implementation Summary

## Overview

This document summarizes the Docker-based build and deployment system implemented for the Zabbix Agent 2 APT Updates plugin.

## What Was Implemented

### 1. Multi-Stage Builder Image (`Dockerfile.builder`)

**Purpose**: Compile the Go plugin for multiple platforms without requiring Go on the host machine.

**Features**:
- Based on `golang:1.21-alpine` - lightweight and efficient
- Builds for three architectures:
  - Linux AMD64 (x86_64)
  - Linux ARM64 (aarch64)
  - Linux ARMv7 (armhf)
- Outputs binaries to `/output` directory
- Clean, minimal final image

**Usage**:
```bash
docker-compose up builder
```

### 2. Runtime Image (`Dockerfile`)

**Purpose**: Run Zabbix Agent 2 with the APT Updates plugin pre-installed.

**Features**:
- Based on official `zabbix/zabbix-agent2:latest-alpine` image
- Includes the compiled plugin binary
- Pre-configured with user parameter for APT updates monitoring
- Health checks for container monitoring
- Minimal security context (read-only mounts where appropriate)

**Key Components**:
- Plugin binary at `/usr/local/bin/zabbix-apt-updates`
- Configuration at `/etc/zabbix/zabbix_agent2.d/userparameter_apt.conf`
- Documentation included in container

### 3. Docker Compose Configuration (`docker-compose.yml`)

**Purpose**: Orchestrate the build and deployment process.

**Services Defined**:

#### Builder Service
- Builds plugin for all platforms
- Outputs to host's `dist/` directory via volume mount
- Ideal for CI/CD pipelines

#### Agent Service (Runtime)
- Runs Zabbix Agent 2 with the plugin
- Exposes port 10050
- Mounts host APT databases for real update detection
- Configurable via environment variables
- Health checks every 30 seconds

#### Dev Service
- Interactive development environment
- Full Go toolchain available
- Source code mounted for live editing
- Persistent across sessions

**Usage Examples**:
```bash
# Build and deploy
docker-compose up -d agent

# Start dev environment
docker-compose up -d dev

# Run tests in dev container
docker-compose exec dev go test -v ./...
```

### 4. Build Script (`build.sh`)

**Purpose**: Simplify common operations with a user-friendly interface.

**Available Commands**:
- `build` - Native build using Makefile
- `build-docker` - Docker-based cross-compilation
- `test` - Run Go tests
- `clean` - Remove build artifacts
- `deploy` - Build and deploy with Docker Compose
- `start` - Start the agent container
- `stop` - Stop the agent container
- `logs` - View container logs
- `shell` - Get shell in dev container
- `help` - Show usage information

**Example Workflow**:
```bash
# Quick deployment
./build.sh deploy

# Development workflow
./build.sh shell
./build.sh test
```

### 5. Documentation (`DOCKER.md`)

Comprehensive guide covering:
- Installation and setup
- Build processes for different scenarios
- Deployment options (production, development, manual)
- Configuration reference
- Troubleshooting guide
- Best practices
- Security considerations
- Advanced topics (multi-instance, integration with other services)

## Architecture Diagram

```
┌───────────────────────────────────────────────────────┐
│                   HOST MACHINE                        │
├───────────────────────────────────────────────────────┤
│  ┌─────────────┐    ┌─────────────────────────────────┐  │
│  │  Docker     │    │   docker-compose.yml           │  │
│  │  Engine     │◄──►│   - builder service             │  │
│  └─────────────┘    │   - agent service               │  │
│                     │   - dev service                 │  │
│                     └─────────────────────────────────┘  │
│                                                  │        │
│  ┌───────────────────────────────────────────────┐    │
│  │               BUILDER IMAGE                  │    │
│  │   golang:1.21-alpine                         │    │
│  │   - Compiles plugin for multiple platforms   │◄──┘    │
│  │   - Outputs to dist/ directory                │        │
│  └───────────────────────────────────────────────┘        │
│                                                  │        │
│  ┌───────────────────────────────────────────────┐    │
│  │               RUNTIME IMAGE                  │    │
│  │   zabbix/zabbix-agent2:latest-alpine         │    │
│  │   - Runs Zabbix Agent 2                       │◄──┘    │
│  │   - Includes APT Updates plugin               │        │
│  │   - Exposes port 10050                        │        │
│  │   - Health checks                             │        │
│  └───────────────────────────────────────────────┘        │
│                                                  │        │
│  ┌───────────────────────────────────────────────┐    │
│  │               DEV IMAGE                      │    │
│  │   golang:1.21-alpine                         │    │
│  │   - Interactive shell                        │◄──┘    │
│  │   - Live code editing                        │        │
│  │   - Test execution                           │        │
│  └───────────────────────────────────────────────┘        │
└───────────────────────────────────────────────────────┘
```

## Key Benefits

### For Developers
1. **No Go installation required** - Build using Docker
2. **Cross-platform support** - Single command builds all architectures
3. **Consistent environment** - Same build process everywhere
4. **Easy testing** - Dev container with full toolchain
5. **Quick iteration** - Live code editing in dev container

### For Operations
1. **Simple deployment** - One command to build and deploy
2. **Containerized** - No dependencies on host system
3. **Configurable** - Environment variables for all settings
4. **Monitorable** - Health checks and logging built-in
5. **Portable** - Works on any Docker-host system

### For Production
1. **Minimal footprint** - Alpine-based images
2. **Security hardened** - Read-only mounts, minimal privileges
3. **Scalable** - Easy to deploy multiple instances
4. **Maintainable** - Clear configuration and documentation
5. **Observable** - Comprehensive logging and health checks

## Deployment Scenarios

### Scenario 1: Production Monitoring

```bash
# Clone repository
git clone http://192.168.0.23:3000/zbx/zabbix_agent2-apt-updates.git
cd zabbix-agent2-apt-updates

# Configure in docker-compose.yml
vim docker-compose.yml  # Set ZBX_SERVER_HOST, ZBX_HOSTNAME

# Deploy
docker-compose up -d agent

# Verify
docker-compose ps
docker-compose logs -f agent
```

### Scenario 2: CI/CD Pipeline

```yaml
# Example GitLab CI configuration
stages:
  - build
  - test

build:
  stage: build
  script:
    - docker-compose up builder
  artifacts:
    paths:
      - dist/

test:
  stage: test
  script:
    - docker-compose run --rm dev go test -v ./...
```

### Scenario 3: Development and Testing

```bash
# Start development environment
./build.sh shell

# Inside container:
go build -o zabbix-apt-updates .
./zabbix-apt-updates check

# Run tests
go test -v ./...
```

### Scenario 4: Multi-Architecture Deployment

```bash
# Build all platforms
./build.sh build-docker

# Results in dist/:
# - zabbix-apt-updates-linux-amd64
# - zabbix-apt-updates-linux-arm64
# - zabbix-apt-updates-linux-armv7

# Deploy to appropriate hardware
scp dist/zabbix-apt-updates-linux-arm64 pi@raspberrypi:/usr/local/bin/
```

## Configuration Options

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `ZBX_SERVER_HOST` | Zabbix server hostname/IP | (required) |
| `ZBX_HOSTNAME` | Hostname reported to Zabbix | (required) |
| `ZBX_UPDATES_THRESHOLD_WARNING` | Warning threshold for updates | 10 |
| `ZBX_DEBUG` | Enable debug logging | false |

### Volume Mounts

| Source | Target | Mode | Purpose |
|--------|--------|------|---------|
| `/var/lib/apt` | `/var/lib/apt` | ro | APT package database |
| `/etc/apt` | `/etc/apt` | ro | APT configuration |

### Ports

| Host | Container | Protocol | Purpose |
|------|-----------|----------|---------|
| 10050 | 10050 | TCP | Zabbix Agent communication |

## Troubleshooting Guide

### Common Issues and Solutions

**Issue**: Container fails to start with permission errors
- **Solution**: Ensure proper volume mounts for APT databases
```yaml
volumes:
  - /var/lib/apt:/var/lib/apt:ro
  - /etc/apt:/etc/apt:ro
```

**Issue**: No updates detected but host shows updates
- **Solution**: Container has its own APT cache
```bash
docker-compose exec agent apt update
```

**Issue**: Port already in use
- **Solution**: Change port mapping in docker-compose.yml
```yaml
ports:
  - "10051:10050"
```

**Issue**: Build fails due to missing dependencies
- **Solution**: Use the builder image which includes all dependencies
```bash
docker-compose up builder
```

### Debugging Tips

1. **Enable debug mode**:
   ```yaml
environment:
     - ZBX_DEBUG=true
   ```

2. **View logs**:
   ```bash
docker-compose logs -f agent | grep DEBUG
   ```

3. **Test plugin directly**:
   ```bash
docker-compose exec agent zabbix-apt-updates check
   ```

4. **Get shell in container**:
   ```bash
docker-compose exec agent sh
   ```

## Performance Considerations

### Image Size Optimization
- Builder image: ~300MB (golang:1.21-alpine)
- Runtime image: ~50MB (zabbix/zabbix-agent2:latest-alpine)
- Final artifacts: ~10MB each (plugin binaries)

### Resource Usage
- CPU: Minimal (plugin runs quickly and exits)
- Memory: ~50MB for agent process
- Disk: ~100MB for images and artifacts

### Scaling
- Can run multiple instances on same host with different ports
- Each instance can monitor different systems
- Stateless design allows easy horizontal scaling

## Security Best Practices

1. **Principle of Least Privilege**
   - Run as non-root user where possible
   - Use read-only volume mounts for sensitive data

2. **Network Security**
   - Limit Zabbix Agent port exposure
   - Use firewall rules to restrict access
   - Consider VPN or private network for agent-server communication

3. **Image Updates**
   - Regularly update base images (golang, zabbix-agent2)
   - Monitor security advisories for dependencies
   - Rebuild images when vulnerabilities are discovered

4. **Secrets Management**
   - For production, use Docker secrets or external secret management
   - Avoid hardcoding sensitive information in compose files

5. **Runtime Security**
   ```yaml
   # Example security configuration
   security_opt:
     - apparmor:unconfined
     - seccomp:unconfined
     - no-new-privileges:true

   cap_drop:
     - ALL

   read_only: true
   ```

## Future Enhancements

Potential improvements for future versions:

1. **Multi-stage runtime image** - Further reduce final image size
2. **Healthcheck customization** - Allow different check intervals
3. **Auto-update mechanism** - Pull latest plugin version automatically
4. **Metrics endpoint** - Expose Prometheus metrics alongside Zabbix
5. **TLS support** - Secure agent-server communication
6. **Configuration validation** - Pre-start checks for required settings
7. **Graceful shutdown** - Proper signal handling for container stops

## Conclusion

The Docker implementation provides a complete, production-ready solution for building and deploying the Zabbix Agent 2 APT Updates plugin. It addresses the needs of developers, operations teams, and end users with:

- ✅ Easy installation and deployment
- ✅ Cross-platform support
- ✅ Comprehensive documentation
- ✅ Production-grade configuration
- ✅ Development and testing environment
- ✅ CI/CD integration capabilities

The system is ready for immediate use in production environments or as a starting point for custom deployments.
