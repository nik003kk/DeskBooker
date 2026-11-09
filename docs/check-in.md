# Check-in (DB-057–DB-062)

Po posodobitvi na `development` iz mape `backend` izvedi:

```bash
go run ./cmd/migrate up
go run ./cmd/migrate status
go run ./cmd/server
```

Pričakovana migracija: **10**, `dirty: false`.

`POST /api/reservations/:id/check-in` zahteva JWT lastnika rezervacije,
`Content-Type: application/json` in `{"pin":"0042"}`. PIN je niz,
zato vodilne ničle ostanejo ohranjene. Uspeh vrne HTTP 200 in
`reservation` s statusom `CHECKED_IN` ter časom `checked_in_at`.
PIN se ne vrne. Ponovitev s pravilnim PIN vrne prvotni uspeh in čas.

Check-in je dovoljen od 10 minut pred začetkom do vključno 15 minut
po začetku, vendar vedno pred koncem rezervacije. Avtoritativen je
čas PostgreSQL po pridobitvi zaklepa rezervacije. Nepravilen PIN ne
spreminja rezervacije. Preklicane in zaključene rezervacije so zavrnjene.

| Odgovor | Pomen |
| --- | --- |
| 400 `invalid_pin` | PIN ni niz štirih ASCII števk |
| 401 | Neveljavna ali manjkajoča prijava |
| 404 | Neobstoječa ali tuja rezervacija |
| 409 `check_in_window` | Prezgodaj, prepozno ali rezervacija že končana |
| 409 `check_in_status` | Status ne dopušča check-in-a |
| 422 `incorrect_pin` | Napačna PIN koda |

`GET /api/reservations/:id` vrne lastniku podatke rezervacije in vira
brez PIN kode. Omogoča osveževanje strani check-in-a; seznam vseh
rezervacij in preklic sta ločeni zahtevi DB-076–DB-079.

`CHECKED_IN` blokira prekrivajoče termine v omejitvi PostgreSQL,
iskanju virov in preverjanju razpoložljivosti. Migracija down vrne
te rezervacije v `CONFIRMED` in odstrani čas check-in-a.

Integracijski test uporablja **prazno namensko testno bazo** prek
`CHECK_IN_TEST_DATABASE_URL`. Sam izvede migracije in preveri lastništvo,
napačen PIN, sočasni check-in, nespremenjen čas pri ponovitvi ter
izločanje zasedenega vira iz iskanja in novih rezervacij. GitHub Actions
za to ustvari ločen PostgreSQL 17 servis.

## Uporabniški vmesnik

Na uspešnem zaslonu rezervacije klikni **Odpri check-in**. Stran
`/reservations/<id>/check-in` pridobi podatke iz API-ja in deluje tudi
po ponovni prijavi ali osvežitvi. Povezavo lahko shraniš med zaznamke.
PIN ostane samo v vnosu modalnega okna in se ne shrani v brskalnik ali URL.
Modal podpira štiri polja, lepljenje kode, premikanje s tipkovnico,
ponovni vnos po napaki in preprečevanje podvojenih zahtevkov.

Odštevalnik se posodablja vsako sekundo in po vrnitvi na zavihek.
Pred odprtjem pokaže preostanek do odprtja, med oknom čas do zaprtja,
po zaprtju pa onemogoči potrditev tudi v že odprtem modalnem oknu.
Ura brskalnika je informativna; API vedno ponovno preveri čas v bazi.
