# CMSC501-1 Practicum Project

Coursework repository for **CMSC501-1 Structure of Programming Language — Summer 2026**.

Shared bill splitter REST API (who owes whom for roommates, trips, and group expenses).

## Repository

https://github.com/uona-cmsc501-1-summer2026-nanlin/practicum-project

## Run (MVP)

```bash
go run .
```

API: `http://localhost:55555/api/v1` — health: `GET /health`

### Tests

```bash
go test ./...
```

### Live demo (Deliverable #2)

Terminal 1:

```bash
go run .
```

Terminal 2 (PowerShell):

```powershell
.\scripts\demo.ps1
```

Sample JSON bodies live in `tmp/` for manual runs.

Postman: import [`docs/swagger/postman/Shared-Bill-Splitter.postman_collection.json`](docs/swagger/postman/Shared-Bill-Splitter.postman_collection.json) — run folder **1. Valid Demo Flow** top to bottom.

Swagger UI (when server is running): **http://localhost:55555/swagger**

## Project layout

| Path | Role |
|------|------|
| `main.go` | Starts Fiber, opens SQLite, registers routes |
| `internal/models/` | DB entities and JSON request/response types |
| `internal/database/` | SQLite connect + AutoMigrate |
| `internal/handlers/` | HTTP + validation via [gobeetle/reply](https://github.com/gobeetle/reply) |
| `internal/settle/` | Pure settlement math (balances + transfers) |
| `docs/swagger/` | OpenAPI 3.1 split spec + merged `generate/openapi.yaml` + Postman |
| `internal/swagger/` | Serves `/swagger` UI and `/swagger/specification` |
| `docs/deliverable2.md` | Code walkthrough + sample API results |

