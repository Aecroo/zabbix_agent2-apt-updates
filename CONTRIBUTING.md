# Contributing to Zabbix Agent 2 APT Updates Plugin

Thank you for considering contributing to this project! We welcome contributions from the community.

## How to Contribute

### Reporting Issues

Before reporting an issue, please:
1. Check the [FAQ](#faq) below
2. Search existing issues to avoid duplicates
3. Provide detailed information about your environment and the problem

When reporting a bug, include:
- Operating system and version (e.g., Ubuntu 22.04)
- Plugin version
- Complete error message
- Steps to reproduce
- Expected vs actual behavior

### Suggesting Features

For feature requests:
1. Explain the use case and why it would be valuable
2. Provide examples of how you envision it working
3. Check if similar functionality already exists

### Submitting Code

1. Fork the repository
2. Create a feature branch from `master`
3. Make your changes with tests
4. Submit a pull request

## Development Setup

### Prerequisites
- Go 1.21 or later
- Git
- Make (optional, for using the Makefile)

### Building

```bash
# Clone the repository
git clone https://github.com/netdata/zabbix-agent-apt-updates.git
cd zabbix-agent-apt-updates

# Build the binary
make build

# Or use go directly
go build -o dist/zabbix-apt-updates
```

### Testing

Run tests:
```bash
go test -v ./...
```

## Code Style Guidelines

### Go Code
- Follow standard Go conventions (https://golang.org/doc/effective_go)
- Use meaningful variable and function names
- Keep functions focused on single responsibilities
- Add comments for non-obvious logic
- Error handling should be clear and consistent

### Commit Messages

Use the following format:
```
<type>(<scope>): <subject>

<body>

<footer>
```

Where:
- **type**: feat, fix, docs, style, refactor, test, chore
- **scope**: optional component (e.g., apt, dnf, config)
- **subject**: short description in present tense
- **body**: detailed explanation if needed
- **footer**: references to issues, etc.

Examples:
```
feat(apt): add security update detection

Add support for checking only security updates using apt-mark.

Fixes #123
```

```
fix(dnf): handle empty output correctly

Previously crashed when dnf check-update returned no output.
Now returns empty result set gracefully.
```

## Testing Requirements

All contributions must include:
- Unit tests for new functionality
- Tests for edge cases
- Documentation updates if the API changes

### Test Coverage

Aim for 80%+ coverage on critical paths. Run coverage report:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Documentation

Update the following files when making changes:
- `README.md` - User-facing documentation
- `CHANGELOG.md` - For new features and fixes
- `zabbix_agent2.conf.example` - Configuration examples

## Pull Request Process

1. Ensure all tests pass: `make test`
2. Update documentation
3. Add changelog entry in CHANGELOG.md
4. Submit PR with clear description
5. Address review feedback promptly

## FAQ

### Common Issues

**Q: Plugin returns "unsupported package manager" on Ubuntu**
A: Ensure `apt` is in PATH and the user has permission to execute it.

**Q: No updates detected but apt shows updates**
A: Run `sudo apt update` first, or check if caching is enabled.

**Q: Binary not found by Zabbix Agent**
A: Verify the path in `userparameter_apt.conf` and file permissions.

### Development Tips

- Use `ZBX_DEBUG=true` for debug output during development
- Test with different APT output formats (Debian vs Ubuntu)
- Consider edge cases like:
  - No network connection
  - APT cache not updated
  - Very large update lists
  - Special characters in package names

## License

By contributing, you agree that your contributions will be licensed under the GPL-2.0 license.
