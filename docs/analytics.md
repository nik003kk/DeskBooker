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


## No-show (DB-070)

`no_show.rate` = EXPIRED / (EXPIRED + CHECKED_IN + COMPLETED) × 100.
Potrjene rezervacije z neznanim izidom, PENDING in CANCELLED so izključene.
Prehod uspešne rezervacije iz CHECKED_IN v COMPLETED zato ne spremeni
razmerja. Vsi seštevki uporabljajo isto obdobje po začetku rezervacije.
Brez znanih izidov je delež null, ne 0 %. API vrne tudi števec `expired`,
`attended` in imenovalec `eligible`, da je izračun preverljiv.


## Stran Analitika (DB-071)

Aktivni vodja ali skrbnik odpre `/analytics` prek povezave Analitika.
Kartice pokažejo zasedenost, vse rezervacije, aktivne rezervacije in no-show
z definicijami ter časom zadnjega podatkovnega posnetka. Stran vključuje
nalaganje, prazno stanje, sporočilo o napaki in ponovni poskus. Podatki
se ob zavrnitvi dostopa odstranijo, prekinjene zahteve ne prepišejo novejših.


## Grafa (DB-072)

Stolpčni graf prikaže pet tehtanih dnevnih deležev iz `weekdays`, krožni
pa vseh šest statusov iz `status_counts`. Krožni deleži uporabljajo vse
rezervacije, zato njihov imenovalec ni isti kot pri KPI no-show. Prazni
podatki ne ustvarijo navideznih deležev. Graf ima oznake, legendo in
natančne vrednosti; pri zasedenosti je na voljo tudi podatkovna tabela.


## Obdobja in primerjava (DB-073)

Stran podpira zadnjih 7 dni, zadnjih 30 dni in poljubni vključujoči
obdobji do 366 dni. Primerjalno obdobje je neposredno pred izbranim
in ima enako število koledarskih dni. Pri deležih je razlika prikazana
v odstotnih točkah, pri številu vseh rezervacij kot relativna sprememba;
ničelna osnova in neznani deleži so posebej označeni. Trenutno aktivnih
rezervacij ne primerjamo z zgodovinskim stanjem, ki ga baza ne hrani.
Gumb Osveži podatke ponovno pridobi obe obdobji. Prekinjene in pozne
zahteve prejšnjega izbora ne smejo prepisati novega prikaza.


## CSV (DB-074)

`GET /api/analytics/export.csv?from=YYYY-MM-DD&to=YYYY-MM-DD` uporablja
isto validacijo obdobja, pravice in izračun kot JSON. Gumb **Izvozi CSV**
izvozi trenutno uporabljeno obdobje. Izvoz med spremembo izbora se prekine.
Poročilo vsebuje metapodatke obdobja in časa izračuna, KPI, vseh šest
statusov, dnevne in tedenske podatke. Ne vsebuje osebnih podatkov ali PIN.

Stolpci so `section,metric,date,weekday,value,unit`. Enote so count,
minutes, percent ali text. Neznani deleži so prazni, decimalno ločilo je
pika, kodiranje UTF-8 z BOM za preglednice. JSON in CSV se izračunata ob
posamezni zahtevi; če se rezervacije med njima spremenijo, se razlikujeta
čas izračuna in podatki. Napake ostanejo JSON in niso ponujene kot CSV.

## Pravice in skupni preizkus (DB-075)

`manager` v bazi pomeni OFFICE_MANAGER iz uporabniške zgodbe, `admin`
pa ADMIN. Na vsaki JSON in CSV zahtevi se ponovno preverita vloga in
stanje računa v bazi, ne v zastarelem žetonu. Brez veljavne prijave velja
401, zaposlenim ali neaktivnim računom 403. Tudi neposreden obisk
`/analytics` v aplikaciji je zaščiten. API ostaja avtoritativen.

Po posodobitvi iz korena projekta:

```bash
git switch development
git pull --ff-only origin development
cd backend
go run ./cmd/migrate up
go run ./cmd/migrate status
go run ./cmd/server
```

Pričakuj **13, dirty: false**. V drugem terminalu iz korena projekta:

```bash
cd frontend
npm ci
npm run dev
```

1. Prijavi se kot aktivni admin ali manager in odpri **Analitika**.
2. Preveri KPI, grafa in tabelo podatkov. Preklopi 7 dni, 30 dni in
   poljubno obdobje ter klikni Osveži podatke. Primerjava se mora
   nanašati na neposredno predhodno enako dolgo obdobje.
3. Izvozi CSV. Preveri datuma in čas izračuna, skupno število rezervacij,
   EXPIRED ter dnevne deleže glede na API in prikaz. Ničelna osnova ni
   neznan delež; neznani deleži so v prikazu —, v CSV pa prazni.
4. Za obdobje brez rezervacij pričakuj ničelne seštevke in jasno prazno
   stanje, brez navideznih rezervacij ali no-show odstotka.
5. Z računom employee preveri, da povezave ni in neposreden obisk
   `/analytics` ne prikaže podatkov. JSON in CSV API morata vrniti 403.
6. Brez žetona oba API endpointa vrneta 401; napačen datum s pravilnimi
   pravicami vrne 400.

Primer s svežim TOKEN:

```bash
curl -i "http://localhost:8080/api/analytics?from=2026-11-01&to=2026-11-09" \
  -H "Authorization: Bearer $TOKEN"
```

Deleže lahko neodvisno preveriš prek status_counts: EXPIRED deliš z
vsoto EXPIRED, CHECKED_IN in COMPLETED. Dnevni delež je booked_minutes /
capacity_minutes × 100. CSV vključuje obe osnovi, da je preverjanje mogoče.
