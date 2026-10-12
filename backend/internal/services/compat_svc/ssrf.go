package compat_svc

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"syscall"
	"time"
)

// Download limits for one mod jar.
const (
	// MaxJarSize is the largest jar that is downloaded. It is checked against
	// Content-Length up front and against the bytes actually read.
	MaxJarSize int64 = 150 << 20
	// DownloadTimeout bounds one jar download, including connecting and
	// following redirects.
	DownloadTimeout = 60 * time.Second
	// maxRedirects is how many redirects a download may follow.
	maxRedirects = 5
	// dialTimeout bounds connecting to one address.
	dialTimeout = 10 * time.Second
)

// errBlockedAddress is returned when a download would connect to an address
// that is not a public internet host.
var errBlockedAddress = errors.New("address is not allowed")

// blockedPrefixes are ranges netip's own predicates do not cover.
var blockedPrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),          // "this network"
	netip.MustParsePrefix("100.64.0.0/10"),      // carrier-grade NAT
	netip.MustParsePrefix("192.0.0.0/24"),       // IETF protocol assignments
	netip.MustParsePrefix("192.0.2.0/24"),       // documentation (TEST-NET-1)
	netip.MustParsePrefix("198.51.100.0/24"),    // documentation (TEST-NET-2)
	netip.MustParsePrefix("203.0.113.0/24"),     // documentation (TEST-NET-3)
	netip.MustParsePrefix("2002::/16"),          // 6to4: embeds an IPv4 address
	netip.MustParsePrefix("2001::/32"),          // Teredo: embeds an IPv4 address
	netip.MustParsePrefix("198.18.0.0/15"),      // benchmarking
	netip.MustParsePrefix("240.0.0.0/4"),        // reserved
	netip.MustParsePrefix("64:ff9b::/96"),       // NAT64: embeds an IPv4 address
	netip.MustParsePrefix("fec0::/10"),          // deprecated site-local
	netip.MustParsePrefix("2001:db8::/32"),      // documentation
	netip.MustParsePrefix("100::/64"),           // discard-only
	netip.MustParsePrefix("255.255.255.255/32"), // broadcast
}

// isBlockedAddr reports whether a download must not connect to addr: loopback,
// private (RFC 1918, ULA), link-local, unspecified, multicast and a few other
// non-public ranges. IPv4-mapped IPv6 addresses are judged as their IPv4 form.
func isBlockedAddr(addr netip.Addr) bool {
	addr = addr.Unmap().WithZone("")
	if !addr.IsValid() {
		return true
	}
	if addr.IsLoopback() || addr.IsPrivate() || addr.IsUnspecified() ||
		addr.IsLinkLocalUnicast() || addr.IsLinkLocalMulticast() ||
		addr.IsInterfaceLocalMulticast() || addr.IsMulticast() {
		return true
	}
	for _, p := range blockedPrefixes {
		if p.Contains(addr) {
			return true
		}
	}
	return false
}

// guardControl returns a net.Dialer.Control hook that rejects the connection
// when the already resolved address is blocked. Control runs after DNS
// resolution and before connect, so it sees the IP actually dialed, which
// defeats DNS rebinding and hostnames that resolve to internal addresses.
// allowLoopback exists for tests that talk to httptest servers.
func guardControl(allowLoopback bool) func(network, address string, _ syscall.RawConn) error {
	return func(_, address string, _ syscall.RawConn) error {
		ap, err := netip.ParseAddrPort(address)
		if err != nil {
			return fmt.Errorf("%w: %q", errBlockedAddress, address)
		}
		if allowLoopback && ap.Addr().Unmap().IsLoopback() {
			return nil
		}
		if isBlockedAddr(ap.Addr()) {
			return fmt.Errorf("%w: %s", errBlockedAddress, ap.Addr())
		}
		return nil
	}
}

// checkRedirect allows at most maxRedirects redirects, to http(s) only.
func checkRedirect(req *http.Request, via []*http.Request) error {
	if len(via) > maxRedirects {
		return fmt.Errorf("stopped after %d redirects", maxRedirects)
	}
	return checkScheme(req.URL.Scheme)
}

func checkScheme(scheme string) error {
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("unsupported URL scheme %q", scheme)
	}
	return nil
}

// newGuardedClient builds the HTTP client used for mod download URLs, which are
// user data. It never uses a proxy (a proxy would connect on our behalf and
// bypass the address check).
func newGuardedClient(allowLoopback bool) *http.Client {
	dialer := &net.Dialer{Timeout: dialTimeout, Control: guardControl(allowLoopback)}
	return &http.Client{
		Timeout:       DownloadTimeout,
		CheckRedirect: checkRedirect,
		Transport: &http.Transport{
			Proxy:                 nil,
			DialContext:           dialer.DialContext,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 30 * time.Second,
			MaxIdleConns:          4,
			IdleConnTimeout:       30 * time.Second,
		},
	}
}
