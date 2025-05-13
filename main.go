package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/backplane/myip/clientip"
	"github.com/urfave/cli/v3"
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
	cli.VersionPrinter = func(cmd *cli.Command) {
		fmt.Fprintf(cmd.Root().Writer, "myip version %s; commit %s; built on %s; by %s\n", version, commit, date, builtBy)
	}

}

func main() {
	logger = slog.New(slog.NewTextHandler(os.Stderr, nil))

	cmd := &cli.Command{
		Name:                  "myip",
		Version:               version,
		Usage:                 "HTTP endpoint that reports the user's IP address back to the user",
		EnableShellCompletion: true,
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
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			setLogLevel(cmd.String("loglevel"))
			return nil, nil
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			logger.Info("starting up",
				"version", version,
				"commit", commit,
				"date", date,
				"builder", builtBy,
			)

			cfg := &HandlerConfig{
				trustedProxies: clientip.NewTrustedProxies(cmd.String("trustedproxies")),
				trustXFF:       cmd.Bool("trustxff"),
			}
			http.HandleFunc("/", cfg.HandleMyIP)

			// Serve forever
			listenAddr := cmd.String("listenaddr")
			logger.Info("listening for API connections", "addr", listenAddr)
			if err := http.ListenAndServe(listenAddr, nil); err != nil {
				return err
			}
			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		logger.Error("error while listening for API connections",
			"error", err)
		os.Exit(1)
	}
	logger.Debug("exiting successfully")
}
