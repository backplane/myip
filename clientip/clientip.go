// clientip remote-client-ip determination which handles complex issues around the X-Forwarded-For
// header, as described here: https://adam-p.ca/blog/2022/03/x-forwarded-for/
package clientip

import (
	"net"
	"net/http"
	"strings"
)

// TrustedProxies holds a list of trusted proxy networks.
type TrustedProxies struct {
	nets []*net.IPNet
}

// NewTrustedProxies parses a comma-separated list of CIDR blocks and returns a TrustedProxies struct.
// Invalid CIDRs are skipped.
func NewTrustedProxies(cidrList string) *TrustedProxies {
	var nets []*net.IPNet
	for _, cidr := range strings.Split(cidrList, ",") {
		cidr = strings.TrimSpace(cidr)
		if cidr == "" {
			continue
		}
		_, netw, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		nets = append(nets, netw)
	}
	return &TrustedProxies{nets: nets}
}

// IsTrusted returns true if ip is in any of the trusted proxy networks.
func (tp *TrustedProxies) IsTrusted(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	for _, netw := range tp.nets {
		if netw.Contains(parsedIP) {
			return true
		}
	}
	return false
}

// GetClientIP determines the real client IP address from an HTTP request,
// following best practices for X-Forwarded-For handling.
//
//   - If trustXFF is false, always use the direct remote address.
//   - If trustXFF is true, concatenate all X-Forwarded-For header values,
//     walk from right to left, skipping over any IPs in trustedProxies, and
//     return the first untrusted IP. If all IPs are trusted, fall back to RemoteAddr.
//   - The remote address's host part is always extracted with net.SplitHostPort
//     if possible.
//
// This function is suitable for use in environments where your app is only reachable
// via trusted proxies. Never set trustXFF=true if your app is internet-facing.
func GetClientIP(req *http.Request, trustXFF bool, trustedProxies *TrustedProxies) string {
	if trustXFF && trustedProxies != nil {
		xffs := req.Header.Values("X-Forwarded-For")
		var ips []string
		if len(xffs) > 0 {
			joined := strings.Join(xffs, ",")
			split := strings.Split(joined, ",")
			for _, ip := range split {
				ip = strings.TrimSpace(ip)
				if ip != "" {
					ips = append(ips, ip)
				}
			}
			// Walk from right to left, skipping trusted proxies
			for i := len(ips) - 1; i >= 0; i-- {
				ip := ips[i]
				if !trustedProxies.IsTrusted(ip) {
					return ip
				}
			}
		}
		// All XFF IPs are trusted, fall back to RemoteAddr
	}

	// Fallback: extract host (IP) from RemoteAddr
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return req.RemoteAddr
}
