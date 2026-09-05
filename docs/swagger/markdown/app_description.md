# Split It API

REST API for splitting shared expenses among roommates, trips, and group bills.

## Workflow

1. **Create people** (global) — Alex, Sam, Jordan
2. **Create a group** — e.g. "July beach trip"
3. **Add members** — attach existing people to the group
4. **Record charges** — who paid, who shares, amount (payer/participants must be members)
5. **Settle** — net balances and simplified transfer list

## Response envelope

All endpoints use [gobeetle/reply](https://github.com/gobeetle/reply) wrappers:

- **Success** — `application/json` with `code`, `status`, and `data`
- **Error** — `application/problem+json` with `code`, `status`, `error`, and `message`

## Demo

Import the Postman collection from `docs/swagger/postman/Shared-Bill-Splitter.postman_collection.json`,
or open the minimal UI at **http://localhost:55555/app/**.
