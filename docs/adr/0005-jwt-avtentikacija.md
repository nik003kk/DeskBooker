# ADR 0005 – JWT za avtentikacijo in preverjanje vlog

- **Status:** Sprejeto
- **Datum:** 2026-10-27

## Kontekst in problem

Aplikacija ima tri vloge: EMPLOYEE, OFFICE_MANAGER in ADMIN. Po prijavi mora API varno prepoznati uporabnika in pri vsakem zaščitenem zahtevku preveriti dostop do ustreznih funkcij.

## Obravnavane alternative

- **JWT žetoni:** API lahko preveri podpis in čas veljavnosti brez branja strežniške seje za vsak zahtevek; zahtevnejši preklic žetonov.
- **Strežniške seje s piškotki:** enostavnejši centralni preklic, vendar zahtevajo shrambo sej.
- **Zunanji ponudnik identitete:** razbremeni implementacijo prijave, vendar vnaša odvisnost od zunanje storitve.

## Sprejeta odločitev in utemeljitev

Izberemo **JWT za avtentikacijo API-ja**. Strežnik po uspešni prijavi izda kratkoživeč podpisan žeton; ob zahtevkih preverja njegov podpis in veljavnost ter vedno uveljavlja pravice dostopa na backendu. Končni način varnega prenosa/shranjevanja žetona (npr. varen piškotek ali glava Authorization) se določi pri izvedbi.

## Posledice

### Pozitivne

- Jasna ločitev prijave in API zahtevkov.
- Primerno za ločen React odjemalec in REST API.
- Centralizirano preverjanje vlog na strežniku.

### Negativne

- Poskrbeti je treba za zaščito pred krajo žetonov in pravilno hrambo ključev.
- Preklic izdanega žetona zahteva dodatno strategijo.
- Podpisan žeton sam po sebi še ne zagotavlja varnega sistema.

## Povezave

- [Arhitektura in tehnološki sklad](../04-arhitektura-in-tehnoloski-sklad.md)
