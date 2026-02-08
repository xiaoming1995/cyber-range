---
description: Guide for developing and extending the Cyber Range backend
---

# Cyber Range Backend Development Guide

This skill provides instructions for working with the Cyber Range backend, which follows **Clean Architecture** principles.

## Architecture Overview

The project is structured into layers:
1.  **`internal/core`**: Domain logic and interfaces (Ports).
2.  **`internal/infrastructure`**: External integrations (Adapters) like Docker, Database.
3.  **`internal/service`**: Business logic orchestration.
4.  **`internal/api`**: HTTP handlers and routing.

## Development Workflows

### 1. Adding a New Dependency (Adapter)

If you need to interact with a new external system (e.g., Redis, K8s):

1.  **Define the Interface (Port)** in `internal/core/ports.go`.
    ```go
    type RedisCache interface {
        Get(ctx context.Context, key string) (string, error)
        Set(ctx context.Context, key string, value string) error
    }
    ```

2.  **Implement the Adapter** in `internal/infrastructure/redis/client.go`.
    Ensure it satisfies the interface defined in Core.

3.  **Inject dependency** in `cmd/api/main.go`.

### 2. Adding a New Feature (Service)

1.  **Create Service** in `internal/service/<feature>_service.go`.
    Inject necessary ports (interfaces) into the service struct.

2.  **Implement Logic** using the injected dependencies.

### 3. Exposing via API

1.  **Create Handler** in `internal/api/handlers/<feature>_handler.go`.
    Inject the Service.

2.  **Register Route** in `cmd/api/main.go`.

## Common Tasks

### Running the Server
```bash
go run cmd/api/main.go
```

### Running Tests
```bash
go test ./...
```

### Dependency Management
Use `go mod tidy` to clean up dependencies.

## 📚 Detailed Guidelines

*   **[Error Handling](./ERROR_HANDLING.md)**: Strategy for domain errors and HTTP mapping.
*   **[Testing](./TESTING.md)**: Unit testing services and generating mocks.
*   **[Logging](./LOGGING.md)**: Structured logging standards.
*   **[Database](./DATABASE.md)**: Entity vs Model separation.
