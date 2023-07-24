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
Note that the use of some of the flags will require that [Tunnelblick](https://tunnelblick.net/index.html) is installed.
|flag|description|requires Tunnelblick|
|-|-|-|
||not adding any flag, simply downloads the ovpn configs in the current working directory|
|`-clear`|removes configs installed with this CLI, i.e. it does not tamper with configs installed through other means|✅|
|`-install`|installs with tunnelblick, when not added, it just downloads the configs|✅|
|`-rm`|removes old configurations before fetching ovpn files, this is useful if the old configs have stopped working. To prevent removal of certain configs, see `-no-rm`|✅|
|`-no-rm`|specifies configs that should not be deleted either with `-clear` or `-rm`|
|`-print`|returns all the configs in Tunnelblick. this is the same as the list of configs shown on Tunnelblick itself|✅|

# Note ⚠️

CLI tool is still under construction

# Roadmap

- [x] downloading profiles
- [x] osascript commands should install the profiles
- [ ] deleting old and invalid profiles (cleanup commands)
- [ ] configuration for profiles that should be static i.e. never deleted by cleanup commands
