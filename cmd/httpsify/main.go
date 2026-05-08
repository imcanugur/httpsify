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
	"github.com/imcanugur/httpsify/internal/netutil"
	"github.com/imcanugur/httpsify/internal/proxy"
	tlsutil "github.com/imcanugur/httpsify/internal/tls"
	"github.com/imcanugur/httpsify/internal/version"
)

const (
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
	colorCyan   = "\033[36m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorDim    = "\033[2m"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.DefaultConfig()
	if err := parseFlags(cfg); err != nil {
		return err
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	logger := logging.NewLogger(cfg.Verbose, cfg.AccessLog)
	// Load or generate certificates

	tlsCfg, err := tlsutil.LoadOrGenerateCert(tlsutil.Config{
		CertPath:   cfg.CertPath,
		KeyPath:    cfg.KeyPath,
		SelfSigned: cfg.SelfSigned,
	})
	if err != nil {
		return fmt.Errorf("TLS configuration error: %w", err)
	}

	// Start server

	server := &http.Server{
		Addr:              cfg.ListenAddr,
		TLSConfig:         tlsCfg,
		ReadHeaderTimeout: time.Duration(cfg.ReadHeaderTimeout) * time.Second,
		IdleTimeout:       time.Duration(cfg.IdleTimeout) * time.Second,
		WriteTimeout:      time.Duration(cfg.WriteTimeout) * time.Second,
	}

	errChan := make(chan error, 1)
	go func() {
		p := proxy.NewServer(cfg, logger)
		server.Handler = p
		
		printStartupBox(cfg.ListenAddr, p.GetListeningPorts())
		if err := server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	return waitForShutdown(server, logger, errChan)
}

func parseFlags(cfg *config.Config) error {
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

	setupFlagUsage()
	flag.Parse()

	if *showVer {
		fmt.Println(version.FullVersion())
		os.Exit(0)
	}

	cfg.LoadFromEnv()
	cfg.ListenAddr, cfg.CertPath, cfg.KeyPath = *listen, *certPath, *keyPath
	cfg.SelfSigned, cfg.Verbose, cfg.AccessLog = *selfSigned, *verbose, *accessLog

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

	return nil
}

func setupFlagUsage() {
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
}

func printStartupBox(listenAddr string, services []proxy.ServiceInfo) {
	ips := netutil.GetLocalIPs()
	
	var webServices []proxy.ServiceInfo
	var systemServices []proxy.ServiceInfo
	for _, svc := range services {
		if svc.IsWeb {
			webServices = append(webServices, svc)
		} else {
			systemServices = append(systemServices, svc)
		}
	}
	
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "  %s%sHTTPSIFY%s %s%s%s\n", colorBold, colorCyan, colorReset, colorDim, version.FullVersion(), colorReset)
	fmt.Fprintf(os.Stderr, "  %s────────────────────────────────────────────────────────────%s\n", colorDim, colorReset)
	
	fmt.Fprintf(os.Stderr, "  %sReady%s    %sServer is up and listening%s\n", colorGreen, colorReset, colorDim, colorReset)
	fmt.Fprintf(os.Stderr, "  %sLocal%s    %shttps://localhost%s%s\n", colorBold, colorReset, colorCyan, listenAddr, colorReset)
	
	for _, ip := range ips {
		fmt.Fprintf(os.Stderr, "  %sNetwork%s  %shttps://%s%s%s\n", colorBold, colorReset, colorCyan, ip, listenAddr, colorReset)
	}
	
	fmt.Fprintf(os.Stderr, "  %s────────────────────────────────────────────────────────────%s\n\n", colorDim, colorReset)
	
	// Web Services
	if len(webServices) > 0 {
		fmt.Fprintf(os.Stderr, "  %sProxy Ready Services:%s\n", colorBold, colorReset)
		for i, svc := range webServices {
			if i >= 5 {
				fmt.Fprintf(os.Stderr, "    %s... and %d more web services%s\n", colorDim, len(webServices)-i, colorReset)
				break
			}
			proc := svc.ProcessName
			if proc == "" { proc = "unknown" }
			fmt.Fprintf(os.Stderr, "    %shttps://%d.localhost%s  %s→%s  %s\n", colorCyan, svc.Port, colorReset, colorDim, colorReset, proc)
		}
		fmt.Fprintf(os.Stderr, "\n")
	}

	// System Services
	if len(systemServices) > 0 {
		fmt.Fprintf(os.Stderr, "  %sOther System Services:%s\n", colorBold, colorReset)
		for i, svc := range systemServices {
			if i >= 5 {
				fmt.Fprintf(os.Stderr, "    %s... and %d more system ports%s\n", colorDim, len(systemServices)-i, colorReset)
				break
			}
			proc := svc.ProcessName
			if proc == "" { proc = "unknown" }
			fmt.Fprintf(os.Stderr, "    %sPort %d%s           %s→%s  %s\n", colorYellow, svc.Port, colorReset, colorDim, colorReset, proc)
		}
		fmt.Fprintf(os.Stderr, "\n")
	}
	
	if len(services) == 0 {
		fmt.Fprintf(os.Stderr, "    %sNo active services detected. Dashboard is ready.%s\n\n", colorDim, colorReset)
	}
	
	fmt.Fprintf(os.Stderr, "  %sDashboard Control Center%s\n", colorBold, colorReset)
	fmt.Fprintf(os.Stderr, "  %sOpen %shttps://localhost%s%s to manage all services and see details.\n", colorDim, colorCyan, colorReset, colorDim)
	
	fmt.Fprintf(os.Stderr, "\n  %s%s[Ctrl+C]%s %sto stop the server%s\n", colorBold, colorYellow, colorReset, colorDim, colorReset)
	fmt.Fprintf(os.Stderr, "\n")
}

func waitForShutdown(server *http.Server, logger *logging.Logger, errChan <-chan error) error {
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
