# Zabbix Agent 2 APT Updates Plugin - Runtime Image
#
# This image includes the Zabbix Agent 2 with the APT updates plugin pre-installed.
# Based on official zabbix/zabbix-agent2 image.

FROM zabbix/zabbix-agent2:latest-alpine

WORKDIR /app

# Install dependencies
RUN apk add --no-cache \
    ca-certificates \
    libstdc++ \
    && mkdir -p /var/lib/zabbix/modules /etc/zabbix/zabbix_agent2.d

# Copy the plugin binary (built separately)
COPY dist/zabbix-agent2-plugin-apt-updates-linux-amd64 /usr/libexec/zabbix/zabbix-agent2-plugin-apt-updates
RUN chmod +x /usr/libexec/zabbix/zabbix-agent2-plugin-apt-updates

# Copy plugin configuration
COPY apt-updates.conf /etc/zabbix/zabbix_agent2.d/apt-updates.conf

# Copy documentation
COPY README.md CHANGELOG.md CONTRIBUTING.md /app/

# Set up Zabbix Agent configuration
RUN echo "PidFile=/tmp/zabbix-agent2.pid" >> /etc/zabbix/zabbix_agent2.conf && \
    echo "LogFile=/tmp/zabbix-agent2.log" >> /etc/zabbix/zabbix_agent2.conf && \
    echo "Include=/etc/zabbix/zabbix_agent2.d/*.conf" >> /etc/zabbix/zabbix_agent2.conf

# Expose Zabbix Agent port
EXPOSE 10050

# Health check
HEALTHCHECK --interval=30s --timeout=3s \
    CMD wget -qO- http://localhost:10050 || exit 1

CMD ["zabbix_agent2", "-", "--config", "/etc/zabbix/zabbix_agent2.conf"]
