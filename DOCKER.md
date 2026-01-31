# Docker Deployment Guide

This guide explains how to use Docker and Docker Compose to build, deploy, and manage the Zabbix Agent 2 APT Updates plugin.

## Overview

The project provides three Docker configurations:

1. **Builder Image** (`Dockerfile.builder`) - Compiles the plugin for multiple platforms
2. **Runtime Image** (`Dockerfile`) - Runs Zabbix Agent 2 with the plugin installed
3. **Development Environment** (`docker-compose.yml` dev service) - For testing and development

## Prerequisites

- Docker Engine (version 20.10+)
- Docker Compose (version 1.29+ or Docker Compose v2)
- Git

## Quick Start

### Build and Deploy in One Command

```bash
# Clone the repository
git clone http://192.168.0.23:3000/zbx/zabbix_agent2-apt-updates.git
cd zabbix-agent2-apt-updates

# Build and start the Zabbix Agent container
docker-compose up -d agent
```

### Verify Deployment

```bash
# Check container status
docker-compose ps

# View logs
docker-compose logs -f agent

# Test the plugin
docker-compose exec agent zabbix-apt-updates check
```

## Building the Plugin

### Using Docker Compose (Recommended)

```bash
# Build for all supported platforms
docker-compose up builder

# View the generated binaries
ls -lh dist/
```

This will create binaries for:
- Linux AMD64 (`dist/zabbix-apt-updates-linux-amd64`)
- Linux ARM64 (`dist/zabbix-apt-updates-linux-arm64`)
- Linux ARMv7 (`dist/zabbix-apt-updates-linux-armv7`)

### Using the Build Script

A convenient build script is provided:

```bash
# Show available commands
./build.sh help

# Build using Docker
./build.sh build-docker

# Build natively (if Go is installed)
./build.sh build
```

## Deployment Options

### Option 1: Docker Compose (Production)

Edit `docker-compose.yml` to configure:

```yaml
services:
  agent:
    environment:
      - ZBX_SERVER_HOST=your-zabbix-server.example.com
      - ZBX_HOSTNAME=monitoring-host-01
      - ZBX_UPDATES_THRESHOLD_WARNING=5
      # - ZBX_DEBUG=true  # Uncomment for debug output
```

Then deploy:

```bash
# Start the container
docker-compose up -d agent

# Stop the container
docker-compose stop agent

# Restart
docker-compose restart agent

# Update to latest version
git pull
docker-compose build --no-cache agent
docker-compose up -d agent
```

### Option 2: Manual Docker Run

```bash
# Build the image
docker build -t zabbix-apt-updates -f Dockerfile .

# Run with custom configuration
docker run -d \
  --name zabbix-agent-apt \
  -p 10050:10050 \
  -v /var/lib/apt:/var/lib/apt:ro \
  -v /etc/apt:/etc/apt:ro \
  -e ZBX_SERVER_HOST=zabbix.example.com \
  -e ZBX_HOSTNAME=production-server-01 \
  -e ZBX_UPDATES_THRESHOLD_WARNING=10 \
  zabbix-apt-updates
```

### Option 3: Development Environment

For testing and development:

```bash
# Start the dev container
docker-compose up -d dev

# Get a shell in the dev container
docker-compose exec dev sh

# Run tests inside the container
go test -v ./...
```

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `ZBX_SERVER_HOST` | (required) | Hostname or IP of Zabbix server/proxy |
| `ZBX_HOSTNAME` | (required) | Unique hostname for this agent |
| `ZBX_UPDATES_THRESHOLD_WARNING` | 10 | Number of updates that triggers warning |
| `ZBX_DEBUG` | false | Enable debug logging |

### Volumes

The runtime image mounts:
- `/var/lib/apt:/var/lib/apt:ro` - APT package database (read-only)
- `/etc/apt:/etc/apt:ro` - APT configuration (read-only)

For RHEL-based systems with DNF, you can also mount:
- `/var/lib/dnf:/var/lib/dnf:ro`
- `/etc/dnf:/etc/dnf:ro`

### Ports

The Zabbix Agent listens on port `10050`. Map this to the host as needed.

## Monitoring and Maintenance

### View Logs

```bash
# Follow logs in real-time
docker-compose logs -f agent

# View previous logs
docker-compose logs --tail=100 agent
```

### Check Health

The container includes a health check that verifies the Zabbix Agent is responding:

```bash
# Check container health
docker inspect --format='{{json .State.Health}}' zabbix-agent-apt-updates_agent_1

# Or use docker-compose
docker-compose ps
```

### Test the Plugin

```bash
# Execute the plugin directly
docker-compose exec agent zabbix-apt-updates check

# Check version
docker-compose exec agent zabbix-apt-updates version
```

## Cross-Platform Builds

The builder image creates binaries for multiple architectures:

```bash
# Build all platforms
docker-compose up builder

# Results in dist/:
# - zabbix-apt-updates-linux-amd64
# - zabbix-apt-updates-linux-arm64
# - zabbix-apt-updates-linux-armv7
```

You can use these binaries to deploy to different hardware platforms.

## Troubleshooting

### Common Issues

**Issue: Container fails with "permission denied" when checking updates**

Solution: Ensure the container has access to APT databases:
```yaml
volumes:
  - /var/lib/apt:/var/lib/apt:ro
  - /etc/apt:/etc/apt:ro
```

**Issue: No updates detected but apt shows updates on host**

Solution: The container has its own APT cache. Run `apt update` inside the container:
```bash
docker-compose exec agent apt update
```

**Issue: Port 10050 already in use**

Solution: Change the port mapping in docker-compose.yml:
```yaml
ports:
  - "10051:10050"  # Map host port 10051 to container port 10050
```

### Debug Mode

Enable debug logging for troubleshooting:

```bash
# In docker-compose.yml:
environment:
  - ZBX_DEBUG=true

# Or when running manually:
-e ZBX_DEBUG=true
```

Then view logs:
```bash
docker-compose logs -f agent | grep DEBUG
```

## Best Practices

### Production Deployment

1. **Use named volumes** for persistent data (if needed)
2. **Set resource limits**:
   ```yaml
   deploy:
     resources:
       limits:
         cpus: '0.5'
         memory: 128M
   ```
3. **Enable health checks** in Zabbix to monitor the agent
4. **Use secrets** for sensitive configuration (if needed)
5. **Consider security context**:
   ```yaml
   security_opt:
     - apparmor:unconfined
     - seccomp:unconfined
   ```

### Security Considerations

1. Run as non-root if possible
2. Use read-only mounts for sensitive directories
3. Limit network access to Zabbix server only
4. Keep the image updated with security patches

## Updating

```bash
# Pull latest changes
git pull origin master

# Rebuild and restart
docker-compose build --no-cache agent
docker-compose up -d agent

# Or if you also need to rebuild the plugin
docker-compose build --no-cache builder
docker-compose up -d builder agent
```

## Docker Compose Reference

The `docker-compose.yml` file defines three services:

### Builder Service

- **Purpose**: Compile the Go plugin for multiple platforms
- **Image**: golang:1.21-alpine
- **Output**: Binaries in the `dist/` directory on the host

### Agent Service (Runtime)

- **Purpose**: Run Zabbix Agent 2 with the APT updates plugin
- **Base Image**: zabbix/zabbix-agent2:latest-alpine
- **Ports**: 10050 (Zabbix Agent)
- **Volumes**: APT databases (read-only)
- **Health Check**: HTTP check on port 10050

### Dev Service

- **Purpose**: Development and testing environment
- **Image**: golang:1.21-alpine
- **Volumes**: Source code mounted for live editing
- **Usage**: Interactive shell for development

## Advanced Configuration

### Custom Zabbix Agent Configuration

You can extend the Zabbix Agent configuration by:

1. Creating a custom `zabbix_agent2.conf` file
2. Mounting it as a volume:
   ```yaml
   volumes:
     - ./custom-zabbix-agent2.conf:/etc/zabbix/zabbix_agent2.conf
   ```

### Multiple Instances

To run multiple instances on the same host:

```bash
# Create separate compose files
docker-compose -f docker-compose.yml -f docker-compose-multi.yml up -d

# Or use different port mappings
ports:
  - "10051:10050"  # First instance
  - "10052:10050"  # Second instance
```

### Integration with Other Services

Example with Zabbix Server and other monitoring tools:

```yaml
version: '3.8'

services:
  zabbix-server:
    image: zabbix/zabbix-server-mysql
    ports:
      - "8080:8080"
    environment:
      - DB_SERVER_HOST=mysql
      - MYSQL_USER=zabbix
      - MYSQL_PASSWORD=password
    depends_on:
      - mysql

  agent:
    build:
      context: .
      dockerfile: Dockerfile
    environment:
      - ZBX_SERVER_HOST=zabbix-server
      - ZBX_HOSTNAME=monitoring-agent-01
    depends_on:
      - zabbix-server

  mysql:
    image: mysql:8.0
    environment:
      - MYSQL_ROOT_PASSWORD=secret
      - MYSQL_DATABASE=zabbix
      - MYSQL_USER=zabbix
      - MYSQL_PASSWORD=password
```

## Support

For issues and questions related to Docker deployment:

1. Check the logs: `docker-compose logs agent`
2. Test manually: `docker-compose exec agent zabbix-apt-updates check`
3. Review this documentation
4. Open an issue in the project repository with relevant logs

## Resources

- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Zabbix Agent 2 Docker Image](https://hub.docker.com/r/zabbix/zabbix-agent2)
- [Go Docker Image](https://hub.docker.com/_/golang)
