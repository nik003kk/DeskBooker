# DeskBooker – frontend

React 19 + Vite 7 + React Router.

## Lokalni zagon

Predpogoji: Node.js 20.19+ ali 22.12+ in npm.

```bash
cd frontend
npm ci
npm test
npm run build
npm run dev
```

V brskalniku odpri http://localhost:5173.

## Navigacija in strani (DB-012)

| Pot | Stran |
| --- | --- |
| `/` | Začetna stran |
| `/desks` | Delovna mesta |
| `/rooms` | Sejne sobe |
| `/parking` | Parkirišča |
| `/equipment` | Oprema |
| `/my-reservations` | Moje rezervacije |
| `/login` | Prijava |
| Neznana pot | 404 |

## UI komponente in odzivna postavitev (DB-013)

- `src/components/ui.jsx`: ponovno uporabne komponente `Button`, `ButtonLink`, `Card`, `Field`, `SelectField`, `StatusBadge`.
- `src/index.css`: skupni design tokens, stanja hover/focus/disabled, prilagodljive mreže in prelomi za tablične ter mobilne naprave.
- Glava z aktivno navigacijo, dostopna preskočna povezava, začetni pregled virov, kartice in osnovni iskalni obrazec.
- Iskanje po imenu in nadstropju uporablja **lokalne demonstracijske podatke**. Ne pošilja zahtevkov na backend in ne shranjuje rezervacij.
- Gumbi »Rezerviraj (kmalu)« so onemogočeni, dokler funkcionalnost ni implementirana.

**Wireframei:** v repozitoriju ob pripravi DB-013 ni bilo priloženih wireframe datotek. Uporabljena je začasna skladna vizualna osnova; po prejemu dejanskih wireframeov je treba preveriti in po potrebi uskladiti komponente.

## Preverjanje

```bash
npm test
npm run build
npm run dev
```

V brskalniku ročno preveri:
1. Poti `/`, `/desks`, `/rooms`, `/parking`, `/equipment`, `/my-reservations`, `/login` in neobstoječo pot.
2. Na strani `/desks` vnesi `D-101` in izberi »Prikaži«: ostati mora ena kartica; `xyz-ne-obstaja` mora prikazati prazno stanje; »Ponastavi« ali »Počisti filter« vrne vse demonstracijske kartice.
3. Na strani `/rooms` izberi 2. nadstropje: vidna morata biti Pohorje in Bled.
4. Zaslon približno 375 px širok: navigacija je vodoravno premakljiva, kartice in obrazci se postavijo v en stolpec in brez vodoravnega razlivanja.
5. Tab navigacija in vidna stanja fokusa, povezava »Preskoči na vsebino«, onemogočeni gumbi rezervacij.

Produkcijska gradnja: `npm run build` ustvari `dist/`, ki je ignoriran v Git.

## Registracijski obrazec (DB-025)

Nova pot `/register` uporablja React obrazec za ime, priimek, službeni
e-naslov, geslo in ponovitev gesla. Preverja zahtevana polja, domeno,
dolžino gesla (**12–72 bajtov**) in ujemanje gesel ter pošlje
`POST /api/auth/register` v obstoječi Go backend.
Ob uspehu pokaže potrditev in povezavo na prijavo, ob že obstoječem
naslovu pa razumljivo sporočilo. Geslo se ne zapisuje v browser storage.

Za lokalni razvoj iz mape `frontend` ustvari `.env.local` (ta datoteka
naj ostane izven Git):

```dotenv
VITE_API_BASE_URL=http://localhost:8080/api
VITE_REGISTRATION_EMAIL_DOMAIN=gmail.com
```

`VITE_REGISTRATION_EMAIL_DOMAIN` mora ustrezati nastavitvi
`REGISTRATION_EMAIL_DOMAIN` v korenski `.env`. Za testiranje lahko
obe nastavita `gmail.com`; v produkciji uporabi dejansko službeno domeno.
Frontend validacija je uporabniška pomoč, **backend je avtoritativen**.
Vsi `VITE_*` podatki so javni v brskalniku: nikoli ne nastavljaj
`JWT_SECRET` ali gesel na frontend strani.

Zagon (dva terminala): iz korena najprej
`docker compose up -d postgres`, nato `cd backend && go run ./cmd/server`
(zahteva korenski `.env` z veljavnim `JWT_SECRET` iz DB-024).
V drugem terminalu izvedi:

```bash
cd frontend
npm ci
npm test
npm run build
npm run dev
```

Odpri <http://localhost:5173/register>. Preveri:
- prazna polja, napačna domena in prekratko geslo prikažejo lokalne napake;
- neusklajena gesla preprečijo oddajo;
- veljaven še neuporabljen e-naslov ustvari račun in prikaže potrditev;
- ponovna registracija istega naslova pokaže napako HTTP 409.
  Za to osveži stran in znova pošlji enake podatke.

**Pozor:** uspešna registracija ustvari resničen zapis v lokalni tabeli
`public.users`. Za ponavljanje uporabite nov testni naslov.
