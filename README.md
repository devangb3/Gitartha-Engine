# Gitartha Engine

Public REST API for serving Bhagavad Gita chapters and verses with English/Hindi translations.

---

## 1. Prerequisites
- Go 1.22+
- PostgreSQL 14+ running locally (default `localhost:5432`)
- `golang-migrate` CLI (for database migrations)

## 2. Repository Setup
```bash
git clone git@github.com:devangb3/Gitartha-Engine.git
cd Gitartha-Engine
go mod tidy
```

## 3. Environment Configuration
Create a `.env` file in the project root:
```bash
cat <<'ENV' > .env
DATABASE_URL=postgres://<user>:<password>@localhost:5432/gitartha?sslmode=disable
PORT=8080
ENV=development
LOG_LEVEL=info
ENV
```
- The database name (`gitartha` in the example) is defined inside the `DATABASE_URL`.
- Ensure the referenced database already exists in PostgreSQL (`createdb gitartha`).

## 4. Database Migrations
Apply the initial schema:
```bash
make migrate-up
```
This runs all SQL files inside `migrations/`. Use `make migrate-down` to roll back.

## 5. Data Ingestion
 Run the Go ingestion CLI to load verses:
   ```bash
   go run ./cmd/ingest --csv bg.csv
   ```
   This reads `bg.csv`, upserts chapters/verses, and updates `verse_count` totals.


## 6. Running the API Server
```bash
make run
```
Output example:
```
[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
[GIN-debug] GET    /healthz                  --> ... (*handler*).health
[GIN-debug] GET    /api/v1/chapters          --> ...
...
```
Visit `http://localhost:8080/healthz` to confirm the service is healthy.

## 7. API Overview
- `GET /api/v1/chapters` — List all chapters.
- `GET /api/v1/chapters/{chapter}` — Chapter metadata + verses.
- `GET /api/v1/chapters/{chapter}/verses/{verse}` — Specific verse with translations.
- `GET /api/v1/search?query=term&lang=en|hi` — Keyword search (English/Hindi).
- `GET /api/v1/random` — Random verse.

Use tools like `curl`, Postman, or `httpie` to exercise the endpoints:
```bash
curl http://localhost:8080/api/v1/chapters/1/verses/1
```

## 8. Testing
Run unit tests (includes database layer tests with `sqlmock`):
```bash
make test
```
Or directly:
```bash
go test ./...
```

## 9. Project Layout (high level)
```
cmd/api          # HTTP server entrypoint
cmd/ingest       # Data ingestion CLI (stub)
internal/config  # Configuration loading (Viper)
internal/db      # PostgreSQL connection helper
internal/data    # DB store for chapters/verses
internal/http    # Gin router & handlers
migrations       # Database schema migrations
```

## 10. Next Steps
- Add Swagger/OpenAPI documentation.
- Containerize (Docker Compose for API + Postgres + Swagger UI).
- Implement integration tests hitting a test Postgres instance.

---

Questions or issues? Open an issue in the GitHub repository or add to the docs.
