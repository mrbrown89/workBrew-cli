# workbrew-cli

![CI](https://github.com/mrbrown89/workbrew-cli/actions/workflows/ci.yml/badge.svg)

A command line interface for querying Workbrew workspaces, devices, packages, and vulnerabilities from the terminal.

## Features

- Store API tokens securely in the macOS Keychain
- Verify API connectivity
- List managed devices
- View detailed device information
- View installed formulae and casks
- View outdated packages and vulnerabilities
- Output data as tables or JSON

## Requirements

- macOS
- A Workbrew API token
- Go 1.26+ (only to build from source)

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

```bash
git clone https://github.com/mrbrown89/workbrew-cli.git
cd workbrew-cli
go build -o workbrew-cli
./workbrew-cli --help
```

Or install with Go:

```bash
go install github.com/mrbrown89/workbrew-cli@latest
```

## Configuration

Configure your Workbrew workspace URL:

```bash
workbrew-cli setup --url https://console.workbrew.com/workspaces/<workspace>
```

The URL must use `https://` and include a host; a trailing `/` is removed. You will then be prompted for your Workbrew API token, which is stored securely in the macOS Keychain. `setup` must be run in a terminal.

Verify authentication:

```bash
workbrew-cli auth status
workbrew-cli auth test
```

`auth status` exits with status 1 if no token is stored.

## Output, errors and exit codes

- Results (tables and JSON) are written to **stdout**. Prompts, warnings and errors are written to **stderr**, so `-o json` output can be piped safely.
- Errors are printed as `Error: <message>`. Mistakes on the command line (unknown flags, wrong arguments, a missing `--url`) also print usage; an unknown subcommand prints any "Did you mean" suggestions and a `--help` hint.
- In JSON output, empty lists are always `[]`, never `null`.

| Exit code | Meaning |
|-----------|---------|
| `0` | Success (including answering "no" at a confirmation prompt) |
| `1` | Any error: bad arguments, missing config or token, API errors, no or several matching devices |
| `130` | Interrupted with Ctrl-C |

API requests time out after 30 seconds. API errors include the HTTP status and the start of the response body.

## Commands

Run `workbrew-cli` with no arguments, or any command with `--help`, to see what is available. `workbrew-cli --version` and `workbrew-cli version` print the version.

Every command that supports `-o` / `--output` accepts `table` (the default) or `json`.

### List Devices

```bash
workbrew-cli devices list
```

Example output:

```text
Workbrew Devices
----------------

Serial Number  Assigned User  macOS           Seen
-------------  -------------  -----           ----
<serial>       <user>         26.5.1 (25F80)  Today
<serial>       <user>         15.7.1          3 days

Total Devices: 2
```

"Seen" is `Today`, `1 day` or `N days`; it is `Unknown` when the date is missing, can't be parsed, or is in the future.

JSON output:

```bash
workbrew-cli devices list -o json
```

### Show Device Details

```bash
workbrew-cli devices get <serial number>
workbrew-cli devices get <user name>
```

The search is case-insensitive. An exact match on the serial number or assigned user is used first; if there is none, a partial match is used. If exactly one device matches it is shown. If several match, they are listed on stderr and the command exits with status 1, so use a full serial number or user name.

Example output:

```text
Workbrew Device
---------------

Serial Number:  <serial number>
Assigned User:  <user name>
Device Type:    MacBook Pro
macOS:          26.5.1 (25F80)
Last Seen:      Today

Formulae:       15
Casks:          9
Outdated:       5
```

JSON output is an object with the device fields plus `formula_count`, `cask_count` and `outdated_count`:

```bash
workbrew-cli devices get <user name> -o json
```

### Show Installed Applications

```bash
workbrew-cli devices get <user name> --apps
```

Outdated apps are listed first, then the rest by name. With `-o json`, `--apps` adds `formulae` and `casks` arrays to the object.

Example output:

```text
Formulae
--------

awscli [OUTDATED]
docker [OUTDATED]
ansible
go

Casks
-----

imazing [OUTDATED]
ghostty
keka
```

### Device Summary Report

`report summary` is an alias for `devices list` and produces the same output:

```bash
workbrew-cli report summary
workbrew-cli report summary -o json
```

### Outdated Package Report

```bash
workbrew-cli report outdated
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

Package  Type     CVEs  Max CVSS  Devices
-------  ----     ----  --------  -------
OpenSSL  formula  2     9.8       Alex Example, Sam
curl     formula  1     7.5       <serial>

Total Vulnerability Findings: 3
Total Unique CVEs: 3
Total Affected Packages: 2
Total Unique Affected Devices: 3
Critical Vulnerabilities: 1
High Vulnerabilities: 1
Medium Vulnerabilities: 1
Low Vulnerabilities: 0
None/Unscored Vulnerabilities: 0
```

Devices are shown by assigned user where known. If device names can't be looked up, a warning is printed to stderr and serial numbers are shown instead.

Filter by severity (`critical`, `high`, `medium` or `low`) and/or by device. Filtering lists one row per CVE:

```bash
workbrew-cli report vulnerabilities --severity critical
workbrew-cli report vulnerabilities --device <serial number>
```

`--device` takes an exact serial number (case-insensitive); partial values don't match.

JSON output:

```bash
workbrew-cli report vulnerabilities -o json
```

### Analytics Report

```bash
workbrew-cli report analytics
```

Sorted by count (highest first), then command, then device, in both table and JSON output.

Filter by device (exact serial number, case-insensitive):

```bash
workbrew-cli report analytics --device <serial number>
```

JSON output:

```bash
workbrew-cli report analytics -o json
```

### Brewfiles Report

```bash
workbrew-cli report brewfiles
workbrew-cli report brewfiles -o json
```

### Brew Commands Report

```bash
workbrew-cli report brew-commands
workbrew-cli report brew-commands -o json
```

Sorted by run count (highest first), then command, in both table and JSON output.

### Create a Brew Update or Upgrade Command

```bash
workbrew-cli command update
workbrew-cli command upgrade
```

You will be asked to confirm on the terminal. To skip the confirmation (and in scripts, where stdin is not a terminal), pass `--yes`:

```bash
workbrew-cli command upgrade --yes
```

Without `--yes`, the command fails if stdin is not a terminal.

## Configuration Storage

Workspace configuration is stored at:

```text
~/Library/Application Support/workbrew-cli/config.yaml
```

API tokens are stored securely in the macOS Keychain.

## Development

Run the same checks as CI:

```bash
gofmt -l .
go vet ./...
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 run
go mod tidy -diff
go run golang.org/x/vuln/cmd/govulncheck@v1.10.0 ./...
go test -race ./...
```

CI runs these on Go 1.26.x (the minimum in `go.mod`) and 1.27.x.

The tests never call the real Workbrew API or Keychain: they use `httptest` and a mocked keyring. Command output is compared with golden files in `testdata/`; after an intended output change, regenerate them with:

```bash
go test -run TestCommandsGolden -update .
```

## License

MIT
