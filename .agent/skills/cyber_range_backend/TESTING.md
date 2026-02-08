# Testing Guidelines

## 1. Unit Testing Strategy
Because we use Clean Architecture, our **Service Layer** contains the pure business logic and is the most important candidate for Unit Tests.

- **Mock Dependencies**: Use `mockgen` to create mocks for interfaces defined in `internal/core`.
- **Table-Driven Tests**: Use Go's standard table-driven testing pattern.

## 2. Generating Mocks
Use `uber-go/mock` (formerly `golang/mock`).

```bash
# Install mockgen
go install go.uber.org/mock/mockgen@latest

# Generate mock for ContainerEngine interface
mockgen -source=internal/core/ports.go -destination=internal/infrastructure/mock/mock_ports.go -package=mock
```

## 3. Example Service Test
```go
func TestStartChallenge(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockEngine := mock.NewMockContainerEngine(ctrl)
    service := NewChallengeService(mockEngine)

    // Expectation
    mockEngine.EXPECT().StartContainer(gomock.Any(), "nginx", gomock.Any()).Return("id-123", 80, nil)

    // Action
    id, err := service.StartChallenge(ctx, "nginx")

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "id-123", id)
}
```
