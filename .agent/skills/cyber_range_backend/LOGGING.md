# Logging Standards

## 1. Structured Logging
Use structured logging (Key-Value pairs) exclusively.
- **Recommended Library**: standard library `log/slog` (Go 1.21+) or `zap`.
- **Forbidden**: `fmt.Printf`, `log.Println` (unless during early startup).

## 2. Contextual Logging
Always include context when possible to enable tracing.

```go
// Good
slog.Info("starting container", 
    "image", image,
    "user_id", userID,
    "trace_id", traceID, // Extract from ctx
)

// Bad
log.Printf("starting container %s for user %s", image, userID)
```

## 3. Log Levels
- **DEBUG**: Fine-grained information for diagnosing issues during development.
- **INFO**: Normal system operation events (startup, shutdown, important business actions).
- **WARN**: Unexpected situations that are recoverable (e.g., fallback to mock engine).
- **ERROR**: Errors that require attention (e.g., database connection failure).
