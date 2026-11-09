# Administracija uporabnikov (DB-099–DB-103)

GET `/api/admin/users?limit=25&offset=0` je dovoljen samo trenutno
aktivnemu administratorju. Vloga manager (OFFICE_MANAGER) ni dovolj.
Brez veljavnega JWT velja 401, brez pravic 403, napačna paginacija 400.

Odgovor vsebuje users, total, limit in offset. Privzeta velikost strani
je 25, največ 100. Vrstni red je po e-pošti in ID-ju. Podatki vključujejo
ime, priimek, e-pošto, oddelek, vlogo, stanje in datum nastanka. Geslo in
njegov hash se nikoli ne prebereta v administratorski seznam.

Spremembe sheme niso potrebne; trenutna migracija ostaja 13.
