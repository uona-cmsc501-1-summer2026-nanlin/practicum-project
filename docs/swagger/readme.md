# Swagger / OpenAPI

Multi-file OpenAPI spec for the Split It API (pattern similar to elevation-package-validation-service).

## Layout

```
docs/swagger/
├── oas.yaml                 # Root spec (paths + $refs)
├── paths/                   # One file per route group
├── components/
│   ├── schemas/
│   ├── parameters/
│   └── responses/
├── markdown/
│   └── app_description.md   # Injected into info.description at runtime
├── generate/
│   └── openapi.yaml         # Merged spec (served by the app)
└── postman/
    └── Shared-Bill-Splitter.postman_collection.json
```

## Regenerate merged spec

```bash
make oas
```

Uses **Redocly CLI** via `npx` (no global install needed). Requires Node.js.

## View when running

1. `go run .`
2. Open **http://localhost:55555/swagger**
3. Live spec JSON: **http://localhost:55555/swagger/specification**

The running service injects `app_description.md` and sets the server URL from `SWAGGER_BASE_URL` (default `http://localhost:55555`).
