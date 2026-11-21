package workbrew

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// writeBody writes a fake API response body, failing the test if it can't.
func writeBody(t *testing.T, w io.Writer, body string) {
	t.Helper()
	if _, err := io.WriteString(w, body); err != nil {
		t.Errorf("writing response body: %v", err)
	}
}

// newTestClient starts an httptest server running handler and returns a
// client pointed at it.
func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	client, err := New(server.URL, "test-token")
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	return client
}

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		token   string
		wantErr string
	}{
		{name: "valid", baseURL: "https://acme.workbrew.com/workspaces/acme", token: "t"},
		{name: "missing scheme", baseURL: "acme.workbrew.com", token: "t", wantErr: "missing scheme or host"},
		{name: "missing host", baseURL: "https://", token: "t", wantErr: "missing scheme or host"},
		{name: "unparsable", baseURL: "://bad", token: "t", wantErr: "invalid workspace URL"},
		{name: "empty token", baseURL: "https://acme.workbrew.com", token: "", wantErr: "API token is empty"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(tt.baseURL, tt.token)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("New() error = %v", err)
				}
				if client.http.Timeout != defaultTimeout {
					t.Errorf("timeout = %v, want %v", client.http.Timeout, defaultTimeout)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("New() error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestGetSendsHeaders(t *testing.T) {
	var got http.Header
	var method string
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Clone()
		method = r.Method
		writeBody(t, w, "[]")
	})

	if _, err := client.Devices(t.Context()); err != nil {
		t.Fatalf("Devices: %v", err)
	}

	if method != http.MethodGet {
		t.Errorf("method = %s, want GET", method)
	}

	want := map[string]string{
		"Authorization":          "Bearer test-token",
		"X-Workbrew-Api-Version": "v0",
		"Accept":                 "application/json",
	}
	for header, value := range want {
		if got.Get(header) != value {
			t.Errorf("%s = %q, want %q", header, got.Get(header), value)
		}
	}
	if ct := got.Get("Content-Type"); ct != "" {
		t.Errorf("GET sent Content-Type %q, want none", ct)
	}
}

func TestCreateBrewCommand(t *testing.T) {
	var method, contentType, auth string
	var body BrewCommandCreateRequest
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		contentType = r.Header.Get("Content-Type")
		auth = r.Header.Get("Authorization")
		if r.URL.Path != "/brew_commands.json" {
			t.Errorf("path = %s, want /brew_commands.json", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decoding request body: %v", err)
		}
		writeBody(t, w, `{"message":"Brew command created"}`)
	})

	result, err := client.CreateBrewCommand(t.Context(), "upgrade")
	if err != nil {
		t.Fatalf("CreateBrewCommand: %v", err)
	}

	if method != http.MethodPost {
		t.Errorf("method = %s, want POST", method)
	}
	if contentType != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", contentType)
	}
	if auth != "Bearer test-token" {
		t.Errorf("Authorization = %q", auth)
	}
	if body.Arguments != "upgrade" {
		t.Errorf("arguments = %q, want upgrade", body.Arguments)
	}
	if result.Message != "Brew command created" {
		t.Errorf("message = %q", result.Message)
	}
}

func TestURLJoining(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		writeBody(t, w, "[]")
	}))
	t.Cleanup(server.Close)

	tests := []struct {
		name     string
		base     string
		wantPath string
	}{
		{name: "host only", base: server.URL, wantPath: "/devices.json"},
		{name: "host with trailing slash", base: server.URL + "/", wantPath: "/devices.json"},
		{name: "path prefix", base: server.URL + "/workspaces/acme", wantPath: "/workspaces/acme/devices.json"},
		{name: "path prefix with trailing slash", base: server.URL + "/workspaces/acme/", wantPath: "/workspaces/acme/devices.json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(tt.base, "test-token")
			if err != nil {
				t.Fatalf("New: %v", err)
			}

			gotPath = ""
			if _, err := client.Devices(t.Context()); err != nil {
				t.Fatalf("Devices: %v", err)
			}
			if gotPath != tt.wantPath {
				t.Errorf("path = %q, want %q", gotPath, tt.wantPath)
			}
		})
	}
}

func TestNon2xxErrors(t *testing.T) {
	long := strings.Repeat("x", 600)

	tests := []struct {
		name    string
		status  int
		body    string
		wantErr string
	}{
		{
			name:    "body excerpt is trimmed",
			status:  http.StatusUnauthorized,
			body:    "  {\"error\":\"bad token\"}\n\n",
			wantErr: `workbrew API returned 401 Unauthorized: {"error":"bad token"}`,
		},
		{
			name:    "body is cut at 512 bytes",
			status:  http.StatusBadGateway,
			body:    long,
			wantErr: "workbrew API returned 502 Bad Gateway: " + long[:512],
		},
		{
			name:    "empty body has no trailing colon",
			status:  http.StatusNotFound,
			body:    "",
			wantErr: "workbrew API returned 404 Not Found",
		},
		{
			name:    "whitespace-only body has no trailing colon",
			status:  http.StatusInternalServerError,
			body:    " \n\t ",
			wantErr: "workbrew API returned 500 Internal Server Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
				writeBody(t, w, tt.body)
			})

			_, err := client.Devices(t.Context())
			if err == nil {
				t.Fatal("Devices() error = nil, want error")
			}
			if err.Error() != tt.wantErr {
				t.Errorf("error =\n  %q\nwant\n  %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestDecodeError(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeBody(t, w, "<html>not json</html>")
	})

	_, err := client.Casks(t.Context())
	if err == nil || !strings.Contains(err.Error(), "decoding casks.json response") {
		t.Fatalf("Casks() error = %v, want decoding error", err)
	}
}

// waitForCancel blocks the handler until the client goes away, so the server
// can shut down cleanly.
func waitForCancel(arrived chan<- struct{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if arrived != nil {
			close(arrived)
		}
		select {
		case <-r.Context().Done():
		case <-time.After(10 * time.Second):
		}
	}
}

func TestTimeout(t *testing.T) {
	client := newTestClient(t, waitForCancel(nil))
	client.http.Timeout = 50 * time.Millisecond

	start := time.Now()
	_, err := client.Devices(t.Context())
	if err == nil {
		t.Fatal("Devices() error = nil, want timeout")
	}

	var netErr net.Error
	if !errors.As(err, &netErr) || !netErr.Timeout() {
		t.Errorf("error = %v, want a timeout", err)
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Errorf("timed out after %v, want about 50ms", elapsed)
	}
}

func TestContextCancel(t *testing.T) {
	arrived := make(chan struct{})
	client := newTestClient(t, waitForCancel(arrived))

	ctx, cancel := context.WithCancel(t.Context())
	go func() {
		<-arrived
		cancel()
	}()

	_, err := client.Devices(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Devices() error = %v, want context.Canceled", err)
	}
}

func TestEmptyLists(t *testing.T) {
	for _, body := range []string{"[]", "null"} {
		t.Run(body, func(t *testing.T) {
			client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				writeBody(t, w, body)
			})

			devices, err := client.Devices(t.Context())
			if err != nil {
				t.Fatalf("Devices: %v", err)
			}
			if devices == nil || len(devices) != 0 {
				t.Fatalf("Devices() = %#v, want empty non-nil slice", devices)
			}

			encoded, err := json.Marshal(devices)
			if err != nil {
				t.Fatal(err)
			}
			if string(encoded) != "[]" {
				t.Errorf("JSON = %s, want []", encoded)
			}
		})
	}
}

func TestTypedListMethodsUseTheirEndpoints(t *testing.T) {
	var gotPath string
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		writeBody(t, w, "[]")
	})

	ctx := t.Context()
	calls := map[string]func() error{
		"/devices.json":         func() error { _, err := client.Devices(ctx); return err },
		"/formulae.json":        func() error { _, err := client.Formulae(ctx); return err },
		"/casks.json":           func() error { _, err := client.Casks(ctx); return err },
		"/vulnerabilities.json": func() error { _, err := client.Vulnerabilities(ctx); return err },
		"/analytics.json":       func() error { _, err := client.Analytics(ctx); return err },
		"/brewfiles.json":       func() error { _, err := client.Brewfiles(ctx); return err },
		"/brew_commands.json":   func() error { _, err := client.BrewCommands(ctx); return err },
	}

	for wantPath, call := range calls {
		if err := call(); err != nil {
			t.Fatalf("%s: %v", wantPath, err)
		}
		if gotPath != wantPath {
			t.Errorf("path = %s, want %s", gotPath, wantPath)
		}
	}
}
