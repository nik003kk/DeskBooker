# DB-018 – tabela reservations

Migracija `000006_create_reservations` doda `public.reservations`.
Vsaka rezervacija povezuje enega uporabnika iz `public.users` z enim
virom iz `public.resources` in ima časovni interval, stanje in
neobvezne podatke za potrditev oziroma prijavo na lokaciji.

## Stolpci

| Stolpec | Tip | Pomen |
| --- | --- | --- |
| `id` | UUID PRIMARY KEY, privzeto `gen_random_uuid()` | Enolična rezervacija |
| `user_id` | UUID NOT NULL, FK na `public.users(id)` | Uporabnik |
| `resource_id` | UUID NOT NULL, FK na `public.resources(id)` | Vir |
| `starts_at` | TIMESTAMPTZ NOT NULL | Začetek |
| `ends_at` | TIMESTAMPTZ NOT NULL | Konec |
| `status` | TEXT NOT NULL, privzeto `PENDING` | Stanje |
| `check_in_code` | TEXT, NULL dovoljen | Koda za check-in |
| `confirmed_at` | TIMESTAMPTZ, NULL dovoljen | Čas potrditve |
| `created_at` | TIMESTAMPTZ NOT NULL, privzeto `NOW()` | Ustvarjeno |

Dovoljena stanja: `PENDING`, `CONFIRMED`, `CANCELLED`, `COMPLETED`.
Podatkovna omejitev zagotavlja `ends_at > starts_at`. Prazna
check-in koda (če je navedena) ni veljavna. Tuja ključa preprečita
rezervacije za neobstoječe uporabnike ali vire; brisanje povezanega
uporabnika/viri je blokirano, dokler obstajajo rezervacije.

**Pojasnilo obsega:** ta migracija namenoma še ne uvaja preprečevanja
časovnega prekrivanja rezervacij, unikatnih check-in kod ali
dodatnih indeksov. To je predmet naslednje naloge DB-019
(relacije, indeksi in omejitve). Pravila življenjskega cikla statusov
bodo obravnavana v API-ju. `confirmed_at` je za zdaj neobvezen in
migracija sama ne nastavi časa ob spremembi statusa.

## Lokalni preizkus

```bash
# v korenu projekta DeskBooker
git fetch origin
git switch --track origin/feature/db-018-reservations-table
docker compose up -d postgres
cd backend
go test ./...
go vet ./...
go run ./cmd/migrate up
go run ./cmd/migrate status
cd ..
```

Pričakovana verzija: **6, dirty: false**. Preveri strukturo:

```bash
docker compose exec postgres psql -U deskbooker -d deskbooker -c '\d public.reservations'
```

Integracijski SQL test (pozitivni in negativni primeri):

```bash
docker compose exec -T postgres psql -X -v ON_ERROR_STOP=1 -U deskbooker -d deskbooker < backend/scripts/reservations-smoke.sql
```

Pričakovan je `NOTICE: DB-018 smoke tests passed` in zadnja vrstica
`ROLLBACK`. Testni uporabnik, vir in rezervacija se ne ohranijo.

## Povrnitev

`go run ./cmd/migrate down` pri verziji 6 odstrani `public.reservations`
in **vse rezervacije**; uporabniki in viri se ohranijo. Ne uporabljaj
na bazi z resničnimi rezervacijami.
