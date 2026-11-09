# DeskBooker – lokalni PostgreSQL (DB-006)

Za lokalni razvoj uporabljamo PostgreSQL v Dockerju. Ta naloga pripravi instanco in podatkovno bazo; tabele bodo ustvarjene v kasnejših Issues.

## Predpogoji

- Docker Desktop ali Docker Engine s Compose v2

## Zagon

Iz korenske mape repozitorija:

```bash
cp .env.example .env
# Nastavi lastno geslo POSTGRES_PASSWORD v .env
docker compose up -d postgres
docker compose ps
```

Povezava iz lokalnega računalnika:

- Host: `localhost`
- Port: `5432` (ali `POSTGRES_PORT`)
- Database: `deskbooker`
- User: `deskbooker`
- Password: vrednost `POSTGRES_PASSWORD` iz `.env`

## Preverjanje

```bash
docker compose exec postgres pg_isready -U deskbooker -d deskbooker
docker compose exec postgres psql -U deskbooker -d deskbooker -c "SELECT current_database(), current_user;"
```

## Ustavitev

```bash
docker compose down
```

**Pozor:** `docker compose down -v` izbriše tudi podatke v razvojni bazi.

Gesla ne dodajaj v Git. PostgreSQL uporablja persistentni Docker volume `deskbooker_pgdata`.

Naslednja naloga DB-007 bo povezala Go REST API s to bazo. Za produkcijo so potrebne drugačne nastavitve dostopa in varnosti.
