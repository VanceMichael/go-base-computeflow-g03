# ComputeFlow

ComputeFlow is a Go backend for operating a regional compute bank and compute marketplace. It models demand waves, three-stage resource attestation, compute jobs and pools, risk checks, emergency response, inter-agency notifications, capacity snapshots and audit history.

The service uses a real SQLite database in WAL mode. Migrations run at startup and all state transitions are performed through service boundaries that carry request context and transaction ownership. The HTTP server exposes health/readiness checks and a small operational API under `/api`.

## Run

```bash
GOTOOLCHAIN=local go run ./cmd/server
```

Environment variables are documented in `.env.example`. The first startup creates the database and applies migrations. The seeded demo user is `coordinator@computeflow.local` with role `coordinator`; tests use an isolated temporary database.

## Verify

```bash
GOTOOLCHAIN=local go test ./... -count=1
GOTOOLCHAIN=local go test -race ./... -count=1
GOTOOLCHAIN=local go vet ./...
GOTOOLCHAIN=local go build ./...
```

The default Docker entrypoint is `./server`, and readiness checks the database connection and migration version.
