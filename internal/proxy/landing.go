package proxy

import (
	"fmt"
	"net/http"
)

const landingPageHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>httpsify &bull; Ready</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600&family=JetBrains+Mono:wght@400;500&display=swap" rel="stylesheet">
    <style>
        :root {
            --bg: #ffffff;
            --fg: #111111;
            --muted: #666666;
            --accent: #000000;
            --border: #eeeeee;
            --success: #10b981;
            --font-sans: 'Inter', -apple-system, system-ui, sans-serif;
            --font-mono: 'JetBrains Mono', monospace;
        }

        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
            -webkit-font-smoothing: antialiased;
        }

        body {
            background: var(--bg);
            color: var(--fg);
            font-family: var(--font-sans);
            height: 100vh;
            display: flex;
            flex-direction: column;
            justify-content: center;
            align-items: center;
            overflow: hidden;
        }

        /* Minimalist Grain Texture */
        body::before {
            content: "";
            position: fixed;
            top: 0;
            left: 0;
            width: 100%%;
            height: 100%%;
            opacity: 0.015;
            z-index: 1000;
            pointer-events: none;
            background-image: url("data:image/svg+xml,%%3Csvg viewBox='0 0 200 200' xmlns='http://www.w3.org/2000/svg'%%3E%%3Cfilter id='noiseFilter'%%3E%%3CfeTurbulence type='fractalNoise' baseFrequency='0.8' numOctaves='4' stitchTiles='stitch'/%%3E%%3C/filter%%3E%%3Crect width='100%%25' height='100%%25' filter='url(%%23noiseFilter)'/%%3E%%3C/svg%%3E");
        }

        .content {
            width: 100%%;
            max-width: 440px;
            padding: 2rem;
            animation: reveal 1s cubic-bezier(0.16, 1, 0.3, 1);
        }

        @keyframes reveal {
            from { opacity: 0; transform: translateY(8px); }
            to { opacity: 1; transform: translateY(0); }
        }

        .status-badge {
            display: inline-flex;
            align-items: center;
            gap: 8px;
            font-size: 11px;
            font-weight: 600;
            letter-spacing: 0.03em;
            text-transform: uppercase;
            color: var(--success);
            margin-bottom: 24px;
            background: #f0fdf4;
            padding: 4px 10px;
            border-radius: 100px;
        }

        .dot {
            width: 6px;
            height: 6px;
            background: var(--success);
            border-radius: 50%%;
            animation: pulse 2s infinite;
        }

        @keyframes pulse {
            0%% { transform: scale(0.9); opacity: 0.5; }
            50%% { transform: scale(1.1); opacity: 1; }
            100%% { transform: scale(0.9); opacity: 0.5; }
        }

        h1 {
            font-size: 28px;
            font-weight: 600;
            letter-spacing: -0.03em;
            margin-bottom: 12px;
            color: var(--accent);
        }

        .description {
            font-size: 15px;
            color: var(--muted);
            margin-bottom: 40px;
            line-height: 1.5;
            font-weight: 400;
        }

        .usage {
            display: flex;
            flex-direction: column;
            gap: 10px;
        }

        .command {
            background: #f9f9f9;
            border: 1px solid var(--border);
            border-radius: 12px;
            padding: 14px 18px;
            display: flex;
            justify-content: space-between;
            align-items: center;
            transition: all 0.2s ease;
        }

        .command:hover {
            background: #f3f3f3;
            transform: scale(1.01);
        }

        .cmd-text {
            font-family: var(--font-mono);
            font-size: 13px;
            font-weight: 500;
            color: #000;
        }

        .cmd-label {
            color: var(--muted);
            font-size: 11px;
            font-weight: 500;
        }

        .footer {
            position: absolute;
            bottom: 40px;
            font-size: 10px;
            color: #bbbbbb;
            font-weight: 500;
            letter-spacing: 0.05em;
            text-transform: uppercase;
            display: flex;
            gap: 16px;
        }

        .footer a {
            color: #bbbbbb;
            text-decoration: none;
            transition: color 0.2s ease;
        }

        .footer a:hover {
            color: #000000;
        }

        .bg-gradient {
            position: fixed;
            top: 0;
            left: 0;
            width: 100%%;
            height: 100%%;
            background: radial-gradient(circle at 50%% -20%%, #f0f0f0 0%%, transparent 60%%);
            z-index: -1;
            pointer-events: none;
        }
    </style>
</head>
<body>
    <div class="bg-gradient"></div>
    
    <div class="content">
        <div class="status-badge">
            <div class="dot"></div>
            System Operational
        </div>
        
        <h1>httpsify</h1>
        <p class="description">Stop configuring. Start building.</p>

        <div class="usage">
            <div class="command">
                <span class="cmd-text">3000.localhost</span>
                <span class="cmd-label">Frontend</span>
            </div>
            <div class="command">
                <span class="cmd-text">8000.localhost</span>
                <span class="cmd-label">Backend</span>
            </div>
            <div class="command">
                <span class="cmd-text">5173.localhost</span>
                <span class="cmd-label">Vite</span>
            </div>
        </div>
    </div>

    <div class="footer">
        <span>Automated SSL Infrastructure &bull; v%s</span>
        <a href="https://github.com/imcanugur/httpsify" target="_blank" rel="noopener noreferrer">View on GitHub</a>
    </div>
</body>
</html>
`

func (s *Server) serveLandingPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.WriteHeader(http.StatusOK)

	version := "1.0.0" 
	fmt.Fprintf(w, landingPageHTML, version)
}
