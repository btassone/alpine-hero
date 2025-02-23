# Alpine Hero

[![Go](https://github.com/btassone/alpine-hero/actions/workflows/go.yml/badge.svg)](https://github.com/btassone/alpine-hero/actions/workflows/go.yml)
[![codecov](https://codecov.io/gh/btassone/alpine-hero/branch/main/graph/badge.svg)](https://codecov.io/gh/btassone/alpine-hero)
[![Go Report Card](https://goreportcard.com/badge/github.com/btassone/alpine-hero)](https://goreportcard.com/report/github.com/btassone/alpine-hero)
[![License](https://img.shields.io/github/license/btassone/alpine-hero)](https://github.com/btassone/alpine-hero/blob/main/LICENSE)
[![Release](https://img.shields.io/github/v/release/btassone/alpine-hero)](https://github.com/btassone/alpine-hero/releases/latest)
[![Go Version](https://img.shields.io/github/go-mod/go-version/btassone/alpine-hero)](https://github.com/btassone/alpine-hero/blob/main/go.mod)

A command-line tool for generating Alpine Linux answer files for automated installation. This tool simplifies the
process of creating configuration files needed for unattended Alpine Linux installations.

## Features

- Generate answer files with customizable configuration
- Validate configuration settings
- Support for common system parameters:
  - Hostname configuration
  - User management (username, password, groups)
  - System settings (timezone, keyboard layout)
  - Network interface configuration
  - Disk device specification
- Cross-platform support (Linux, macOS)
- Automated CI/CD with GitHub Actions
- Comprehensive test coverage

## Prerequisites

- Go 1.23 or higher
- Make (optional, for using Makefile commands)

## Installation

### From Release

Download the latest release from the [releases page](https://github.com/username/alpine-template/releases/latest) for
your platform.

### From Source

Clone the repository and build the project:

```bash
# Build for current platform
make build

# Cross-compile for specific platforms
make build-linux  # For Linux ARM64 (e.g., Raspberry Pi)
make build-mac    # For macOS (both AMD64 and ARM64)
```

## Usage

### Basic Usage

Generate an answer file with default settings:

```bash
./alpine-template generate
```

### Custom Configuration

Specify custom settings using command-line flags:

```bash
./alpine-template generate \
  --hostname myhostname \
  --username myuser \
  --password mypassword \
  --timezone Europe/London \
  --keymap uk \
  --interface eth0 \
  --disk /dev/sda \
  --groups "audio,video,netdev,docker"
```

### Available Commands

- `generate`: Create an answers file
- `validate`: Check if the current configuration is valid

### Configuration Options

| Flag        | Short | Description                   | Default            |
|-------------|-------|-------------------------------|--------------------|
| --hostname  | -n    | System hostname               | alpinehost         |
| --username  | -u    | Main user account name        | alpine             |
| --password  | -p    | User password                 | changeme           |
| --timezone  | -t    | System timezone               | UTC                |
| --keymap    | -k    | Keyboard layout               | us                 |
| --interface | -i    | Network interface             | eth0               |
| --disk      | -d    | Installation disk device      | /dev/mmcblk0       |
| --groups    |       | User groups (comma-separated) | audio,video,netdev |
| --output    | -o    | Output file path              | answers.txt        |

## Development

### Available Make Commands

- `make help`: Display help information
- `make all`: Run tests and build
- `make build`: Build for current platform
- `make clean`: Clean build artifacts
- `make test`: Run tests
- `make coverage`: Generate test coverage report
- `make coverage-html`: Generate and open HTML coverage report
- `make fmt`: Format Go code
- `make lint`: Run linter
- `make deps`: Install dependencies

### Project Structure

```
.
├── .github/
│   └── workflows/    # GitHub Actions workflow files
├── templates/
│   └── answers.tmpl  # Answer file template
├── main.go          # Main application code
├── main_test.go     # Test files
├── go.mod          # Go module file
├── Makefile        # Build and development commands
└── README.md       # This file
```

## CI/CD Pipeline

The project uses GitHub Actions for continuous integration and deployment:

- Automated testing on every push and pull request
- Code coverage reporting via Codecov
- Automated linting with golangci-lint
- Automated releases when tags are pushed

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

Before submitting a PR, please ensure:

- All tests pass (`make test`)
- Code is properly formatted (`make fmt`)
- Linter shows no issues (`make lint`)
- Test coverage remains high (`make coverage`)

## License

[MIT License](LICENSE)

## Security

- Default configuration values are provided for demonstration only
- Change default passwords before deployment
- Review and customize all settings before using in production
- Regular security updates are provided through releases

## Support

For issues, questions, or contributions, please open an issue in the project repository.

## Acknowledgments

Thanks to all contributors who have helped with code, documentation, and testing.