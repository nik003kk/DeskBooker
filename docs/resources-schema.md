# DB-016 – tabela resources

Migracija `000004_create_resources` ustvari prvo pravo tabelo virov:
`public.resources`. Ni povezana z izoliranimi razvojnimi tabelami v
`demo_seed`.

## Stolpci

| Stolpec | PostgreSQL | Namen |
| --- | --- | --- |
| `id` | `UUID PRIMARY KEY DEFAULT gen_random_uuid()` | Enoličen ID vira |
| `name` | `TEXT NOT NULL` | Ime vira, unikatno tudi ne glede na velikost črk |
| `type` | `TEXT NOT NULL` | Samo `DESK`, `MEETING_ROOM`, `PARKING_SPOT` |
| `floor` | `INTEGER`, lahko NULL | Nadstropje; NULL npr. za zunanje parkirišče |
| `capacity` | `INTEGER NOT NULL DEFAULT 1` | Pozitivna zmogljivost |
| `is_active` | `BOOLEAN NOT NULL DEFAULT TRUE` | Ali je vir aktiven |
| `created_at` | `TIMESTAMPTZ NOT NULL DEFAULT NOW()` | Čas nastanka |

Podatkovne omejitve preprečujejo prazna imena, neznane tipe in
`capacity <= 0`. Unikatni indeks na `LOWER(name)` prepreči
podvojena imena ne glede na velike/male črke. Ime je enolično
*med vsemi tipi virov*. Dodan je tudi delni indeks na `(type, floor)`
za aktivne vire; to še ni rezervacijska logika.

`capacity` ima privzeto vrednost 1. Za sejne sobe lahko nastaviš večjo.
`floor` nima omejitve predznaka: negativne vrednosti lahko pomenijo
kletne etaže. Oprema za zdaj **ni tip v tej tabeli**, ker issue DB-016
navaja le tri tipe; razvojna tabela `demo_seed.equipment` ostaja ločena.

## Lokalni preizkus

V korenu repozitorija:

```bash
git fetch origin
git switch --track origin/feature/db-016-resources-table
docker compose up -d postgres
cd backend
go test ./...
go vet ./...
go run ./cmd/migrate up
go run ./cmd/migrate status
cd ..
```

Pričakovan rezultat migracije: **verzija 4, dirty: false**.
Struktura:

```bash
docker compose exec postgres psql -U deskbooker -d deskbooker -c '\d public.resources'
```

Preizkus dejanskih SQL omejitev:

```bash
docker compose exec -T postgres psql -X -v ON_ERROR_STOP=1 -U deskbooker -d deskbooker < backend/scripts/resources-smoke.sql
```

Izpisati mora `NOTICE: DB-016 smoke tests passed`, nato `ROLLBACK`.
Test v transakciji vstavi tri tipe virov, preveri privzete vrednosti
in negativne primere. Testni podatki se ne ohranijo.

## Varnost povrnitve

`go run ./cmd/migrate down` povrne **eno** migracijo, zato pri verziji 4
odstrani `public.resources` in vse njegove podatke, ne pa `public.users`
ali `demo_seed`. Tega **ne zaganjaj na bazi z dejanskimi viri**.
Povrnitev ne uporablja `CASCADE`, da ne izbriše prihodnjih odvisnih tabel.
