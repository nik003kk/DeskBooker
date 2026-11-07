# DB-017 – tabela resource_features

Migracija `000005_create_resource_features` doda tabelo
`public.resource_features` za povezavo enega vira z več lastnostmi
(npr. monitor, docking_station, projector, EV_CHARGING).

## Model

| Stolpec | Tip | Pravilo |
| --- | --- | --- |
| `resource_id` | UUID NOT NULL | Tuji ključ na `public.resources(id)` |
| `feature` | TEXT NOT NULL | Neprazno ime lastnosti brez robnih presledkov |

En vir ima lahko poljubno število različnih lastnosti, ista lastnost pa
lahko pripada več virom. Kombinacija `resource_id` in `LOWER(feature)`
je **unikatna**, zato iste lastnosti za isti vir ne moremo shraniti dvakrat,
niti z drugačno velikostjo črk.

Tuji ključ preprečuje osirotele lastnosti. `ON DELETE CASCADE` ob
izbrisu *vira* samodejno odstrani njegove lastnosti, ne posega pa v
lastnosti drugih virov.

To je preprost katalog *imen lastnosti*, ne samostojna inventarna tabela
opreme. Če bo kasnejša naloga zahtevala fizične kose opreme ali skupen
katalog lastnosti, ga bova dopolnila z ločenim modelom/migracijo.

## Lokalni preizkus

Iz korena repozitorija:

```bash
git fetch origin
git switch --track origin/feature/db-017-resource-features
docker compose up -d postgres
cd backend
go test ./...
go vet ./...
go run ./cmd/migrate up
go run ./cmd/migrate status
cd ..
```

Pričakovana verzija migracij je **5, dirty: false**.

Preveri tabelo:

```bash
docker compose exec postgres psql -U deskbooker -d deskbooker -c '\d public.resource_features'
```

Izvedi integracijski SQL preizkus z uspešnimi in neuspešnimi primeri:

```bash
docker compose exec -T postgres psql -X -v ON_ERROR_STOP=1 -U deskbooker -d deskbooker < backend/scripts/resource-features-smoke.sql
```

Izpis: `NOTICE: DB-017 smoke tests passed` in na koncu `ROLLBACK`.
Preizkus preveri več lastnosti na vir, deljenje lastnosti med viri,
prepoved podvajanja in praznih vrednosti, tuji ključ ter brisanje lastnosti
ob izbrisu vira. Testni podatki ne ostanejo v bazi.

## Povrnitev

`go run ./cmd/migrate down` pri verziji 5 izbriše
`public.resource_features` **in vse njene zapise**, ne pa virov ali
uporabnikov. Na bazi z resničnimi podatki pred povrnitvijo naredi
varnostno kopijo.
