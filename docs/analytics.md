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
