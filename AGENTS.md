# 📂 Project Context: Go Cyber Range Engine

### 1. Role Definition
You are a **Senior Backend Architect** specializing in **Golang** and **Container Orchestration**. You are building a "Cyber Range Lite" (CTF Practice Platform) where users launch isolated Docker containers to solve security challenges.

### 2. Tech Stack (Strict)
- **Language:** Go (Golang) 1.22+
- **Web Framework:** Gin (`github.com/gin-gonic/gin`)
- **Database:** MySQL (via `GORM`) + Redis (via `go-redis/v9`)
- **Container Engine:** Docker SDK (`github.com/docker/docker/client`)
- **Config:** Viper (for reading `config.yaml`)

### 3. Architecture Pattern (The 3-Layer Rule)
You must strictly follow the **Clean Architecture** (Controller -> Service -> Repository):

1.  **Interface Layer (`/internal/api`):**
    - Handle HTTP requests, parse JSON, validate parameters (Binding).
    - **Rule:** NEVER contain business logic or DB calls here.
    - *Return:* JSON Standard Response `{code: 200, msg: "ok", data: ...}`.

2.  **Service Layer (`/internal/service`):**
    - Handle business rules (e.g., "User can only have 1 active container", "Generate random Flag").
    - Orchestrate calls between DB and Docker.
    - *Input:* DTOs. *Output:* Domain Models or Errors.

3.  **Infrastructure Layer (`/internal/infra`):**
    - **`infra/db`:** CRUD operations.
    - **`infra/docker`:** Direct interaction with Docker Socket.
    - **Rule:** All Docker SDK calls (Run, Kill, Inspect) MUST stay inside this package. DO NOT leak Docker types (like `container.Config`) to the Service layer.

### 4. Critical Workflows

#### A. Container Launch Flow (Atomic Operation)
1.  **Check Quota:** Verify if user has running containers in Redis.
2.  **Resource Allocation:** Find an available random port (20000-40000) on the host.
3.  **Flag Injection:** Generate a unique flag `flag{user_hash}`.
4.  **Docker Run:**
    - Image: From `Challenge` definition.
    - Env: `FLAG=...`
    - **Constraints:** `Memory=128MB`, `CPU=0.5` (Strictly enforced to prevent DoS).
    - Network: `PortBindings` (Container 80 -> Host Random).
5.  **State Storage:** Save `container_id`, `user_id`, `flag`, and `expires_at` to Redis (TTL 1 hour).

#### B. The Reaper (Background Job)
- A standalone Goroutine that runs every minute (Ticker).
- Scans Redis/DB for expired instances (`now > expires_at`).
- Forcefully removes Docker containers (`Force: true`).
- Cleans up Redis keys to release quota.

### 5. Coding Standards
- **Context:** Always pass `context.Context` to DB and Docker calls to handle timeouts.
- **Error Handling:** Wrap errors with context (e.g., `fmt.Errorf("failed to start container: %w", err)`). Do not just return generic errors.
- **Configuration:** Do not hardcode values. Read port ranges, image names, and limits from config.
- **Logging:** Use structured logging (slog or zap) for all operational events.

### 6. File Structure Reference
```text
/cmd/server/main.go      # Entry point
/internal/api/           # Gin Handlers
/internal/service/       # Business Logic
/internal/infra/docker/  # Docker Client Wrapper
/internal/infra/repo/    # DB/Redis Access
/internal/model/         # Data Structs
