# Example plugin for Zabbix agent 2

This is an example repository meant to serve as a starting point for creating your own plugin. It has 2 functional
items used as examples for testing basic loadable plugin functionality. This readme is written as if for a 
real plugin - when creating your own plugin, it is suggested to rename instances of `example` to the correct name of your
new plugin.

## Requirements

- Zabbix agent 2 version 6.0.0 or newer
- If building from source: Latest or second-latest stable Go version (see go.mod for exact version requirements)

## Supported operating systems and architectures

The plugin works on all operating systems and architectures that the Go programming language and Zabbix agent 2 supports.

## Installation

The plugin can be compiled using `go build`. However, on Unix-based systems, it is suggested to use `make`;
on Windows-based systems, while it is suggested to use `mingw32-make`, it is required to use `windres.exe`.

## Setup

Set the `Plugins.Example.System.Path` setting in the Zabbix agent 2 configuration file with the path to the example plugin executable.

We recommend creating an `example.conf` file and placing all plugin-related configurations there.
Then, import the plugin configuration file in the Zabbix agent 2 configuration file `zabbix_agent2.conf`.

Add the following setting to the plugin configuration file `example.conf`:

```conf
Plugins.Example.System.Path=/path/to/executable/example
```
To import the plugin configuration file in Zabbix agent 2, add the following line to `zabbix_agent2.conf`:

```conf
Include=/path/to/config/example.conf
```

This is the bare minimum required to get the plugin running.
More information about available configuration settings is available below, under *Configuration options*.

## Command line options

The example plugin is not intended to be used as a command line utility; however, it does provide the following command line options:

- `-h`, `--help` - displays a help message
- `-V`, `--version` - prints the program version and license information

## Connection configuration

The example plugin mimics connection by requiring metrics to have a username and password.
The allowed usernames and passwords are hardcoded and used to simulate connection parameter requirements.

Allowed values

- Username: "Zabbix", "Admin", "User", "Test"
- Password: "", "Foo", "Bar", "Test"

### In metric key parameters

Each metric that the plugin provides has parameters for connection configuration. All keys have parameters for connection configuration.

### As a named session

Named sessions allow grouping the example connection settings under a name. Define the named session configuration parameters like so:

```conf
Plugins.Example.Sessions.StagingEnv.Username=Zabbix
Plugins.Example.Sessions.TestEnv.Username=Test
```

The example above defines sessions named `StagingEnv` and `TestEnv`.
The sessions can then be used as the first parameter to a metric key (`example.get[StagingEnv]` and `example.get[TestEnv]`) as opposed to defining each parameter separately.

## Configuration options

### Plugin settings

Below are the global settings for the example plugin. Applicable to all connections.

#### `Plugins.Example.System.Path`

Path to the example plugin executable.

Example:

```conf
Plugins.Example.System.Path=/usr/libexec/zabbix/zabbix-agent2-plugin-example
```

#### `Plugins.Example.Timeout`

Specifies the wait time (in seconds) for a device to respond when first connecting and on follow-up operations in the session.
Range: 1-30 seconds. If not specified, the value defaults to a global timeout value defined in Zabbix agent 2 configuration.

Example:

```conf
Plugins.Example.Timeout=10
```

### Session settings

For the following session configuration option, the `*` symbol in the field name implies a session name.
Replace `*` with the actual session name (for example, `StagingEnv` or `TestEnv`).

#### `Plugins.Example.Sessions.*.Username`

Specifies the username for the session `*`; the username will be checked against the hardcoded values.

Default: ``

Example:

```conf
Plugins.Example.Sessions.exampleSession.Username=Test
```

#### `Plugins.Example.Sessions.*.Password`

Specifies the password for the session `*`; the password will be checked against the hardcoded values.

Default: ``

Example:

```conf
Plugins.Example.Sessions.exampleSession.Password=foobar
```

### Default settings

The `Plugins.Example.Default.*` fields define the default values that will be used if no other value is specified.
The `*` symbol implies a specific configuration field.

#### `Plugins.Example.Default.Username`

Specifies the default username; the username will be checked against the hardcoded values.

Default: `Zabbix`

Example:

```conf
Plugins.Example.Default.Username=Test
```

#### `Plugins.Example.Default.Password`

Specifies the default password; the password will be checked against the hardcoded values.

Default: ``

Example:

```conf
Plugins.Example.Default.Password=Foobar
```

## Metric keys

Every metric key has the following parameters (further referred to as `<commonParameters>`):

- User - Username to be tested against the hardcoded values.
- Password - Password to be tested against the hardcoded values.

### `example.myip[<commonParameters>]`

Returns the provided caller's IP address. Tests against "https://api.ipify.org".

### `example.go.env[<commonParameters>,<environmentNames>...]`

Returns either all environments or those with names provided.

`<environmentNames>` - Environment names to return; returns all if empty, returns an error if an environment is not found.

## Troubleshooting

The plugin sends all of its logs to Zabbix agent 2, which further logs them wherever the agent 2 log file location is configured to.

For debugging Zabbix agent 2, the log level setting can be increased either in the configuration file field `DebugLevel` or by runtime control, running:

```sh
zabbix_agent2 -R log_level_increase
```

For more information about Zabbix agent 2, see [Zabbix documentation](https://www.zabbix.com/documentation/current/en/manual/concepts/agent2).

## Feedback

Noticed a bug or have an idea for improvement?
Feel free to open an issue in the [Zabbix support system](https://support.zabbix.com/secure/Dashboard.jspa).
