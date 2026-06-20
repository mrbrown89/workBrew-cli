# workbrew-cli

![CI](https://github.com/mrbrown89/workbrew-cli/actions/workflows/ci.yml/badge.svg)

A command line interface for querying Workbrew workspaces, devices, packages, and vulnerabilities from the terminal.

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
- Go 1.26.4+
- A Workbrew API token

## Installation

### Homebrew

```bash

brew install mrbrown89/tap/workbrew-cli

```

Verify installation:

```bash

workbrew-cli version

```

### Build from Source

Clone the repository:

```bash

git clone https://github.com/mrbrown89/workbrew-cli.git

cd workbrew-cli

```

Build the binary:

```bash

go build -o workbrew-cli

```

Run:

```bash

workbrew-cli --help

```

## Configuration

Configure your Workbrew workspace URL:

```bash
workbrew-cli setup --url https://console.workbrew.com/workspaces/<workspace>
```

You will be prompted to enter your Workbrew API token. The token is stored securely in the macOS Keychain.

Verify authentication:

```bash
workbrew-cli auth status
workbrew-cli auth test
```

## Commands

### List Devices

```bash
workbrew-cli devices list
```

Example output:

```text
Workbrew Devices
----------------

Serial Number      Assigned User                  macOS          Seen
-------------      -------------                  -----          ----
<serial number>    <user>                      <macOS version>  <last seen date>     

Total Devices: <number of devices>
```

JSON output:

```bash
workbrew-cli devices list -o json
```

### Show Device Details

Search by serial number:

```bash
workbrew-cli devices get <serial number>
```

Search by assigned user:

```bash
workbrew-cli devices get <user name>
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
workbrew-cli devices get <user name> -o json
```

### Show Installed Applications

```bash
workbrew-cli devices get <user name> --apps
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
workbrew-cli report summary
```

JSON output:

```bash
workbrew-cli report summary -o json
```

### Outdated Package Report

```bash
workbrew-cli report outdated
```

JSON output:

```bash
workbrew-cli report outdated -o json
```

### Vulnerability Report

```bash
workbrew-cli report vulnerabilities
```

Example output:

```text
Workbrew Vulnerabilities
------------------------

Formula                        CVEs       Max CVSS Devices
-------                        ----       -------- -------
curl                           2          8.2      2
openssl                        1          9.8      1

Total Vulnerable Formulae: 2
Total CVEs: 3
```

JSON output:

```bash
workbrew-cli report vulnerabilities -o json
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
go build -o workbrew-cli
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
