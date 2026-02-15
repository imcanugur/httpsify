package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var DefaultDenyPorts = []string{
	"22",      // SSH
	"25",      // SMTP
	"135-139", // NetBIOS/SMB
	"445",     // SMB
	"3389",    // RDP
	"5900",    // VNC
}

type Config struct {
	ListenAddr string

	CertPath string
	KeyPath  string

	SelfSigned bool

	DenyPorts  []PortRange
	AllowRange PortRange

	Verbose   bool
	AccessLog bool

	ReadHeaderTimeout int
	IdleTimeout       int
	WriteTimeout      int
	DialTimeout       int
}

type PortRange struct {
	Start int
	End   int
}

func (pr PortRange) Contains(port int) bool {
	return port >= pr.Start && port <= pr.End
}

func (pr PortRange) String() string {
	if pr.Start == pr.End {
		return strconv.Itoa(pr.Start)
	}
	return fmt.Sprintf("%d-%d", pr.Start, pr.End)
}

func DefaultConfig() *Config {
	cfg := &Config{
		ListenAddr:        ":443",
		CertPath:          "./cert/localhost.pem",
		KeyPath:           "./cert/localhost-key.pem",
		SelfSigned:        true,
		AllowRange:        PortRange{Start: 1024, End: 65535},
		Verbose:           false,
		AccessLog:         true,
		ReadHeaderTimeout: 10,
		IdleTimeout:       120,
		WriteTimeout:      30,
		DialTimeout:       10,
	}

	cfg.DenyPorts, _ = ParsePortRanges(strings.Join(DefaultDenyPorts, ","))

	return cfg
}

func (c *Config) LoadFromEnv() {
	if v := os.Getenv("HTTPSIFY_LISTEN"); v != "" {
		c.ListenAddr = v
	}
	if v := os.Getenv("HTTPSIFY_CERT"); v != "" {
		c.CertPath = v
	}
	if v := os.Getenv("HTTPSIFY_KEY"); v != "" {
		c.KeyPath = v
	}
	if v := os.Getenv("HTTPSIFY_SELF_SIGNED"); v != "" {
		c.SelfSigned = v == "true" || v == "1"
	}
	if v := os.Getenv("HTTPSIFY_DENY_PORTS"); v != "" {
		if ranges, err := ParsePortRanges(v); err == nil {
			c.DenyPorts = ranges
		}
	}
	if v := os.Getenv("HTTPSIFY_ALLOW_RANGE"); v != "" {
		if ranges, err := ParsePortRanges(v); err == nil && len(ranges) > 0 {
			c.AllowRange = ranges[0]
		}
	}
	if v := os.Getenv("HTTPSIFY_VERBOSE"); v != "" {
		c.Verbose = v == "true" || v == "1"
	}
	if v := os.Getenv("HTTPSIFY_ACCESS_LOG"); v != "" {
		c.AccessLog = v == "true" || v == "1"
	}
}

func ParsePortRanges(s string) ([]PortRange, error) {
	if s == "" {
		return nil, nil
	}

	var ranges []PortRange
	parts := strings.Split(s, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		pr, err := ParsePortRange(part)
		if err != nil {
			return nil, fmt.Errorf("invalid port range %q: %w", part, err)
		}
		ranges = append(ranges, pr)
	}

	return ranges, nil
}

func ParsePortRange(s string) (PortRange, error) {
	s = strings.TrimSpace(s)

	if strings.Contains(s, "-") {
		parts := strings.SplitN(s, "-", 2)
		start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			return PortRange{}, fmt.Errorf("invalid start port: %w", err)
		}
		end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			return PortRange{}, fmt.Errorf("invalid end port: %w", err)
		}

		if err := ValidatePort(start); err != nil {
			return PortRange{}, fmt.Errorf("invalid start port: %w", err)
		}
		if err := ValidatePort(end); err != nil {
			return PortRange{}, fmt.Errorf("invalid end port: %w", err)
		}
		if start > end {
			return PortRange{}, errors.New("start port cannot be greater than end port")
		}

		return PortRange{Start: start, End: end}, nil
	}

	port, err := strconv.Atoi(s)
	if err != nil {
		return PortRange{}, fmt.Errorf("invalid port: %w", err)
	}
	if err := ValidatePort(port); err != nil {
		return PortRange{}, err
	}

	return PortRange{Start: port, End: port}, nil
}

func ValidatePort(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("port %d out of valid range (1-65535)", port)
	}
	return nil
}

func (c *Config) IsPortAllowed(port int) bool {
	for _, pr := range c.DenyPorts {
		if pr.Contains(port) {
			return false
		}
	}

	return c.AllowRange.Contains(port)
}

func (c *Config) Validate() error {
	if c.ListenAddr == "" {
		return errors.New("listen address is required")
	}

	if !c.SelfSigned {
		if c.CertPath == "" || c.KeyPath == "" {
			return errors.New("certificate and key paths are required unless --self-signed is set")
		}
	}

	if c.ReadHeaderTimeout < 1 {
		return errors.New("read header timeout must be at least 1 second")
	}
	if c.IdleTimeout < 1 {
		return errors.New("idle timeout must be at least 1 second")
	}
	if c.WriteTimeout < 1 {
		return errors.New("write timeout must be at least 1 second")
	}
	if c.DialTimeout < 1 {
		return errors.New("dial timeout must be at least 1 second")
	}

	return nil
}
