# Changelog

## 0.4.0 (unreleased)

### Changed
- Errors are now printed to stderr as `Error: <message>` and the CLI exits with
  status 1. Previously most failures printed to stdout and exited 0.
- Usage text is only printed for command-line mistakes (unknown flags, wrong
  argument counts, missing required flags), not for runtime errors. An unknown
  subcommand prints the error, any "Did you mean" suggestions and
  "Run 'workbrew-cli --help' for usage."
- Commands that take no arguments now reject them (e.g. `devices list foo`)
  instead of silently ignoring them.
- An unknown subcommand of a group (`devices`, `report`, `command`, `auth`) is an
  error (exit 1) with "Did you mean" suggestions and a `--help` hint; it
  previously printed the group's help and exited 0. Running a group on its own
  still shows its help.
- `auth status` exits 1 when no API token is stored.
- `devices get` exits 1 when no device matches.
- `report vulnerabilities --severity <invalid>` and `report summary -o <invalid>`
  exit 1.
- Workbrew API errors now include up to 512 bytes of the response body, e.g.
  `workbrew API returned 401 Unauthorized: {"error":"..."}`.
- A missing config or keychain token gives a "run `workbrew-cli setup`" message;
  other keychain errors are shown in full instead of being reported as "no token".
- API requests time out after 30 seconds.
- Ctrl-C cancels in-flight API requests and prompts, prints `interrupted` to
  stderr and exits 130; a second Ctrl-C exits immediately.
- `setup --url` must be an `https://` URL with a host. It is checked before the
  token prompt, and trailing slashes are removed.
- `setup` fails with a clear error when stdin is not a terminal.
- `command update` / `command upgrade` fail when stdin is not a terminal and
  `--yes` was not given, instead of reading an empty answer and cancelling.
- The `command update` / `command upgrade` confirmation prompt, the `setup`
  token prompt and "Cancelled" are written to stderr, so stdout only carries results.
- `-o/--output` is checked while flags are parsed on every command that has it
  (`devices list` and `devices get` previously accepted any value and printed a
  table). An invalid value prints usage and exits 1.
- Tables are aligned with tabwriter, so columns fit their contents instead of
  using fixed widths (long names no longer push columns out of line).
- `report summary` is now an alias for `devices list`: same output, including the
  "Workbrew Devices" heading (was "Workbrew Device Summary").
- The "Seen" / "Last Seen" column shows "Today", "1 day" or "N days" (was a bare
  number). Dates up to 5 minutes in the future (clock skew) count as "Today";
  dates further in the future, or that can't be parsed, show "Unknown".
- `report analytics` is sorted by count (highest first), then command, then
  device, in both table and JSON output (JSON was previously in API order).
- `report brew-commands` is sorted by run count (highest first), then command,
  in both table and JSON output (JSON was previously in API order).
- `report analytics --device` now needs an exact serial number (case-insensitive);
  it previously matched any device containing the value.
- `devices get` prefers an exact (case-insensitive) serial or user match, then
  falls back to a partial match. If several devices match, they are listed on
  stderr and the command exits 1 (it previously showed the first partial match).
- `devices get -o json` returns an object with the device fields plus
  `formula_count`, `cask_count` and `outdated_count`, and `formulae` / `casks`
  arrays when `--apps` is given (it previously returned only the device fields).
- Empty lists in JSON output are `[]`, never `null` (`report outdated`,
  `report analytics --device`, `devices get --apps`).
- `report vulnerabilities` prints a warning to stderr if device names can't be
  looked up, and falls back to serial numbers (previously silent).
- `--version` prints the version (`workbrew-cli version x.y.z`); the `version`
  subcommand still prints just the number.
- Running `workbrew-cli` with no arguments shows help (it previously printed
  "Workbrew CLI").
- `devices get` fetches devices, formulae and casks in parallel, and
  `report vulnerabilities` fetches device names alongside the report (table
  output only), so both are faster.
- Failing to write table output to stdout (for example, redirected to a full disk)
  is now an error (`Error: writing output: ...`, exit 1). JSON output already
  reported this.
