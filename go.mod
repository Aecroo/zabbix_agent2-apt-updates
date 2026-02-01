module github.com/netdata/zabbix-agent-apt-updates

go 1.21

require (
	golang.zabbix.com/sdk v1.2.2-0.20251205121637-3b95c058c0e4
)

replace golang.zabbix.com/sdk => ../zabbix_example
