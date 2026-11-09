# DB-023 – prijava uporabnikov

Endpoint `POST /api/auth/login` preveri e-poštni naslov in geslo
z uporabo `repository.GetUserByEmail` ter `auth.VerifyPassword`.

Zahtevek:

```http
POST /api/auth/login
Content-Type: application/json

{"email":"ana@gmail.com","password":"VeryStrongPass123!"}
```

Uspeh: **HTTP 200 OK** in `{"user":{...}}` z varnim
`UserDTO` brez gesla ali hasha. Odgovor vsebuje glavo
`Cache-Control: no-store`. Endpoint **ne izdaja JWT, seje ali piškotkov**:
preveri identiteto, ne omogoča še prijavljenega dostopa do zaščitenih
API poti. To je predmet nadaljnjih opravil.

Napačno geslo in neobstoječi uporabnik dobita enak odgovor:
`401 {"error":{"code":"invalid_credentials","message":"Invalid email or password"}}`.
Neveljavno JSON telo ali manjkajoča polja: **400**;
napačen Content-Type: **415**; DB napaka: **500** brez notranjih
podrobnosti.

Geslo je občutljivo: ne pošiljaj ga po nezavarovani povezavi izven
lokalnega razvojnega okolja. Pri dejanski uporabi je obvezna HTTPS
povezava. Endpoint za zdaj nima omejevanja poskusov prijave; pred javno
objavo je potreben rate limiting in varna uprava sej.

## Lokalni preizkus

```bash
git fetch origin
git switch --track origin/feature/db-023-login-api
docker compose up -d postgres
cd backend
go test ./...
go vet ./...
go build ./...
go run ./cmd/migrate status
go run ./cmd/server
```

Migracije ostajajo **7, dirty: false**. V drugem terminalu preveri že
registriran lokalni račun (na primer uporabnika iz DB-021, če še obstaja):

```bash
curl -i -X POST http://localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"ana@gmail.com","password":"VeryStrongPass123!"}'
```

Pričakuj HTTP **200** in podatke o uporabniku brez občutljivih polj.
Pri napačnem geslu:

```bash
curl -i -X POST http://localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"ana@gmail.com","password":"WrongPassword123!"}'
```

Pričakuj **401**. Tudi za neobstoječ e-naslov mora biti odgovor 401.
Pri lokalnem preverjanju nikoli ne uporabljaj resničnih produkcijskih
gesel ali objavljaj občutljivih podatkov iz zahtevkov.
