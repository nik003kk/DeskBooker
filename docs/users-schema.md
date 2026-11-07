# DB-015 – tabela users

Prva dejanska poslovna tabela DeskBookerja je `public.users`. Ne zamenjuj je s
predhodno razvojno tabelo `demo_seed.users`, ki ostane nespremenjena.

## Stolpci

| Stolpec | PostgreSQL | Pomen |
| --- | --- | --- |
| `id` | `UUID PRIMARY KEY DEFAULT gen_random_uuid()` | Samodejni enolični identifikator |
| `email` | `TEXT NOT NULL` | E-pošta, unikatna ne glede na velikost črk |
| `password_hash` | `TEXT NOT NULL` | Samo zgoščena vrednost gesla, nikoli surovo geslo |
| `first_name` | `TEXT NOT NULL` | Ime |
| `last_name` | `TEXT NOT NULL` | Priimek |
| `role` | `TEXT NOT NULL DEFAULT 'employee'` | `admin`, `manager`, `employee` |
| `department` | `TEXT` | Neobvezen oddelek |
| `created_at` | `TIMESTAMPTZ NOT NULL DEFAULT NOW()` | Čas ustvaritve v časovno ozaveščenem tipu |

Dodatne omejitve preprečujejo prazne vrednosti za geselski hash in
ime/priimek, neveljavne vloge ter očitno neveljavne e-poštne naslove.
`LOWER(email)` v unikatnem indeksu preprečuje podvojene e-pošte, tudi
če se črke razlikujejo v velikosti. Ta preprost pogoj `LIKE '%@%'` ni popolna
validacija e-pošte; dodatna validacija bo sodila v kasnejšo registracijo.

**Varnost:** migracija ne ustvarja uporabnikov ali gesel. `password_hash`
mora vsebovati rezultat varnega algoritma za gesla (npr. Argon2id ali
bcrypt), nikoli golega besedila. V tem koraku še ne izvajamo prijave.

## Lokalni preizkus

```bash
# v korenski mapi DeskBooker
docker compose up -d postgres
cd backend
go test ./...
go vet ./...
go run ./cmd/migrate up
go run ./cmd/migrate status
```

Po migraciji mora biti verzija **3, dirty: false**. Preveri strukturo:

```bash
docker compose exec postgres psql -U deskbooker -d deskbooker -c '\\d public.users'
```

Za **dejanski integracijski preizkus** omejitev iz korena projekta:

```bash
docker compose exec -T postgres psql -X -v ON_ERROR_STOP=1 -U deskbooker -d deskbooker < backend/scripts/users-smoke.sql
```

Ta preizkus poskusi uspešen vnos, samodejni UUID ter zavrnitev podvojene
e-pošte (različna velikost črk), neveljavne vloge, praznega hasha in
neveljavne e-pošte. Celoten preizkus se izvede v transakciji z `ROLLBACK`,
zato **ne pušča testnih vrstic** v `public.users`.

## Povrnitev migracije

```bash
cd backend
go run ./cmd/migrate down
go run ./cmd/migrate status
go run ./cmd/migrate up
```

`down` povrne samo **eno** migracijo (verzija 3 → 2), a **izbriše vse
uporabnike v `public.users`**. Tega ne izvajaj v bazi z resničnimi podatki.
Migracija brez `CASCADE` preprečuje nenameren izbris bodočih odvisnih tabel.
