# Administracija uporabnikov (DB-099–DB-103)

GET `/api/admin/users?limit=25&offset=0` je dovoljen samo trenutno
aktivnemu administratorju. Vloga manager (OFFICE_MANAGER) ni dovolj.
Brez veljavnega JWT velja 401, brez pravic 403, napačna paginacija 400.

Odgovor vsebuje users, total, limit in offset. Privzeta velikost strani
je 25, največ 100. Vrstni red je po e-pošti in ID-ju. Podatki vključujejo
ime, priimek, e-pošto, oddelek, vlogo, stanje in datum nastanka. Geslo in
njegov hash se nikoli ne prebereta v administratorski seznam.

Spremembe sheme niso potrebne; trenutna migracija ostaja 13.


## Zamenjava vloge (DB-100)

PATCH `/api/admin/users/:id/role` sprejme samo JSON z `role`.

| Vloga v zgodbi in zahtevku | Shranjena/vrnjena vloga |
| --- | --- |
| EMPLOYEE | employee |
| OFFICE_MANAGER | manager |
| ADMIN | admin |

Za združljivost so dovoljene tudi tri shranjene vrednosti. Neznane
vloge, neveljavni UUID, dodatna polja in dodatni JSON objekti so zavrnjeni.
Uspeh vrne HTTP 200 in varen `user`, neobstoječ uporabnik 404.

V eni transakciji se zakleneta administrator in ciljni račun v
stabilnem vrstnem redu ID-jev. Pravice administratorja se ponovno
preverijo po zaklepu, zato stara preveritev middleware ne dovoljuje
spremembe po sočasnem odvzemu vloge. Lastna vloga je zaščitena že ob
prvi uvedbi pisalnega endpointa. Gesla niso del tega API-ja.
