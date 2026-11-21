package main

import (
	"strings"
	"testing"
)

func TestValidateWorkspaceURL(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr string
	}{
		{name: "host", in: "https://acme.workbrew.com", want: "https://acme.workbrew.com"},
		{name: "trailing slash", in: "https://acme.workbrew.com/", want: "https://acme.workbrew.com"},
		{name: "several trailing slashes", in: "https://acme.workbrew.com///", want: "https://acme.workbrew.com"},
		{name: "path", in: "https://console.workbrew.com/workspaces/acme/", want: "https://console.workbrew.com/workspaces/acme"},
		{name: "port", in: "https://acme.workbrew.com:8443", want: "https://acme.workbrew.com:8443"},
		{name: "surrounding whitespace", in: "  https://acme.workbrew.com/ \n", want: "https://acme.workbrew.com"},
		{name: "upper-case scheme", in: "HTTPS://acme.workbrew.com", want: "HTTPS://acme.workbrew.com"},
		{name: "http", in: "http://acme.workbrew.com", wantErr: "must use https"},
		{name: "no scheme", in: "acme.workbrew.com", wantErr: "must use https"},
		{name: "other scheme", in: "ftp://acme.workbrew.com", wantErr: "must use https"},
		{name: "no host", in: "https://", wantErr: "missing host"},
		{name: "path but no host", in: "https:///workspaces/acme", wantErr: "missing host"},
		{name: "port but no host", in: "https://:443", wantErr: "missing host"},
		{name: "empty", in: "", wantErr: "must use https"},
		{name: "unparsable", in: "https://acme.workbrew.com/%zz", wantErr: "invalid --url"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateWorkspaceURL(tt.in)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("validateWorkspaceURL(%q) error = %v, want containing %q", tt.in, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("validateWorkspaceURL(%q) error = %v", tt.in, err)
			}
			if got != tt.want {
				t.Errorf("validateWorkspaceURL(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
