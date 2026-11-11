# Docker Compose (DB-111)

## Prvi zagon

Potrebuješ Docker Desktop z vključenim Docker Engine in Docker Compose v2.
Ukaze izvajaj iz korenske mape DeskBooker.

1. Če `.env` še ne obstaja: `cp .env.example .env`. Obstoječega `.env` ne prepisuj.
2. Nastavi `POSTGRES_PASSWORD`, `JWT_SECRET` in `REGISTRATION_EMAIL_DOMAIN`.
   JWT ključ ustvari z `openssl rand -hex 32` in rezultat shrani kot `JWT_SECRET=...`.
   Domena naj bo ista kot pri dosedanjem lokalnem zagonu.
3. Ustavi lokalni `go run ./cmd/server` in `npm run dev` (Ctrl+C), saj uporabljata ista vrata.
4. Zaženi:

```bash
docker compose up -d --build
docker compose ps -a
docker compose logs --tail=50 migrate backend frontend
```

Odpri http://localhost:5173. API je tudi neposredno na http://localhost:8080/api/health/db.
Frontend streže prevedeno React aplikacijo z nginxom; `/api/` posreduje Go storitvi,
zato brskalnik ne potrebuje dostopa do Docker imen `backend` ali `postgres`.
Osveževanje podstrani, npr. `/desks`, deluje prek SPA fallbacka.

## Vrstni red in podatki

PostgreSQL mora prestati healthcheck, nato enkratna storitev `migrate` izvede
obstoječe SQL migracije. API se zažene šele po uspešnih migracijah, frontend pa
po zdravem API-ju. `migrate` v stanju `Exited (0)` je pričakovano.
Neuspešne migracije preprečijo zagon API-ja; vzrok preveri v `docker compose logs migrate`.
Obstoječe migracije ostanejo nespremenjene (trenutno verzija 13).

Ohranjen je obstoječi volume `deskbooker_pgdata`, ime storitve `postgres`, ime baze
in uporabnik `deskbooker`. Ob zagonu iz iste mape oziroma z istim Compose project
imenom se uporabi dosedanja baza. Uporabi tudi isto geslo za PostgreSQL; sprememba
`.env` ne spremeni gesla v že inicializirani bazi. Drugo project ime pomeni drug volume.
Samodejnega seedanja, brisanja ali ustvarjanja administratorskega računa ni.

```bash
# Ustavitev, podatki ostanejo:
docker compose down
# Ponovni zagon:
docker compose up -d
# Stanje migracij:
docker compose run --rm migrate /app/migrate status
# Ponovitev migracij (varno tudi brez novih migracij):
docker compose run --rm migrate
# Dnevnik API-ja:
docker compose logs -f backend
```

`docker compose down -v` izbriše podatkovni volume. Za običajno ustavitev ga ne uporabljaj.

## Konfiguracija

- `.env` ostane na gostitelju in se ne kopira v slike. Compose posreduje izbrane
  nastavitve procesom. Geslo in JWT ključ nista frontend build argumenta.
- `POSTGRES_PORT` v korenskem `.env` določa samo vrata baze na gostitelju
  (npr. 5433); znotraj Dockerja sta naslov in vrata vedno `postgres:5432`.
- Lokalni `DATABASE_URL` in `POSTGRES_HOST=localhost` se ne prenašata v zabojnike.
- Frontend je na 5173, API na 8080, oboje vezano na lokalni računalnik.
- `REGISTRATION_EMAIL_DOMAIN` se uporabi tudi pri gradnji frontend prikaza.
  Po spremembi nastavitve znova uporabi `docker compose up -d --build`.
- Frontend uporablja relativni `/api`; dosedanji `npm run dev` ohrani privzeti
  API naslov `http://localhost:8080/api` in obstoječo CORS konfiguracijo.
- Listmonk ni dodan: trenutni MVP nima povezave za pošiljanje pošte prek njega.
  Dodamo ga skupaj s poštno funkcionalnostjo, ko bo potrebna.

To je okolje za lokalni MVP. Javna produkcijska namestitev potrebuje še domeno,
HTTPS in ustrezno upravljanje produkcijskih skrivnosti.

## Preverjanje

Celotno okolje preveri ročno: zgradi in zaženi storitve, preveri stanje migracij,
prijavo, zaščiten API prek nginx proxyja ter osvežitev SPA podstrani.
Za Docker smoke skripto uporabi ločeno testno okolje po [navodilih](ci.md).
Samodejni CI/CD je odložen do naslednje naloge.

Ročno preveri obstoječo prijavo, prikaz dejanskih virov in rezervacijo. Po ponovnem
zagonu morajo uporabniki in rezervacije ostati v bazi.
