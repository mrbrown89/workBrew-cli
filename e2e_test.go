package main

import (
	"bytes"
	"errors"
	"flag"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/zalando/go-keyring"
)

var update = flag.Bool("update", false, "rewrite golden files in testdata/")

// apiResponse is a canned response for one endpoint on the fake API.
type apiResponse struct {
	status int
	body   string
}

type recordedRequest struct {
	method string
	path   string
	header http.Header
	body   string
}

// testEnv is a fake Workbrew API plus an isolated config directory and a
// mocked keychain. Nothing in it talks to the real API or keychain.
type testEnv struct {
	t      *testing.T
	server *httptest.Server

	// placeholders replaces the generated (time-relative) timestamps with
	// fixed text, so golden files stay stable.
	placeholders *strings.Replacer

	mu        sync.Mutex
	responses map[string]apiResponse // keyed by "METHOD /path"
	requests  []recordedRequest
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()

	responses, placeholders := defaultResponses(time.Now())
	env := &testEnv{t: t, responses: responses, placeholders: placeholders}
	env.server = httptest.NewServer(http.HandlerFunc(env.handle))
	t.Cleanup(env.server.Close)

	// os.UserConfigDir reads HOME on macOS and XDG_CONFIG_HOME on Linux.
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))

	keyring.MockInit()

	env.writeConfig("url: " + env.server.URL + "/\n")
	env.setToken("test-token")

	return env
}

func (env *testEnv) handle(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)

	env.mu.Lock()
	env.requests = append(env.requests, recordedRequest{
		method: r.Method,
		path:   r.URL.Path,
		header: r.Header.Clone(),
		body:   string(body),
	})
	response, ok := env.responses[r.Method+" "+r.URL.Path]
	env.mu.Unlock()

	if !ok {
		http.Error(w, "no fixture for "+r.Method+" "+r.URL.Path, http.StatusNotFound)
		return
	}

	if response.status != 0 {
		w.WriteHeader(response.status)
	}
	if _, err := io.WriteString(w, response.body); err != nil {
		env.t.Errorf("writing response body: %v", err)
	}
}

func (env *testEnv) respond(methodPath string, status int, body string) {
	env.mu.Lock()
	defer env.mu.Unlock()
	env.responses[methodPath] = apiResponse{status: status, body: body}
}

func (env *testEnv) writeConfig(content string) {
	env.t.Helper()

	path, err := getConfigPath()
	if err != nil {
		env.t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		env.t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		env.t.Fatal(err)
	}
}

func (env *testEnv) removeConfig() {
	env.t.Helper()

	path, err := getConfigPath()
	if err != nil {
		env.t.Fatal(err)
	}
	if err := os.Remove(path); err != nil {
		env.t.Fatal(err)
	}
}

func (env *testEnv) setToken(token string) {
	env.t.Helper()

	if err := keyring.Set(keychainService, keychainAccount, token); err != nil {
		env.t.Fatal(err)
	}
}

func (env *testEnv) deleteToken() {
	env.t.Helper()

	if err := keyring.Delete(keychainService, keychainAccount); err != nil {
		env.t.Fatal(err)
	}
}

func (env *testEnv) recorded() []recordedRequest {
	env.mu.Lock()
	defer env.mu.Unlock()
	return append([]recordedRequest(nil), env.requests...)
}

// cliResult is what one run of the CLI produced. stderr includes what main
// would print for the returned error.
type cliResult struct {
	stdout string
	stderr string
	err    error
}

// runCLI runs the command tree the way main does, with separate stdout and
// stderr buffers and stdin from the given string (so it is never a terminal).
func runCLI(t *testing.T, stdin string, args ...string) cliResult {
	t.Helper()

	cmd := newRootCmd()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetIn(strings.NewReader(stdin))
	cmd.SetArgs(args)

	executed, err := cmd.ExecuteContextC(t.Context())
	if err != nil {
		printError(&stderr, executed, err)
	}

	return cliResult{stdout: stdout.String(), stderr: stderr.String(), err: err}
}

// trimLines removes trailing whitespace from every line, so tabwriter padding
// doesn't make golden comparisons fragile.
func trimLines(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t\r")
	}
	return strings.Join(lines, "\n")
}

// checkGolden compares got with testdata/<name>, after trimming trailing
// whitespace on each line. Run `go test -update` to rewrite the files.
func checkGolden(t *testing.T, name, got string) {
	t.Helper()

	path := filepath.Join("testdata", name)
	got = trimLines(got)

	if *update {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}

	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading golden file (run go test -update to create it): %v", err)
	}
	if got != trimLines(string(want)) {
		t.Errorf("output does not match %s\n--- got ---\n%s\n--- want ---\n%s", path, got, want)
	}
}

// defaultResponses is a small workspace. last_seen_at values are relative to
// now so the "Seen" column is stable (26 hours ago is always "1 day"); the
// returned replacer swaps them for placeholders in golden output.
func defaultResponses(now time.Time) (map[string]apiResponse, *strings.Replacer) {
	seenOneDay := now.Add(-26 * time.Hour).UTC().Format(time.RFC3339)
	seenTenDays := now.Add(-10*24*time.Hour - time.Hour).UTC().Format(time.RFC3339)
	placeholders := strings.NewReplacer(seenOneDay, "<SEEN 26 HOURS AGO>", seenTenDays, "<SEEN 10 DAYS 1 HOUR AGO>")

	return map[string]apiResponse{
		"GET /devices.json": {body: `[
			{"serial_number":"ABC123","mdm_user_or_device_name":"Alex Example","device_type":"MacBook Pro","os_version":"macOS 15.1","last_seen_at":"` + seenOneDay + `"},
			{"serial_number":"ABC1234","mdm_user_or_device_name":"Alexandra Long-Name Person","device_type":"Mac mini","os_version":"macOS 14.7","last_seen_at":"` + seenTenDays + `"},
			{"serial_number":"XYZ999","mdm_user_or_device_name":"Sam","device_type":"iMac","os_version":"","last_seen_at":""}
		]`},
		"GET /formulae.json": {body: `[
			{"name":"wget","homebrew_core_version":"1.24.5","outdated":true,"devices":["ABC123"]},
			{"name":"git","homebrew_core_version":"2.47.0","outdated":false,"devices":["ABC123","XYZ999"]},
			{"name":"jq","homebrew_core_version":"1.7.1","outdated":true,"devices":["XYZ999"]}
		]`},
		"GET /casks.json": {body: `[
			{"name":"iterm2","homebrew_cask_version":"3.5.4","outdated":false,"devices":["ABC123"]},
			{"name":"firefox","homebrew_cask_version":"132.0","outdated":true,"devices":["ABC123"]}
		]`},
		"GET /analytics.json": {body: `[
			{"device":"ABC123","command":"brew upgrade","last_run":"2026-11-20","count":3},
			{"device":"XYZ999","command":"brew install","last_run":"2026-11-19","count":3},
			{"device":"ABC1234","command":"brew upgrade","last_run":"2026-11-18","count":3},
			{"device":"XYZ999","command":"brew doctor","last_run":"2026-11-17","count":9}
		]`},
		"GET /brewfiles.json": {body: `[
			{"label":"Base","slug":"base","run_count":4,"started_at":"2026-11-01","finished_at":"2026-11-02","devices":["ABC123","XYZ999"]}
		]`},
		"GET /brew_commands.json": {body: `[
			{"command":"brew update","last_updated_by_user":"admin","finished_at":"2026-11-20","run_count":2,"devices":["ABC123"]},
			{"command":"brew cleanup","last_updated_by_user":"admin","finished_at":"2026-11-19","run_count":2,"devices":[]},
			{"command":"brew upgrade","last_updated_by_user":"ops","finished_at":"2026-11-18","run_count":5,"devices":["ABC123","XYZ999"]}
		]`},
		"POST /brew_commands.json": {body: `{"message":"Brew command created"}`},
		"GET /vulnerabilities.json": {body: `[
			{"package":"openssl","display_name":"OpenSSL","package_type":"formula","affected_devices":["abc123","XYZ999"],
			 "vulnerabilities":[{"clean_id":"CVE-2026-0001","cvss_score":9.8},{"clean_id":"CVE-2026-0002","cvss_score":5.0}]},
			{"package":"curl","package_type":"formula","affected_devices":["ABC1234"],
			 "vulnerabilities":[{"clean_id":"CVE-2026-0003","cvss_score":7.5},{"clean_id":"CVE-2026-0004","cvss_score":null}]}
		]`},
	}, placeholders
}

func TestCommandsGolden(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		golden string
	}{
		{"devices list", []string{"devices", "list"}, "devices_list.txt"},
		{"report summary is devices list", []string{"report", "summary"}, "devices_list.txt"},
		{"devices list json", []string{"devices", "list", "-o", "json"}, "devices_list.json"},
		{"devices get", []string{"devices", "get", "abc123"}, "devices_get.txt"},
		{"devices get apps", []string{"devices", "get", "ABC123", "--apps"}, "devices_get_apps.txt"},
		{"devices get apps json", []string{"devices", "get", "abc123", "--apps", "-o", "json"}, "devices_get_apps.json"},
		{"devices get json", []string{"devices", "get", "sam", "-o", "json"}, "devices_get.json"},
		{"report outdated", []string{"report", "outdated"}, "report_outdated.txt"},
		{"report outdated json", []string{"report", "outdated", "-o", "json"}, "report_outdated.json"},
		{"report analytics", []string{"report", "analytics"}, "report_analytics.txt"},
		{"report analytics json", []string{"report", "analytics", "-o", "json"}, "report_analytics.json"},
		{"report brewfiles", []string{"report", "brewfiles"}, "report_brewfiles.txt"},
		{"report brew-commands", []string{"report", "brew-commands"}, "report_brew_commands.txt"},
		{"report brew-commands json", []string{"report", "brew-commands", "-o", "json"}, "report_brew_commands.json"},
		{"report vulnerabilities", []string{"report", "vulnerabilities"}, "report_vulnerabilities.txt"},
		{"report vulnerabilities filtered", []string{"report", "vulnerabilities", "--severity", "CRITICAL", "--device", "abc123"}, "report_vulnerabilities_filtered.txt"},
		{"report vulnerabilities json", []string{"report", "vulnerabilities", "--severity", "high", "-o", "json"}, "report_vulnerabilities.json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := newTestEnv(t)

			result := runCLI(t, "", tt.args...)
			if result.err != nil {
				t.Fatalf("error = %v\nstderr:\n%s", result.err, result.stderr)
			}
			if result.stderr != "" {
				t.Errorf("stderr = %q, want empty", result.stderr)
			}

			checkGolden(t, tt.golden, env.placeholders.Replace(result.stdout))
		})
	}
}

func TestCommands(t *testing.T) {
	tests := []struct {
		name  string
		setup func(env *testEnv)
		stdin string
		args  []string

		wantErr    string // substring of the returned error; empty means no error
		wantStdout string // exact stdout; checked unless anyStdout
		anyStdout  bool
		wantStderr []string // substrings that must all appear in stderr
	}{
		{
			name:       "empty JSON list is []",
			args:       []string{"report", "analytics", "-o", "json", "--device", "abc"},
			wantStdout: "[]\n",
		},
		{
			name:       "several devices match",
			args:       []string{"devices", "get", "ABC12"},
			wantErr:    `"ABC12" matches 2 devices`,
			wantStderr: []string{"ABC123", "ABC1234", "Alexandra Long-Name Person", "Error: \"ABC12\" matches 2 devices"},
		},
		{
			name:       "no device matches",
			args:       []string{"devices", "get", "nobody"},
			wantErr:    `no device found matching "nobody"`,
			wantStderr: []string{`Error: no device found matching "nobody"`},
		},
		{
			name:       "empty device search",
			args:       []string{"devices", "get", "  "},
			wantErr:    "device search value cannot be empty",
			wantStderr: []string{"Error: device search value cannot be empty"},
		},
		{
			name: "device name lookup failure warns and falls back to serials",
			setup: func(env *testEnv) {
				env.respond("GET /devices.json", http.StatusBadGateway, `{"error":"upstream"}`)
			},
			args:      []string{"report", "vulnerabilities"},
			anyStdout: true,
			wantStderr: []string{
				`warning: could not look up device names, showing serial numbers: workbrew API returned 502 Bad Gateway: {"error":"upstream"}`,
			},
		},
		{
			name: "API error includes body excerpt",
			setup: func(env *testEnv) {
				env.respond("GET /brewfiles.json", http.StatusUnauthorized, `{"error":"invalid token"}`)
			},
			args:       []string{"report", "brewfiles"},
			wantErr:    `getting brewfiles: workbrew API returned 401 Unauthorized: {"error":"invalid token"}`,
			wantStderr: []string{"Error: getting brewfiles: workbrew API returned 401 Unauthorized"},
		},
		{
			name:       "auth status with token",
			args:       []string{"auth", "status"},
			wantStdout: "API token found in keychain\nToken length: 10 characters\n",
		},
		{
			name:       "auth status without token",
			setup:      func(env *testEnv) { env.deleteToken() },
			args:       []string{"auth", "status"},
			wantErr:    "no API token found in keychain; run `workbrew-cli setup",
			wantStderr: []string{"Error: no API token found"},
		},
		{
			name:       "missing token before any API call",
			setup:      func(env *testEnv) { env.deleteToken() },
			args:       []string{"devices", "list"},
			wantErr:    "no API token found in keychain",
			wantStderr: []string{"Error: no API token found"},
		},
		{
			name:       "missing config",
			setup:      func(env *testEnv) { env.removeConfig() },
			args:       []string{"report", "outdated"},
			wantErr:    "no config found at",
			wantStderr: []string{"Error: no config found at", "run `workbrew-cli setup --url <workspace-url>`"},
		},
		{
			name:       "config without URL",
			setup:      func(env *testEnv) { env.writeConfig("url: \"\"\n") },
			args:       []string{"report", "outdated"},
			wantErr:    "has no workspace URL",
			wantStderr: []string{"Error: config"},
		},
		{
			name:       "auth test",
			args:       []string{"auth", "test"},
			anyStdout:  true,
			wantStderr: nil,
		},
		{
			name:       "command update with --yes",
			args:       []string{"command", "update", "--yes"},
			wantStdout: "Brew command created\n",
		},
		{
			name:       "command update without --yes and no terminal",
			args:       []string{"command", "update"},
			stdin:      "y\n",
			wantErr:    "stdin is not a terminal; pass --yes",
			wantStderr: []string{"Error: stdin is not a terminal; pass --yes"},
		},
		{
			name:       "setup rejects http before prompting",
			args:       []string{"setup", "--url", "http://acme.workbrew.com"},
			wantErr:    "must use https",
			wantStderr: []string{`Error: invalid --url "http://acme.workbrew.com": must use https`},
		},
		{
			name:       "setup needs a terminal for the token",
			args:       []string{"setup", "--url", "https://acme.workbrew.com"},
			wantErr:    "stdin is not a terminal",
			wantStderr: []string{"Error: reading API token: stdin is not a terminal"},
		},
		{
			name:       "invalid severity",
			args:       []string{"report", "vulnerabilities", "--severity", "urgent"},
			wantErr:    `invalid severity "urgent"`,
			wantStderr: []string{`Error: invalid severity "urgent"`},
		},
		{
			name:       "version",
			args:       []string{"version"},
			wantStdout: version + "\n",
		},
		{
			name:       "--version flag",
			args:       []string{"--version"},
			wantStdout: "workbrew-cli version " + version + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := newTestEnv(t)
			if tt.setup != nil {
				tt.setup(env)
			}

			result := runCLI(t, tt.stdin, tt.args...)

			switch {
			case tt.wantErr == "" && result.err != nil:
				t.Fatalf("error = %v, want nil\nstderr:\n%s", result.err, result.stderr)
			case tt.wantErr != "" && (result.err == nil || !strings.Contains(result.err.Error(), tt.wantErr)):
				t.Fatalf("error = %v, want containing %q", result.err, tt.wantErr)
			}

			if !tt.anyStdout && result.stdout != tt.wantStdout {
				t.Errorf("stdout =\n%s\nwant\n%s", result.stdout, tt.wantStdout)
			}

			if len(tt.wantStderr) == 0 && result.stderr != "" {
				t.Errorf("stderr = %q, want empty", result.stderr)
			}
			for _, want := range tt.wantStderr {
				if !strings.Contains(result.stderr, want) {
					t.Errorf("stderr missing %q\nstderr:\n%s", want, result.stderr)
				}
			}
		})
	}
}

func TestCommandSendsRequest(t *testing.T) {
	env := newTestEnv(t)

	result := runCLI(t, "", "command", "upgrade", "-y")
	if result.err != nil {
		t.Fatalf("error = %v", result.err)
	}

	var posts []recordedRequest
	for _, request := range env.recorded() {
		if request.method == http.MethodPost {
			posts = append(posts, request)
		}
	}
	if len(posts) != 1 {
		t.Fatalf("got %d POST requests, want 1", len(posts))
	}

	post := posts[0]
	if post.path != "/brew_commands.json" {
		t.Errorf("path = %s", post.path)
	}
	if post.body != `{"arguments":"upgrade"}` {
		t.Errorf("body = %s", post.body)
	}
	if got := post.header.Get("Authorization"); got != "Bearer test-token" {
		t.Errorf("Authorization = %q", got)
	}
}

func TestNoAPICallsForLocalCommands(t *testing.T) {
	env := newTestEnv(t)
	env.deleteToken()
	env.removeConfig()

	for _, args := range [][]string{{"version"}, {"--help"}, {"devices", "--help"}} {
		result := runCLI(t, "", args...)
		if result.err != nil {
			t.Errorf("%v: error = %v (should not need config or token)", args, result.err)
		}
	}

	if requests := env.recorded(); len(requests) != 0 {
		t.Errorf("made %d API requests, want 0", len(requests))
	}
}

func TestUsageErrors(t *testing.T) {
	const rootHint = "Run 'workbrew-cli --help' for usage."

	// Cobra writes usage to OutOrStderr(), which is the SetOut writer when one
	// is set, so in these tests usage lands in the stdout buffer. In the real
	// binary SetOut isn't called and usage goes to stderr.
	tests := []struct {
		name      string
		args      []string
		wantErr   string
		wantUsage bool
		wantHint  string // the "Run ... --help" line expected on stderr, if any
	}{
		{"unknown flag", []string{"devices", "list", "--bogus"}, "unknown flag: --bogus", true, ""},
		{"invalid output format", []string{"report", "outdated", "-o", "yaml"}, `invalid argument "yaml" for "-o, --output" flag: must be "table" or "json"`, true, ""},
		{"missing argument", []string{"devices", "get"}, "accepts 1 arg(s), received 0", true, ""},
		{"unexpected argument", []string{"devices", "list", "extra"}, `unknown command "extra" for "workbrew-cli devices list"`, true, ""},
		{"missing required flag", []string{"setup"}, `required flag(s) "url" not set`, true, ""},
		{"unknown subcommand", []string{"bogus"}, `unknown command "bogus" for "workbrew-cli"`, false, rootHint},
		{"unknown subcommand with suggestion", []string{"repot"}, "Did you mean this?\n\treport\n", false, rootHint},
		{"unknown devices subcommand", []string{"devices", "bogus"}, `unknown command "bogus" for "workbrew-cli devices"`, false, "Run 'workbrew-cli devices --help' for usage."},
		{"unknown report subcommand with typo suggestion", []string{"report", "vulnerabilites"}, "unknown command \"vulnerabilites\" for \"workbrew-cli report\"\n\nDid you mean this?\n\tvulnerabilities\n", false, "Run 'workbrew-cli report --help' for usage."},
		{"unknown report subcommand with prefix suggestions", []string{"report", "brew"}, "Did you mean this?\n\tbrewfiles\n\tbrew-commands\n", false, "Run 'workbrew-cli report --help' for usage."},
		{"unknown command subcommand", []string{"command", "install"}, `unknown command "install" for "workbrew-cli command"`, false, "Run 'workbrew-cli command --help' for usage."},
		{"unknown auth subcommand with suggestion", []string{"auth", "stauts"}, "Did you mean this?\n\tstatus\n", false, "Run 'workbrew-cli auth --help' for usage."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newTestEnv(t)

			result := runCLI(t, "", tt.args...)
			if result.err == nil || !strings.Contains(result.err.Error(), tt.wantErr) {
				t.Fatalf("error = %v, want containing %q", result.err, tt.wantErr)
			}

			if got := strings.Contains(result.stdout+result.stderr, "Usage:"); got != tt.wantUsage {
				t.Errorf("usage shown = %v, want %v\nstdout:\n%s\nstderr:\n%s", got, tt.wantUsage, result.stdout, result.stderr)
			}

			gotHint := strings.Contains(result.stderr, "--help' for usage.")
			switch {
			case tt.wantHint == "" && gotHint:
				t.Errorf("unexpected --help hint\nstderr:\n%s", result.stderr)
			case tt.wantHint != "" && !strings.HasSuffix(result.stderr, tt.wantHint+"\n"):
				t.Errorf("stderr should end with %q\nstderr:\n%s", tt.wantHint, result.stderr)
			}

			if !strings.Contains(result.stderr, "Error: ") || !strings.Contains(result.stderr, tt.wantErr) {
				t.Errorf("stderr missing \"Error: \" or %q\nstderr:\n%s", tt.wantErr, result.stderr)
			}
		})
	}
}

func TestRuntimeErrorsDoNotShowUsage(t *testing.T) {
	env := newTestEnv(t)
	env.respond("GET /devices.json", http.StatusInternalServerError, "")

	result := runCLI(t, "", "devices", "list")
	if result.err == nil {
		t.Fatal("error = nil, want API error")
	}
	if strings.Contains(result.stdout+result.stderr, "Usage:") {
		t.Errorf("usage shown for a runtime error\nstdout:\n%s\nstderr:\n%s", result.stdout, result.stderr)
	}
	if want := "Error: getting devices: workbrew API returned 500 Internal Server Error\n"; result.stderr != want {
		t.Errorf("stderr = %q, want %q", result.stderr, want)
	}
}

func TestBareCommandsShowHelp(t *testing.T) {
	tests := []struct {
		args []string
		want string
	}{
		{[]string{}, "workbrew-cli [command]"},
		{[]string{"devices"}, "workbrew-cli devices [command]"},
		{[]string{"report"}, "workbrew-cli report [command]"},
		{[]string{"command"}, "workbrew-cli command [command]"},
		{[]string{"auth"}, "workbrew-cli auth [command]"},
	}

	for _, tt := range tests {
		t.Run(strings.Join(append([]string{"workbrew-cli"}, tt.args...), " "), func(t *testing.T) {
			env := newTestEnv(t)

			result := runCLI(t, "", tt.args...)
			if result.err != nil {
				t.Fatalf("error = %v, want nil (exit 0)\nstderr:\n%s", result.err, result.stderr)
			}
			if !strings.Contains(result.stdout, "Available Commands:") || !strings.Contains(result.stdout, tt.want) {
				t.Errorf("stdout is not help for %v:\n%s", tt.args, result.stdout)
			}
			if result.stderr != "" {
				t.Errorf("stderr = %q, want empty", result.stderr)
			}
			if requests := env.recorded(); len(requests) != 0 {
				t.Errorf("made %d API requests, want 0", len(requests))
			}
		})
	}
}

func TestVulnerabilitiesFetchesDevicesOnlyForTable(t *testing.T) {
	for _, tt := range []struct {
		format      string
		wantDevices int
	}{
		{"table", 1},
		{"json", 0},
	} {
		t.Run(tt.format, func(t *testing.T) {
			env := newTestEnv(t)

			if result := runCLI(t, "", "report", "vulnerabilities", "-o", tt.format); result.err != nil {
				t.Fatalf("error = %v", result.err)
			}

			got := 0
			for _, request := range env.recorded() {
				if request.path == "/devices.json" {
					got++
				}
			}
			if got != tt.wantDevices {
				t.Errorf("devices requests = %d, want %d", got, tt.wantDevices)
			}
		})
	}
}

func TestStdoutWriteFailureIsAnError(t *testing.T) {
	for _, args := range [][]string{
		{"devices", "list"},                  // table output
		{"report", "outdated", "-o", "json"}, // JSON output
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			newTestEnv(t)

			cmd := newRootCmd()
			var stderr bytes.Buffer
			cmd.SetOut(&failingWriter{limit: 10})
			cmd.SetErr(&stderr)
			cmd.SetArgs(args)

			_, err := cmd.ExecuteContextC(t.Context())
			if !errors.Is(err, errDiskFull) {
				t.Fatalf("error = %v, want %v", err, errDiskFull)
			}
		})
	}
}
