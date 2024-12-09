# Echolog

Echolog is a comprehensive HTTP request/response logging middleware for the Echo framework, providing robust logging capabilities with rotation support and structured output formats.

## Features

- Structured logging with JSON support
- Request/Response body and header capture
- Log rotation with compression
- Request ID tracking
- Configurable logging levels
- Path-based logging skip
- Stack trace capture for errors
- Thread-safe logging
- Customizable body size limits
- Client IP and User-Agent logging

## Installation

```bash
go get github.com/jiharal/echolog
```

## Quick Start

```go
package main

import (
    "github.com/labstack/echo/v4"
    "github.com/jiharal/echolog"
)

func main() {
    e := echo.New()

    // Create basic logger configuration
    config := Echolog.LoggerConfig{
        Filename: "app.log",
        MaxSize:  10, // 10MB
    }

    // Initialize and use the logger
    logger := Echolog.NewLogger(config)
    e.Use(logger.Middleware())

    e.Start(":8080")
}
```

## Configuration

### LoggerConfig Options

```go
type LoggerConfig struct {
    // File output configuration
    Filename   string // Log file path
    MaxSize    int    // Maximum size in megabytes before rotation
    MaxBackups int    // Maximum number of old log files to retain
    MaxAge     int    // Maximum number of days to retain old log files
    Compress   bool   // Compress rotated files

    // Logger behavior configuration
    LogLevel           LogLevel // Minimum log level to record
    SkipPaths         []string // Paths to exclude from logging
    MaxBodySize       int64    // Maximum size of body to log
    RequestIDHeader   string   // Header to use for request ID
    DisableRequestLog bool     // Disable request body logging
    DisableStackTrace bool     // Disable stack trace for errors

    // Output options
    JSONOutput bool   // Output logs in JSON format
}
```

### Log Levels

```go
const (
    DEBUG LogLevel = iota
    INFO
    WARN
    ERROR
)
```

## Advanced Usage

### Complete Configuration Example

```go
config := Echolog.LoggerConfig{
    // File configuration
    Filename:   "/var/log/app.log",
    MaxSize:    10,    // 10MB
    MaxBackups: 5,     // Keep 5 old files
    MaxAge:     30,    // 30 days
    Compress:   true,  // Compress old files

    // Logger configuration
    LogLevel:           Echolog.INFO,
    SkipPaths:         []string{"/health", "/metrics"},
    MaxBodySize:       1024 * 1024, // 1MB
    RequestIDHeader:   "X-Request-ID",
    DisableRequestLog: false,
    DisableStackTrace: false,
    JSONOutput:        true,
}
```

### JSON Output Format

When `JSONOutput` is enabled, logs are structured as follows:

```json
{
    "timestamp": "2024-12-09T10:00:00Z",
    "level": "INFO",
    "request_id": "req-123",
    "method": "POST",
    "uri": "/api/users",
    "status": 200,
    "latency": "145ms",
    "request_headers": {...},
    "request_body": "...",
    "response_headers": {...},
    "response_body": "...",
    "client_ip": "192.168.1.1",
    "user_agent": "Mozilla/5.0...",
    "error": "...",
    "stack_trace": "..."
}
```

### Skipping Paths

```go
config := Echolog.LoggerConfig{
    // ... other config
    SkipPaths: []string{
        "/health",
        "/metrics",
        "/public/",
    },
}
```

### Custom Request ID Header

```go
config := Echolog.LoggerConfig{
    // ... other config
    RequestIDHeader: "X-Custom-Request-ID",
}
```

## Best Practices

1. **Log Rotation**

   - Set appropriate `MaxSize`, `MaxBackups`, and `MaxAge` for your needs
   - Enable compression for long-term storage

2. **Body Size Limits**

   - Set reasonable `MaxBodySize` to prevent memory issues
   - Consider disabling request logging for file uploads

3. **Performance**

   - Use `SkipPaths` for high-traffic endpoints that don't need logging
   - Disable stack traces in production if not needed

4. **Security**
   - Be cautious with logging sensitive data
   - Consider masking sensitive headers and body fields

## Error Handling

The logger handles errors gracefully and includes them in the log output:

```go
if err != nil {
    // Logs will include:
    // - Error message
    // - Stack trace (if enabled)
    // - Request context
    // - Response status
}
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For issues, questions, or contributions, please visit:
https://github.com/jiharal/echolog/issues
