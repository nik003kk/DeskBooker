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
