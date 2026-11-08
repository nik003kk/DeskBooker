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

## Prijavna stran (DB-026)

`/login` vsebuje obrazec za e-pošto in geslo, kliče `POST /api/auth/login`
in ob pravilnih podatkih preusmeri na osnovno ploščo `/`.
Prikazuje validacijske napake ter HTTP 401, ne shranjuje gesla.
Trajno stanje prijave in odjava prideta v DB-027.

Preverjanje: `npm test && npm run build && npm run dev`, nato odpri
`http://localhost:5173/login`; z obstoječim računom preveri preusmeritev
na domov, z napačnim geslom pa sporočilo o napaki.

## Stanje prijave in odjava (DB-027)

Po uspešni prijavi se JWT in potek veljavnosti hranita v `sessionStorage`
trenutnega zavihka (ne v `localStorage`). Ob obnovitvi strani se
uporabnik preveri prek `GET /api/auth/me`, neveljaven ali potekel
žeton pa se izbriše. Navigacija prikaže gumb **Odjava**, ki odstrani
lokalno sejo. JWT ima veljavnost 15 minut in se ne osvežuje samodejno.

Za produkcijo velja opozorilo: JavaScript dosegljivi JWT v
`sessionStorage` zahtevajo močno zaščito pred XSS. Dolgoročno je
priporočljivo oceniti varno strežniško sejo ali HttpOnly piškotke.

## Zaščitene poti in preverjanje dovoljenj (DB-028)

React `/my-reservations` je dostopen samo ob veljavni prijavi,
sicer preusmeri na `/login`. `authorizedFetch` samodejno priloži
Bearer JWT: pri HTTP 401 odstrani lokalno sejo, pri HTTP 403 pa
prikaže napako dovoljenj, ne da bi uporabnika odjavil.

Backend `auth.RequireJWT` preverja podpis in veljavnost, novi
`auth.RequireRoles` pa pri dovoljenjih prebere dejansko vlogo iz baze
(ne iz poljubne zahteve). Vse prihodnje administrativne API poti
lahko uporabljajo `RequireJWT` in `RequireRoles` skupaj.

## Moj profil (DB-031)

Zaščitena pot `/profile` prebere `GET /api/users/me` z JWT ter
prikaže ime, priimek, e-pošto, oddelek, vlogo in stanje računa.
Neprijavljene preusmeri na prijavo, HTTP 401 odstrani sejo in HTTP
403 pokaže sporočilo o dovoljenjih. `npm test && npm run build`.

## Urejanje profila (DB-032)

Na `/profile` lahko spremeniš ime, priimek in oddelek. Obrazec
validira podatke pred pošiljanjem, uporablja zaščiten `PATCH
/api/users/me` in po uspehu osveži prikazane podatke. Prazni oddelek
se pošlje kot `null`; e-pošte ali vloge ni mogoče urejati.

## Dodajanje virov (DB-036)

Upravitelj ali skrbnik lahko na `/admin/resources/new` doda mizo, sejno sobo
ali parkirišče, nastavi nadstropje, kapaciteto, aktivnost in opremo.
Obrazec preveri polja; za ustvarjanje uporabi `POST /api/resources`,
za opremo pa `POST /api/resources/:id/features`. Ker sta to ločena zahtevka,
ob napaki shranjevanja opreme sporočilo pove, da je bil vir že ustvarjen.
