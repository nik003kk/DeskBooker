# ADR 0002 – React kot enostranska spletna aplikacija

- **Status:** Sprejeto
- **Datum:** 2026-10-27

## Kontekst in problem

Zaposleni potrebujejo interaktivno filtriranje virov, pregled rezervacij in check-in brez nepotrebnih ponovnih nalaganj strani. Rešitev je predvidena kot odzivna spletna aplikacija, ne kot nativna mobilna aplikacija.

## Obravnavane alternative

- **React SPA:** komponentni pristop in dinamičen prikaz stanja, a zahteva odjemalsko upravljanje stanja.
- **Vue SPA:** podoben koncept z drugo sintakso in ekosistemom.
- **Strežniško izrisane HTML strani:** manj odjemalske kompleksnosti, a več usklajevanja pri interaktivnih tokovih.

## Sprejeta odločitev in utemeljitev

Izberemo **React SPA**, ker omogoča ponovno uporabo komponent za filtre, kartice virov, modalno okno za check-in in upravljavske poglede. Komponente pridobivajo podatke prek ločenega REST API-ja (ADR 0001).

## Posledice

### Pozitivne

- Ponovno uporabne komponente za zaposlene in upravitelje.
- Dinamično posodabljanje uporabniškega vmesnika.
- En spletni vmesnik za namizne in mobilne brskalnike.

### Negativne

- Potrebni so odjemalsko usmerjanje, obravnava napak in stanje prijave.
- Dodatna orodja za gradnjo in testiranje frontend dela.

## Povezave

- [Arhitektura in tehnološki sklad](../04-arhitektura-in-tehnoloski-sklad.md)
