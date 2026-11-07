# DeskBooker: podatkovne migracije (DB-010)

Za verzionirane migracije uporabljamo **golang-migrate/migrate v4** s PostgreSQL gonilnikom.
SQL datoteke se hranijo v `backend/migrations/`, po parih `NNNNNN_ime.up.sql` in `NNNNNN_ime.down.sql`.

## Prva namestitev in izvedba

V korenu projekta mora obstajati lokalna datoteka `.env` (ne sodi v Git) z
`POSTGRES_PASSWORD` in `POSTGRES_PORT=5433` (če lokalni PostgreSQL teče na teh vratih).
Zagnan mora biti Docker kontejner PostgreSQL:

```bash
docker compose up -d postgres
docker compose ps
cd backend
go mod tidy
go test ./...
go run ./cmd/migrate status
go run ./cmd/migrate up
go run ./cmd/migrate status
```

Skripta sama prebere korensko `.env` prek obstoječega config paketa; ni treba uporabljati `source ../.env`.
Če imaš `DATABASE_URL` že izvozno nastavljeno v okolju, ima prednost pred `POSTGRES_*`.

`up` uporabi vse še neizvedene migracije, `status` pokaže verzijo in stanje
`dirty`. Ob ponovnem zagonu `up` brez novih migracij ne naredi sprememb.

## Povrnitev

```bash
go run ./cmd/migrate down
go run ./cmd/migrate status
go run ./cmd/migrate up
```

**Pozor:** `down` namenoma povrne **samo eno** migracijo na izvedbo.
Migracije za nazaj so lahko destruktivne, zato jih pred uporabo
v okolju z dejanskimi podatki preglej in naredi varnostno kopijo.
Povrnitve ne izvajaj samodejno pri zagonu aplikacije.

Prva migracija `000001_baseline` uporabi neškodljiv `SELECT 1`;
ne ustvarja aplikacijskih tabel. Naloga je vzpostaviti in preveriti
infrastrukturo migracij, ne prehitevati podatkovnega modela.
Orodje samodejno vzdržuje tabelo `schema_migrations`.

## Dodajanje nove migracije

Ustvari nov par z naraščajočo šestmestno številko, npr.:

- `backend/migrations/000002_create_users.up.sql`
- `backend/migrations/000002_create_users.down.sql`

`up.sql` mora dodati spremembo; `down.sql` jo mora v obratnem vrstnem
redu povrniti. Za PostgreSQL DDL uporabi transakcijsko varne stavke;
če orodje označi stanje `dirty=true`, ne uporabljaj samodejnega
`force` ali popravka verzije brez preverjanja stanja baze.

## Preverjanje SQL

Ko se izvede `up`, lahko v drugem terminalu preveriš:

```bash
docker compose exec postgres psql -U deskbooker -d deskbooker -c "SELECT version, dirty FROM schema_migrations;"
```

Po `down` (če je to edina izvedena migracija) običajno ostane
vrstica z verzijo `-1` oziroma `nil version`, odvisno od vrste
izvedbe; `status` to prikaže kot brez aktivnih migracij.

## Varnost

Migracije izvajaj ločeno od zagona HTTP strežnika. Uporabljaj omejene
uporabniške pravice v produkciji in varne skrivnosti. Napake v
povezavi z bazo se v CLI namenoma ne izpisujejo s polnim DSN, da se
geslo ne pojavi v terminalu.
