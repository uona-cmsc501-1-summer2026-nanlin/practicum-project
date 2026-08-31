# Shared Bill Splitter API

REST API for splitting shared expenses among roommates, trips, and group bills.

## Workflow

1. **Create a group** — e.g. "July beach trip"
2. **Add people** — Alex, Sam, Jordan
3. **Record charges** — who paid, who shares, amount
4. **Settle** — net balances and simplified transfer list

## Response envelope

All endpoints use [gobeetle/reply](https://github.com/gobeetle/reply) wrappers:

- **Success** — `application/json` with `code`, `status`, and `data`
- **Error** — `application/problem+json` with `code`, `status`, `error`, and `message`

## Demo

Import the Postman collection from `docs/swagger/postman/Shared-Bill-Splitter.postman_collection.json`.
