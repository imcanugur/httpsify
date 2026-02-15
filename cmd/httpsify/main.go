package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/imcanugur/httpsify/internal/config"
	"github.com/imcanugur/httpsify/internal/logging"
	"github.com/imcanugur/httpsify/internal/proxy"
	tlsutil "github.com/imcanugur/httpsify/internal/tls"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.DefaultConfig()

	var (
		listen     = flag.String("listen", cfg.ListenAddr, "Listen address (e.g., :443)")
		certPath   = flag.String("cert", cfg.CertPath, "Path to TLS certificate (PEM)")
		keyPath    = flag.String("key", cfg.KeyPath, "Path to TLS private key (PEM)")
		selfSigned = flag.Bool("self-signed", true, "Generate self-signed certificate if missing (enabled by default)")
		denyPorts  = flag.String("deny-ports", strings.Join(config.DefaultDenyPorts, ","), "Comma-separated list of denied ports/ranges")
		allowRange = flag.String("allow-range", fmt.Sprintf("%d-%d", cfg.AllowRange.Start, cfg.AllowRange.End), "Allowed port range")
		verbose    = flag.Bool("verbose", cfg.Verbose, "Enable verbose/debug logging")
		accessLog  = flag.Bool("access-log", cfg.AccessLog, "Enable access logging")
		showVer    = flag.Bool("version", false, "Show version information")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `httpsify - Dynamic HTTPS reverse proxy for local development

Usage: httpsify [options]

Routes requests based on subdomain:
  https://<port>.localhost  ->  http://127.0.0.1:<port>

Examples:
  https://8000.localhost  ->  http://127.0.0.1:8000
  https://3000.localhost  ->  http://127.0.0.1:3000

Options:
`)
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
Environment Variables:
  HTTPSIFY_LISTEN       Listen address
  HTTPSIFY_CERT         Certificate path
  HTTPSIFY_KEY          Key path
  HTTPSIFY_SELF_SIGNED  Generate self-signed cert (true/false)
  HTTPSIFY_DENY_PORTS   Denied ports list
  HTTPSIFY_ALLOW_RANGE  Allowed port range
  HTTPSIFY_VERBOSE      Verbose logging (true/false)
  HTTPSIFY_ACCESS_LOG   Access logging (true/false)

`)
	}

	flag.Parse()

	if *showVer {
		fmt.Printf("httpsify %s (commit: %s, built: %s)\n", version, commit, date)
		return nil
	}

	cfg.LoadFromEnv()

	cfg.ListenAddr = *listen
	cfg.CertPath = *certPath
	cfg.KeyPath = *keyPath
	cfg.SelfSigned = *selfSigned
	cfg.Verbose = *verbose
	cfg.AccessLog = *accessLog

	if *denyPorts != "" {
		ranges, err := config.ParsePortRanges(*denyPorts)
		if err != nil {
			return fmt.Errorf("invalid deny-ports: %w", err)
		}
		cfg.DenyPorts = ranges
	}

	if *allowRange != "" {
		pr, err := config.ParsePortRange(*allowRange)
		if err != nil {
			return fmt.Errorf("invalid allow-range: %w", err)
		}
		cfg.AllowRange = pr
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	logger := logging.NewLogger(cfg.Verbose, cfg.AccessLog)

	logger.ServerStarting(cfg.ListenAddr, cfg.CertPath, cfg.KeyPath, cfg.SelfSigned)

	tlsCfg, err := tlsutil.LoadOrGenerateCert(tlsutil.Config{
		CertPath:   cfg.CertPath,
		KeyPath:    cfg.KeyPath,
		SelfSigned: cfg.SelfSigned,
	})
	if err != nil {
		return fmt.Errorf("TLS configuration error: %w", err)
	}

	if cfg.SelfSigned {
		logger.CertGenerated(cfg.CertPath, cfg.KeyPath)
	}

	proxyHandler := proxy.NewServer(cfg, logger)

	server := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           proxyHandler,
		TLSConfig:         tlsCfg,
		ReadHeaderTimeout: time.Duration(cfg.ReadHeaderTimeout) * time.Second,
		IdleTimeout:       time.Duration(cfg.IdleTimeout) * time.Second,
		WriteTimeout:      time.Duration(cfg.WriteTimeout) * time.Second,
	}

	errChan := make(chan error, 1)

	go func() {
		logger.ServerStarted(cfg.ListenAddr)
		
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "╭─────────────────────────────────────────────────────────────╮\n")
		fmt.Fprintf(os.Stderr, "│                                                             │\n")
		fmt.Fprintf(os.Stderr, "│   🔒 httpsify is running                                    │\n")
		fmt.Fprintf(os.Stderr, "│                                                             │\n")
		fmt.Fprintf(os.Stderr, "│   Listening on: %-42s│\n", "https://localhost"+cfg.ListenAddr)
		fmt.Fprintf(os.Stderr, "│                                                             │\n")
		fmt.Fprintf(os.Stderr, "│   Usage examples:                                           │\n")
		fmt.Fprintf(os.Stderr, "│     https://8000.localhost  →  http://127.0.0.1:8000        │\n")
		fmt.Fprintf(os.Stderr, "│     https://3000.localhost  →  http://127.0.0.1:3000        │\n")
		fmt.Fprintf(os.Stderr, "│     https://5173.localhost  →  http://127.0.0.1:5173        │\n")
		fmt.Fprintf(os.Stderr, "│                                                             │\n")
		fmt.Fprintf(os.Stderr, "│   Press Ctrl+C to stop                                      │\n")
		fmt.Fprintf(os.Stderr, "│                                                             │\n")
		fmt.Fprintf(os.Stderr, "╰─────────────────────────────────────────────────────────────╯\n")
		fmt.Fprintf(os.Stderr, "\n")

		if err := server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		return fmt.Errorf("server error: %w", err)
	case sig := <-sigChan:
		logger.Info("shutting down", "signal", sig.String())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown error: %w", err)
	}

	logger.Info("server stopped")
	return nil
}
