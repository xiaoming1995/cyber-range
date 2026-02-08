# Database & Models

## 1. Model Separation (Crucial)
Strictly separate **Domain Entities** from **Database Models**.

- **Domain Entities (`internal/core`)**: Pure Go structs, no tags (like `gorm:` or `json:`), no framework dependencies. Represent business data.
- **DB Models (`internal/infrastructure/persistence/models`)**: Structs with DB tags, specific to the ORM (GORM/SQLX).

## 2. Conversion
The **Repository** (Adapter) is responsible for converting between them.

```go
// Infrastructure Layer
func (r *SQLRepository) GetUser(ctx context.Context, id string) (*core.User, error) {
    var dbModel UserDBModel
    // 1. Fetch from DB
    if err := r.db.First(&dbModel, "id = ?", id).Error; err != nil {
        return nil, err
    }
    
    // 2. Convert to Domain Entity
    return &core.User{
        ID:   dbModel.ID,
        Name: dbModel.Username,
    }, nil
}
```

## 3. Transaction Boundary
Transactions should ideally be managed by a `UnitOfWork` pattern or carefully orchestrated in the Service layer without leaking DB details (e.g., passing a `context` that carries the transaction).
