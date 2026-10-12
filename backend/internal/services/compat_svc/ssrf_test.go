package compat_svc

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"testing"
)

func TestIsBlockedAddr(t *testing.T) {
	tests := []struct {
		addr    string
		blocked bool
	}{
		{"127.0.0.1", true},
		{"127.8.9.10", true},
		{"::1", true},
		{"10.0.0.1", true},
		{"172.16.5.4", true},
		{"172.31.255.255", true},
		{"192.168.1.1", true},
		{"fd00::1", true},
		{"fc00::1", true},
		{"169.254.169.254", true}, // cloud metadata
		{"fe80::1", true},
		{"0.0.0.0", true},
		{"0.1.2.3", true},
		{"::", true},
		{"224.0.0.1", true},
		{"ff02::1", true},
		{"100.64.0.1", true},
		{"255.255.255.255", true},
		{"240.0.0.1", true},
		{"::ffff:127.0.0.1", true}, // IPv4-mapped
		{"::ffff:10.1.1.1", true},
		{"::ffff:8.8.8.8", false},
		{"64:ff9b::7f00:1", true},
		{"192.0.2.10", true},                           // TEST-NET-1
		{"198.51.100.7", true},                         // TEST-NET-2
		{"203.0.113.99", true},                         // TEST-NET-3
		{"2002:7f00:1::1", true},                       // 6to4 embedding 127.0.0.1
		{"2001:0:4136:e378:8000:63bf:3fff:fdd2", true}, // Teredo
		{"2001:db8::1", true},
		{"192.0.3.1", false},
		{"2001:1::1", false},
		{"2003::1", false},
		{"172.15.0.1", false},
		{"172.32.0.1", false},
		{"8.8.8.8", false},
		{"1.1.1.1", false},
		{"93.184.216.34", false},
		{"2606:4700:4700::1111", false},
	}
	for _, tt := range tests {
		t.Run(tt.addr, func(t *testing.T) {
			if got := isBlockedAddr(netip.MustParseAddr(tt.addr)); got != tt.blocked {
				t.Errorf("isBlockedAddr(%s) = %v, want %v", tt.addr, got, tt.blocked)
			}
		})
	}
	if !isBlockedAddr(netip.Addr{}) {
		t.Error("invalid address must be blocked")
	}
}

func TestGuardControl(t *testing.T) {
	tests := []struct {
		name          string
		address       string
		allowLoopback bool
		wantErr       bool
	}{
		{"loopback blocked", "127.0.0.1:80", false, true},
		{"loopback allowed in tests", "127.0.0.1:80", true, false},
		{"private not allowed by loopback option", "10.0.0.1:80", true, true},
		{"public", "8.8.8.8:443", false, false},
		{"hostname rejected", "example.com:80", false, true},
		{"ipv6 loopback", "[::1]:80", false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := guardControl(tt.allowLoopback)("tcp", tt.address, nil)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !errors.Is(err, errBlockedAddress) {
				t.Errorf("err = %v, want errBlockedAddress", err)
			}
		})
	}
}

func TestGuardedClientBlocksLoopbackByDefault(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("secret"))
	}))
	defer srv.Close()

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL, nil)
	resp, err := newGuardedClient(false).Do(req)
	if err == nil {
		_ = resp.Body.Close()
		t.Fatal("request to loopback succeeded")
	}
	if !strings.Contains(err.Error(), errBlockedAddress.Error()) {
		t.Errorf("err = %v, want blocked address", err)
	}

	resp, err = newGuardedClient(true).Do(req)
	if err != nil {
		t.Fatalf("test client: %v", err)
	}
	_ = resp.Body.Close()
}

func TestCheckRedirect(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/loop", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/loop", http.StatusFound)
	})
	mux.HandleFunc("/ftp", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "ftp://example.com/x", http.StatusFound)
	})
	mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/done", http.StatusFound)
	})
	mux.HandleFunc("/done", func(w http.ResponseWriter, _ *http.Request) {})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	client := newGuardedClient(true)

	tests := []struct {
		path    string
		wantErr string
	}{
		{"/loop", "stopped after 5 redirects"},
		{"/ftp", "unsupported URL scheme"},
		{"/ok", ""},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+tt.path, nil)
			resp, err := client.Do(req)
			if resp != nil {
				_ = resp.Body.Close()
			}
			switch {
			case tt.wantErr == "" && err != nil:
				t.Fatalf("unexpected error: %v", err)
			case tt.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErr)):
				t.Fatalf("err = %v, want %q", err, tt.wantErr)
			}
		})
	}
}
