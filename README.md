<p align="center">
  <img src="https://img.shields.io/badge/httpsify-🔒-blueviolet?style=for-the-badge&labelColor=1a1a2e&color=6c63ff" alt="httpsify" width="300"/>
</p>

<h1 align="center">httpsify</h1>

<p align="center">
  <strong>One command. Every port. Instant HTTPS.</strong>
</p>

<p align="center">
  <a href="https://github.com/imcanugur/httpsify/releases"><img src="https://img.shields.io/github/v/release/imcanugur/httpsify?style=flat-square&color=6c63ff" alt="Release"></a>
  <a href="https://github.com/imcanugur/httpsify/actions"><img src="https://img.shields.io/github/actions/workflow/status/imcanugur/httpsify/release.yml?style=flat-square&label=build" alt="Build"></a>
  <a href="https://goreportcard.com/report/github.com/imcanugur/httpsify"><img src="https://goreportcard.com/badge/github.com/imcanugur/httpsify?style=flat-square" alt="Go Report"></a>
  <a href="https://github.com/imcanugur/httpsify/blob/main/LICENSE"><img src="https://img.shields.io/github/license/imcanugur/httpsify?style=flat-square&color=6c63ff" alt="License"></a>
  <a href="https://github.com/imcanugur/httpsify"><img src="https://img.shields.io/badge/platform-linux%20%7C%20macos%20%7C%20windows-informational?style=flat-square" alt="Platform"></a>
</p>

<br/>

<p align="center">
  <code>https://&lt;PORT&gt;.localhost</code> &nbsp;→&nbsp; <code>http://127.0.0.1:&lt;PORT&gt;</code>
</p>

<br/>

---

## The Problem

You're building a modern app. Your API runs on `:8000`, your frontend on `:3000`, your WebSocket server on `:8080`. Then you need HTTPS because:

- OAuth providers require it
- Service Workers demand it  
- Secure cookies won't work without it
- Your staging environment uses it
- Browser APIs like Geolocation, Camera, Clipboard need a secure context

So you waste **30 minutes** configuring nginx, generating certs, editing configs for each port...

## The Solution

```bash
sudo httpsify
```

**That's it.** Every local port is now available over HTTPS. No config files. No per-port setup. No restart needed when you spin up a new service.

```
https://3000.localhost  →  React app
https://8000.localhost  →  Laravel API
https://5173.localhost  →  Vite dev server
https://8080.localhost  →  Go backend
https://4200.localhost  →  Angular app
https://8888.localhost  →  Jupyter notebook
```

All at the same time. Zero configuration.

---

## Quick Start

### Install (30 seconds)

**One-liner:**

```bash
curl -fsSL https://raw.githubusercontent.com/imcanugur/httpsify/main/install.sh | bash
```

**Or with Go:**

```bash
go install github.com/imcanugur/httpsify/cmd/httpsify@latest
```

**Or from source:**

```bash
git clone https://github.com/imcanugur/httpsify.git
cd httpsify && make build
```

### Run

```bash
sudo httpsify
```

Certificates are generated automatically on first run. Open `https://8000.localhost` and you're done.

---

## Why httpsify?

<table>
<tr>
<th></th>
<th>httpsify</th>
<th>nginx + mkcert</th>
<th>Caddy</th>
</tr>
<tr>
<td><strong>Setup time</strong></td>
<td>🟢 10 seconds</td>
<td>🔴 15+ minutes</td>
<td>🟡 5 minutes</td>
</tr>
<tr>
<td><strong>New port</strong></td>
<td>🟢 Nothing to do</td>
<td>🔴 Edit config + reload</td>
<td>🔴 Edit Caddyfile + reload</td>
</tr>
<tr>
<td><strong>Config files</strong></td>
<td>🟢 Zero</td>
<td>🔴 nginx.conf + certs</td>
<td>🟡 Caddyfile</td>
</tr>
<tr>
<td><strong>WebSocket</strong></td>
<td>🟢 Built-in</td>
<td>🟡 Extra config</td>
<td>🟢 Built-in</td>
</tr>
<tr>
<td><strong>Binary size</strong></td>
<td>🟢 ~7 MB</td>
<td>🔴 ~50 MB</td>
<td>🟡 ~40 MB</td>
</tr>
<tr>
<td><strong>Dependencies</strong></td>
<td>🟢 None</td>
<td>🔴 nginx + openssl</td>
<td>🟢 None</td>
</tr>
</table>

---

## Features

### 🔀 Dynamic Routing — Zero Config

No config files. No YAML. No restart. Just use the port number as a subdomain:

```
https://3000.localhost  →  http://127.0.0.1:3000
https://8000.localhost  →  http://127.0.0.1:8000
```

Spin up a new service on port 9000? It's already available at `https://9000.localhost`.

### 🔒 Automatic TLS Certificates

Certificates are generated automatically on first run. Supports two modes:

| Mode | Command | Browser Warning? |
|------|---------|-----------------|
| **Self-signed** (default) | `sudo httpsify` | Yes (import CA to fix) |
| **mkcert** (recommended) | `sudo httpsify --cert cert/localhost.pem --key cert/localhost-key.pem` | No |

**Trusted certs with mkcert:**

```bash
# One-time setup
brew install mkcert    # or: apt install libnss3-tools && go install filippo.io/mkcert@latest
mkcert -install

# Generate wildcard cert
mkdir -p cert
mkcert -cert-file cert/localhost.pem -key-file cert/localhost-key.pem \
    localhost "*.localhost" 127.0.0.1 ::1

# Run with trusted certs
sudo httpsify --cert cert/localhost.pem --key cert/localhost-key.pem
```

### 🌐 WebSocket Support

Full `wss://` pass-through with connection hijacking. Just works:

```javascript
const ws = new WebSocket('wss://8080.localhost/ws');
ws.onopen = () => ws.send('Hello!');
ws.onmessage = (e) => console.log(e.data);
```

### 🛡️ Security Built-in

- **TLS 1.2+** with modern cipher suites (ECDHE, AES-GCM, ChaCha20)
- **Perfect Forward Secrecy** enabled by default
- **Port denylist** blocks dangerous ports (SSH, SMB, RDP, VNC)
- **Port range control** — only 1024-65535 allowed by default

### 📊 Structured JSON Logging

Every request is logged with timing, status codes, and request IDs:

```json
{"time":"2026-02-15T10:30:45Z","level":"INFO","msg":"request completed","request_id":"abc123","method":"GET","host":"8000.localhost","target_port":8000,"status":200,"latency":"2.3ms","bytes":4521}
```

### ⚡ Production Quality

- Proper timeouts and keep-alive
- Connection pooling
- Graceful shutdown (30s drain)
- Cross-platform: Linux, macOS, Windows

---

## Real-World Examples

### Full-Stack Development

```bash
# Terminal 1 — Backend
cd api && php artisan serve  # :8000

# Terminal 2 — Frontend
cd frontend && npm run dev   # :5173

# Terminal 3 — httpsify (once, covers everything)
sudo httpsify
```

Now your frontend at `https://5173.localhost` can call `https://8000.localhost/api` with proper CORS, secure cookies, and zero SSL errors.

### OAuth / Social Login Testing

```bash
# Google OAuth callback: https://3000.localhost/auth/callback
# No more "redirect_uri_mismatch" errors
sudo httpsify
```

### Mobile App Development

```bash
# Expose your API over HTTPS for mobile testing
sudo httpsify --listen :8443
# Point your mobile app to: https://8443-api.localtest.me
```

---

## Configuration

### CLI Options

```
httpsify [options]

  --listen string        Listen address (default ":443")
  --cert string          TLS certificate path (default "./cert/localhost.pem")
  --key string           TLS private key path (default "./cert/localhost-key.pem")
  --self-signed          Auto-generate self-signed cert (default: true)
  --deny-ports string    Blocked ports (default "22,25,135-139,445,3389,5900")
  --allow-range string   Allowed port range (default "1024-65535")
  --verbose              Debug logging
  --access-log           Access logging (default: true)
  --version              Show version
```

### Environment Variables

```bash
export HTTPSIFY_LISTEN=":443"
export HTTPSIFY_CERT="./cert/localhost.pem"
export HTTPSIFY_KEY="./cert/localhost-key.pem"
export HTTPSIFY_SELF_SIGNED="true"
export HTTPSIFY_DENY_PORTS="22,25,135-139,445,3389,5900"
export HTTPSIFY_ALLOW_RANGE="1024-65535"
export HTTPSIFY_VERBOSE="true"
export HTTPSIFY_ACCESS_LOG="true"
```

### Run Without sudo

```bash
# Option 1: Use a high port
httpsify --listen :8443

# Option 2: Grant capability (Linux)
sudo setcap CAP_NET_BIND_SERVICE=+eip ./httpsify
httpsify

# Option 3: authbind (Linux)
sudo apt install authbind
sudo touch /etc/authbind/byport/443 && sudo chmod 500 /etc/authbind/byport/443
authbind --deep httpsify
```

---

## Run as a Service

### systemd (Linux)

```bash
sudo cp httpsify.service /etc/systemd/system/
sudo systemctl enable --now httpsify
sudo systemctl status httpsify
```

### launchd (macOS)

```bash
sudo cp com.httpsify.plist /Library/LaunchDaemons/
sudo launchctl load /Library/LaunchDaemons/com.httpsify.plist
```

---

## Error Responses

httpsify returns helpful JSON errors so your frontend can handle them:

```json
// Port is blocked
{"error": "Port 22 is not allowed", "hint": "This port is either denied or outside the allowed range", "example": "https://8000.localhost"}

// No service running
{"error": "Connection refused", "hint": "No service is listening on port 8000"}

// Bad format
{"error": "Invalid host format", "hint": "Use format: https://<port>.localhost", "example": "https://8000.localhost"}
```

---

## Troubleshooting

<details>
<summary><strong>Certificate warning in browser</strong></summary>

**With mkcert (recommended):** Run `mkcert -install` to trust the local CA.

**With self-signed:** Import `./cert/ca.pem` into your browser:
- **Chrome:** `chrome://settings/certificates` → Authorities → Import
- **Firefox:** `about:preferences#privacy` → Certificates → View → Authorities → Import
- **Windows:** Double-click `ca.pem` → Install → Local Machine → Trusted Root
</details>

<details>
<summary><strong>Connection refused</strong></summary>

Your backend isn't running on that port:
```bash
lsof -i :8000          # macOS/Linux
netstat -ano | findstr :8000  # Windows
```
</details>

<details>
<summary><strong>Address already in use</strong></summary>

Another process is using port 443:
```bash
sudo lsof -i :443
# Or use a different port:
httpsify --listen :8443
```
</details>

<details>
<summary><strong>WebSocket not connecting</strong></summary>

1. Verify your WS server is running on the target port
2. Enable verbose logging: `httpsify --verbose`
3. Check that your server handles the Upgrade handshake
</details>

---

## Architecture

```
                     ┌─────────────────────────────┐
                     │         httpsify             │
                     │    (TLS Termination)         │
                     └──────────┬──────────────────┘
                                │
         ┌──────────────────────┼──────────────────────┐
         │                      │                      │
    ┌────▼────┐           ┌─────▼────┐          ┌──────▼─────┐
    │ :3000   │           │  :8000   │          │   :5173    │
    │ React   │           │ Laravel  │          │   Vite     │
    └─────────┘           └──────────┘          └────────────┘

  https://3000.localhost  https://8000.localhost  https://5173.localhost
```

## Project Structure

```
httpsify/
├── cmd/httpsify/          # CLI entry point
│   └── main.go
├── internal/
│   ├── config/            # Configuration & validation
│   ├── logging/           # Structured JSON logging
│   ├── proxy/             # Reverse proxy + WebSocket
│   └── tls/               # TLS & certificate generation
├── install.sh             # One-line installer
├── httpsify.service       # systemd unit
├── Makefile               # Build targets
└── .github/workflows/     # CI/CD
```

---

## Contributing

```bash
git clone https://github.com/imcanugur/httpsify.git
cd httpsify
make test        # Run tests
make build       # Build binary
make lint        # Run linter
```

PRs welcome! Please run `make test` before submitting.

## Author

<p>
  <a href="https://github.com/imcanugur">
    <img src="https://img.shields.io/badge/built%20by-imcanugur-6c63ff?style=flat-square&logo=github" alt="imcanugur"/>
  </a>
</p>

**Can Ugur** — Full-stack developer & open-source enthusiast

- GitHub: [@imcanugur](https://github.com/imcanugur)
- Project: [httpsify](https://github.com/imcanugur/httpsify)

> *"I was tired of configuring nginx for every new port. So I built httpsify — one binary, zero config, infinite ports."*

If this tool saved you time, consider giving it a ⭐ on [GitHub](https://github.com/imcanugur/httpsify)!

## License

MIT License — see [LICENSE](LICENSE) for details.

---

<p align="center">
  <strong>Stop configuring. Start building.</strong>
  <br/><br/>
  Made with ☕ by <a href="https://github.com/imcanugur"><strong>@imcanugur</strong></a>
  <br/><br/>
  <a href="https://github.com/imcanugur/httpsify"><img src="https://img.shields.io/github/stars/imcanugur/httpsify?style=social" alt="Stars"></a>
</p>
