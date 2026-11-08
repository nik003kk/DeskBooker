# DB-022 – varno shranjevanje in preverjanje gesel

Modul `backend/internal/auth/password.go` uvaja centralizirani funkciji:
- `HashPassword(plain)`: zahteva 12–72 bajtov, izdela soljeni bcrypt
  hash (privzeti strošek `bcrypt.DefaultCost`).
- `VerifyPassword(storedHash, plain)`: preveri geslo s standardnim
  bcrypt primerjanjem. Za napačno geslo, pokvarjen hash ali geslo,
  daljše od 72 bajtov, vrne isto napako `ErrInvalidCredentials`.

Meja 72 bajtov prepreči, da bi bcrypt pri preverjanju ignoriral
dodatne bajte gesla. Gre za bajte, ne za število Unicode znakov.

Registracija iz DB-021 uporablja `HashPassword` pred zapisom v
`public.users.password_hash`, nikoli ne shrani surovega gesla.
Javni `UserDTO` ne vsebuje hasha. Uporabniški repository doda
`GetUserByEmail`, ki poizveduje s parametriziranim SQL in omogoča
nalogi DB-023 preverjanje gesla ob prijavi.

## Obseg

Endpoint `POST /api/auth/login`, seje in žetoni **niso** del DB-022;
spadajo v DB-023 in nadaljnje naloge. Ta naloga vsebuje geselsko
logiko in podatkovno podporo za preverjanje ob prijavi.

## Lokalni testi

```bash
git fetch origin
git switch --track origin/feature/db-022-password-security
cd backend
go test ./...
go vet ./...
go build ./...
```

Testi preverijo različne bcrypt soli, pravilen in napačen vnos,
neveljaven hash, prekratka/predolga gesla, UTF-8 bajtno mejo in
parametrizirano poizvedbo za e-poštni naslov.

## Preverjanje na dejanski lokalni bazi

Migracije morajo ostati na verziji **7, dirty: false**. Za neinvaziven
integracijski preizkus brez trajnih testnih podatkov zaženi:

```bash
docker compose up -d postgres
cd backend
go run ./cmd/migrate status
go run ./cmd/password-smoke
```

Integracijski ukaz ustvari testnega uporabnika v transakciji,
preveri, da je v podatkovni bazi shranjen bcrypt hash, nato
primerja pravilno in napačno geslo. Na koncu vedno izvede
`ROLLBACK`, zato testni račun ne ostane v bazi.
