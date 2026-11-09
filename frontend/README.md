# DeskBooker – frontend

React 19, Vite 7 in React Router. Celotna namestitev, vloge in Docker zagon
so opisani v [glavnem README-ju](../README.md).

## Lokalni razvoj

Potrebuješ Node.js 22.12+ in npm ter delujoč Go API na http://localhost:8080.
Za nov lokalni frontend iz korena kopiraj `frontend/.env.example` v
`frontend/.env.local`, če ta še ne obstaja. Nastavi:

```dotenv
VITE_API_BASE_URL=http://localhost:8080/api
VITE_REGISTRATION_EMAIL_DOMAIN=example.com
```

Domena mora biti ista kot `REGISTRATION_EMAIL_DOMAIN` v korenskem `.env`.
V `VITE_*` ne vpisuj gesel ali ključev. Po spremembi znova zaženi Vite.
Naslednje ukaze izvajaj iz `frontend`:

```bash
npm ci
npm run dev
```

Odpri http://localhost:5173. Pri Docker zagonu frontend streže nginx;
`/api/` posreduje backendu in omogoča neposredno osveževanje SPA podstrani.
Docker ne nalaga lokalnih `.env*` datotek frontenda.

## Testi in gradnja

Iz `frontend`:

```bash
npm test
npm run build
```

Testi vključujejo API funkcije in UI tokove prek jsdom. Gradnja ustvari `dist/`.
Ročno preveri prijavo, dejanske vire, filtriranje in rezervacijo po postopku
v glavnem README-ju.

## Glavne strani

| Pot | Namen / dostop |
| --- | --- |
| `/register`, `/login` | Registracija in prijava |
| `/profile` | Profil prijavljenega uporabnika |
| `/desks`, `/rooms`, `/parking` | Dejanski aktivni viri, filtri razpoložljivosti in rezerviranje |
| `/search` | Iskanje po vseh vrstah virov |
| `/reservations/confirm`, `/reservations/success` | Potrditev izbora in rezultat rezervacije s PIN-om |
| `/reservations/:id/check-in` | Check-in lastne rezervacije |
| `/admin/resources`, `/admin/resources/new`, `/admin/resources/:id/edit` | Upravljanje virov za vodjo in administratorja |
| `/analytics` | Analitika in CSV za vodjo in administratorja |
| `/admin/users` | Administracija uporabnikov samo za administratorja |

Zavihka `/equipment` in `/my-reservations` sta še predstavitvena; ne predstavljata
končanega samostojnega rezerviranja opreme ali pregleda lastnih rezervacij.
Rezerviranje desk/sejnih sob/parkirišč in dostop do check-ina uporabljata zgornje poti.

JWT je shranjen v `sessionStorage` trenutnega zavihka in velja 15 minut. Ob HTTP 401
se seja odstrani, HTTP 403 pomeni manjkajoče pravice. Zaščito vedno uveljavlja tudi API.
PIN uspešne rezervacije je v začasnem Router state; po osvežitvi zaslona lahko izgine.

Podrobnosti: [kategorije virov](../docs/resource-categories.md),
[check-in](../docs/check-in.md), [analitika](../docs/analytics.md) in
[administracija uporabnikov](../docs/user-administration.md).
