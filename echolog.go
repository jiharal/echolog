package applog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"gopkg.in/natefinch/lumberjack.v2"
)

// LogLevel represents the severity of the log entry
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp   time.Time       `json:"timestamp"`
	Level       string          `json:"level"`
	RequestID   string          `json:"request_id"`
	Method      string          `json:"method"`
	URI         string          `json:"uri"`
	Status      int             `json:"status"`
	Latency     time.Duration   `json:"latency"`
	ReqHeaders  json.RawMessage `json:"request_headers,omitempty"`
	ReqBody     string          `json:"request_body,omitempty"`
	RespHeaders json.RawMessage `json:"response_headers,omitempty"`
	RespBody    string          `json:"response_body,omitempty"`
	Error       string          `json:"error,omitempty"`
	Stack       string          `json:"stack_trace,omitempty"`
	ClientIP    string          `json:"client_ip"`
	UserAgent   string          `json:"user_agent"`
}

// LoggerConfig provides configuration options for the logger middleware
type LoggerConfig struct {
	// File output configuration
	Filename   string
	MaxSize    int  // megabytes
	MaxBackups int  // number of backups
	MaxAge     int  // days
	Compress   bool // compress old files

	// Logger behavior configuration
	LogLevel          LogLevel
	SkipPaths         []string
	MaxBodySize       int64  // maximum size of body to log
	RequestIDHeader   string // header to use for request ID
	DisableRequestLog bool   // disable request body logging
	DisableStackTrace bool   // disable stack trace for errors

	// Output options
	JSONOutput bool // output logs in JSON format
}

type Logger struct {
	config LoggerConfig
	lumber *lumberjack.Logger
	mu     sync.Mutex
}

// NewLogger creates a new logger instance with the given configuration
func NewLogger(config LoggerConfig) *Logger {
	// Set defaults if not provided
	if config.MaxBodySize == 0 {
		config.MaxBodySize = 1024 * 1024 // 1MB default
	}
	if config.RequestIDHeader == "" {
		config.RequestIDHeader = "X-Request-ID"
	}

	logger := &Logger{
		config: config,
		lumber: &lumberjack.Logger{
			Filename:   config.Filename,
			MaxSize:    config.MaxSize,
			MaxBackups: config.MaxBackups,
			MaxAge:     config.MaxAge,
			Compress:   config.Compress,
		},
	}

	return logger
}

// Middleware returns an Echo middleware handler
func (l *Logger) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip logging for specified paths
			if l.shouldSkip(c.Request().URL.Path) {
				return next(c)
			}

			start := time.Now()
			req := c.Request()
			res := c.Response()

			// Create log entry
			entry := &LogEntry{
				Timestamp: start,
				RequestID: req.Header.Get(l.config.RequestIDHeader),
				Method:    req.Method,
				URI:       req.RequestURI,
				ClientIP:  c.RealIP(),
				UserAgent: req.UserAgent(),
			}

			// Log request headers and body
			if !l.config.DisableRequestLog {
				l.captureRequest(req, entry)
			}

			// Create custom response writer to capture response
			resWriter := &responseWriter{
				ResponseWriter: res.Writer,
				body:           new(bytes.Buffer),
			}
			res.Writer = resWriter

			// Process request
			err := next(c)

			// Capture response details
			entry.Status = res.Status
			entry.Latency = time.Since(start)

			if err != nil {
				entry.Error = err.Error()
				if !l.config.DisableStackTrace {
					entry.Stack = l.getStackTrace()
				}
			}

			// Log response headers and body
			l.captureResponse(res, resWriter, entry)

			// Write log entry
			l.writeLog(entry)

			return err
		}
	}
}

// Helper methods

func (l *Logger) shouldSkip(path string) bool {
	for _, p := range l.config.SkipPaths {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

func (l *Logger) captureRequest(req *http.Request, entry *LogEntry) {
	// Capture headers
	if headers, err := json.Marshal(req.Header); err == nil {
		entry.ReqHeaders = headers
	}

	// Capture body if not multipart
	if req.Header.Get("Content-Type") != "multipart/form-data" {
		body, err := io.ReadAll(io.LimitReader(req.Body, l.config.MaxBodySize))
		if err == nil {
			entry.ReqBody = string(body)
			req.Body = io.NopCloser(bytes.NewBuffer(body))
		}
	}
}

func (l *Logger) captureResponse(res *echo.Response, rw *responseWriter, entry *LogEntry) {
	// Capture headers
	if headers, err := json.Marshal(res.Header()); err == nil {
		entry.RespHeaders = headers
	}

	// Capture body
	if rw.body.Len() > 0 {
		entry.RespBody = rw.body.String()
	}
}

func (l *Logger) writeLog(entry *LogEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.config.JSONOutput {
		if data, err := json.Marshal(entry); err == nil {
			l.lumber.Write(append(data, '\n'))
		}
	} else {
		// Format as text
		fmt.Fprintf(l.lumber, "[%s] %s %s %s %d %v\n",
			entry.Timestamp.Format(time.RFC3339),
			entry.RequestID,
			entry.Method,
			entry.URI,
			entry.Status,
			entry.Latency,
		)
		if entry.Error != "" {
			fmt.Fprintf(l.lumber, "Error: %s\n", entry.Error)
			if entry.Stack != "" {
				fmt.Fprintf(l.lumber, "Stack: %s\n", entry.Stack)
			}
		}
	}
}

func (l *Logger) getStackTrace() string {
	buf := make([]byte, 1024)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// responseWriter captures the response body while writing it
type responseWriter struct {
	http.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}
