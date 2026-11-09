# DeskBooker

Sistem za rezervacijo delovnih mest, sejnih sob in parkirnih mest v hibridni pisarni.
React aplikacija uporablja Go REST API in PostgreSQL. MVP vključuje registracijo in
prijavo, upravljanje virov, iskanje razpoložljivosti, rezervacije s PIN check-inom,
samodejno sproščanje ob neprihodu, analitiko in administracijo uporabniških vlog.

## Predpogoji

| Način | Potrebuješ |
| --- | --- |
| Celotna aplikacija v Dockerju | Git, Docker Desktop z zagnanim Engine in Docker Compose v2 |
| Lokalni razvoj | Dodatno Go 1.25+ ter Node.js 22.12+ in npm |

Ukazi spodaj so za terminal macOS/Linux. Izhodišče je vedno **korenska mapa repozitorija**,
razen kjer blok izrecno začne s `cd backend` ali `cd frontend`.

```bash
git clone https://github.com/nik003kk/DeskBooker.git
cd DeskBooker
git switch development
git pull --ff-only origin development
```

Pri obstoječem checkoutu začni v njegovi korenski mapi in uporabi zadnja dva ukaza.

## Konfiguracija

Samo pri novi namestitvi kopiraj primer:

```bash
cp .env.example .env
openssl rand -hex 32
```

Obstoječega `.env` ne prepisuj. Rezultat `openssl` shrani kot `JWT_SECRET` v korenski
`.env`; ostale vrednosti prilagodi svojemu okolju:

```dotenv
POSTGRES_PASSWORD=izberi-lastno-lokalno-geslo
POSTGRES_PORT=5433
JWT_SECRET=sem-prilepi-nakljucni-kljuc-iz-ukaza-openssl
REGISTRATION_EMAIL_DOMAIN=example.com
CORS_ALLOWED_ORIGINS=http://localhost:5173
```

| Nastavitev | Pomen |
| --- | --- |
| `POSTGRES_PASSWORD` | Geslo baze. Pri obstoječem volume uporabi dosedanje geslo. |
| `POSTGRES_PORT` | Vrata baze na gostitelju; primer uporablja 5433. V Docker omrežju je vedno `postgres:5432`. |
| `JWT_SECRET` | Naključni ključ, najmanj 32 bajtov; zgornjo nadomestno vrednost zamenjaj. |
| `REGISTRATION_EMAIL_DOMAIN` | Dovoljena domena registracije brez `@`, npr. `example.com`. Uporabi domeno svojih testnih računov. |
| `CORS_ALLOWED_ORIGINS` | Dovoljeni spletni izvori, ločeni z vejicami, brez zaključne poševnice. |

Za lokalni Go proces so možni še `PORT`, `POSTGRES_HOST`, `POSTGRES_USER`,
`POSTGRES_DB`, `POSTGRES_SSLMODE` in `DATABASE_URL`. `DATABASE_URL` ima prednost pred
`POSTGRES_*`; že nastavljene procesne spremenljivke imajo prednost pred `.env`.
Compose posreduje izbrane nastavitve in sam nastavi notranji naslov baze.

`.env` je ignoriran v Gitu. Gesla in JWT ključa ne dodajaj v `VITE_*`, saj so te
vrednosti javne v brskalniku. Podrobnosti: [Docker konfiguracija](docs/docker.md).

## Zagon celotne aplikacije z Dockerjem

Ustavi morebitna lokalna `go run ./cmd/server` in `npm run dev` s Ctrl+C,
saj Docker uporablja ista vrata. Nato iz korena:

```bash
docker compose up -d --build
docker compose ps -a
```

| Storitev | Naslov / pričakovano stanje |
| --- | --- |
| Frontend | http://localhost:5173, `healthy` |
| Go API | http://localhost:8080/api/health, `healthy` |
| PostgreSQL | `localhost:5433` pri zgornji konfiguraciji, `healthy` |
| Migracije | `migrate`: `Exited (0)` po uspešnem zaključku |

Prva gradnja lahko traja nekaj minut. `health: starting` je začetno stanje.
Compose počaka na zdravo bazo, izvede migracije, nato zažene API in frontend.
Preveri povezavo in dnevnike:

```bash
curl -i http://localhost:8080/api/health/db
curl -i http://localhost:5173/api/health/db
docker compose logs --tail=50 migrate backend frontend
```

Oba klica morata vrniti HTTP 200. Drugi preveri tudi nginx posredovanje do API-ja.
Podstrani, kot je `/desks`, delujejo tudi ob neposrednem odpiranju ali osvežitvi.

```bash
# Ustavi zabojnike in ohrani podatke:
docker compose down
# Ponovno zaženi:
docker compose up -d
# Po posodobitvi kode znova zgradi slike:
git pull --ff-only origin development
docker compose up -d --build
```

Ob zagonu iz iste mape z istim Compose project imenom ostane obstoječi volume
`deskbooker_pgdata`. `docker compose down -v` ga izbriše, zato ga za običajno
ustavitev ne uporabljaj. Listmonk za trenutni MVP ni potreben.

## Lokalni razvoj: PostgreSQL v Dockerju, Go in React na gostitelju

Po pripravi korenskega `.env` iz korena zaženi samo bazo. Če celotno Docker okolje
že teče, ga najprej ustavi z `docker compose down` (brez `-v`).

```bash
docker compose up -d postgres
docker compose ps
```

Ko je baza `healthy`, v prvem terminalu iz korena:

```bash
cd backend
go mod download
go run ./cmd/migrate up
go run ./cmd/migrate status
go run ./cmd/server
```

Go ukaze izvajaj v `backend`, kjer je `go.mod`. Proces sam prebere korenski `.env`;
`source .env` ni potreben. Privzeti naslov baze je `localhost`; vrata prebere iz `.env`.
API je na http://localhost:8080.

Za nov frontend konfiguracijo pripravi **samo, če `frontend/.env.local` še ne obstaja**:

```bash
cp frontend/.env.example frontend/.env.local
```

V njej nastavi `VITE_API_BASE_URL=http://localhost:8080/api` in
`VITE_REGISTRATION_EMAIL_DOMAIN` na isto domeno kot v korenskem `.env`.
Nato v drugem terminalu iz korena:

```bash
cd frontend
npm ci
npm run dev
```

Odpri http://localhost:5173. Po spremembi frontend `.env.local` ponovno zaženi Vite.
Docker gradnja te datoteke ne uporablja; javno domeno dobi iz korenskega `.env`,
API pa kliče prek relativnega `/api`.

## Migracije in razvojni podatki

```bash
# Docker, iz korena:
docker compose run --rm migrate /app/migrate status
# Izvedi čakajoče migracije:
docker compose run --rm migrate
```

Pri lokalnem Go zagonu uporabi `go run ./cmd/migrate up` oziroma `status` iz
`backend`. Trenutna verzija je **13**, `dirty: false`. Ponovni `up` brez novih
migracij uspešno sporoči, da ni sprememb. Lokalni API sam migracij ne zaganja;
celotni Compose zagon jih izvede prek storitve `migrate`.

`down` povrne eno migracijo in lahko izbriše podatke; ni del običajnega zagona.
Pri `dirty: true` preveri neuspešno migracijo in bazo pred nadaljevanjem.
Več: [migracije](docs/migrations.md).

Izbirni `APP_ENV=development go run ./cmd/seed` iz `backend` polni le ločeno shemo
`demo_seed`. Ne ustvari prijavnih računov ali dejanskih virov za MVP.
Več: [razvojni podatki](docs/seed-data.md).

## Prvi uporabnik in administrator

Na novi bazi na http://localhost:5173/register registriraj račun z dovoljeno domeno
in geslom dolžine 12–72 bajtov. Registracija ustvari zaposlenega, prijaviš se na `/login`.
Privzetih administratorskih poverilnic ni.

Samo za začetno nastavitev nove lokalne baze odpri PostgreSQL iz korena:

```bash
docker compose exec postgres psql -X -U deskbooker -d deskbooker
```

V psql zamenjaj primer s svojim registriranim e-naslovom in izvedi:

```sql
UPDATE public.users
SET role = 'admin'
WHERE lower(email) = lower('tvoj.racun@example.com')
RETURNING id, email, role;
```

Ukaz mora vrniti en račun z vlogo `admin`; če ne vrne nobenega, preveri e-naslov
in uspešno registracijo. Zapusti psql z `\q` in se ponovno prijavi.
Obstoječega administratorskega računa ni treba nastavljati znova. Nadaljnje vloge
urejaj prek strani **Uporabniki**. API ob vsaki zaščiteni zahtevi preveri aktualne pravice.

| Vloga v bazi / odgovoru API | Pravice |
| --- | --- |
| `employee` | Iskanje, lastna rezervacija in check-in, profil |
| `manager` | Dodatno upravljanje virov ter analitika in CSV |
| `admin` | Dodatno administracija uporabniških vlog |

API za spremembo vloge sprejema `EMPLOYEE`, `OFFICE_MANAGER`, `ADMIN`.
Administrator ne more spremeniti lastne vloge ali izbrisati lastnega računa.

## Testi in gradnja

Go preverjanje iz korena:

```bash
cd backend
go test ./...
go vet ./...
go build ./...
```

Frontend preverjanje v drugem terminalu iz korena:

```bash
cd frontend
npm ci
npm test
npm run build
```

`npm run build` ustvari `frontend/dist`. Običajni Go testi brez nastavljene
`CHECK_IN_TEST_DATABASE_URL` preskočijo integracijski test s pravo bazo.
Ta izvaja migracije in podatkovne spremembe, zato ga izvajaj **samo na prazni,
ločeni testni bazi**, nikoli na svoji razvojni ali produkcijski bazi.

CI v [`.github/workflows/ci.yml`](.github/workflows/ci.yml) sam pripravi tako bazo
in izvede `go test -race ./internal/repository -run TestCheckInIntegration -count=1 -v`.
Preveri tudi Go teste/vet/gradnjo, frontend teste/gradnjo in celotno Docker okolje:
proxy, prijavo, dovoljenja, migracije ter ohranitev podatkov po ponovni postavitvi.
Docker smoke skripta ustvarja testni račun in je namenjena temu izoliranemu CI okolju.

### Ročni preizkus MVP-ja

1. Prijavi se kot administrator in dodaj aktivno mizo, sejno sobo ali parkirišče.
2. Odpri ustrezni zavihek in preveri dejanske podatke; po spremembi klikni **Osveži vire**.
3. Izberi prihodnji datum in veljaven čas, klikni **Poišči vire** ter rezerviraj.
   Shrani referenco in štirimestni PIN z zaslona uspeha.
4. Ob razpoložljivem oknu izvedi check-in; napačen PIN se zavrne.
5. Kot vodja/administrator preveri analitiko in CSV, kot zaposleni pa zavrnitev dostopa.
6. Po `docker compose down` in `docker compose up -d` preveri ohranitev podatkov.

Podrobna pravila in preizkusi: [check-in](docs/check-in.md),
[no-show sproščanje](docs/no-show-release.md), [analitika](docs/analytics.md),
[uporabniške vloge](docs/user-administration.md), [kategorije virov](docs/resource-categories.md).

## Pogoste težave

| Težava | Preveri / rešitev |
| --- | --- |
| `cannot find main module` | Go ukaze zaženi iz `backend`, ne iz korena. |
| `port is already allocated` | Ustavi lokalni API/Vite na 8080/5173; za bazo lahko v `.env` uporabiš `POSTGRES_PORT=5433`. |
| Compose zahteva manjkajočo spremenljivko | Korenski `.env` mora imeti geslo baze, JWT ključ in domeno registracije. |
| Ni povezave ali napačno geslo baze | Preveri `docker compose ps -a`, vrata in dosedanje geslo volume; sprememba `.env` ne spremeni gesla že ustvarjene baze. |
| Migracije odpovejo / API se ne zažene | `docker compose logs migrate backend`; preveri tudi `migrate status`. |
| HTTP 401 po nekaj časa | JWT velja 15 minut; ponovno se prijavi. |
| HTTP 403 | Račun nima pravic; preveri njegovo trenutno vlogo in stanje. |
| Registracija zavrne domeno / HTTP 503 | Nastavi `REGISTRATION_EMAIL_DOMAIN`; uskladi frontend namig ter ponovno zaženi/obnovi gradnjo. |
| V kategoriji ni virov | Dodaj aktiven vir ustreznega tipa; pri časovnih filtrih preveri zasedenost. |
| CORS pri lokalnem razvoju | Izvor brskalnika se mora ujemati z `CORS_ALLOWED_ORIGINS`; `localhost` in `127.0.0.1` sta različna izvora. |

## Struktura projekta in nadaljnja dokumentacija

| Pot | Vsebina |
| --- | --- |
| [`backend/`](backend/README.md) | Go API, migracije, modeli in repository |
| [`frontend/`](frontend/README.md) | React aplikacija, filtri, administracija in UI testi |
| [`compose.yaml`](compose.yaml) | Skupni zagon PostgreSQL, migracij, API-ja in frontenda |
| [`docs/docker.md`](docs/docker.md) | Podrobna Docker navodila in ohranitev podatkov |
| [`docs/04-arhitektura-in-tehnoloski-sklad.md`](docs/04-arhitektura-in-tehnoloski-sklad.md) | Arhitektura in tehnološki sklad |
| [`docs/05-uporabniske-zgodbe.md`](docs/05-uporabniske-zgodbe.md) | Uporabniške zgodbe |

Navodila opisujejo lokalni zagon MVP-ja. Javna namestitev potrebuje tudi HTTPS,
produkcijsko upravljanje skrivnosti in varnostne kopije baze.
