# Migration and Schema Changes (suggested-new-db)

## What changed
- Introduced `golang-migrate`-based migrations under `API/pkg/db/migrations/sqlite/000001_create_base.{up,down}.sql`.
- Added migration runner `API/pkg/db/sqlite/sqlite.go` and switched `InitDB` to call it (see `// NEW:` comments in `API/models/database_model.go`).
- Expanded schema to cover missing social features: extended user profile/privacy fields, follower relationships, post visibility and allow-lists, comment media, groups/members/invites/requests, group events/RSVPs, 1:1 and group chat (with images), and richer notifications.
- Kept category seeding; it now runs after migrations.
- Added dependency `github.com/golang-migrate/migrate/v4 v4.17.0` to `API/go.mod`.

## How to apply migrations locally
1) Ensure network is available, then install deps:
   - `cd API`
   - `go mod tidy` (fetches golang-migrate; failed earlier due to no network)
2) Run the API normally (`go run cmd/main.go`). Migrations auto-apply on startup via `InitDB` -> `sqlite.Migrate`.
3) Verify DB state (optional):
   - `sqlite3 database/forum.db ".schema user"` (or inspect other tables).

## Testing
- Automated tests were **not run** (network restricted; `go mod tidy` could not fetch migrate dependency). After fetching deps, rerun:
  - `go test ./...`
- Manual test recommended after deps are in place:
  - Start API, ensure it boots without migration errors.
  - Confirm schema creation by hitting a simple endpoint (e.g., categories) and inspecting `database/forum.db`.

## Push the branch
- From repo root:
  - `git push -u origin suggested-new-db`

