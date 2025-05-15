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
func NewTrustedProxies(cidrList string) *TrustedProxies {
	var nets []*net.IPNet
	for cidr := range strings.SplitSeq(cidrList, ",") {
		cidr = strings.TrimSpace(cidr)
		if cidr == "" {
			continue
		}
		if !strings.Contains(cidr, "/") {
			cidr = cidr + "/32"
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

func (tp *TrustedProxies) String() string {
	var proxies []string
	for _, net := range tp.nets {
		proxies = append(proxies, net.String())
	}
	return "[" + strings.Join(proxies, ", ") + "]"
}

// FlattenDelimitedInputs processes a slice of multiple delimited-value strings by splitting them on the delimiter,
// trimming whitespace, removing empty strings, and removing duplicates while preserving the original order. An
// empty separator results in a split between every utf-8 character. The result is a single slice of strings. For
// example, given:
// ["1.1.1.1, "2.2.2.2, 3.3.3.3, 4.4.4.4", "4.4.4.4, 5.5.5.5"]
// return:
// ["1.1.1.1", "2.2.2.2", "3.3.3.3", "4.4.4.4", "5.5.5.5"]
func FlattenDelimitedInputs(input []string, sep string) []string {
	seen := make(map[string]struct{})
	var result []string

	for _, delimitedStr := range input {
		for str := range strings.SplitSeq(delimitedStr, sep) {
			// Trim whitespace from the string
			trimmed := strings.TrimSpace(str)

			// Skip empty strings and duplicates
			if trimmed != "" {
				if _, exists := seen[trimmed]; !exists {
					seen[trimmed] = struct{}{}
					result = append(result, trimmed)
				}
			}
		}
	}

	return result
}

// GetClientIP determines the real client IP address from an HTTP request,
// following best practices for X-Forwarded-For handling.
//
//   - If trustedHeader is true, we use it
//   - If trustXFF is false, use the direct remote address.
//   - If trustXFF is true and trustedProxies is non-empty, walk through the X-Forwarded-For
//     header from right to left, skipping over any IPs in trustedProxies, and return the first untrusted IP.
//   - If trustXFF is true and trustedProxies is nil or empty, return the first IP in X-Forwarded-For (from the left).
//   - If no suitable IP is found, fall back to RemoteAddr.
//   - The remote address's host part is extracted with net.SplitHostPort if possible.
func GetClientIP(req *http.Request, trustXFF bool, trustedProxies *TrustedProxies, trustedHeader string) string {

	// 1. Use a trustedHeader if provided
	if trustedHeader != "" {
		if ip := req.Header.Get(trustedHeader); ip != "" {
			return strings.TrimSpace(ip)
		}
	}

	// 2. Handle X-Forwarded-For if trustXFF is true
	if trustXFF {
		if xffs := FlattenDelimitedInputs(req.Header.Values("X-Forwarded-For"), ","); len(xffs) > 0 {

			if trustedProxies != nil && len(trustedProxies.nets) > 0 {
				// we have some trustedProxies, exclude those and choose the right-most xff ip
				for i := len(xffs) - 1; i >= 0; i-- {
					ip := xffs[i]
					if trustedProxies.IsTrusted(ip) {
						continue
					}
					return ip
				}
				// if everything is trusted we fallback to RemoteAddr
			} else {
				// without any trustedProxies (but with trustXFF), we choose the left-most xff ip
				return xffs[0]
			}

		}
		// without any xff headers, we fallback to RemoteAddr
	}

	// 3. Fallback: extract host (IP) from RemoteAddr
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return req.RemoteAddr
}
