# ADR 0001 – Ločena tro-nivojska arhitektura

- **Status:** Sprejeto
- **Datum:** 2026-10-27

## Kontekst in problem

DeskBooker je spletna aplikacija za rezervacije z različnimi uporabniškimi vlogami. Podatke mora varno shranjevati, preprečevati časovno prekrivajoče se rezervacije in omogočati razvoj uporabniškega vmesnika neodvisno od poslovnih pravil.

## Obravnavane alternative

- **Ločeni React odjemalec, REST API in relacijska baza:** jasna razdelitev odgovornosti, vendar zahteva usklajevanje API-ja.
- **Monolitna strežniško izrisana aplikacija:** enostavnejša začetna postavitev, vendar manj ločen razvoj interaktivnega odjemalca.
- **Mikrostoritve:** ločeno razmeščanje komponent, vendar nepotrebna operativna zahtevnost za MVP.

## Sprejeta odločitev in utemeljitev

Izberemo **tro-nivojsko arhitekturo**: React SPA, Go REST API in PostgreSQL. Za MVP to prinaša jasno razmejitev med prikazom, poslovno logiko in podatki brez dodatnih stroškov upravljanja mikrostoritev.

## Posledice

### Pozitivne

- Odjemalec in API lahko razvijamo ter testiramo ločeno.
- Poslovna pravila za rezervacije ostanejo centralizirana na strežniku.
- Ena relacijska baza olajša doslednost podatkov.

### Negativne

- Potrebno je načrtovati in verzionirati API pogodbe.
- Razvojno okolje potrebuje usklajeno delovanje več komponent.

## Povezave

- [Arhitektura in tehnološki sklad](../04-arhitektura-in-tehnoloski-sklad.md)
