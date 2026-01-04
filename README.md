# eero-cli

A command-line utility for controlling Eero mesh WiFi networks.

## Installation

### Using Homebrew

```bash
brew tap dciobanu/tap
brew install eero-cli
```

### Building from source-code

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
eero-cli devices                        # List all devices
eero-cli devices --online --wireless    # Filter by status/type
eero-cli devices --profile Kids         # Filter by profile
eero-cli devices --paused               # Show paused devices
eero-cli devices --private              # Show private (hidden MAC) devices
eero-cli devices monitor                # Monitor for state changes
eero-cli devices monitor --interval 5   # Custom poll interval
eero-cli devices inspect <id>           # Show full device JSON
eero-cli devices pause <id>             # Pause internet access
eero-cli devices unpause <id>           # Restore internet access
eero-cli devices block <id>             # Block from network
eero-cli devices unblock <id>           # Unblock device
eero-cli devices rename <id> <name>     # Set nickname
```

### Profiles

```bash
eero-cli profiles                           # List all profiles
eero-cli profiles inspect <id>              # Show full profile JSON
eero-cli profiles pause <id>                # Pause a profile
eero-cli profiles unpause <id>              # Unpause a profile
eero-cli profiles add <profile> <device>    # Add device to profile
eero-cli profiles remove <profile> <device> # Remove device from profile
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

## Later

API endpoints to explore for future features:

### Eero Nodes (High Value)
```
GET  /2.2/networks/{id}/eeros      - List mesh nodes (status, clients, mesh quality)
POST /2.2/eeros/{id}/reboot        - Reboot individual node
POST /2.2/eeros/{id}/led           - LED control
```

### Port Forwarding (High Value)
```
GET  /2.2/networks/{id}/forwards   - List rules (ip, ports, protocol, enabled)
POST /2.2/networks/{id}/forwards   - Create rule
PUT  /2.2/networks/{id}/forwards/{id} - Update/enable/disable
DELETE /2.2/networks/{id}/forwards/{id} - Delete rule
```

### DHCP Reservations (High Value)
```
GET  /2.2/networks/{id}/reservations - List (mac, ip, description)
POST /2.2/networks/{id}/reservations - Create reservation
DELETE /2.2/networks/{id}/reservations/{id} - Delete
```

### Profile Schedules
```
GET /2.2/networks/{id}/profiles/{id}/schedules - Bedtime/scheduled pauses
```

### Firmware Updates
```
GET /2.2/networks/{id}/updates     - Status (has_update, target_firmware, can_update_now)
```

## License

MIT
