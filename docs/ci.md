# Ročno preverjanje – CI/CD je odložen

Workflow `.github/workflows/ci.yml` je odstranjen. GitHub ob pushu ali PR-ju
ne izvaja več samodejnih testov, gradenj ali nameščanja aplikacije.
CI/CD bo dodan pri naslednji nalogi. Pretekli zagoni ostanejo v GitHub zgodovini.

## Frontend in Go

Iz korena projekta:

```bash
cd frontend
npm ci
npm test
npm run build
cd ../backend
go mod download
go test ./...
go vet ./...
go build ./...
```

Brez `CHECK_IN_TEST_DATABASE_URL` se integracijski test s pravo bazo preskoči.

## PostgreSQL integracija

Potrebuješ prazno namensko PostgreSQL 17 bazo. Test sam izvaja migracije in
podatkovne spremembe, zato ne uporabi razvojne ali produkcijske baze.
V terminalu nastavi `CHECK_IN_TEST_DATABASE_URL` na povezovalni niz te ločene baze,
nato iz `backend` izvedi:

```bash
go test -race ./internal/repository -run TestCheckInIntegration -count=1 -v
```

Za vsak ponovni zagon pripravi prazno testno bazo. Preizkus pokriva rezervacije,
sočasne zahteve, check-in, no-show sproščanje, analitiko, pravice in dejanske vire.

## Docker smoke

Skripta `scripts/docker-smoke.py` ostane na voljo za ročno preverjanje.
Ustvari testnega zaposlenega z e-naslovom `compose-smoke@<domena>` in testnim geslom,
zato jo uporabi samo v ločenem testnem checkoutu z lastnim Compose project imenom,
lastnim `.env` in praznim podatkovnim volume. Vrata 5173, 8080 in izbrana vrata baze
morajo biti prosta. Ne izvajaj je na običajnem MVP okolju.

Iz korena tega testnega checkouta najprej nastavi ločeno geslo baze, JWT ključ in
domeno `example.com` v `.env`. V istem terminalu izvedi:

```bash
export COMPOSE_PROJECT_NAME=deskbooker-smoke
export REGISTRATION_EMAIL_DOMAIN=example.com
docker compose up -d --build --wait --wait-timeout 120
python3 scripts/docker-smoke.py
docker compose run --rm migrate
docker compose down
docker compose up -d --wait --wait-timeout 120
python3 scripts/docker-smoke.py --existing
docker compose down
```

Prvi zagon skripte zahteva, da testni uporabnik še ne obstaja. `--existing` preveri
njegovo ohranitev po ponovni postavitvi. Zadnji `down` ohrani namenski testni volume.
Skripta preveri SPA podstran, nginx proxy, povezavo z bazo, prijavo, zaščiten API
in zavrnitev administratorskega dostopa zaposlenemu. Za novi sveži preizkus pripravi
nov ločen testni project/volume. Pregled zdravstvenih stanj in ročni poslovni tok
sta opisana v [Docker navodilih](docker.md) in [README-ju](../README.md).
