# DeskBooker – Go backend

REST API za DeskBooker (Go + Gin + PostgreSQL/pgx).

## Predpogoji

- Go 1.25 ali novejši
- Docker Desktop z zagnanim PostgreSQL (glej `docs/database-setup.md`)

## Lokalni zagon (DB-008)

V **korenu** repozitorija enkrat ustvari lokalno konfiguracijo:

```bash
cp .env.example .env
```

V `.env` nastavi lastno geslo `POSTGRES_PASSWORD`; `POSTGRES_PORT=5433` ustreza našemu lokalnemu okolju, kjer so vrata 5432 že zasedena.

Zaženi PostgreSQL:

```bash
docker compose up -d postgres
docker compose ps
```

Zaženi backend (brez `source .env`):

```bash
cd backend
go run ./cmd/server
```

Strežnik sam poišče **korensko** datoteko `.env`, ne glede na to, ali ga poženeš iz korena ali iz mape `backend`. Če datoteke ni, pričakuje okoljske spremenljivke, nastavljene v procesu (npr. v produkciji). Že nastavljene okoljske spremenljivke imajo prednost pred datoteko `.env`.

Podprte nastavitve:
- `PORT` (HTTP, privzeto 8080)
- `POSTGRES_HOST` (privzeto localhost), `POSTGRES_PORT` (iz `.env`), `POSTGRES_USER`, `POSTGRES_DB`, `POSTGRES_PASSWORD`, `POSTGRES_SSLMODE`
- `DATABASE_URL`, ki ima prednost pred `POSTGRES_*`

Datoteka uporablja preprost zapis `KEY=VALUE` (lahko tudi enojni/dvojni narekovaji, brez shell interpolacije). Če jo urejaš, ne vstavljaj ukazov ali skrivnosti v Git. `backend/.env.example` je informativni primer; za lokalni zagon je edini vir datoteka `.env` v korenu. Gesel in ključev nikoli ne nastavljaj v `frontend/VITE_*`, ker so te vrednosti javne v brskalniku.

## Preverjanje

```bash
curl -i http://localhost:8080/api/health
curl -i http://localhost:8080/api/health/db
```

Ob uspešni povezavi drugi endpoint vrne HTTP 200:

```json
{"status":"ok","database":"connected"}
```

Če povezava ob zagonu ni na voljo, backend vrne napako in se ne zažene.

## Testi

```bash
cd backend
go test ./...
go vet ./...
```

Testi vključujejo preverjanje branja lokalne konfiguracije, njenega prednostnega vrstnega reda in obravnave napak. Domenske migracije in endpointi bodo dodani v nadaljevanju.

## CORS in REST API (DB-009)

Gin vrača JSON na `GET /api/health` in `GET /api/health/db`. Neznana pot vrne JSON napako HTTP 404, nedovoljena metoda pa JSON HTTP 405.

Brskalniški klici so dovoljeni samo iz konfiguriranih izvorov. Privzeto je dovoljen `http://localhost:5173` (React Vite); dodatne izvore nastavi v **korenskem** `.env`:

```dotenv
CORS_ALLOWED_ORIGINS=http://localhost:5173,http://127.0.0.1:5173
```

Ne uporabljaj `*`, ne dodajaj zaključne poševnice (`/`), poti ali poverilnic. Če `CORS_ALLOWED_ORIGINS` ni nastavljen, se uporablja samo `http://localhost:5173`. Nastavitev vpliva na brskalniški CORS; **ne nadomešča avtentikacije** in neposrednih API zahtevkov brez `Origin` ne blokira.

Preizkus dovoljenega izvora:

```bash
curl -i -H 'Origin: http://localhost:5173' http://localhost:8080/api/health
```

Odgovor mora vsebovati `Access-Control-Allow-Origin: http://localhost:5173`.

Preizkus preflight zahtevka:

```bash
curl -i -X OPTIONS \
  -H 'Origin: http://localhost:5173' \
  -H 'Access-Control-Request-Method: GET' \
  http://localhost:8080/api/health
```

Pričakuj HTTP 204. Za neznani izvor (`https://evil.example`) se dovoljenje CORS ne doda; preflight vrne HTTP 403.

V mapi `backend` izvedi `go test ./...` in `go vet ./...`.

## Podatkovne migracije (DB-010)

Za nadzor različic PostgreSQL sheme uporabljamo `golang-migrate`. Podrobna navodila:
[`docs/migrations.md`](../docs/migrations.md).

Iz mape `backend`:

```bash
go mod tidy
go run ./cmd/migrate status
go run ./cmd/migrate up
go run ./cmd/migrate status
go run ./cmd/migrate down
```

`up` izvede vse čakajoče verzije, `down` pa povrne **eno** verzijo.
Migracije se ne izvajajo samodejno ob zagonu API strežnika.
Prva verzija je neinvazivna baseline migracija; aplikacijskih tabel še ne ustvari.

## Razvojni testni podatki (DB-011)

Navodila: [`docs/seed-data.md`](../docs/seed-data.md).

Po zagonu PostgreSQL in migracij:

```bash
go run ./cmd/migrate up
APP_ENV=development go run ./cmd/seed
```

Ponovljen seed posodobi iste stabilne ID-je brez podvajanja podatkov.
Ukaz deluje **samo** ob `APP_ENV=development`, uporablja obstoječo korensko
`.env` in zapisuje podatke v ločeno shemo `demo_seed`.
To še niso končne poslovne tabele ali dejanski uporabniški računi.

## Uporabniki – podatkovna tabela (DB-015)

Migracija `000003_create_users` doda dejansko tabelo `public.users` z
UUID, e-pošto, geselskim hashom, imenom, priimkom, vlogo, oddelkom in
časom ustvaritve. Podrobnosti: [`docs/users-schema.md`](../docs/users-schema.md).

```bash
go run ./cmd/migrate up
go run ./cmd/migrate status
```

Za integracijski preizkus omejitev (iz **korena repozitorija**):

```bash
docker compose exec -T postgres psql -X -v ON_ERROR_STOP=1 -U deskbooker -d deskbooker < backend/scripts/users-smoke.sql
```

Preizkus izvede `ROLLBACK` in ne pusti testnih vrstic. V produkciji
ne zaganjaj migracije `down` brez varnostne kopije, saj izbriše uporabnike.

## Rezervacijski viri – tabela resources (DB-016)

Migracija `000004_create_resources` ustvari poslovno tabelo `public.resources`
za mize, sejne sobe in parkirna mesta. Vse podrobnosti, omejitve in primeri
so v [`docs/resources-schema.md`](../docs/resources-schema.md).

```bash
go run ./cmd/migrate up
go run ./cmd/migrate status
```

Pričakovana verzija je `4`. Lokalni integracijski test iz **korenske mape**:

```bash
docker compose exec -T postgres psql -X -v ON_ERROR_STOP=1 -U deskbooker -d deskbooker < backend/scripts/resources-smoke.sql
```

Test izvede `ROLLBACK` in ne pusti podatkov. Povrnitev migracije
`down` izbriše vse dejanske vire; ne uporabljaj je na produkcijski bazi.

## Lastnosti rezervacijskih virov (DB-017)

Migracija `000005_create_resource_features` doda tabelo `public.resource_features`
za pripenjanje več lastnosti/opreme istemu viru. Podrobnosti:
[`docs/resource-features-schema.md`](../docs/resource-features-schema.md).

```bash
go run ./cmd/migrate up
go run ./cmd/migrate status
```

Pričakovana verzija je `5`. Integracijski test za lokalni PostgreSQL
(iz korena repozitorija):

```bash
docker compose exec -T postgres psql -X -v ON_ERROR_STOP=1 -U deskbooker -d deskbooker < backend/scripts/resource-features-smoke.sql
```

Test ne pušča podatkov (`ROLLBACK`). Povrnitev migracije `down`
izbriše podatke o lastnostih in je zato destruktivna.

## Rezervacije – tabela reservations (DB-018)

Migracija `000006_create_reservations` doda `public.reservations`
s povezavama na uporabnike in vire, začetkom/koncem, statusom,
check-in kodo in časom potrditve. Več informacij:
[`docs/reservations-schema.md`](../docs/reservations-schema.md).

```bash
go run ./cmd/migrate up
go run ./cmd/migrate status
```

Pričakovana verzija je `6`. Lokalni integracijski preizkus iz korena:

```bash
docker compose exec -T postgres psql -X -v ON_ERROR_STOP=1 -U deskbooker -d deskbooker < backend/scripts/reservations-smoke.sql
```

Smoke test preveri uspešen vnos, potrditvena polja in zavrnitve neveljavnih
podatkov; na koncu izvede `ROLLBACK` in ne pušča testnih zapisov.
Migracija `down` izbriše vse rezervacije, zato je ne izvajaj na
produkcijski bazi brez ustreznega načrta in varnostne kopije.

## Pravila za rezervacije in indeksi (DB-019)

Migracija `000007_reservation_constraints` uvede preprečevanje časovno
prekrivajočih se aktivnih rezervacij istega vira na ravni PostgreSQL,
enolične check-in kode in indekse za uporabnikov pregled in urnik virov.
Podrobnosti: [`docs/reservation-constraints.md`](../docs/reservation-constraints.md).

```bash
go run ./cmd/migrate up
go run ./cmd/migrate status
```

Pričakovana verzija je `7`. SQL integracijski preizkus izvedi iz
**korena repozitorija**:

```bash
docker compose exec -T postgres psql -X -v ON_ERROR_STOP=1 -U deskbooker -d deskbooker < backend/scripts/reservation-constraints-smoke.sql
```

Testni zapisi se povrnejo z `ROLLBACK`. Pred migracijo na bazi z resničnimi
podatki preveri morebitna že obstoječa prekrivanja in podvojene check-in
kode, saj migracija ob konfliktu ne more uspeti.

## Go modeli, DTO in repository (DB-020)

`internal/models` definira uporabnike, vire, lastnosti in rezervacije z
varnimi javnimi DTO-ji. `internal/repository` uporablja PostgreSQL prek
obstoječega `pgxpool`. Podrobnosti:
[`docs/go-models-repository.md`](../docs/go-models-repository.md).

```bash
go test ./...
go vet ./...
go build ./...
go run ./cmd/repository-smoke
```

Zadnji ukaz poveže repository z lokalno bazo (zahteva migracije do
verzije **7**) in izvede vse CRUD smoke korake v transakciji, ki se
na koncu povrne z `ROLLBACK`.

## Registracija službenih uporabnikov (DB-021)

`POST /api/auth/register` sprejme ime, priimek, službeni e-naslov
in geslo. V korenski `.env` nastavi `REGISTRATION_EMAIL_DOMAIN` na
**dejansko** podjetniško domeno; brez nastavitve registracija vrne 503.
Gesla se zgoščujejo z bcrypt, podvojen e-naslov vrne HTTP 409.
API ne vrne gesla ali hasha in ne izda prijavne seje.

Primeri, zahteve za geslo in lokalni preizkus:
[`docs/registration-api.md`](../docs/registration-api.md).

## Varnost gesel (DB-022)

Centralizirani bcrypt funkciji `auth.HashPassword` in
`auth.VerifyPassword` zagotavljata soljeno zgoščevanje ter preverjanje
pravilnega/napačnega gesla. Registracija uporablja isti modul.
Repository metoda `GetUserByEmail` omogoča bodočo prijavo; dejanski
`POST /api/auth/login` je predmet DB-023.

```bash
go test ./...
go vet ./...
go build ./...
go run ./cmd/password-smoke
```

Zadnji ukaz potrebuje zagnan lokalni PostgreSQL z migracijo **7**
in preveri shranjeni hash ter gesla. Testni podatki se povrnejo
z `ROLLBACK`. Podrobnosti:
[`docs/password-security.md`](../docs/password-security.md).

## Prijava – preverjanje e-pošte in gesla (DB-023)

`POST /api/auth/login` preveri lokalno registrirani račun prek
`GetUserByEmail` in `VerifyPassword`. Na uspeh vrne HTTP 200
in varni `UserDTO`, na neznan e-naslov ali napačno geslo HTTP 401.
API ne vrača hasha in ne ustvarja seje ali žetona.

```bash
go test ./...
go vet ./...
go build ./...
go run ./cmd/server
```

Za ročni preizkus na lokalni bazi (migracije **7**) glej
[`docs/login-api.md`](../docs/login-api.md), kjer so primeri
za HTTP 200 in 401. Za javno uporabo bosta potrebna HTTPS in
omejevanje pogostosti poskusov prijave.

## JWT avtentikacija (DB-024)

Obvezen je naključno ustvarjen `JWT_SECRET` (vsaj 32 bajtov) v **korenski** `.env`:

```bash
openssl rand -hex 32
```

Izpis shrani v `.env` kot `JWT_SECRET=<izpis>`. Ne v Git in ne v `VITE_*`.
Strežnik se brez veljavnega ključa ne zažene.

`POST /api/auth/login` ob pravilnih poverilnicah vrne kratkoživeči
podpisani `access_token` (HS256, **15 minut**) in `token_type: Bearer`.
Zaščitena pot `GET /api/auth/me` zahteva
`Authorization: Bearer <access_token>` in vrne varne podatke prijavljenega
uporabnika. Manjkajoč, spremenjen ali potekel žeton vrne HTTP **401**.

Primeri in varnostne omejitve: [`docs/jwt-authentication.md`](../docs/jwt-authentication.md).
