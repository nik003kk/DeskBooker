# Dejanski viri v kategorijah (DB-116–DB-118)

Kategorije uporabljajo isto API iskanje, kartice in potrditev rezervacije
kot stran Iskanje virov. API se bere z veljavnim žetonom prijavljenega
uporabnika. Ni potrebna nova migracija; trenutna verzija ostaja 13.

| Zavihek | Pot | Tip v GET /api/resources |
| --- | --- | --- |
| Delovna mesta | /desks | DESK |
| Sejne sobe | /rooms | MEETING_ROOM |
| Parkirišča | /parking | PARKING_SPOT |

Ob prvem obisku se pokažejo aktivni viri te vrste z dejanskim imenom,
nadstropjem, kapaciteto in opremo. To je pregled virov, ne trditev, da
so prosti v poljubnem terminu. Za rezervacijo izberi datum, začetek in
konec ter klikni Poišči vire. API nato izloči zasedene in neaktivne vire
ter upošteva vse izbrane lastnosti. Ponastavitev ohrani vrsto kategorije.

Gumb Rezerviraj prenese dejanski UUID vira in uporabljeni termin na
obstoječo stran potrditve. Po spremembi še neuporabljenih filtrov je
gumb onemogočen do ponovnega iskanja. Prikazani uporabljeni termin
zato ne more biti tiho nadomeščen z nedokončanim vnosom v obrazcu.
Menjava kategorije odpre njen lasten prazen nabor filtrov.

Splošna stran /search še naprej omogoča iskanje po vseh treh vrstah.

## Osveževanje in stanja (DB-117)

Gumb Osveži vire in vrnitev na zavihek ponovno prebereta podatke iz API-ja.
Dodan ali urejen vir se tako prikaže brez ponovnega zagona aplikacije,
deaktiviran pa izgine. Uporabljeni tip, termin in oprema se pri osvežitvi
in ponovnem poskusu ohranijo. Osvežitev ne potrdi še nedokončanih sprememb
v obrazcu: če so filtri spremenjeni, ostane rezervacija onemogočena do
klika Poišči vire. Med nalaganjem in ob napaki rezervacija ni mogoča.

Pri praznem rezultatu se pokaže ustrezno prazno stanje, pri napaki gumb
Poskusi znova. Prikaz se nikoli ne zapolni s predstavitvenimi podatki.
Predhodne zahteve se prekinejo; njihov pozen odgovor ne prepiše novega
izbora ali druge kategorije.

## Preizkusi (DB-118)

Frontend DOM preizkusi naložijo dejanske poti aplikacije. Preverijo vse
tri kategorije, izbrani tip, kartice, ponastavitev, osveževanje, napake,
pozne odgovore stare kategorije ter prehod do uspešne rezervacije s
pravilnim UUID, terminom v Europe/Ljubljana in PIN-om. Preverjeno je tudi,
da še neuporabljena sprememba termina ostane nepotrjena po osvežitvi.

PostgreSQL integracijski preizkus uporabi pravi administratorski API za
ustvarjanje, urejanje in deaktivacijo vseh treh vrst. Iskanje bere z JWT
zaposlenega, preveri opremo in spremenjena polja, premik v drugo kategorijo
po spremembi tipa ter izločanje zasedenega vira po dejanskem vnosu rezervacije.
Podatki so namenski testni primeri v ločeni testni bazi;
[navodila za ročni zagon](ci.md).

## Skupni lokalni preizkus

Iz korena DeskBooker posodobi development in ponovno zaženi frontend.
Če backend že teče na zadnji različici prejšnjega sklopa, ni potrebna
sprememba sheme; migracija ostaja **13, dirty: false**.

```bash
git switch development
git pull --ff-only origin development
cd frontend
npm ci
npm run dev
```

1. Kot admin dodaj namensko mizo, sejno sobo in parkirno mesto z
   različnimi imeni, nadstropji, kapacitetami in opremo.
2. Kot prijavljen zaposleni odpri Delovna mesta, Sejne sobe in Parkirišča.
   Vsak nov vir se mora pojaviti samo v ustrezni kategoriji s pravimi podatki.
3. Kot admin spremeni ime ali opremo vira. Na strani zaposlenega klikni
   Osveži vire ali se vrni na zavihek; sprememba se mora prikazati.
4. Deaktiviraj namenski vir in osveži seznam. Ne sme se več prikazati.
   Ponovna aktivacija ga vrne med aktivne vire.
5. Izberi prihodnji datum in čas, klikni Poišči vire ter rezerviraj prost
   namenski vir. Potrditev mora pokazati pravilen vir in uporabljeni termin.
6. Po novi rezervaciji isto kategorijo ponovno preišči za isti termin.
   Zasedeni vir mora izginiti iz razpoložljivih, v pregledu brez termina
   pa ostati viden kot aktivni vir.
7. Spremeni datum/čas brez novega iskanja: Rezerviraj mora ostati
   onemogočen tudi po osvežitvi, dokler ne klikneš Poišči vire.
8. Ponastavi filtre in preklopi kategorijo. Vrsta mora ostati pravilna;
   generično Iskanje virov mora še vedno omogočati vse vrste.
9. Če parkirišč še ni, pričakuj prazno stanje. Brez prijave so te tri
   strani zaščitene. Pri nedosegljivem backendu pokažejo napako in ponovni
   poskus, nikoli starih predstavitvenih kartic.
