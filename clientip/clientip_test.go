package clientip

import (
	"net/http"
	"testing"
)

func makeRequest(xffHeaders []string, remoteAddr string, extraHeaders *map[string]string) *http.Request {
	req, _ := http.NewRequest("GET", "/", nil)
	for _, h := range xffHeaders {
		req.Header.Add("X-Forwarded-For", h)
	}
	if extraHeaders != nil {
		for k, v := range *(extraHeaders) {
			req.Header.Add(k, v)
		}
	}
	req.RemoteAddr = remoteAddr
	return req
}

func TestGetClientIP(t *testing.T) {
	trusted := NewTrustedProxies("127.0.0.1/32,10.0.0.0/8,192.168.0.0/16")

	tests := []struct {
		name           string
		xff            []string
		remote         string
		trustXFF       bool
		trustedProxies *TrustedProxies
		trustedHeader  string
		extraHeaders   *map[string]string
		want           string
	}{
		{
			name:           "no XFF, trustXFF=false",
			xff:            nil,
			remote:         "8.8.8.8:1234",
			trustXFF:       false,
			trustedProxies: trusted,
			trustedHeader:  "",
			extraHeaders:   nil,
			want:           "8.8.8.8",
		},
		{
			name:           "no XFF, trustXFF=true",
			xff:            nil,
			remote:         "8.8.8.8:1234",
			trustXFF:       true,
			trustedProxies: trusted,
			trustedHeader:  "",
			extraHeaders:   nil,
			want:           "8.8.8.8",
		},
		{
			name:           "all trusted, trustXFF=true",
			xff:            []string{"10.0.1.1, 192.168.1.1, 127.0.0.1"},
			remote:         "9.9.9.9:1234",
			trustXFF:       true,
			trustedProxies: trusted,
			trustedHeader:  "",
			extraHeaders:   nil,
			want:           "9.9.9.9",
		},
		{
			name:           "last untrusted is returned, trustXFF=true",
			xff:            []string{"8.8.8.8, 10.0.1.1, 192.168.1.1"},
			remote:         "1.2.3.4:1234",
			trustXFF:       true,
			trustedProxies: trusted,
			trustedHeader:  "",
			extraHeaders:   nil,
			want:           "8.8.8.8",
		},
		{
			name:           "multiple XFF headers concatenated, trustXFF=true",
			xff:            []string{"1.1.1.1, 2.2.2.2", "10.0.0.2, 192.168.0.1"},
			remote:         "3.3.3.3:1234",
			trustXFF:       true,
			trustedProxies: trusted,
			trustedHeader:  "",
			extraHeaders:   nil,
			want:           "2.2.2.2",
		},
		{
			name:           "RemoteAddr without port",
			xff:            nil,
			remote:         "4.4.4.4",
			trustXFF:       true,
			trustedProxies: trusted,
			trustedHeader:  "",
			extraHeaders:   nil,
			want:           "4.4.4.4",
		},
		{
			name:           "no trustedProxies, trustXFF=true (should act as trust no proxies)",
			xff:            []string{"1.1.1.1, 2.2.2.2"},
			remote:         "3.3.3.3:1234",
			trustXFF:       true,
			trustedProxies: nil,
			trustedHeader:  "",
			extraHeaders:   nil,
			want:           "3.3.3.3",
		},
		{
			name:           "trustedHeader",
			xff:            nil,
			remote:         "3.3.3.3:1234",
			trustXFF:       true,
			trustedProxies: nil,
			trustedHeader:  "X-Client-IP-For-Real",
			extraHeaders: &map[string]string{
				"X-Client-IP-For-Real": "4.4.4.4",
				"X-Client-IP":          "7.6.5.4",
			},
			want: "4.4.4.4",
		},
		{
			name:           "trustedHeaderMismatch",
			xff:            nil,
			remote:         "3.3.3.3:1234",
			trustXFF:       true,
			trustedProxies: nil,
			trustedHeader:  "X-Client-IP-Wrong",
			extraHeaders: &map[string]string{
				"X-Client-IP-For-Real": "4.4.4.4",
				"X-Client-IP":          "7.6.5.4",
			},
			want: "3.3.3.3",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := makeRequest(tc.xff, tc.remote, tc.extraHeaders)
			got := GetClientIP(req, tc.trustXFF, tc.trustedProxies, tc.trustedHeader)
			if got != tc.want {
				t.Errorf("want %q, got %q", tc.want, got)
			}
		})
	}
}
