# DB-021 – registracija uporabnikov

HTTP endpoint: `POST /api/auth/register`. Ustvari uporabnika z vlogo
`employee` (dodeli jo PostgreSQL). Registracija ne ustvarja seje/JWT in
ne pomeni samodejne prijave.

**Obvezna nastavitev:** V korenski lokalni `.env` dodaj
`REGISTRATION_EMAIL_DOMAIN=example.com`, kjer **example.com zamenjaš z
dejansko domeno službenih e-naslovov**. Brez konfigurirane domene
endpoint vrne HTTP 503, ne dovoli registracije. Dovoljena je točno ena
domena. To je začetna varovalka, ni preverjanje lastništva e-naslova:
verifikacija po e-pošti, rate limiting in prijava pridejo v prihodnjih
nalogah.

## Zahtevek

```http
POST /api/auth/register
Content-Type: application/json

{
  "first_name": "Ana",
  "last_name": "Novak",
  "email": "ana@example.com",
  "password": "VeryStrongPass123!"
}
```

Validacija: ime/priimek neprazna, največ 100 Unicode znakov in brez
kontrolnih znakov; e-naslov skladen z dovoljeno domeno, normaliziran
v male črke; geslo 12–72 **bajtov**. Neznana JSON polja se zavrnejo,
zato stranka ne more sama zahtevati vloge `admin` ali poslati
`password_hash`. Geslo se pred zapisom zgošči z bcrypt. Surovo geslo
ali hash nista nikoli v odgovoru.

Uspeh: HTTP **201 Created**, `{"user": {...}}` s podatki iz
`UserDTO` (ID, email, ime, priimek, vloga, oddelek, created_at).
Napake: HTTP **400** za neveljavne podatke, **409** za že obstoječi
e-naslov (tudi različna velikost črk), **415** za napačen Content-Type,
**503** za neaktivno konfiguracijo, **500** za nepredvideno DB napako.
Napake imajo obliko `{"error":{"code":"...","message":"..."}}`.

## Lokalni preizkus

Iz korena repozitorija:

```bash
git fetch origin
git switch --track origin/feature/db-021-registration-api
docker compose up -d postgres
cd backend
go test ./...
go vet ./...
go build ./...
go run ./cmd/migrate status
go run ./cmd/server
```

Migracije: **7, dirty: false**. Zadnji ukaz teče v odprtem terminalu.
V drugem terminalu izvedi (glej domensko nastavitev zgoraj):

```bash
curl -i -X POST http://localhost:8080/api/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"first_name":"Ana","last_name":"Novak","email":"ana@example.com","password":"VeryStrongPass123!"}'
```

Prvič pričakuj HTTP **201**, ob ponovitvi istega zahtevka pa **409**.
Za neveljaven e-naslov:

```bash
curl -i -X POST http://localhost:8080/api/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"first_name":"Ana","last_name":"Novak","email":"ana@gmail.com","password":"VeryStrongPass123!"}'
```

Pričakuj HTTP **400**. **Pozor:** uspešen ročni preizkus
ustvari dejanskega uporabnika v `public.users`; pri ponovljenih
testih uporabi drugo ime predalčka ali ga v razvojni bazi
varno odstrani po preverjanju. Testi `go test ./...` uporabljajo
lažen repository in ne ustvarijo podatkov v PostgreSQL.

Strežnik pred zagonom zahteva lokalno povezavo s PostgreSQL.
