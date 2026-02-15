package proxy

import (
	"net/http"
	"testing"
)

func TestParseHost(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		wantPort int
		wantErr  bool
	}{
		{
			name:     "valid localhost port 8000",
			host:     "8000.localhost",
			wantPort: 8000,
			wantErr:  false,
		},
		{
			name:     "valid localhost port 3000",
			host:     "3000.localhost",
			wantPort: 3000,
			wantErr:  false,
		},
		{
			name:     "valid localhost port 5173",
			host:     "5173.localhost",
			wantPort: 5173,
			wantErr:  false,
		},
		{
			name:     "valid localtest.me",
			host:     "8080.localtest.me",
			wantPort: 8080,
			wantErr:  false,
		},
		{
			name:     "valid with explicit port",
			host:     "8000.localhost:443",
			wantPort: 8000,
			wantErr:  false,
		},
		{
			name:     "port 1 (minimum)",
			host:     "1.localhost",
			wantPort: 1,
			wantErr:  false,
		},
		{
			name:     "port 65535 (maximum)",
			host:     "65535.localhost",
			wantPort: 65535,
			wantErr:  false,
		},
		{
			name:    "invalid - no port",
			host:    "localhost",
			wantErr: true,
		},
		{
			name:    "invalid - plain domain",
			host:    "example.com",
			wantErr: true,
		},
		{
			name:    "invalid - non-numeric port",
			host:    "abc.localhost",
			wantErr: true,
		},
		{
			name:    "invalid - port 0",
			host:    "0.localhost",
			wantErr: true,
		},
		{
			name:    "invalid - port too high",
			host:    "65536.localhost",
			wantErr: true,
		},
		{
			name:    "invalid - negative port",
			host:    "-1.localhost",
			wantErr: true,
		},
		{
			name:    "invalid - wrong domain",
			host:    "8000.example.com",
			wantErr: true,
		},
		{
			name:    "invalid - subdomain format",
			host:    "app.8000.localhost",
			wantErr: true,
		},
	}

	s := &Server{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port, err := s.parseHost(tt.host)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseHost(%q) error = %v, wantErr %v", tt.host, err, tt.wantErr)
				return
			}
			if !tt.wantErr && port != tt.wantPort {
				t.Errorf("parseHost(%q) = %d, want %d", tt.host, port, tt.wantPort)
			}
		})
	}
}

func TestIsWebSocketRequest(t *testing.T) {
	tests := []struct {
		name       string
		connection string
		upgrade    string
		want       bool
	}{
		{
			name:       "valid websocket upgrade",
			connection: "Upgrade",
			upgrade:    "websocket",
			want:       true,
		},
		{
			name:       "valid websocket upgrade lowercase",
			connection: "upgrade",
			upgrade:    "websocket",
			want:       true,
		},
		{
			name:       "valid with keep-alive",
			connection: "keep-alive, Upgrade",
			upgrade:    "websocket",
			want:       true,
		},
		{
			name:       "missing upgrade header",
			connection: "Upgrade",
			upgrade:    "",
			want:       false,
		},
		{
			name:       "missing connection header",
			connection: "",
			upgrade:    "websocket",
			want:       false,
		},
		{
			name:       "wrong upgrade type",
			connection: "Upgrade",
			upgrade:    "h2c",
			want:       false,
		},
		{
			name:       "no upgrade",
			connection: "keep-alive",
			upgrade:    "",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &http.Request{
				Header: make(http.Header),
			}
			if tt.connection != "" {
				req.Header.Set("Connection", tt.connection)
			}
			if tt.upgrade != "" {
				req.Header.Set("Upgrade", tt.upgrade)
			}

			if got := isWebSocketRequest(req); got != tt.want {
				t.Errorf("isWebSocketRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHostPatternRegex(t *testing.T) {
	tests := []struct {
		host  string
		match bool
		port  string
	}{
		{"8000.localhost", true, "8000"},
		{"3000.localhost", true, "3000"},
		{"80.localhost", true, "80"},
		{"8080.localtest.me", true, "8080"},
		{"5173.localtest.me", true, "5173"},
		{"8000.localhost:443", true, "8000"},
		{"localhost", false, ""},
		{"example.com", false, ""},
		{"abc.localhost", false, ""},
		{"8000.example.com", false, ""},
		{".localhost", false, ""},
		{"8000.", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			matches := hostPattern.FindStringSubmatch(tt.host)
			if (matches != nil) != tt.match {
				t.Errorf("hostPattern.FindStringSubmatch(%q) match = %v, want %v", tt.host, matches != nil, tt.match)
			}
			if tt.match && matches[1] != tt.port {
				t.Errorf("hostPattern.FindStringSubmatch(%q) port = %q, want %q", tt.host, matches[1], tt.port)
			}
		})
	}
}
