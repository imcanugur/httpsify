# httpsify - Dynamic HTTPS Reverse Proxy for Local Development

A production-grade Go reverse proxy that provides HTTPS for **any** local port dynamically via subdomain routing.

```
https://<PORT>.localhost  →  http://127.0.0.1:<PORT>
```

## Features

- 🔒 **Automatic TLS** - Use mkcert certificates or auto-generate self-signed certs
- 🔀 **Dynamic Routing** - No manual configuration per port
- 🌐 **WebSocket Support** - Full WebSocket pass-through with connection hijacking
- ⚡ **Production Quality** - Proper timeouts, keep-alive, connection pooling
- 🛡️ **Security** - Port denylists, input validation, TLS 1.2+
- 📊 **Structured Logging** - JSON logs with request IDs, timing, and status codes
- 🎯 **Zero Dependencies** - Only stdlib + golang.org/x/net for edge cases

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/imcanugur/httpsify.git
cd httpsify

# Build
go build -o httpsify ./cmd/httpsify

# Or use make
make build
```

### Using Go Install

```bash
go install github.com/imcanugur/httpsify/cmd/httpsify@latest
```

## Quick Start

### Trusted Certificates with mkcert

[mkcert](https://github.com/FiloSottile/mkcert) creates locally-trusted development certificates.

```bash
# Install mkcert
# macOS
brew install mkcert

# Linux
sudo apt install libnss3-tools
go install filippo.io/mkcert@latest

# Windows
choco install mkcert

# Install local CA
mkcert -install

# Generate certificates
mkdir -p cert
mkcert -cert-file cert/localhost.pem -key-file cert/localhost-key.pem \
    localhost "*.localhost" localtest.me "*.localtest.me" 127.0.0.1 ::1

# Run httpsify
sudo ./httpsify --cert cert/localhost.pem --key cert/localhost-key.pem
```

### Self-Signed Certificates

```bash
# Generate and use self-signed certificates
sudo ./httpsify --self-signed

# Certificates are saved to ./cert/ directory
# You'll need to trust ./cert/ca.pem in your browser/system
```

### Why sudo?

Port 443 requires root privileges on Linux/macOS. Alternatives:

```bash
# Option 1: Use a high port
./httpsify --listen :8443

# Option 2: Use setcap (Linux only)
sudo setcap CAP_NET_BIND_SERVICE=+eip ./httpsify
./httpsify --listen :443

# Option 3: Use authbind (Linux)
sudo apt install authbind
sudo touch /etc/authbind/byport/443
sudo chmod 500 /etc/authbind/byport/443
sudo chown $USER /etc/authbind/byport/443
authbind --deep ./httpsify
```

## Usage Examples

### Laravel Development (Port 8000)

```bash
# Terminal 1: Start Laravel
cd my-laravel-app
php artisan serve

# Terminal 2: Start httpsify
sudo ./httpsify

# Access via HTTPS
# https://8000.localhost
```

### React Development (Port 3000)

```bash
# Terminal 1: Start React
cd my-react-app
npm start

# Terminal 2: httpsify is already running

# Access via HTTPS
# https://3000.localhost
```

### Vite Development (Port 5173)

```bash
# Terminal 1: Start Vite
cd my-vite-app
npm run dev

# Access via HTTPS
# https://5173.localhost
```

### Multiple Services at Once

```bash
# All these work simultaneously:
# https://8000.localhost  →  Laravel API
# https://3000.localhost  →  React frontend  
# https://5173.localhost  →  Vite dev server
# https://8080.localhost  →  Go backend
# https://6379.localhost  →  (blocked - not in allow range by default)
```

## Command-Line Options

```
Usage: httpsify [options]

Options:
  --listen string        Listen address (default ":443")
  --cert string          Path to TLS certificate PEM (default "./cert/localhost.pem")
  --key string           Path to TLS private key PEM (default "./cert/localhost-key.pem")
  --self-signed          Generate self-signed certificate if missing
  --deny-ports string    Comma-separated denied ports/ranges (default "22,25,135-139,445,3389,5900")
  --allow-range string   Allowed port range (default "1024-65535")
  --verbose              Enable verbose/debug logging
  --access-log           Enable access logging (default true)
  --version              Show version information
```

## Environment Variables

All options can also be set via environment variables:

```bash
export HTTPSIFY_LISTEN=":443"
export HTTPSIFY_CERT="./cert/localhost.pem"
export HTTPSIFY_KEY="./cert/localhost-key.pem"
export HTTPSIFY_SELF_SIGNED="true"
export HTTPSIFY_DENY_PORTS="22,25,135-139,445"
export HTTPSIFY_ALLOW_RANGE="1024-65535"
export HTTPSIFY_VERBOSE="true"
export HTTPSIFY_ACCESS_LOG="true"
```

## WebSocket Support

WebSocket connections are fully supported. Example test:

```javascript
// Connect to a WebSocket server running on port 8080
const ws = new WebSocket('wss://8080.localhost/ws');

ws.onopen = () => {
    console.log('Connected!');
    ws.send('Hello, WebSocket!');
};

ws.onmessage = (event) => {
    console.log('Received:', event.data);
};
```

### WebSocket Test Server (Go)

```go
package main

import (
    "fmt"
    "net/http"
    "golang.org/x/net/websocket"
)

func main() {
    http.Handle("/ws", websocket.Handler(func(ws *websocket.Conn) {
        var msg string
        for {
            if err := websocket.Message.Receive(ws, &msg); err != nil {
                break
            }
            fmt.Printf("Received: %s\n", msg)
            websocket.Message.Send(ws, "Echo: "+msg)
        }
    }))
    
    fmt.Println("WebSocket server on :8080")
    http.ListenAndServe(":8080", nil)
}
```

## Windows + WSL2 Setup

### Running in WSL2

```bash
# In WSL2 terminal
cd /path/to/httpsify

# Build
go build -o httpsify ./cmd/httpsify

# Generate certs and run
sudo ./httpsify --self-signed
```

### Accessing from Windows

1. **Method 1: Use `.localhost` domain** (works automatically in most browsers)
   - Chrome and Edge support `*.localhost` out of the box

2. **Method 2: Edit Windows hosts file**
   ```powershell
   # Run as Administrator
   notepad C:\Windows\System32\drivers\etc\hosts
   
   # Add these lines (where 172.x.x.x is your WSL IP)
   172.x.x.x 8000.localhost
   172.x.x.x 3000.localhost
   ```

3. **Get WSL IP**
   ```bash
   # In WSL
   hostname -I | awk '{print $1}'
   ```

### Trusting Self-Signed Certificates on Windows

1. Copy `cert/ca.pem` from WSL to Windows
2. Double-click the file
3. Click "Install Certificate"
4. Select "Local Machine" → "Trusted Root Certification Authorities"

## Error Responses

httpsify returns helpful JSON errors:

```json
{
  "error": "Port 22 is not allowed",
  "hint": "This port is either denied or outside the allowed range",
  "example": "https://8000.localhost"
}
```

```json
{
  "error": "Connection refused",
  "hint": "No service is listening on port 8000",
  "example": ""
}
```

```json
{
  "error": "invalid host format: example.com",
  "hint": "Use format: https://<port>.localhost",
  "example": "https://8000.localhost"
}
```

## Security

### Denied Ports (Default)

These ports are blocked by default to prevent accidental exposure:

| Port | Service |
|------|---------|
| 22 | SSH |
| 25 | SMTP |
| 135-139 | NetBIOS/SMB |
| 445 | SMB |
| 3389 | RDP |
| 5900 | VNC |

### Port Allow Range

By default, only ports 1024-65535 are allowed. To allow privileged ports:

```bash
./httpsify --allow-range "80-65535"
```

### TLS Configuration

- Minimum TLS 1.2
- Modern cipher suites only (ECDHE, AES-GCM, ChaCha20)
- Perfect forward secrecy (PFS) enabled
- X25519 preferred for key exchange

## Logging

JSON structured logging with request tracking:

```json
{"time":"2024-01-15T10:30:45Z","level":"INFO","msg":"request completed","request_id":"1705315845123456789-1","method":"GET","host":"8000.localhost","target_port":8000,"status":200,"latency":"12.345ms","bytes":1234}
```

Enable debug logging:

```bash
./httpsify --verbose
```

## Troubleshooting

### "connection refused" errors

Your backend service isn't running on that port:

```bash
# Check what's running
lsof -i :8000  # macOS/Linux
netstat -ano | findstr :8000  # Windows
```

### Certificate warnings in browser

#### For mkcert users
Make sure you ran `mkcert -install` to install the local CA.

#### For self-signed
Import `./cert/ca.pem` into your browser/OS trust store.

**Chrome:**
1. Navigate to `chrome://settings/certificates`
2. Go to "Authorities" tab
3. Click "Import" and select `ca.pem`
4. Check "Trust this certificate for identifying websites"

**Firefox:**
1. Navigate to `about:preferences#privacy`
2. Scroll to "Certificates" → "View Certificates"
3. Go to "Authorities" tab
4. Click "Import" and select `ca.pem`

### "address already in use" error

Another process is using port 443:

```bash
# Find the process
sudo lsof -i :443

# Or use a different port
./httpsify --listen :8443
```

### WebSocket connections failing

1. Make sure your backend WebSocket server is running
2. Check that the backend responds to the WebSocket upgrade handshake
3. Enable verbose logging to see what's happening:

```bash
./httpsify --verbose
```

### Permission denied (port 443)

See the "Why sudo?" section above for alternatives.

## Running as a Service

### systemd (Linux)

```bash
# Copy the unit file
sudo cp httpsify.service /etc/systemd/system/

# Edit paths as needed
sudo systemctl edit httpsify

# Enable and start
sudo systemctl enable --now httpsify

# Check status
sudo systemctl status httpsify
```

### launchd (macOS)

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.httpsify.server</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/httpsify</string>
        <string>--cert</string>
        <string>/etc/httpsify/localhost.pem</string>
        <string>--key</string>
        <string>/etc/httpsify/localhost-key.pem</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
</dict>
</plist>
```

## Testing

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific package tests
go test -v ./internal/config/
go test -v ./internal/proxy/
```

## Project Structure

```
httpsify/
├── cmd/
│   └── httpsify/
│       └── main.go           # CLI entry point
├── internal/
│   ├── config/
│   │   ├── config.go         # Configuration parsing
│   │   └── config_test.go    # Config tests
│   ├── logging/
│   │   └── logging.go        # Structured logging
│   ├── proxy/
│   │   ├── proxy.go          # Reverse proxy handler
│   │   └── proxy_test.go     # Proxy tests
│   └── tls/
│       └── tls.go            # TLS configuration
├── cert/                     # Generated certificates (gitignored)
├── go.mod
├── go.sum
├── Makefile
├── README.md
└── httpsify.service          # systemd unit template
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `go test ./...`
5. Submit a pull request

## License

MIT License - see LICENSE file for details.
