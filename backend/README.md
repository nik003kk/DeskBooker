# DeskBooker – Go backend

Go 1.25+, Gin, PostgreSQL 17 in pgx. Osrednja navodila za konfiguracijo,
prvi račun, Docker ter odpravljanje težav so v [glavnem README-ju](../README.md).

## Lokalni zagon

Najprej pripravi korenski `.env` in zaženi `docker compose up -d postgres` iz
korena, kot opisuje glavni README. Naslednje ukaze izvajaj iz mape `backend`:

```bash
go mod download
go run ./cmd/migrate up
go run ./cmd/migrate status
go run ./cmd/server
```

API posluša na http://localhost:8080. `GET /api/health/db` preveri tudi bazo.
Konfiguracija se bere iz korenskega `.env`; procesne spremenljivke imajo prednost.
JWT ključ mora imeti najmanj 32 bajtov. Registracija zahteva dovoljeno e-poštno domeno.

Lokalni API sam ne izvaja migracij. Trenutna verzija je 13 (`dirty: false`).
Pri celotnem Docker zagonu jih izvede storitev `migrate` pred zagonom API-ja.
Ukaz `down` povrne eno migracijo in lahko izbriše podatke.

## Preverjanje

Iz `backend`:

```bash
go test ./...
go vet ./...
go build ./...
```

Brez `CHECK_IN_TEST_DATABASE_URL` se integracijski test s PostgreSQL preskoči.
CI ga izvede s `-race` na prazni, ločeni bazi. Ne usmerjaj ga na razvojno ali
produkcijsko bazo: sam izvaja migracije in podatkovne spremembe.

Izbirni `APP_ENV=development go run ./cmd/seed` polni samo shemo `demo_seed`;
ne ustvari prijavnih računov ali dejanskih virov MVP-ja.

## Dokumentacija API-ja in podatkov

- [Migracije](../docs/migrations.md) in [razvojni podatki](../docs/seed-data.md)
- [Registracija](../docs/registration-api.md), [prijava](../docs/login-api.md) in [JWT](../docs/jwt-authentication.md)
- [Modeli in repository](../docs/go-models-repository.md)
- [Rezervacijske omejitve](../docs/reservation-constraints.md)
- [Check-in](../docs/check-in.md) in [no-show sproščanje](../docs/no-show-release.md)
- [Analitika](../docs/analytics.md) in [administracija uporabnikov](../docs/user-administration.md)
- [Docker okolje](../docs/docker.md)
