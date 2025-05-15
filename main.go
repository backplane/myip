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

type Config struct {
	TrustedProxies *clientip.TrustedProxies
	TrustedHeader  string
	TrustXFF       bool
	ListenAddr     string
}

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
		_, err := fmt.Fprintf(cmd.Root().Writer, "myip version %s; commit %s; built on %s by %s\n", version, commit, date, builtBy)
		if err != nil {
			panic("failed to write out version information")
		}
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
				Name:    "loglevel",
				Value:   "INFO",
				Usage:   "how verbosely to log, one of: DEBUG, INFO, WARN, ERROR",
				Sources: cli.EnvVars("LOG_LEVEL"),
			},
			&cli.StringFlag{
				Name:    "listenaddr",
				Value:   "0.0.0.0:8000",
				Usage:   "IP address and port to listen on",
				Sources: cli.EnvVars("LISTEN_ADDR"),
			},
			&cli.StringFlag{
				Name:    "trustedheader",
				Value:   "",
				Usage:   "the name of a trusted header which provides the client IP address directly",
				Sources: cli.EnvVars("TRUSTED_HEADER"),
			},
			&cli.BoolFlag{
				Name:    "trustxff",
				Value:   false,
				Usage:   "trust X-Forwarded-For headers in the request (only enable if running behind a proxy)",
				Sources: cli.EnvVars("TRUST_XFF"),
			},
			&cli.StringFlag{
				Name:    "trustedproxies",
				Value:   "",
				Usage:   "comma-separated list of IP blocks (in CIDR-notation) that upstream proxy request come from",
				Sources: cli.EnvVars("TRUSTED_PROXIES"),
			},
		},
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			setLogLevel(cmd.String("loglevel"))
			return nil, nil
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			cfg := &Config{
				TrustedProxies: clientip.NewTrustedProxies(cmd.String("trustedproxies")),
				TrustedHeader:  cmd.String("trustedheader"),
				TrustXFF:       cmd.Bool("trustxff"),
				ListenAddr:     cmd.String("listenaddr"),
			}

			logger.Info("starting up",
				"version", version,
				"commit", commit,
				"date", date,
				"builder", builtBy,
				"config", cfg,
			)

			http.HandleFunc("/", cfg.HandleMyIP)

			// Serve forever
			logger.Info("listening for API connections")
			if err := http.ListenAndServe(cfg.ListenAddr, nil); err != nil {
				return err
			}
			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		logger.Error("error while listening for connections",
			"error", err)
		os.Exit(1)
	}
	logger.Debug("exiting successfully")
}
