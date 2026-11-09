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


## Zaščita lastnega računa (DB-101)

Lastne vloge ni mogoče nastaviti ali spremeniti, tudi če je nova vloga
enaka ali je UUID zapisan z velikimi črkami. Vrne se 403
`self_role_protected`. Identiteta administratorja izhaja iz preverjenega
JWT; polj actor_id, user_id, status ali email v zahtevku ni dovoljeno dodati.

DELETE `/api/admin/users/:id` pri lastnem računu vrne 403
`self_account_protected`; ostalo brisanje je onemogočeno in vrne 403
`account_deletion_disabled`. Splošno brisanje računov ni del MVP in ne
obstaja pisalna operacija za brisanje. Tudi profilna pot ne podpira DELETE.

Sočasni preizkus dveh administratorjev, ki drug drugemu odvzemata vlogo,
mora dovoliti samo en prehod in ohraniti enega aktivnega administratorja
iz tega para. Razlog je ponovni pregled vloge po zaklepu obeh vrstic.


## Stran Uporabniki (DB-102)

Aktiven administrator odpre `/admin/users` prek povezave Uporabniki.
Manager in employee te povezave ne vidita, neposreden obisk pa je
zavrnjen pred nalaganjem seznama. Tabela vsebuje paginacijo, osveževanje,
varne podatke računa, oznake stanja in izbiro vloge. Lastna vrstica je
zaklenjena in označena kot Tvoj račun. Brisanja ni v uporabniškem vmesniku.

Pred shranjevanjem modal pokaže ime, e-pošto ter staro in novo vlogo.
Ob napaki se uspeh ne prikaže; tabela se posodobi šele po potrditvi API-ja.
Podvojeni kliki ne pošljejo več mutacij. Ob izgubi administratorskih pravic
se podatki odstranijo in prikaže napaka. Neuspešno branje je mogoče ponoviti.
