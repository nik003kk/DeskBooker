# DeskBooker – frontend

React 19 + Vite 7 + React Router (DB-012).

## Predpogoji

- Node.js, združljiv z Vite 7 (20.19+ ali 22.12+)
- npm

## Lokalni zagon

Iz korenske mape projekta:

```bash
cd frontend
npm install
npm run dev
```

Odpri naslov Vite (običajno http://localhost:5173).

## Navigacija

| Pot | Stran |
| --- | --- |
| `/` | Začetna stran in hitre povezave |
| `/desks` | Delovna mesta |
| `/rooms` | Sejne sobe |
| `/parking` | Parkirišča |
| `/equipment` | Oprema |
| `/my-reservations` | Moje rezervacije |
| `/login` | Prijava |
| Neznana pot | Stran 404 in povezava domov |

Navigacija uporablja `NavLink` brez osveževanja strani. Aktivna povezava je vizualno označena;
usmerjanje upravlja `BrowserRouter`. Poti so centralizirane v `src/routes.js`.
Navigacija je prilagojena ozkim zaslonom in dostopna s tipkovnico.

**Pomembno:** DB-012 implementira samo usmerjanje, osnovno navigacijo in ogrodje strani.
Podatki, prijava in rezervacije še niso implementirani; prikazane so strani »V pripravi«.

## Testi in gradnja

```bash
npm test
npm run build
```

`npm test` preverja poti, unikatnost, povezave med navigacijo in glavnimi sekcijami.
Ročno v brskalniku preveri premikanje med stranmi, osvežitev neposredne povezave
(npr. `http://localhost:5173/rooms`), neobstoječo pot in obnašanje navigacije na telefonu.

`npm install` bo posodobil `frontend/package-lock.json`. **Commitaj in pushaj
lockfile na feature vejo**, preden se PR združi. Ne dodajaj `node_modules/` ali `dist/`.
