# DeskBooker – Go backend

REST API za DeskBooker (Go + Gin + PostgreSQL/pgx).

## Predpogoji

- Go 1.25 ali novejši
- Docker Desktop z zagnanim PostgreSQL (glej `docs/database-setup.md`)

## Lokalni zagon (DB-008)

V **korenu** repozitorija enkrat ustvari lokalno konfiguracijo:

```bash
cp .env.example .env
```

V `.env` nastavi lastno geslo `POSTGRES_PASSWORD`; `POSTGRES_PORT=5433` ustreza našemu lokalnemu okolju, kjer so vrata 5432 že zasedena.

Zaženi PostgreSQL:

```bash
docker compose up -d postgres
docker compose ps
```

Zaženi backend (brez `source .env`):

```bash
cd backend
go run ./cmd/server
```

Strežnik sam poišče **korensko** datoteko `.env`, ne glede na to, ali ga poženeš iz korena ali iz mape `backend`. Če datoteke ni, pričakuje okoljske spremenljivke, nastavljene v procesu (npr. v produkciji). Že nastavljene okoljske spremenljivke imajo prednost pred datoteko `.env`.

Podprte nastavitve:
- `PORT` (HTTP, privzeto 8080)
- `POSTGRES_HOST` (privzeto localhost), `POSTGRES_PORT` (iz `.env`), `POSTGRES_USER`, `POSTGRES_DB`, `POSTGRES_PASSWORD`, `POSTGRES_SSLMODE`
- `DATABASE_URL`, ki ima prednost pred `POSTGRES_*`

Datoteka uporablja preprost zapis `KEY=VALUE` (lahko tudi enojni/dvojni narekovaji, brez shell interpolacije). Če jo urejaš, ne vstavljaj ukazov ali skrivnosti v Git. `backend/.env.example` je informativni primer; za lokalni zagon je edini vir datoteka `.env` v korenu. Gesel in ključev nikoli ne nastavljaj v `frontend/VITE_*`, ker so te vrednosti javne v brskalniku.

## Preverjanje

```bash
curl -i http://localhost:8080/api/health
curl -i http://localhost:8080/api/health/db
```

Ob uspešni povezavi drugi endpoint vrne HTTP 200:

```json
{"status":"ok","database":"connected"}
```

Če povezava ob zagonu ni na voljo, backend vrne napako in se ne zažene.

## Testi

```bash
cd backend
go test ./...
go vet ./...
```

Testi vključujejo preverjanje branja lokalne konfiguracije, njenega prednostnega vrstnega reda in obravnave napak. Domenske migracije in endpointi bodo dodani v nadaljevanju.
