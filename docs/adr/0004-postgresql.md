# ADR 0004 – PostgreSQL kot relacijska podatkovna baza

- **Status:** Sprejeto
- **Datum:** 2026-10-27

## Kontekst in problem

Podatkovni model vsebuje povezane entitete User, Resource, ResourceFeature in Reservation. Pri ustvarjanju rezervacij je pomembno dosledno preverjanje časovnih prekrivanj ter ohranjanje referenčne integritete.

## Obravnavane alternative

- **PostgreSQL:** relacijske omejitve in transakcije; zahteva upravljanje baze in migracij.
- **MySQL/MariaDB:** prav tako relacijska alternativa z transakcijami, vendar bi morali enako načrtovati zaklepanje in omejitve.
- **MongoDB:** prilagodljiv dokumentni model, vendar manj naraven za trenutno povezane entitete in pravila rezervacij.

## Sprejeta odločitev in utemeljitev

Izberemo **PostgreSQL**, ker relacijski model in transakcije ustrezajo rezervacijam. Preprečevanje dvojnih rezervacij bomo zasnovali z ustreznim transakcijskim preverjanjem oziroma podatkovnimi omejitvami; sama izbira baze brez teh pravil še ne prepreči prekrivanj.

## Posledice

### Pozitivne

- Tuji ključi ščitijo povezave med uporabniki, viri in rezervacijami.
- Transakcije so podlaga za dosledno ustvarjanje rezervacij.
- SQL je primeren za poročila o zasedenosti.

### Negativne

- Zahteva migracije sheme in vzdrževanje baze.
- Pravila časovnih prekrivanj zahtevajo posebno implementacijo in teste.

## Povezave

- [Arhitektura in tehnološki sklad](../04-arhitektura-in-tehnoloski-sklad.md)
