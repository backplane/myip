package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/backplane/myip/clientip"
)

type HandlerConfig struct {
	trustedProxies *clientip.TrustedProxies
	trustXFF       bool
}

// HandleMyIP is an http endpoint that returns the IP address of the requester
func (cfg *HandlerConfig) HandleMyIP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Accepting HTTP GETs only
	if r.Method != http.MethodGet {
		logger.Warn("invalid request method", "method", r.Method)
		w.WriteHeader(http.StatusBadRequest)
		if _, err := io.WriteString(w, "Invalid request method\n"); err != nil {
			logger.Error("failed to write error response", "error", err)
		}
		return
	}

	ip := clientip.GetClientIP(r, cfg.trustXFF, cfg.trustedProxies)

	w.Header().Add(`Content-Type`, `application/json`)
	fmt.Fprintf(w, "{\"ip\": \"%s\"}\n", ip)
}
