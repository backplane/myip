package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/backplane/myip/clientip"
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
			&cli.BoolFlag{
				Name:  "trustxff",
				Value: false,
				Usage: "trust X-Forwarded-For headers in the request (only enable if running behind a proxy)",
			},
			&cli.StringFlag{
				Name:  "trustedproxies",
				Value: "",
				Usage: "comma-separated list of IP blocks (in CIDR-notation) that upstream proxy request come from",
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

			cfg := &HandlerConfig{
				trustedProxies: clientip.NewTrustedProxies(ctx.String("trustedproxies")),
				trustXFF:       ctx.Bool("trustxff"),
			}
			http.HandleFunc("/", cfg.HandleMyIP)

			// Serve forever
			listenAddr := ctx.String("listenaddr")
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
