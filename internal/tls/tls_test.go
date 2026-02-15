package tlsutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateSelfSignedCert(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "httpsify-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	certPath := filepath.Join(tmpDir, "localhost.pem")
	keyPath := filepath.Join(tmpDir, "localhost-key.pem")

	cfg := Config{
		CertPath:   certPath,
		KeyPath:    keyPath,
		SelfSigned: true,
	}

	tlsCfg, err := LoadOrGenerateCert(cfg)
	if err != nil {
		t.Fatalf("LoadOrGenerateCert() error = %v", err)
	}

	if tlsCfg == nil {
		t.Fatal("LoadOrGenerateCert() returned nil config")
	}

	if !fileExists(certPath) {
		t.Errorf("Certificate file not created: %s", certPath)
	}
	if !fileExists(keyPath) {
		t.Errorf("Key file not created: %s", keyPath)
	}

	caPath := filepath.Join(tmpDir, "ca.pem")
	if !fileExists(caPath) {
		t.Errorf("CA certificate file not created: %s", caPath)
	}

	if tlsCfg.MinVersion < 0x0303 { 
		t.Errorf("MinVersion = %x, want >= TLS 1.2 (0x0303)", tlsCfg.MinVersion)
	}

	if len(tlsCfg.Certificates) != 1 {
		t.Errorf("Expected 1 certificate, got %d", len(tlsCfg.Certificates))
	}
}

func TestLoadExistingCert(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "httpsify-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	certPath := filepath.Join(tmpDir, "localhost.pem")
	keyPath := filepath.Join(tmpDir, "localhost-key.pem")

	cfg := Config{
		CertPath:   certPath,
		KeyPath:    keyPath,
		SelfSigned: true,
	}

	_, err = LoadOrGenerateCert(cfg)
	if err != nil {
		t.Fatalf("First LoadOrGenerateCert() error = %v", err)
	}

	tlsCfg, err := LoadOrGenerateCert(cfg)
	if err != nil {
		t.Fatalf("Second LoadOrGenerateCert() error = %v", err)
	}

	if tlsCfg == nil {
		t.Fatal("LoadOrGenerateCert() returned nil config")
	}
}

func TestLoadNonExistentCert(t *testing.T) {
	cfg := Config{
		CertPath:   "/nonexistent/path/cert.pem",
		KeyPath:    "/nonexistent/path/key.pem",
		SelfSigned: false,
	}

	_, err := LoadOrGenerateCert(cfg)
	if err == nil {
		t.Error("LoadOrGenerateCert() expected error for non-existent cert")
	}
}

func TestFileExists(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "httpsify-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	if !fileExists(tmpFile.Name()) {
		t.Errorf("fileExists(%q) = false, want true", tmpFile.Name())
	}

	if fileExists("/nonexistent/path/file") {
		t.Error("fileExists() = true for non-existent file")
	}
}
