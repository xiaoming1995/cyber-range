# Error Handling Strategy

## 1. Domain Errors (Core Layer)
Define sentinel errors or error types in `internal/core/errors.go` (or near the interface definitions). These should be agnostic of the transport layer (HTTP/gRPC).

```go
// internal/core/errors.go
var (
    ErrResourceNotFound = errors.New("resource not found")
    ErrInvalidInput     = errors.New("invalid input")
    ErrConflict         = errors.New("resource conflict")
)
```

## 2. Service Layer Responsibility
The service layer **MUST** return errors defined in the Core layer.
- If an infrastructure adapter returns an error (e.g., specific SQL error), the Service or Repository **MUST** wrap or map it to a Core error.
- **Do not** leak implementation details (like `sql.ErrNoRows`) to the API layer.

```go
if err != nil {
    if errors.Is(err, sql.ErrNoRows) {
        return core.ErrResourceNotFound
    }
    return fmt.Errorf("failed to get user: %w", err)
}
```

## 3. API Layer (Handlers)
The Handler is responsible for mapping Core errors to HTTP Status Codes.

| Core Error | HTTP Status |
| :--- | :--- |
| `ErrResourceNotFound` | `404 Not Found` |
| `ErrInvalidInput` | `400 Bad Request` |
| `ErrConflict` | `409 Conflict` |
| Other/Unknown | `500 Internal Server Error` |

Use a centralized helper (e.g., `api.ErrorResponse(c, err)`) to standardize error JSON responses.
