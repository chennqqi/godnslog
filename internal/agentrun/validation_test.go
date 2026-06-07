package agentrun

import (
	"context"
	"fmt"
	"net"
	"testing"
)

func TestValidateWebhookURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantError bool
	}{
		{
			name:      "valid https URL with IP",
			url:       "https://8.8.8.8/review",
			wantError: false,
		},
		{
			name:      "empty URL",
			url:       "",
			wantError: true,
		},
		{
			name:      "http URL rejected",
			url:       "http://hooks.example.com/review",
			wantError: true,
		},
		{
			name:      "localhost rejected",
			url:       "https://localhost:8080/hook",
			wantError: true,
		},
		{
			name:      "127.0.0.1 rejected",
			url:       "https://127.0.0.1:8080/hook",
			wantError: true,
		},
		{
			name:      "::1 rejected",
			url:       "https://[::1]:8080/hook",
			wantError: true,
		},
		{
			name:      "private IP 10.0.0.0/8 rejected",
			url:       "https://10.0.0.1/hook",
			wantError: true,
		},
		{
			name:      "private IP 172.16.0.0/12 rejected",
			url:       "https://172.16.0.1/hook",
			wantError: true,
		},
		{
			name:      "private IP 192.168.0.0/16 rejected",
			url:       "https://192.168.1.1/hook",
			wantError: true,
		},
		{
			name:      "AWS metadata IP rejected",
			url:       "https://169.254.169.254/hook",
			wantError: true,
		},
		{
			name:      "GCP metadata hostname rejected",
			url:       "https://metadata.google.internal/hook",
			wantError: true,
		},
		{
			name:      "invalid URL format",
			url:       "not-a-url",
			wantError: true,
		},
		{
			name:      "URL without hostname",
			url:       "https://",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWebhookURL(tt.url)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateWebhookURL() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidateWebhookURLWithResolver(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		resolver  IPResolver
		wantError bool
	}{
		{
			name: "hostname resolves to public IP - allowed",
			url:  "https://safe.example.com/hook",
			resolver: func(ctx context.Context, hostname string) ([]net.IP, error) {
				return []net.IP{net.ParseIP("8.8.8.8")}, nil
			},
			wantError: false,
		},
		{
			name: "hostname resolves to private IP - rejected",
			url:  "https://internal.example.com/hook",
			resolver: func(ctx context.Context, hostname string) ([]net.IP, error) {
				return []net.IP{net.ParseIP("10.0.0.1")}, nil
			},
			wantError: true,
		},
		{
			name: "hostname resolves to metadata IP - rejected",
			url:  "https://metadata.example.com/hook",
			resolver: func(ctx context.Context, hostname string) ([]net.IP, error) {
				return []net.IP{net.ParseIP("169.254.169.254")}, nil
			},
			wantError: true,
		},
		{
			name: "hostname resolves to localhost - rejected",
			url:  "https://local.example.com/hook",
			resolver: func(ctx context.Context, hostname string) ([]net.IP, error) {
				return []net.IP{net.ParseIP("127.0.0.1")}, nil
			},
			wantError: true,
		},
		{
			name: "hostname resolves to link-local - rejected",
			url:  "https://link.example.com/hook",
			resolver: func(ctx context.Context, hostname string) ([]net.IP, error) {
				return []net.IP{net.ParseIP("169.254.1.1")}, nil
			},
			wantError: true,
		},
		{
			name: "hostname resolves to multiple IPs, one private - rejected",
			url:  "https://multi.example.com/hook",
			resolver: func(ctx context.Context, hostname string) ([]net.IP, error) {
				return []net.IP{net.ParseIP("8.8.8.8"), net.ParseIP("10.0.0.1")}, nil
			},
			wantError: true,
		},
		{
			name: "resolver fails - fail-closed",
			url:  "https://safe.example.com/hook",
			resolver: func(ctx context.Context, hostname string) ([]net.IP, error) {
				return nil, fmt.Errorf("resolver error")
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWebhookURLWithResolver(context.Background(), tt.url, tt.resolver)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateWebhookURLWithResolver() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidateWebhookHeaders(t *testing.T) {
	tests := []struct {
		name      string
		headers   map[string]string
		wantError bool
	}{
		{
			name:      "valid Content-Type header",
			headers:   map[string]string{"Content-Type": "application/json"},
			wantError: false,
		},
		{
			name:      "valid X-* header",
			headers:   map[string]string{"X-GODNSLOG-Source": "operator"},
			wantError: false,
		},
		{
			name:      "multiple valid headers",
			headers:   map[string]string{"Content-Type": "application/json", "X-GODNSLOG-Source": "operator"},
			wantError: false,
		},
		{
			name:      "empty headers",
			headers:   map[string]string{},
			wantError: false,
		},
		{
			name:      "nil headers",
			headers:   nil,
			wantError: false,
		},
		{
			name:      "Authorization header rejected",
			headers:   map[string]string{"Authorization": "Bearer token"},
			wantError: true,
		},
		{
			name:      "Cookie header rejected",
			headers:   map[string]string{"Cookie": "session=abc"},
			wantError: true,
		},
		{
			name:      "Set-Cookie header rejected",
			headers:   map[string]string{"Set-Cookie": "session=abc"},
			wantError: true,
		},
		{
			name:      "Proxy-Authorization header rejected",
			headers:   map[string]string{"Proxy-Authorization": "Basic token"},
			wantError: true,
		},
		{
			name:      "non-X-* custom header rejected",
			headers:   map[string]string{"Custom-Header": "value"},
			wantError: true,
		},
		{
			name:      "hop-by-hop header TE rejected",
			headers:   map[string]string{"TE": "trailers"},
			wantError: true,
		},
		{
			name:      "hop-by-hop header Trailer rejected",
			headers:   map[string]string{"Trailer": "custom"},
			wantError: true,
		},
		{
			name:      "hop-by-hop header Transfer-Encoding rejected",
			headers:   map[string]string{"Transfer-Encoding": "chunked"},
			wantError: true,
		},
		{
			name:      "hop-by-hop header Upgrade rejected",
			headers:   map[string]string{"Upgrade": "websocket"},
			wantError: true,
		},
		{
			name:      "case insensitive header check",
			headers:   map[string]string{"authorization": "Bearer token"},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWebhookHeaders(tt.headers)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateWebhookHeaders() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}
