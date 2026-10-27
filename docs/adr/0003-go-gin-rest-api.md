# ADR 0003 – Go in Gin za strežniški REST API

- **Status:** Sprejeto
- **Datum:** 2026-10-27

## Kontekst in problem

API DeskBooker obdeluje prijavo, pravice glede na vloge, iskanje virov in transakcijske rezervacije. Potreben je predvidljiv strežniški del, ki lahko poganja tudi periodično opravilo za sproščanje nepotrjenih rezervacij.

## Obravnavane alternative

- **Go z ogrodjem Gin:** preprost proces, učinkovit HTTP API in podpora sočasnim opravilom.
- **Node.js z Expressom:** skupni jezik s frontendom, vendar drugačen model izvajanja in orodja.
- **Python z Django REST Framework:** veliko vgrajenih funkcij, vendar težje ogrodje glede na predviden obseg.

## Sprejeta odločitev in utemeljitev

Izberemo **Go z ogrodjem Gin** za REST API z JSON komunikacijo. Go omogoča razmeroma preprost strežniški proces, Gin pa usmerjanje zahtevkov in middleware za avtentikacijo; ključna preverjanja rezervacij ostanejo na strežniku.

## Posledice

### Pozitivne

- Enostavna zasnova in razmeščanje strežniške komponente.
- Dobra podpora sočasnemu izvajanju in periodičnim opravilom.
- Poslovna pravila so neodvisna od React odjemalca.

### Negativne

- Ekipa mora vzdrževati dva programska jezika (JavaScript in Go).
- Manj vnaprej pripravljenih aplikacijskih funkcij kot pri bolj celovitih ogrodjih.

## Povezave

- [Arhitektura in tehnološki sklad](../04-arhitektura-in-tehnoloski-sklad.md)
