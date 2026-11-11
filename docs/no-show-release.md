# Samodejno sproščanje rezervacij (DB-063–DB-067)

Backend ima en Go worker, ki ob zagonu in nato vsako minuto pregleda
rezervacije brez potrjene prisotnosti. Posamezen pregled ima največ
30 sekund in se ne prekriva s prejšnjim. Napaka ne ustavi naslednjih
pregledov. Ob ustavitvi strežnika se worker prekliče in zaključi,
preden se zapre povezovalni bazen PostgreSQL.

Pred zagonom izvedi `go run ./cmd/migrate up`. Migracija 11 doda
status `EXPIRED` in indeks za zapadle rezervacije.

Worker z enim atomskim SQL stavkom spremeni največ 500 rezervacij na
pregled. Veljajo samo `CONFIRMED`, `checked_in_at IS NULL` in začetek
strogo pred časom PostgreSQL minus 15 minut. Pri točno 15 minutah je
check-in še dovoljen. `CHECKED_IN`, `CANCELLED`, `COMPLETED` in `PENDING`
ostanejo nespremenjene. Zaklenjene vrstice preskoči do naslednjega pregleda.

Več procesov lahko izvaja worker hkrati: `FOR UPDATE SKIP LOCKED`
prepreči dvojno obdelavo. Ponovitev že obdelane rezervacije ne spremeni.
Napaka povrne celoten SQL stavek. Če je več kot 500 zapadlih rezervacij,
se preostale obdelajo v naslednjih minutnih pregledih.

Pri običajni obremenitvi se rezervacija sprosti ob prvem pregledu po
izteku 15-minutnega okna, torej približno v naslednji minuti. Po ponovnem
zagonu se stare zapadle rezervacije preverijo takoj.


## Ponovna rezervacija (DB-065)

`EXPIRED` ne blokira iskanja, neposrednega preverjanja razpoložljivosti
ali nove rezervacije v prekrivajočem se terminu. Iskanje in preverjanje
uporabljata isti seznam blokirajočih stanj: `PENDING`, `CONFIRMED` in
`CHECKED_IN`, usklajen z omejitvijo baze. Po novi rezervaciji vir
ponovno postane zaseden. Prejšnja rezervacija ostane `EXPIRED` v zgodovini.


## Revizijska sled (DB-066)

Migracija 12 ustvari `public.reservation_audit_log`. Vsak samodejni
prehod zapiše ID rezervacije, dogodek `NO_SHOW_EXPIRED`, prejšnje stanje
`CONFIRMED`, novo stanje `EXPIRED`, čas PostgreSQL in sporočilo
**Rezervacija potekla zaradi odsotnosti check-in-a**. PIN in osebni
podatki niso del revizijskega zapisa.

Sprememba stanja in zapis dogodka sta del istega atomskega SQL stavka.
Če zapis v revizijsko tabelo ne uspe, se stanje ne spremeni. Unikatna
omejitev nad rezervacijo in tipom dogodka prepreči podvojeno sled.
Migracija down odstrani revizijsko tabelo skupaj z revizijsko zgodovino,
rezervacije pa ohrani.

Primer pregleda (samo ID namenske testne rezervacije):

```sql
SELECT reservation_id,event_type,previous_status,new_status,occurred_at,message
FROM public.reservation_audit_log
WHERE reservation_id = 'UUID_TESTNE_REZERVACIJE'
ORDER BY occurred_at;
```


## Avtomatizirani preizkusi (DB-067)

Preizkuse zaženi ročno z ločenim PostgreSQL 17 in `go test -race`
po [navodilih](ci.md).
Poleg check-in-a preveri večkratne preglede, osem sočasnih workerjev,
en sam revizijski dogodek, mejo točno 15 minut in mikrosekundo za njo.
Preizkusi z nadzorovanimi transakcijskimi zaklepi preverijo oba vrstna
reda check-in-a in poteka brez čakanja na resnično časovno mejo.
Namerno zavrnjen audit INSERT mora povrniti spremembo statusa;
ponovni pregled po odstranitvi napake mora uspeti natanko enkrat.

## Skupni lokalni preizkus

1. Posodobi development, ustavi backend, izvedi `go run ./cmd/migrate up`
   in `go run ./cmd/migrate status`. Pričakovana verzija je **12**,
   `dirty: false`. Ponovno zaženi `go run ./cmd/server`.
2. Prijavi se s svojim testnim računom in prek `GET /api/auth/me`
   pridobi svoj uporabniški ID.
3. Iz korena repozitorija ustvari namenski zamujeni testni primer:

```bash
docker compose exec -T postgres psql -X -v ON_ERROR_STOP=1 \
  -U deskbooker -d deskbooker -v user_id=TVOJ_UPORABNISKI_UUID \
  < backend/scripts/no-show-demo.sql
```

Skript ustvari nov, posebej poimenovan vir in novo rezervacijo za
prijavljenega testnega uporabnika. Začetek postavi 16 minut nazaj,
zato ni treba čakati 15 minut. Namenjen je lokalni razvojni bazi.
Drugih rezervacij ne spremeni. Izpiše `resource_id` in `reservation_id`.

4. V naslednji minuti mora worker izpisati `processed 1 reservations`
   (ali več, če so zapadle tudi druge rezervacije). Po ponovnem zagonu
   strežnika se pregled izvede takoj.
5. Zaščiten `GET /api/reservations/<id>` mora pokazati `EXPIRED` brez
   `checked_in_at`. Stran `/reservations/<id>/check-in` po osvežitvi
   pokaže ta status in ne omogoča potrditve prihoda.
6. Preveri en dogodek v `reservation_audit_log`. Po naslednjem pregledu
   mora ostati natanko en dogodek in stanje `EXPIRED`.
7. V iskanju rezerviraj ustvarjeni vir za termin v prihodnosti znotraj
   naslednje ure. Nova rezervacija mora uspeti, stara ostati `EXPIRED`.
8. Predhodna rezervacija s potrjeno prisotnostjo mora ostati
   `CHECKED_IN` in še naprej blokirati svoj termin.

Skript ne spreminja sistemske ure, ne dodaja testnih API obvozov in
ne izbriše svojih testnih zapisov. Namenski testni vir lahko po
preizkusu deaktiviraš v upravljanju virov.
