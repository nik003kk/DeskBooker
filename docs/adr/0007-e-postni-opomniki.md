# 0007 – Uporaba listmonk za e-poštne opomnike

- **Status:** Sprejeto
- **Datum:** 2026-10-27
- **Odločevalci:** Razvojna ekipa DeskBooker

## Kontekst in problem

DeskBooker mora zaposlenemu približno 10 minut pred začetkom rezervacije poslati e-poštni opomnik z datumom, časom, nazivom rezerviranega vira, 4-mestno PIN kodo in povezavo do check-in strani (US-14). Če je bila rezervacija preklicana, opomnika ne sme poslati. Backend je zasnovan v jeziku Go, uporablja PostgreSQL in že predvideva periodične naloge.

Želimo odprtokodno rešitev z dobro skupnostno podporo, ki jo lahko gostimo sami in povezujemo z Go aplikacijo prek API-ja. Razvoj lastnega sistema za pošiljanje e-pošte, upravljanje predlog in dostave bi povečal obseg MVP.

## Obravnavane alternative

1. **listmonk (odprtokodna, samostojno gostovana storitev):** Go aplikacija z REST API-jem za transakcijska sporočila, predlogami in PostgreSQL. Je zelo priljubljena v kategoriji odprtokodnih orodij za poštne sezname (približno 24.000 GitHub zvezdic oktobra 2026). Za dejansko dostavo potrebuje nastavljen SMTP strežnik oziroma SMTP relay.
2. **Postal (odprtokodna platforma za e-poštno dostavo):** Omogoča pošiljanje prek SMTP/API-ja in več nadzora nad dostavno infrastrukturo; zahteva več operativne konfiguracije in vzdrževanja.
3. **Neposredno pošiljanje SMTP iz Go backenda:** Manj dodatnih komponent, vendar je potrebno v DeskBookerju samostojno implementirati predloge, beleženje napak in ponovne poskuse.
4. **Komercialna storitev (npr. SendGrid ali Mailgun):** Enostavnejša upravljana dostava, vendar lahko uvede stroške in odvisnost od zunanjega ponudnika.

## Sprejeta odločitev

Za pripravo in pošiljanje **transakcijskih e-poštnih opomnikov** izberemo **listmonk**, ki ga gostimo kot ločeno storitev. Go backend se z njim povezuje prek zaščitenega HTTP API-ja `POST /api/tx`, z vnaprej določeno predlogo sporočila.

Periodični proces v DeskBookerju preveri prihajajoče rezervacije in sproži opomnik 10 minut pred začetkom samo za rezervacije, ki so še v statusu `CONFIRMED`. Podatki v sporočilu obsegajo datum, čas, naziv vira, PIN in povezavo do check-in strani. Vsaka rezervacija mora imeti zabeležen status oziroma čas poslanega opomnika, da ponovno izvajanje opravila ne povzroči podvojenih sporočil.

Za dejansko dostavo listmonk povežemo z **SMTP relayem**, katerega konkretna konfiguracija bo določena ob uvedbi. listmonk sam ni nadomestilo za infrastrukturo, ki e-pošto preda prejemnikovemu strežniku. Dostopni žetoni in SMTP poverilnice se hranijo v okoljskih spremenljivkah oziroma sistemu za skrivnosti, ne v Git repozitoriju.

### Razlogi za odločitev

- listmonk je odprtokoden (AGPLv3), aktiven in dobro uveljavljen projekt;
- njegov transakcijski API omogoča integracijo z Go brez dodatnega SDK-ja;
- že uporablja PostgreSQL, ki je del tehnološkega sklada DeskBooker;
- predloge sporočil se upravljajo ločeno od rezervacijske poslovne logike.

Popularnost je pomožen kriterij, ne dokaz, da je storitev optimalna za vsako aplikacijo; pred odločitvijo smo upoštevali tudi obseg in potrebo po SMTP infrastrukturi.

## Pozitivne posledice

- Ni treba implementirati lastnega sistema za oblikovanje e-poštnih predlog.
- DeskBooker ohrani ločeno poslovno logiko rezervacij in infrastrukturo za sporočila.
- Integracija temelji na dokumentiranem HTTP API-ju.
- Podatki in storitev lahko ostanejo na infrastrukturi, ki jo upravljamo sami.

## Negativne posledice in tveganja

- Dodatna storitev pomeni več zahtev pri nameščanju, nadzoru in posodabljanju.
- listmonk še vedno potrebuje SMTP relay ter pravilno nastavitev dostavljivosti e-pošte.
- Nedosegljivost listmonk ali SMTP strežnika lahko zamakne opomnike; potrebujemo beleženje napak in omejene ponovne poskuse.
- AGPLv3 zahteva preverjanje licenčnih obveznosti pri načinu uporabe in morebitnih spremembah programske opreme.
- Za samo nekaj opomnikov na dan bi bil lahko preprost neposredni SMTP odjemalec manj kompleksen.

## Povezane funkcionalnosti

- US-14: e-poštni opomnik 10 minut pred začetkom rezervacije.
- US-07: check-in s PIN kodo.
- US-08: periodična obdelava rezervacij (ločena od logike pošiljanja opomnikov).

## Viri

- listmonk – uradna stran: https://listmonk.app/
- listmonk – API za transakcijska sporočila: https://listmonk.app/docs/apis/transactional/
- listmonk – dokumentacija: https://listmonk.app/docs/
- Postal – odprtokodna alternativa: https://github.com/postalserver/postal
- Pregled priljubljenosti odprtokodnih e-poštnih orodij (oktober 2026): https://appsgit.com/best/open-source-email-marketing
