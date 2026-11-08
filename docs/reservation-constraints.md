# DB-019 – rezervacijske relacije, indeksi in omejitve

Migracija `000007_reservation_constraints` razširi `public.reservations`
iz DB-018. Ohranita se tuja ključa za `user_id` in `resource_id`,
pravila statusov in pravilnega intervala iz DB-018.

## Dodatne omejitve

- **Brez dvojne rezervacije istega vira:** PostgreSQL `EXCLUDE USING gist`
  kombinira `resource_id WITH =` in prekrivanje
  `tstzrange(starts_at, ends_at, '[)') WITH &&`.
  Omejitev velja samo za `PENDING` in `CONFIRMED` rezervacije.
- **Intervali [start, end):** rezervacija od 9:00 do 10:00 in naslednja
  od 10:00 do 11:00 sta dovoljeni (brez prekrivanja).
- **Statusi `CANCELLED` in `COMPLETED`:** ne blokirajo novih rezervacij
  istega intervala; sprememba iz `CANCELLED` v `CONFIRMED` ponovno
  sproži preverjanje konfliktov.
- **Enoličnost check-in kode:** unikatni delni indeks na
  `LOWER(check_in_code)` prepreči podvojitve ne glede na velikost črk;
  več rezervacij brez kode (`NULL`) ostane dovoljenih.
- **Indeksi za poizvedbe:** `(user_id, starts_at DESC)` za uporabnikovo
  zgodovino in `(resource_id, starts_at)` za pregled urnika vira.

PostgreSQL za GiST primerjavo UUID uporabi razširitev `btree_gist`
(`CREATE EXTENSION IF NOT EXISTS btree_gist`). To je standardna
razširitev podprta v PostgreSQL 17, ki ne zahteva posebnega API-ja.
Migracija `down` je namenoma ne odstrani, ker jo utegnejo uporabljati
tudi drugi objekti.

**Pozor:** migracija `up` ne bo uspela, če so v tabeli že obstoječe
nasprotujoče si aktivne rezervacije ali podvojene check-in kode.
Pred izvajanjem v okolju z resničnimi podatki je treba pregledati in
razrešiti konflikte. Ne uporabljaj `migrate force` brez analize.

## Lokalni preizkus

Iz korenske mape:

```bash
git fetch origin
git switch --track origin/feature/db-019-reservation-constraints
docker compose up -d postgres
cd backend
go test ./...
go vet ./...
go run ./cmd/migrate up
go run ./cmd/migrate status
cd ..
```

Pričakovana verzija migracije: **7, dirty: false**.

Preveri relacije in indekse:

```bash
docker compose exec postgres psql -U deskbooker -d deskbooker -c '\d public.reservations'
```

Integracijski PostgreSQL smoke test (uspešni in neuspešni primeri):

```bash
docker compose exec -T postgres psql -X -v ON_ERROR_STOP=1 -U deskbooker -d deskbooker < backend/scripts/reservation-constraints-smoke.sql
```

Pričakovan izpis: `NOTICE: DB-019 smoke tests passed` in `ROLLBACK`.
SQL preizkus uporabi transakcijo, zato ne pušča uporabnikov,
virov ali rezervacij v bazi.

## Povrnitev

`go run ./cmd/migrate down` (samo ena migracija) pri verziji 7
odstrani nove indekse in omejitev prekrivanja ter vrne verzijo 6,
**ne izbriše pa rezervacij**. Po povrnitvi baza ne preprečuje
dvojnih aktivnih rezervacij; tega ne izvajaj brez pregleda posledic.
