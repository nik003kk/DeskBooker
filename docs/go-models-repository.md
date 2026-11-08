# DB-020 – Go podatkovni modeli, DTO in repository

Nove datoteke:
- `backend/internal/models/models.go`: strukture za `User`, `Resource`,
  `ResourceFeature`, `Reservation` in tipizirane vloge, tipe ter statuse.
- `backend/internal/models/dto.go`: vhodni DTO-ji za ustvarjanje in varni
  izhodni DTO-ji; `User.Public()` nikoli ne vrne `password_hash`,
  `Reservation.Public()` ne vrača check-in kode.
- `backend/internal/repository/repository.go`: podatkovni sloj, ki uporablja
  obstoječi PostgreSQL `pgxpool.Pool` prek ozkega vmesnika `DB`.

Vse poizvedbe uporabljajo `$1`, `$2` itd. namesto sestavljanja SQL
z uporabniškimi podatki. Metode sprejmejo `context.Context` in vrnejo
napake iz PostgreSQL, vključno s `pgx.ErrNoRows` za manjkajoče zapise.
Klicatelj upravlja povezavo in življenjski cikel bazena.

## Podprte operacije

| Entiteta | Operacije |
| --- | --- |
| Uporabnik | CreateUser, GetUser |
| Vir | CreateResource, GetResource, ListActiveResources |
| Lastnosti | AddFeature, ListFeatures |
| Rezervacija | CreateReservation, GetReservation, ListUserReservations, ConfirmReservation |

Konstruktorski primer z obstoječim poolom:

```go
pool, err := database.Open(ctx)
if err != nil { return err }
defer pool.Close()
store := repository.New(pool)
resource, err := store.GetResource(ctx, resourceID)
```

Pred uporabo prej navedeni primer potrebuje ustrezne uvoze, lokalni
`ctx` in `resourceID`. `CreateUserInput.PasswordHash` je namenoma
obvezna vrednost, ki jo mora predhodno ustvariti varen avtentikacijski
sloj; repository **ne izvaja zgoščevanja gesel**. Hash se nikoli ne
izpostavi v javnem `UserDTO`.

`ConfirmReservation` potrdi samo `PENDING` rezervacijo in nastavi
`confirmed_at`. Konflikt časa/istega vira zavrne exclusion constraint
iz DB-019. Stanje `CANCELLED`/`COMPLETED` in pravila prijave bodo
nadaljevala prihodnja API opravila.

## Preverjanje

```bash
cd backend
go test ./...
go vet ./...
go build ./...
```

Ti testi uporabljajo lažne DB odgovore; realni PostgreSQL integracijski
preizkusi predhodnih migracij ostanejo v `backend/scripts/`.
Za bazo mora biti migracija na verziji **7** (`dirty: false`).
DB-020 ne dodaja novih migracij in ne spreminja API poti.

## Integracijski preizkus z PostgreSQL

Po zagonu lokalnega Docker PostgreSQL in migracij do verzije 7:

```bash
docker compose up -d postgres
cd backend
go run ./cmd/repository-smoke
```

Pričakovan izpis:
`DB-020 repository smoke tests passed: user, resource, features, reservation, confirmation, rollback`.

Preizkus uporabi **pravi repository** nad PostgreSQL transakcijo,
ustvari in prebere uporabnika/vir/lastnost/rezervacijo, potrdi rezervacijo
in preveri neuspeh pri ponovni potrditvi. Transakcija se vedno povrne
z `ROLLBACK`; testni podatki se ne ohranijo.
