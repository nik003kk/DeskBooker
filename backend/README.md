# DeskBooker – Go backend

REST API za DeskBooker (Go + Gin + PostgreSQL/pgx).

## Predpogoji

- Go 1.25 ali novejši (zahteva pgx v5.12.0)
- Docker Desktop z zagnanim PostgreSQL (glej `docs/database-setup.md`)

## Lokalni zagon

Najprej iz **korenske mape** repozitorija zaženi bazo:

```bash
docker compose up -d postgres
docker compose ps
```

Preveri, da root `.env` vsebuje `POSTGRES_PASSWORD` in `POSTGRES_PORT=5433` (ali druga prosta vrata).
Backend neposredno ne bere datoteke `.env`; spremeljivke je treba izvoziti v proces.

Iz korenske mape lahko podatke iz root `.env` preneseš v okolje in zaženeš backend:

```bash
set -a
source .env
set +a
cd backend
go mod tidy
go run ./cmd/server
```

Privzeto se poveže na `localhost:5433`, kadar je `POSTGRES_PORT=5433` v root `.env`. Geslo je vrednost `POSTGRES_PASSWORD`. Izogibaj se zapisovanju gesla v ukazno zgodovino ali repozitorij.

Za druga okolja lahko namesto `POSTGRES_*` nastaviš `DATABASE_URL`; ta ima prednost. Za produkcijo nastavi TLS (npr. `sslmode=require`) in uporabi varno shranjevanje skrivnosti.

## Preverjanje

```bash
curl -i http://localhost:8080/api/health
curl -i http://localhost:8080/api/health/db
```

Prvi endpoint preverja HTTP strežnik; drugi preveri povezavo s PostgreSQL:

```json
{"status":"ok","database":"connected"}
```

Če baza ni dosegljiva že ob zagonu, se backend ne zažene. Če povezava odpove kasneje, `/api/health/db` vrne HTTP 503 brez internih podatkov o povezavi.

## Testi

```bash
cd backend
go test ./...
go vet ./...
```

Testi preverijo sestavo povezovalnega niza in uspešne/neuspešne odgovore health endpointa. Za preverjanje dejanske povezave mora biti lokalna PostgreSQL baza zagnana.

## Obseg DB-007

- PostgreSQL pool z začetnim preverjanjem povezave in zapiranjem ob ustavitvi.
- Nastavitev iz okoljskih spremenljivk, brez gesel v izvorni kodi.
- HTTP endpoint `GET /api/health/db`.
- Enotni testi konfiguracije in health endpointa.

Podatkovne migracije, tabele in domenska logika so ločene naloge.
