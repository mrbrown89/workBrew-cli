# workbrew-cli

A command line interface for querying Workbrew workspaces from the terminal.

## Features

- Configure a Workbrew workspace
- Store API tokens securely in the macOS Keychain
- Verify API connectivity
- List managed devices
- View detailed device information
- View installed formulae and casks
- View outdated packages
- Output data as tables or JSON

## Requirements

- macOS
- Go 1.24+
- A Workbrew API token

## Installation

Clone the repository:

```bash
git clone https://github.com/mrbrown89/workBrew-cli.git
cd workbrew-cli
```

Build the binary:

```bash
go build -o workbrew
```

Run:

```bash
./workbrew --help
```

## Configuration

Configure your Workbrew workspace URL:

```bash
./workbrew setup --url https://console.workbrew.com/workspaces/<workspace>
```

You will be prompted to enter your Workbrew API token. The token is stored securely in the macOS Keychain.

Verify authentication:

```bash
./workbrew auth status
./workbrew auth test
```

## Commands

### List Devices

```bash
./workbrew devices list
```

Example output:

```text
Workbrew Devices
----------------

Serial Number      Assigned User                  macOS          Seen
-------------      -------------                  -----          ----
<serial number>    <user>                      <macOS verison>  <last seen date>     

Total Devices: <number of devices>
```

JSON output:

```bash
./workbrew devices list -o json
```

### Show Device Details

Search by serial number:

```bash
./workbrew devices get <serial number>
```

Search by assigned user:

```bash
./workbrew devices get <user name>
```

Example output:

```text
Workbrew Device
---------------

Serial Number: <serial number>
Assigned User: <user name>
Device Type:   MacBook Pro
macOS:         26.5.1 (25F80)
Last Seen:     Today

Formulae:      15
Casks:         9
Outdated:      5
```

JSON output:

```bash
./workbrew devices get <user name> -o json
```

### Show Installed Applications

```bash
./workbrew devices get <user name> --apps
```

Example output:

```text
Formulae
--------

awscli [OUTDATED]
docker [OUTDATED]
ollama [OUTDATED]
ansible
go

Casks
-----

imazing [OUTDATED]
postman [OUTDATED]
ghostty
keka
```

### Device Summary Report

```bash
./workbrew report summary
```

JSON output:

```bash
./workbrew report summary -o json
```

### Outdated Package Report

```bash
./workbrew report outdated
```

JSON output:

```bash
./workbrew report outdated -o json
```

## Configuration Storage

Workspace configuration is stored at:

```text
~/Library/Application Support/workbrew-cli/config.yaml
```

API tokens are stored securely in the macOS Keychain.

## Development

Format source code:

```bash
go fmt ./...
```

Build:

```bash
go build
```

Run:

```bash
go run .
```

Run tests:

```bash
go test ./...
```

## License

MIT
