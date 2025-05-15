package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/backplane/myip/clientip"
)

// HandleMyIP is an http endpoint that returns the IP address of the requester
func (cfg *Config) HandleMyIP(w http.ResponseWriter, r *http.Request) {
	// Accepting HTTP GETs only
	if r.Method != http.MethodGet {
		logger.Warn("invalid request method", "method", r.Method)
		w.WriteHeader(http.StatusBadRequest)
		if _, err := io.WriteString(w, "Invalid request method\n"); err != nil {
			logger.Error("failed to write error response", "error", err)
		}
		return
	}

	ip := clientip.GetClientIP(r, cfg.trustXFF, cfg.trustedProxies, cfg.trustedHeader)

	w.Header().Add(`Content-Type`, `application/json`)
	_, err := fmt.Fprintf(w, "{\"ip\": \"%s\"}\n", ip)
	if err != nil {
		logger.Error("error writing response", "error", err)
	}
}
