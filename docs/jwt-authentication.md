# DB-024 – JWT avtentikacija (US-01)

JWT dostopni žetoni so podpisani z **HMAC SHA-256 (HS256)**, imajo
omejeno veljavnost **15 minut** in so namenjeni API-ju DeskBooker.
Podpis in časovne trditve preverja backend; v HTTP odgovoru nikoli ne
pošilja gesla ali shranjenega geselskega hasha.

## Obvezna lokalna konfiguracija

Iz korenske mape repozitorija generiraj skrivni ključ:

```bash
openssl rand -hex 32
```

Pridobljeni rezultat (64 naključnih šestnajstiških znakov) shrani v
**lokalno** korensko datoteko `.env`:

```dotenv
JWT_SECRET=<TUKAJ_NAKLJUČNI_IZPIS>
```

Zgornja vrstica je vzorec, ne dejanski ključ. `JWT_SECRET` mora
vsebovati vsaj 32 bajtov; API se brez ustrezne nastavitve ne zažene.
Ključa ne objavljaj v Git, logih ali `VITE_*` spremenljivkah.
Zamenjava skrivnosti razveljavi vse prej izdane žetone.

`REGISTRATION_EMAIL_DOMAIN=gmail.com` lahko ostane za lokalno
registracijo, ker JWT deluje neodvisno od nastavljenih domen.

## Prijava

```http
POST /api/auth/login
Content-Type: application/json

{"email":"ana@gmail.com","password":"VeryStrongPass123!"}
```

Pravilna prijava vrne HTTP **200**:

```json
{
  "user": { "id": "...", "email": "ana@gmail.com", "role": "employee" },
  "access_token": "<signed-JWT>",
  "token_type": "Bearer",
  "expires_at": "2030-01-01T12:15:00Z"
}
```

Čas v primeru je zgolj prikaz strukture. Dejanski žeton poteče 15 minut
po izdaji. Zastavica `Cache-Control: no-store` preprečuje predpomnjenje
običajnih HTTP odgovorov. JWT ni shranjen v bazi in ne vsebuje gesla.

## Preverjanje prijavljenega uporabnika

```http
GET /api/auth/me
Authorization: Bearer <signed-JWT>
```

API preveri podpis, izključno algoritem HS256, `iss`, `aud`, `sub`,
`iat`, `nbf` in `exp`. Nato uporabnika ponovno poišče v PostgreSQL
in vrne varni `UserDTO` s HTTP **200**.

Če žeton manjka, je neveljaven, potekel ali uporabnik več ne obstaja,
API vrne HTTP **401**. Napake baze vrnejo HTTP **500**, brez podrobnosti
povezave. Zaščitena pot ne vrača hasha in ne uporablja piškotkov.

## Lokalni preizkus

Iz korenske mape:

```bash
git fetch origin
git switch --track origin/feature/db-024-jwt-authentication
docker compose up -d postgres
cd backend
go test ./...
go vet ./...
go build ./...
go run ./cmd/migrate status
go run ./cmd/server
```

Preveri verzijo migracij **7, dirty: false**.
V drugem terminalu pridobi žeton iz odgovora (primer izogiba se
zapisovanju v Git, ne izpisuje žetona samodejno v zgodovino ukazov):

```bash
TOKEN=$(curl -fsS -X POST http://localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"ana@gmail.com","password":"VeryStrongPass123!"}' \
  | python3 -c 'import sys,json; print(json.load(sys.stdin)["access_token"])')
curl -i http://localhost:8080/api/auth/me -H "Authorization: Bearer $TOKEN"
curl -i http://localhost:8080/api/auth/me
curl -i http://localhost:8080/api/auth/me -H "Authorization: Bearer ${TOKEN}invalid"
unset TOKEN
```

Prvi `/api/auth/me` mora vrniti **200**, naslednja dva pa **401**.
Če `ana@gmail.com` v lokalni bazi ne obstaja, ga prej registriraj prek
DB-021 oziroma uporabi svoj obstoječi testni račun.

## Varnostne omejitve

- JWT je **bearer credential**: kdor dobi žeton, se do poteka lahko
  predstavi kot uporabnik. Ne objavljaj ga in ne zapisuj v lokalni
  pomnilnik brskalnika brez varnostnega pregleda.
- V produkciji je obvezen HTTPS in varna centralizirana hramba skrivnosti.
- Ta naloga **ne vključuje** refresh žetonov, odjave, rotacije žetonov,
  revokacije že izdanih žetonov ali omejevanja števila poskusov prijave.
  To je treba rešiti pred javno produkcijsko uporabo.
- Omejitve pravic glede na vlogo in zaščita drugih API poti bosta
  ločeni funkcionalnosti.
