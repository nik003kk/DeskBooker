# Analitika (DB-068–DB-075)

Po posodobitvi izvedi migracije do **13**, `dirty: false`, in ponovno
zaženi backend. `GET /api/analytics?from=2026-11-01&to=2026-11-09`
vrne zbirne podatke brez osebnih podatkov in PIN kod. API od začetka
zahteva veljaven JWT ter trenutno aktivno vlogo `manager` ali `admin`.

Datuma sta vključujoča v Europe/Ljubljana. Privzeto obdobje je zadnjih
7 koledarskih dni vključno z današnjim, največ 366 dni na zahtevo.
Rezervacija šteje v skupni seštevek po svojem začetku, statusi pa po
trenutnem stanju. Aktivne imajo status CONFIRMED ali CHECKED_IN in
konec v prihodnosti glede na čas PostgreSQL. Pretekle, še neobdelane
CONFIRMED niso aktivne. Vsi statusi se vrnejo tudi pri ničelnem številu.
Izračuni ene zahteve uporabljajo isti posnetek baze. Odgovori imajo
Cache-Control: no-store; napačni datumi vrnejo 400, napaka baze 500.


## Zasedenost (DB-069)

Zasedenost je rezervirani čas / razpoložljivi čas trenutno aktivnih
virov med **09:00 in 17:00**, od ponedeljka do petka, v Europe/Ljubljana.
Vsak vir je ena enota; kapaciteta sejne sobe ni število samostojno
rezervabilnih mest. Upoštevajo se CONFIRMED, CHECKED_IN in COMPLETED.
PENDING, CANCELLED in EXPIRED ne prispevajo k rezerviranemu času.
Rezervacije čez polnoč se razdelijo po dnevih; čas izven 09–17 se odreže.
Prekrivajoča zaključena zgodovina se združi po viru, da delež ne preseže 100 %.

`daily` vsebuje delovne dneve obdobja, `weekdays` pet tehtanih povprečij
od ponedeljka do petka, `occupancy_rate` pa tehtani delež celotnega obdobja.
Deleži so odstotki, zaokroženi na dve decimalki. Brez kapacitete je delež
null, pri razpoložljivi kapaciteti brez rezervacij pa 0.

Kapaciteta upošteva datum nastanka vira, tudi del dneva, če je nastal
med 09–17. Zgodovina aktivacij in deaktivacij se ne shranjuje, zato je
obseg kapacitete trenutni aktivni nabor, ne rekonstrukcija preteklega
števila virov. Skupni statusni seštevki vključujejo vse vire.
