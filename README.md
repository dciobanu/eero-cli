# eero-cli

A command-line utility for controlling Eero mesh WiFi networks.

## Installation

```bash
# Build
make build

# Install to /usr/local/bin
make install
```

## Usage

### Authentication

```bash
eero-cli login     # Authenticate with email/phone + verification code
eero-cli logout    # Clear saved token
eero-cli status    # Show authentication status
```

### Devices

```bash
eero-cli devices                    # List all devices
eero-cli devices pause <id>         # Pause internet access
eero-cli devices unpause <id>       # Restore internet access
eero-cli devices block <id>         # Block from network
eero-cli devices unblock <id>       # Unblock device
eero-cli devices rename <id> <name> # Set nickname
```

### Profiles

```bash
eero-cli profiles              # List all profiles
eero-cli profiles pause <id>   # Pause a profile
eero-cli profiles unpause <id> # Unpause a profile
```

### Guest Network

```bash
eero-cli guest                 # Show guest network status
eero-cli guest enable          # Enable guest network
eero-cli guest disable         # Disable guest network
eero-cli guest password <pass> # Set password
```

### Network

```bash
eero-cli reboot    # Reboot the network
```

## Configuration

Tokens are stored in:
- **macOS**: `~/Library/Application Support/eero-cli/config.json`
- **Linux**: `~/.config/eero-cli/config.json`

## Development

```bash
make build      # Build binary
make test       # Run tests
make build-all  # Cross-compile for all platforms
make clean      # Remove build artifacts
```

## License

MIT
