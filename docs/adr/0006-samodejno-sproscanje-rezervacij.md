# ADR 0006 – Periodično sproščanje nepotrjenih rezervacij

- **Status:** Sprejeto
- **Datum:** 2026-10-27

## Kontekst in problem

DeskBooker mora omejiti »ghost bookings«. Če rezervacija ni potrjena s check-in v 15 minutah po začetku, mora vir postati ponovno razpoložljiv. MVP potrebuje preprosto avtomatizacijo brez ločene infrastrukture.

## Obravnavane alternative

- **Periodično opravilo v Go procesu (goroutine / cron):** malo infrastrukture, vendar odvisno od delovanja procesa.
- **Ločena vrsta opravil in worker:** zanesljivejša porazdeljena obdelava, a več komponent.
- **Sproščanje samo ob branju razpoložljivosti:** brez ozadnega opravila, vendar statusi in revizijska sled lahko zaostajajo.

## Sprejeta odločitev in utemeljitev

Izberemo **periodično opravilo znotraj Go strežnika**, ki približno vsako minuto poišče nepotrjene rezervacije po 15-minutnem roku in jim transakcijsko spremeni status iz `CONFIRMED` v `EXPIRED`. Operacija mora biti idempotentna; pri več strežniških instancah bo potrebno usklajevanje izvajanja.

## Posledice

### Pozitivne

- Ni potrebe po dodatnem strežniku za vrsto opravil v MVP.
- Vir je samodejno sproščen brez ročnega posega.
- Preverjanje je preprosto testirati na rezervacijah z različnimi roki.

### Negativne

- Sprostitev ni nujno izvedena natanko ob izteku 15. minute.
- Ob zaustavljenem strežniku opravilo ne teče.
- Več instanc zahteva previdnost pri sočasni obdelavi.

## Povezave

- [Arhitektura in tehnološki sklad](../04-arhitektura-in-tehnoloski-sklad.md)
