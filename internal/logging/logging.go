package logging

import (
	"context"
	"log/slog"
	"os"
	"time"
)

type ContextKey string

const (
	RequestIDKey ContextKey = "request_id"
)

type Logger struct {
	*slog.Logger
	verbose   bool
	accessLog bool
}

func NewLogger(verbose, accessLog bool) *Logger {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				if t, ok := a.Value.Any().(time.Time); ok {
					a.Value = slog.StringValue(t.Format(time.RFC3339))
				}
			}
			return a
		},
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)

	return &Logger{
		Logger:    logger,
		verbose:   verbose,
		accessLog: accessLog,
	}
}

func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{
		Logger:    l.Logger.With(slog.String("request_id", requestID)),
		verbose:   l.verbose,
		accessLog: l.accessLog,
	}
}

func (l *Logger) AccessLog() bool {
	return l.accessLog
}

func (l *Logger) LogRequest(ctx context.Context, method, host string, targetPort int, statusCode int, latency time.Duration, bytesWritten int64, err error) {
	if !l.accessLog {
		return
	}

	attrs := []slog.Attr{
		slog.String("method", method),
		slog.String("host", host),
		slog.Int("target_port", targetPort),
		slog.Int("status", statusCode),
		slog.Duration("latency", latency),
		slog.Int64("bytes", bytesWritten),
	}

	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		attrs = append([]slog.Attr{slog.String("request_id", requestID)}, attrs...)
	}

	if err != nil {
		attrs = append(attrs, slog.String("error", err.Error()))
		l.LogAttrs(ctx, slog.LevelError, "request failed", attrs...)
	} else {
		l.LogAttrs(ctx, slog.LevelInfo, "request completed", attrs...)
	}
}

func (l *Logger) ServerStarting(addr, certPath, keyPath string, selfSigned bool) {
	l.Info("server starting",
		slog.String("listen", addr),
		slog.String("cert", certPath),
		slog.String("key", keyPath),
		slog.Bool("self_signed", selfSigned),
	)
}

func (l *Logger) ServerStarted(addr string) {
	l.Info("server listening",
		slog.String("addr", addr),
	)
}

func (l *Logger) CertGenerated(certPath, keyPath string) {
	l.Info("self-signed certificate generated",
		slog.String("cert", certPath),
		slog.String("key", keyPath),
	)
}

func (l *Logger) PortDenied(requestID string, port int, reason string) {
	l.Warn("port access denied",
		slog.String("request_id", requestID),
		slog.Int("port", port),
		slog.String("reason", reason),
	)
}

func (l *Logger) InvalidHost(requestID string, host string, reason string) {
	l.Warn("invalid host header",
		slog.String("request_id", requestID),
		slog.String("host", host),
		slog.String("reason", reason),
	)
}

func (l *Logger) ProxyError(requestID string, targetPort int, err error) {
	l.Error("proxy error",
		slog.String("request_id", requestID),
		slog.Int("target_port", targetPort),
		slog.String("error", err.Error()),
	)
}

func (l *Logger) WebSocketUpgrade(requestID string, targetPort int) {
	l.Debug("websocket upgrade",
		slog.String("request_id", requestID),
		slog.Int("target_port", targetPort),
	)
}
