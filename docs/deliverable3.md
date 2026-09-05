# Practicum Project — Final Deliverable (plan)

**Student:** Nanlin  
**Course:** CMSC501-1 Structure of Programming Language — Summer 2026  
**Repository:** https://github.com/uona-cmsc501-1-summer2026-nanlin/practicum-project  
**Language:** Go (Golang) + Fiber v2 + GORM + SQLite  
**Builds on:** [Deliverable #1](deliverable1.md) (design), [Deliverable #2](deliverable2.md) (MVP API)

---

## 1. Goal

Ship a **simple minimal UI** to manage **people**, **groups**, and **charges**, and refactor the data model so **people are global** (not duplicated per group).

**Current limitation (D2):** each `Person` row has a required `group_id`. The same human in two groups must be created twice. This is acceptable for the MVP but is marked for update here.

**Target:** one person record reused across groups; membership is a separate link; charges still belong to a group and reference global user IDs.

---

## 2. Phase 1 — Backend: separate people from groups

### Schema

- [x] Add `users` table (`id`, `name`, `email`) — global identity, no `group_id`
- [x] Add `group_members` join table (`group_id`, `user_id`, unique pair)
- [x] Migrate `charges` to reference `user_id` (payer + participants) instead of group-scoped person IDs
- [x] Drop old `people` table (wipe SQLite DB on upgrade; AutoMigrate creates new tables)
- [x] Update [deliverable2-doc.md](deliverable2-doc.md) (schema and API sections)

### API — global people

- [x] `POST /api/v1/people` — create person
- [x] `GET /api/v1/people` — list all people
- [x] `GET /api/v1/people/:id` — get one
- [x] `PUT /api/v1/people/:id` — update name/email
- [x] `DELETE /api/v1/people/:id` — delete (block if still in any group or on charges)

### API — group membership

- [x] `GET /api/v1/groups/:groupId/members` — list members in group
- [x] `POST /api/v1/groups/:groupId/members` — add existing person (`{ "userId": 1 }`)
- [x] `DELETE /api/v1/groups/:groupId/members/:userId` — remove from group

### API — charges and settle

- [x] Update charge validation: payer + participants must be **members** of that group
- [x] Update settle to resolve names from global `users`
- [x] Update Swagger + Postman collection
- [x] Update tests (`handlers_test.go`, `settle_test.go`)

---

## 3. Phase 2 — Minimal UI

### Setup

- [x] Add `web/` folder — plain HTML + CSS + vanilla JS
- [x] Serve static files from Fiber at `/app`
- [x] Small `api.js` helper: `fetch` wrapper to `/api/v1`

### People screen

- [x] List all people (name, email)
- [x] Form: add person
- [x] Delete person (with confirm)

### Groups screen

- [x] List groups
- [x] Form: create group (name, currency)
- [x] Click group → group detail view
- [x] Delete group

### Group detail screen

- [x] **Members** — show who is in the group; dropdown to add existing person; remove member
- [x] **Charges** — list charges; form: description, amount, payer (dropdown), participants (checkboxes), split rule
- [x] **Settle** — button calls `POST .../settle`; show balances + who-pays-whom table

### Polish (still minimal)

- [x] Basic error messages from API (`400` / `404`)
- [x] Empty states (“No people yet”, “No charges”)
- [x] Simple nav: **People** | **Groups**

---

## 4. Phase 3 — Optional later (not required for final MVP)

- [ ] Edit person / edit group name
- [ ] Edit or delete individual charges from UI
- [ ] Search/filter people when adding to group
- [ ] Dark mode / mobile layout

---

## 5. Screen flow

```
People          Groups              Group detail
────────        ──────              ────────────
[+ Add]         [+ New group]       Members: Alex, Sam … [+ Add]
Alex            > July trip         Charges: Dinner $90 … [+ Add]
Sam             > Apartment         [Settle] → balances + transfers
```

---

## 6. Suggested order of work

1. **Backend refactor (Phase 1)** — UI depends on global people + membership endpoints
2. **People page first** — seed data without tying to a group
3. **Groups + members** — wire “add existing person to group”
4. **Charges + settle last** — needs members in place

---

## 7. Target schema (after Phase 1)

```mermaid
erDiagram
    users ||--o{ group_members : "joins"
    groups ||--o{ group_members : "has"
    groups ||--o{ charges : "has"
    users ||--o{ charges : "paid_by"

    users {
        integer id PK
        text name "NOT NULL"
        text email
    }

    group_members {
        integer group_id FK
        integer user_id FK
    }

    groups {
        integer id PK
        text name "NOT NULL"
        text currency "NOT NULL"
    }

    charges {
        integer id PK
        integer group_id FK
        text description
        real amount
        integer paid_by_user_id FK
        text participant_ids "JSON array of user IDs"
    }
```

Settle logic stays largely the same: net balances and transfers are still computed per group; only identity resolution changes.

---

## Personal notes — points to talk about / put on PPT slides

**Slide — Why refactor people**
- D2: person tied to one group → duplicate entries for the same human
- Final: global `users` + `group_members` join table

**Slide — UI scope**
- Minimal: three areas (People, Groups, Group detail)
- No auth, no payments — same REST API as Postman demo

**Slide — Demo path**
- Create people once → create group → add members → add charge → settle in browser

**Slide — Closing**
- API-first MVP (D2) + thin UI (final) + cleaner data model for real-world reuse
