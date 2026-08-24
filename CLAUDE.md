# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Run locally (requires MySQL running)
go run cmd/main.go

# Build binary
go build -o bin/short_url cmd/main.go

# Docker build & run
make run         # build + docker compose up -d
make down        # docker compose down
make clean       # docker compose down -v (deletes volumes)

# Swagger doc generation (after editing annotations)
swag init --parseDependency --parseInternal -g cmd/main.go -o cmd/docs
```

## Architecture

**Request flow:** `main.go` → `config.InitConfig()` (loads YAML + connects DB + auto-migrates) → `routers.SetupRouter()` (Gin engine) → `handler/` functions (business logic).

### Package roles

| Package | Responsibility |
|---------|---------------|
| `cmd/` | Entry point (`main.go`) + Swagger generated files (`docs/`) |
| `config/` | Viper-based YAML config loading + MySQL/GORM initialization |
| `global/` | Shared `*gorm.DB` instance, accessed by handlers |
| `model/` | GORM struct (`ShortUrl`) + request DTO (`CreateShortUrlRequest`) |
| `handler/` | Gin handlers — `GenerateShortUrl` (POST) and `VisitShortUrl` (GET) |
| `routers/` | Gin route registration + Swagger UI mounting |
| `utils/` | `GenerateShortCode(id)` — Base62 encoding with offset |

### Key patterns

- **Short code generation:** Record inserted first to get auto-increment ID, then `GenerateShortCode(id + 100000)` encodes to Base62. Custom codes skip encoding.
- **Concurrent click counting:** `go func()` + `UpdateColumn(..., gorm.Expr("click_count + 1"))` — atomic increment without full-row save.
- **Soft delete:** `DeletedAt gorm.DeletedAt` on the model — GORM filters these by default.
- **Schema management:** `AutoMigrate()` runs on startup — no manual migration files.
- **Config driven:** All environment-specific values in `config/config.yaml` loaded via Viper.

### API routes

| Method | Path | Handler |
|--------|------|---------|
| GET | `/s/:shorturl` | `VisitShortUrl` — 302 redirect, 404/410 on missing/expired |
| POST | `/api/v1/shorten` | `GenerateShortUrl` — creates short link, returns JSON |
| GET | `/swagger/*any` | Swagger UI |
| GET/POST | `/test` | Connectivity check |
