# 02 - Funkcionalnosti in obseg

## Uporabniške vloge

### 1. Zaposleni / Član (Employee / Member)
* **Pravice:** 
- Pregled razpoložljivosti delovnih miz, sejnih sob in parkirnih mest
- filtriranje virov po parametrih opreme
- ustvarjanje, pregled in preklic lastnih rezervacij
- izvedba "check-in" prijave ob prihodu
- pregled javnega profila sodelavcev (kje sedijo določen dan)

* **Omejitve:** 
- Ne more dodajati, urejati ali brisati virov (miz/sob/parkirišč)
- ne more spreminjati parametrov tlorisa ali nastavitev sistema
- ne more urejati ali preklicovati rezervacij drugih uporabnikov.

### 2. Upravitelj pisarne (Office Manager)
* **Pravice:** 
- Dodajanje, urejanje in deaktivacija delovnih miz, sejnih sob in parkirnih mest
- definiranje opreme posamezne mize (npr. monitor, USB-C dock) 
- ročni pregled in preklic poljubne rezervacije; vpogled v analitiko in poročila zasedenosti pisarne.

* **Omejitve:** 
- Ne more upravljati s sistemskimi uporabniškimi računi drugih adminov in sistemskimi nastavitvami avtentikacije.

### 3. Administrator (Admin)
* **Pravice:** 
- Celoten nadzor nad sistemom: upravljanje uporabniških računov in dodeljevanje vlog (Zaposleni, Office Manager, Admin)
- nastavljanje sistemskih parametrov (npr. časovno okno za check-in, pravila odpovedi)
- vpogled v celotno revizijsko sled (audit log).
* **Omejitve:** Ne more ustvarjati rezervacij v imenu drugih uporabnikov brez njihove avtentikacije (zaradi zagotavljanja sledljivosti).

---

## Predvidene funkcionalnosti

1. **[MVP] Avtentikacija in profil uporabnika** – Registracija, prijava (z e-pošto in geslom ali JWT) in pregled lastnega profila z vlogo.
2. **[MVP] [Domenska] Dodajanje in upravljanje delovnih mest ter virov** – Upravljanje delovnih miz, sejnih sob in parkirišč s pripadajočimi lastnostmi (npr. številka mize, nadstropje, tip opreme).
3. **[MVP] [Domenska] Napredno filtriranje in iskanje po specifikacijah opreme** – Iskanje proste mize/sobe po datumu, časovnem reScope-u in specifični opremi (npr. "dva monitorja", "dvižna miza", "projektor").
4. **[MVP] [Domenska] Delovni potek rezervacije mesta/sobe (Workflow)** – Izbira vira, časovnega intervala ter potrditev rezervacije z preverjanjem preklapljanja terminov (double-booking protection).
5. **[MVP] [Domenska] Avtomatski proces potrditve prisotnosti (Check-in Workflow)** – Možnost potrditve prisotnosti na s PIN kodo znotraj 15-minutnega okna od začetka rezervacije.
6. **[MVP] [Domenska] Logika avtomatskega sproščanja nepotrjenih mest ("No-show auto-release")** – Sistemski proces, ki ob neizvedenem check-in-u po 15 minutah spremeni status rezervacije v `EXPIRED` in sprosti vir.
7. **[MVP] Nadzorna plošča in analitika zasedenosti pisarne** – Grafični prikaz najbolj zasedenih dni v tednu, izkoriščenosti posameznih nadstropij ter deleža "no-show" neizkoriščenih rezervacij za Office Managera.
8.  **Moji zahtevki in pregled zgodovine rezervacij** – Pregledno upravljanje aktivnih, preteklih in poteklih rezervacij z možnostjo predčasnega preklica.
9. **Iskanje mesta sodelavca ("Find my colleague")** – Iskanje, v katerem delu pisarne ali na kateri mizi bo določen dan sedel izbrani sodelavec.
10. **Koledarski pregled zasedenosti sejnih sob** – Dnevni in tedenski interaktivni koledarski prikaz zasedenosti skupnih prostorov.
11. **Rezervacija parkirnega mesta vezana na delovno mesto** – Možnost hkratne rezervacije parkirnega mesta ob uspešni rezervaciji delovne mize.
12. **Obvestila in opomniki (Email / In-app)** – Pošiljanje opomnika uporabniku 10 minut pred začetkom rezervacije s povezavo za check-in.


---

## Kasnejše nadgradnje
1. **Interaktivni grafični tloris pisarne (2D Floor Plan Canvas):** Vizualno izbiranje in rezervacija mize s klikom na interaktivno shemo nadstropja v realnem času.
2. **Integracija s koledarji (Google Calendar / Outlook):** Avtomatska sinhronizacija rezervacije sejne sobe ali mize v uporabnikov osebni koledar.
3. **QR koda na mizi za hitri Check-in:** Fizične QR kode na mizah, ki ob skeniranju s pametnim telefonom samodejno izvedejo check-in za prijavljenega uporabnika (odvisno, če lahko kasneje delamo integracijo z mobilnimi napravami - vseeno dostop kot web aplikacija na mobilni napravi).
4. **Pametna pravila in omejitve (Booking quotas):** Nastavitev največjega števila dni rezervacij na teden na posameznega zaposlenega za zagotavljanje enakomerne dostopnosti.
5. **Zamenjava in odstop rezervacije med sodelavci ("Seat Swap"):** Če zaposleni ugotovi, da ne potrebuje več svoje rezervirane mize ali sejne sobe, jo lahko z enim klikom da na voljo v interno "tržnico zamenjav", kjer jo lahko prevzame drug sodelavec ali ponudi zamenjavo za drug termin.

### Bonus nadgradnja k 4. zahtevi:
* **Pravila za pravično uporabo in kvote (Booking Quotas & Fair-use):** Nastavitev naprednih poslovnih pravil, npr. omejitev največjega števila dni rezervacij na teden na posameznega zaposlenega ali omejitev trajanja rezervacije sejne sobe (npr. največ 2 uri na dan) za zagotavljanje enakomerne dostopnosti virov vsem zaposlenim.

---

## Out of scope (Kaj projekt NE bo)
* **Sledenje lokaciji uporabnika v realnem času preko GPS ali Wi-Fi:** Sistem ne bo sledil premikanju zaposlenih po pisarni ali izven nje.
* **Plačilni modul za izposojo ali najem:** Znotraj aplikacije ne bo integriranega plačilnega sistema (npr. Stripe) za obračunavanje sejnih sob.
* **Strojna oprema ali fizični zasloni pred sejnimi sobami:** Projekt ne vključuje razvoja ali integracije s fizičnimi tablicami/zasloni na vratih sob.
* **Nativna mobilna aplikacija (iOS/Android):** Aplikacija bo razvita izključno kot odzivna spletna aplikacija (Responsive Web App).