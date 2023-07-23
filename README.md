# Content

Little go script to download latest .ovpn files from vpngate

## Building

### Prerequisites

- Go should be installed
- Tunnelblick should be installed if using certain commands. See [flags](#configurations)

### Distro (Binary)

Run `go install github.com/lordvidex/vpn-gate` to install the binary in `$PATH`

### Source

To build from source, clone this repository and `go run main.go`

## Configurations

Run `vpn-gate --help` to view available options and their descriptions
Note that the use of any of the flags will require that [Tunnelblick](https://tunnelblick.net/index.html) is installed.

# Note ⚠️

CLI tool is still under construction

# Roadmap

- [x] downloading profiles
- [x] osascript commands should install the profiles
- [ ] deleting old and invalid profiles (cleanup commands)
- [ ] configuration for profiles that should be static i.e. never deleted by cleanup commands
