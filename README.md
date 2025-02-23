# Alpine Template Generator

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

## Prerequisites

- Go 1.23 or higher
- Make (optional, for using Makefile commands)

## Installation

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
- `make fmt`: Format Go code
- `make lint`: Run linter
- `make deps`: Install dependencies

### Project Structure

```
.
├── templates/
│   └── answers.tmpl    # Answer file template
├── main.go             # Main application code
├── go.mod             # Go module file
├── Makefile           # Build and development commands
└── README.md          # This file
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

[Insert License Information Here]

## Security

- Default configuration values are provided for demonstration only
- Change default passwords before deployment
- Review and customize all settings before using in production

## Support

For issues, questions, or contributions, please open an issue in the project repository.