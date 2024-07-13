package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/urfave/cli/v2"
)

// command-line option init & defaults
var (
	// version, commit, date, builtBy are provided by goreleaser during build
	version = "dev"
	commit  = "dev"
	date    = "unknown"
	builtBy = "unknown"

	logger *slog.Logger
)

// HandleMyIP is an http endpoint that returns the IP address of the requester
func HandleMyIP(w http.ResponseWriter, r *http.Request) {
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

	ip := r.Header.Get(`X-Forwarded-For`)
	if ip == "" {
		ip = r.RemoteAddr
	} else {
		commaIdx := strings.Index(ip, ",")
		if commaIdx > -1 {
			ip = ip[0:commaIdx]
		}
	}

	logger.Info("req",
		"remote_addr", r.RemoteAddr,
		"forwarded_for", r.Header.Get(`X-Forwarded-For`))

	w.Header().Add(`Content-Type`, `application/json`)
	fmt.Fprintf(w, "{\"ip\": \"%s\"}\n", ip)
}

// setLogLevel sets the log level
func setLogLevel(level string) {
	switch strings.ToUpper(level) {
	case "INFO":
		slog.SetLogLoggerLevel(slog.LevelInfo)
	case "WARN":
		slog.SetLogLoggerLevel(slog.LevelWarn)
	case "ERROR":
		slog.SetLogLoggerLevel(slog.LevelError)
	default:
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}
}

func init() {
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Printf("myip version %s; commit %s; built on %s; by %s\n", version, commit, date, builtBy)
	}

}

func main() {
	logger = slog.New(slog.NewTextHandler(os.Stderr, nil))

	app := &cli.App{
		Name:    "myip",
		Version: version,
		Usage:   "HTTP endpoint that reports the user's IP address back to the user",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "loglevel",
				Value: "INFO",
				Usage: "how verbosely to log, one of: DEBUG, INFO, WARN, ERROR",
			},
			&cli.StringFlag{
				Name:  "listenaddr",
				Value: "0.0.0.0:8000",
				Usage: "IP address and port to listen on",
			},
		},
		Before: func(ctx *cli.Context) error {
			setLogLevel(ctx.String("loglevel"))
			return nil
		},
		Action: func(ctx *cli.Context) error {
			logger.Info("starting up",
				"version", version,
				"commit", commit,
				"date", date,
				"builder", builtBy,
			)

			// setup webhook listener
			listenAddr := ctx.String("listenaddr")
			http.HandleFunc("/", HandleMyIP)

			// Serve forever
			logger.Info("listening for API connections", "addr", listenAddr)
			if err := http.ListenAndServe(listenAddr, nil); err != nil {
				return err
			}
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		logger.Error("error while listening for API connections",
			"error", err)
		os.Exit(1)
	}
	logger.Debug("exiting successfully")
}
