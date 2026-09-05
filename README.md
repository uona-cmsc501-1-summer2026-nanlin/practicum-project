# CMSC501-1 Practicum Project

Coursework repository for **CMSC501-1 Structure of Programming Language — Summer 2026**.

**Split It** — REST API + minimal UI (who owes whom for roommates, trips, and group expenses).

## Repository

https://github.com/uona-cmsc501-1-summer2026-nanlin/practicum-project

## Run

```bash
go run .
# or: make run
```

| Resource | URL |
|----------|-----|
| **App UI** | http://localhost:55555/app/ |
| **API** | http://localhost:55555/api/v1 |
| **Health** | http://localhost:55555/health |
| **Swagger** | http://localhost:55555/swagger |

### Upgrading from Deliverable #2

Delete the old SQLite file so AutoMigrate creates `users` + `group_members`:

```powershell
Remove-Item billsplitter.db -ErrorAction SilentlyContinue
```

### Tests

```bash
go test ./...
```

Postman: import [`docs/swagger/postman/Shared-Bill-Splitter.postman_collection.json`](docs/swagger/postman/Shared-Bill-Splitter.postman_collection.json) — run folder **1. Valid Demo Flow** top to bottom.

## Project layout

| Path | Role |
|------|------|
| `main.go` | Starts Fiber, opens SQLite, serves `/app`, registers routes |
| `web/` | Minimal vanilla UI (People, Groups, Group detail) |
| `internal/models/` | DB entities and JSON request/response types |
| `internal/database/` | SQLite connect + AutoMigrate |
| `internal/handlers/` | HTTP + validation via [gobeetle/reply](https://github.com/gobeetle/reply) |
| `internal/settle/` | Pure settlement math (balances + transfers) |
| `docs/swagger/` | OpenAPI 3.1 split spec + merged `generate/openapi.yaml` + Postman |
| `docs/deliverable3.md` | Final deliverable plan |

## Typical flow

1. Create **people** once (global)
2. Create a **group**
3. **Add members** from existing people
4. Add **charges** (payer + participants must be members)
5. **Settle** — balances and who-pays-whom
