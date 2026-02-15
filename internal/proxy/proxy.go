package proxy

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/imcanugur/httpsify/internal/config"
	"github.com/imcanugur/httpsify/internal/logging"
)

var (
	hostPattern = regexp.MustCompile(`^(\d+)\.(localhost|localtest\.me)(?::\d+)?$`)
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Hint    string `json:"hint,omitempty"`
	Example string `json:"example,omitempty"`
}

type Server struct {
	cfg           *config.Config
	logger        *logging.Logger
	requestIDGen  atomic.Uint64
	transportPool sync.Pool
}

func NewServer(cfg *config.Config, logger *logging.Logger) *Server {
	return &Server{
		cfg:    cfg,
		logger: logger,
		transportPool: sync.Pool{
			New: func() interface{} {
				return &http.Transport{
					DialContext: (&net.Dialer{
						Timeout:   time.Duration(cfg.DialTimeout) * time.Second,
						KeepAlive: 30 * time.Second,
					}).DialContext,
					MaxIdleConns:          100,
					MaxIdleConnsPerHost:   10,
					IdleConnTimeout:       time.Duration(cfg.IdleTimeout) * time.Second,
					TLSHandshakeTimeout:   10 * time.Second,
					ExpectContinueTimeout: 1 * time.Second,
					DisableCompression:    true, 
				}
			},
		},
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	requestID := s.generateRequestID()
	ctx := context.WithValue(r.Context(), logging.RequestIDKey, requestID)
	r = r.WithContext(ctx)

	w.Header().Set("X-Request-ID", requestID)

	port, err := s.parseHost(r.Host)
	if err != nil {
		s.handleError(w, r, requestID, http.StatusBadRequest, err.Error(), 
			"Use format: https://<port>.localhost",
			"https://8000.localhost")
		s.logger.InvalidHost(requestID, r.Host, err.Error())
		return
	}

	if !s.cfg.IsPortAllowed(port) {
		s.handleError(w, r, requestID, http.StatusForbidden, 
			fmt.Sprintf("Port %d is not allowed", port),
			"This port is either denied or outside the allowed range",
			"https://8000.localhost")
		s.logger.PortDenied(requestID, port, "port in deny list or outside allow range")
		return
	}

	rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

	if isWebSocketRequest(r) {
		s.handleWebSocket(rw, r, requestID, port)
	} else {
		s.handleHTTP(rw, r, requestID, port)
	}

	latency := time.Since(start)
	s.logger.LogRequest(ctx, r.Method, r.Host, port, rw.statusCode, latency, rw.bytesWritten, rw.err)
}

func (s *Server) parseHost(host string) (int, error) {
	matches := hostPattern.FindStringSubmatch(host)
	if matches == nil {
		return 0, fmt.Errorf("invalid host format: %s", host)
	}

	portStr := matches[1]
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, fmt.Errorf("invalid port number: %s", portStr)
	}

	if err := config.ValidatePort(port); err != nil {
		return 0, err
	}

	return port, nil
}

func (s *Server) handleHTTP(w *responseWriter, r *http.Request, requestID string, port int) {
	target, _ := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", port))

	transport := s.transportPool.Get().(*http.Transport)
	defer s.transportPool.Put(transport)

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.Host = r.Host

			if clientIP, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
				if prior := req.Header.Get("X-Forwarded-For"); prior != "" {
					clientIP = prior + ", " + clientIP
				}
				req.Header.Set("X-Forwarded-For", clientIP)
			}
			req.Header.Set("X-Forwarded-Proto", "https")
			req.Header.Set("X-Forwarded-Host", r.Host)
			req.Header.Set("X-Request-ID", requestID)

			s.logger.Debug("proxying request",
				"request_id", requestID,
				"method", req.Method,
				"path", req.URL.Path,
				"target", target.String(),
			)
		},
		Transport: transport,
		ErrorHandler: func(rw http.ResponseWriter, req *http.Request, err error) {
			s.logger.ProxyError(requestID, port, err)
			w.err = err

			errMsg := "Backend service unavailable"
			hint := fmt.Sprintf("Make sure a service is running on port %d", port)
			
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				errMsg = "Request timed out"
				hint = "The backend service took too long to respond"
			} else if strings.Contains(err.Error(), "connection refused") {
				errMsg = "Connection refused"
				hint = fmt.Sprintf("No service is listening on port %d", port)
			}

			s.writeJSONError(rw, http.StatusBadGateway, errMsg, hint, "")
		},
		ModifyResponse: func(resp *http.Response) error {
			if resp.Header.Get("Access-Control-Allow-Origin") == "" {
			}
			return nil
		},
	}

	proxy.ServeHTTP(w, r)
}

func (s *Server) handleWebSocket(w *responseWriter, r *http.Request, requestID string, port int) {
	s.logger.WebSocketUpgrade(requestID, port)

	backendAddr := fmt.Sprintf("127.0.0.1:%d", port)
	backendConn, err := net.DialTimeout("tcp", backendAddr, time.Duration(s.cfg.DialTimeout)*time.Second)
	if err != nil {
		s.logger.ProxyError(requestID, port, err)
		s.handleError(w.ResponseWriter, r, requestID, http.StatusBadGateway,
			"Failed to connect to backend",
			fmt.Sprintf("Make sure a service is running on port %d", port),
			"")
		return
	}
	defer backendConn.Close()

	hijacker, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		s.handleError(w.ResponseWriter, r, requestID, http.StatusInternalServerError,
			"WebSocket hijacking not supported",
			"",
			"")
		return
	}

	clientConn, clientBuf, err := hijacker.Hijack()
	if err != nil {
		s.logger.ProxyError(requestID, port, err)
		s.handleError(w.ResponseWriter, r, requestID, http.StatusInternalServerError,
			"Failed to hijack connection",
			"",
			"")
		return
	}
	defer clientConn.Close()

	if err := r.Write(backendConn); err != nil {
		s.logger.ProxyError(requestID, port, fmt.Errorf("failed to write request to backend: %w", err))
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		io.Copy(clientConn, backendConn)
		clientConn.Close()
	}()

	go func() {
		defer wg.Done()
		if clientBuf.Reader.Buffered() > 0 {
			io.CopyN(backendConn, clientBuf, int64(clientBuf.Reader.Buffered()))
		}
		io.Copy(backendConn, clientConn)
		backendConn.Close()
	}()

	wg.Wait()
}

func isWebSocketRequest(r *http.Request) bool {
	connection := strings.ToLower(r.Header.Get("Connection"))
	upgrade := strings.ToLower(r.Header.Get("Upgrade"))
	return strings.Contains(connection, "upgrade") && upgrade == "websocket"
}

func (s *Server) handleError(w http.ResponseWriter, r *http.Request, requestID string, statusCode int, errMsg, hint, example string) {
	s.writeJSONError(w, statusCode, errMsg, hint, example)
}

func (s *Server) writeJSONError(w http.ResponseWriter, statusCode int, errMsg, hint, example string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	resp := ErrorResponse{
		Error:   errMsg,
		Hint:    hint,
		Example: example,
	}

	json.NewEncoder(w).Encode(resp)
}

func (s *Server) generateRequestID() string {
	id := s.requestIDGen.Add(1)
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), id)
}

type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
	err          error
	wroteHeader  bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.wroteHeader {
		rw.statusCode = code
		rw.wroteHeader = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += int64(n)
	return n, err
}

func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := rw.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, fmt.Errorf("hijacking not supported")
}

func (rw *responseWriter) Flush() {
	if flusher, ok := rw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}
