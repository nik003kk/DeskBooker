# DB-011 – Začetni testni podatki

Testni podatki se shranjujejo v **ločeno shemo `demo_seed`**.
To so izključno razvojne testne tabele, **ne končni podatkovni model DeskBookerja**.
Prava poslovna shema in API endpointi bodo urejeni v naslednjih nalogah.
Trenutni API zato še ne prikazuje podatkov iz `demo_seed`.

## Vsebina testnih podatkov

| Vrsta | Primeri | Število |
| --- | --- | ---: |
| Uporabniki | Ana Novak (admin), Marko Kranjc (manager), Nina Zajc in Luka Horvat (employee) | 4 |
| Mize | D-101, D-102, D-201, D-202, D-203 | 5 |
| Sejne sobe | Triglav, Pohorje, Bled | 3 |
| Parkirna mesta | P-01, P-02, P-03, P-04 | 4 |
| Oprema | monitor, priklopna postaja, projektor, kamera | 4 |

Uporabniki so izmišljeni, z naslovi `example.test`; gesel ni.
Stabilni ID-ji in `ON CONFLICT ... DO UPDATE` omogočajo ponovljiv zagon
brez podvajanja podatkov.

## Lokalna uporaba

Iz korena repozitorija:

```bash
docker compose up -d postgres
cd backend
go test ./...
go vet ./...
go run ./cmd/migrate up
APP_ENV=development go run ./cmd/seed
```

Iz varnostnih razlogov ukaz zahteva **izrecno `APP_ENV=development`**.
Sam naloži konfiguracijo iz korenske `.env`, tako kot API in migracijski ukaz.
Uporaba je prepovedana, če je `APP_ENV` nastavljen na `production` oziroma ni nastavljen.

## Preverjanje podatkov

Iz korenske mape projekta zaženi:

```bash
docker compose exec postgres psql -U deskbooker -d deskbooker -c "SELECT 'users' AS type, COUNT(*) FROM demo_seed.users UNION ALL SELECT 'desks', COUNT(*) FROM demo_seed.desks UNION ALL SELECT 'meeting_rooms', COUNT(*) FROM demo_seed.meeting_rooms UNION ALL SELECT 'parking_spaces', COUNT(*) FROM demo_seed.parking_spaces UNION ALL SELECT 'equipment', COUNT(*) FROM demo_seed.equipment;"
```

Ponovi `APP_ENV=development go run ./cmd/seed` in preveri,
da število vrstic ostane **4 / 5 / 3 / 4 / 4**.

## Čiščenje / povrnitev

Migracija `000002_demo_seed_schema` ustvari shemo `demo_seed`.
Če je trenutno zadnja migracija, lahko za njen umik uporabiš:

```bash
go run ./cmd/migrate down
```

**Pozor:** `down` izbriše celotno `demo_seed` shemo in njene
testne podatke. Preden izvedeš `down`, preveri stanje z `status`.
Podatki v `public` ostanejo nespremenjeni.

## Varnost

Ne uporabljaj testnih e-poštnih naslovov ali imen kot pravih računov,
ne shranjuj gesel v fixture SQL in nikoli ne izvajaj seed ukaza v produkciji.
