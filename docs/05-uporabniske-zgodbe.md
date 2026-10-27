# 05 - Uporabniške zgodbe in merila sprejemljivosti (Acceptance Criteria)

Dokument vsebuje 15 uporabniških zgodb za sistem **DeskBooker**, ki pokrivajo ključne funkcionalnosti aplikacije. Vse zgodbe sledijo standardni strukturi ter kriterijem INVEST (Independent, Negotiable, Valuable, Estimable, Small, Testable).

---

### US-01:   Kot uporabnik želim ustvariti račun in se varno prijaviti v aplikacijo DeskBooker z e-pošto in geslom, da lahko dostopam do funkcionalnosti rezervacije virov.
* **Povezana funkcionalnost:** [MVP] Avtentikacija in profil uporabnika
* **Uporabniška vloga:** Uporabnik (Zaposleni / Office Manager / Admin)
* **Acceptance Criteria (Merila sprejemljivosti):**
  1. Sistem omogoča registracijo z vnosom imena, priimka, službenega e-naslova in gesla (vsaj 8 znakov, vsaj ena številka in velika črka).
  2. Ob uspešni prijavi z veljavnimi podatki sistem generira JWT žeton in uporabnika preusmeri na osnovno nadzorno ploščo.
  3. Ob vnosu neveljavnega e-naslova ali napačnega gesla sistem prikaže jasno sporočilo o napaki "Neveljavno uporabniško ime ali geslo".
  4. Če e-poštni naslov ob registraciji že obstaja v bazi, sistem izpiše opozorilo "Uporabnik s tem e-naslovom je že registriran".

---

### US-02:   Kot zaposleni želim imeti vpogled v svoje profilne podatke in sistemsko vlogo, da vem, kateri oddelek predstavljam in kakšne pravice imam.
* **Povezana funkcionalnost:** [MVP] Avtentikacija in profil uporabnika
* **Uporabniška vloga:** Zaposleni / Član
* **Acceptance Criteria (Merila sprejemljivosti):**
  1. Na strani profila se prikažejo ime, priimek, e-naslov, oddelek ter dodeljena vloga (npr. `EMPLOYEE`).
  2. Uporabnik lahko spremeni svoje osebne nastavitve (npr. ime, priimek in oddelek), e-naslova pa ne more samostojno urejati.
  3. Spremembe profila se ob kliku na gumb "Shrani" takoj posodobijo v podatkovni bazi.
  4. Profil jasno prikazuje trenutno aktiven status uporabniškega računa.

---

### US-03:   Kot upravitelj pisarne želim v sistem dodati novo delovno mizo ali sejno sobo z definiranimi lastnostmi, da jo bodo zaposleni lahko rezervirali.
* **Povezana funkcionalnost:** [MVP] Dodajanje in upravljanje delovnih mest ter virov
* **Uporabniška vloga:** Upravitelj pisarne (Office Manager)
* **Acceptance Criteria (Merila sprejemljivosti):**
  1. Obrazec za dodajanje vira zahteva vnos naziva vira (npr. "Miza B12"), tipa vira (`DESK`, `MEETING_ROOM`, `PARKING_SPOT`), nadstropja in kapacitete.
  2. Upravitelj lahko ob ustvarjanju označi specifično opremo vira (npr. Dva monitorja, Dvižna miza, Projektor).
  3. Nov vir se po potrditvi takoj shrani in pojavi na seznamu aktivnih virov.
  4. Sistem prepreči shranjevanje vira, če je polje za naziv prazno ali je nadstropje negativno število.

---

### US-04: Kot upravitelj pisarne želim urediti podatke ali spremeniti status delovnega mesta v neaktivno, da zaposleni ne morejo rezervirati opreme, ki je v popravilu.
* **Povezana funkcionalnost:** [MVP] Dodajanje in upravljanje delovnih mest ter virov
* **Uporabniška vloga:** Upravitelj pisarne (Office Manager)
* **Acceptance Criteria (Merila sprejemljivosti):**
  1. Upravitelj lahko spremeni katerokoli polje obstoječega vira ali doda/odstrani opremo.
  2. Upravitelj lahko preklopi stikalo `is_active` na `false`, s čimer se vir označi kot neaktiven.
  3. Neaktivni viri se ne prikazujejo med rezultati iskanja za običajne zaposlene pri novih rezervacijah.
  4. Deaktivacija vira ne izbriše pretekle zgodovine rezervacij tega vira.

---

### US-05: Kot zaposleni želim filtrirati proste mize po datumu, času in specifični opremi (npr. dva monitorja in dvižna miza), da najdem ustrezno delovno mesto za svoje potrebe.
* **Povezana funkcionalnost:** [MVP] Napredno filtriranje in iskanje po specifikacijah opreme
* **Uporabniška vloga:** Zaposleni / Član
* **Acceptance Criteria (Merila sprejemljivosti):**
  1. Uporabnik lahko izbere datum, začetni in končni čas ter označi poljubne filtre opreme (`DUAL_MONITOR`, `ADJUSTABLE_DESK` itd.).
  2. Sistem vrne samo tiste vire, ki so prosti v izbranem časovnem oknu in vsebujejo vse izbrane lastnosti opreme.
  3. Če za izbrane filtre ni prostih virov, sistem prikaže opozorilo "Za izbrane kriterije ni prostih virov" ter ponudi možnost ponastavitve filtrov.
  4. Rezultati iskanja prikažejo naziv vira, nadstropje, ikone opreme ter gumb za hitro rezervacijo.

---

### US-06: Kot zaposleni želim izbrati prost vir ter potrditi rezervacijo za določen termin, da si zagotovim mesto v pisarni.
* **Povezana funkcionalnost:** [MVP] Delovni potek rezervacije mesta/sobe (Workflow)
* **Uporabniška vloga:** Zaposleni / Član
* **Acceptance Criteria (Merila sprejemljivosti):**
  1. Ob kliku na "Rezerviraj" sistem preveri razpoložljivost vira in v bazi ustvari rezervacijo s statusom `CONFIRMED`.
  2. Sistem ob uspešni rezervaciji samodejno generira edinstveno 4-mestno PIN kodo za check-in.
  3. Uporabnik na zaslonu prejme potrditveno sporočilo z podrobnostmi rezervacije in prikazano PIN kodo.
  4. Če je vir medtem že rezerviral drug uporabnik, sistem prepreči dvojno rezervacijo (double-booking) in izpiše napako "Izbrani vir je v tem terminu že zaseden".

---

### US-07: Kot zaposleni želim ob prihodu v pisarno vnesti svojo PIN kodo ali klikniti gumb za check-in, da potrdim svojo prisotnost in ohranim rezervacijo.
* **Povezana funkcionalnost:** [MVP] Avtomatski proces potrditve prisotnosti (Check-in Workflow)
* **Uporabniška vloga:** Zaposleni / Član
* **Acceptance Criteria (Merila sprejemljivosti):**
  1. Možnost check-in-a postane aktivna presatno 10 minut pred začetkom in ostane na voljo do 15 minut po začetku rezervacije.
  2. Ob pravilnem vnosu 4-mestne PIN kode ali potrditvi v aplikaciji se status rezervacije spremeni v `CHECKED_IN`, zabeleži pa se tudi čas `checked_in_at`.
  3. Ob vnosu napačne PIN kode sistem onemogoči potrditev in izpiše sporočilo "Napačna PIN koda".
  4. Po uspešnem check-in-u se na kartici rezervacije prikaže zelena značka "Prisotnost potrjena".

---

### US-08:   Kot sistem želim samodejno preklicati rezervacijo, če uporabnik ne potrdi prisotnosti v 15 minutah po začetku, da se vir sprosti za ostale sodelavce.
* **Povezana funkcionalnost:** [MVP] Logika avtomatskega sproščanja nepotrjenih mest ("No-show auto-release")
* **Uporabniška vloga:** Sistemski proces / Admin
* **Acceptance Criteria (Merila sprejemljivosti):**
  1. Ozadnji delovni proces (Cron Goroutine) preveri vse rezervacije v statusu `CONFIRMED` vsako minuto.
  2. Če je trenutni čas za več kot 15 minut večji od `start_time` in `checked_in_at` ostaja prazno, sistem spremeni status rezervacije v `EXPIRED`.
  3. Sprostitev rezervacije avtomatsko omogoči novim uporabnikom, da sproščeni vir takoj ponovno rezervirajo.
  4. V revizijsko sled (audit log) se zapiše dogodek "Rezervacija potekla zaradi odsotnosti check-in-a".

---

### US-09:   Kot upravitelj pisarne želim vpogled v grafične prikaze zasedenosti pisarniških kapacitet in deleža "no-show" rezervacij, da optimiziram izkoriščenost prostorov.
* **Povezana funkcionalnost:** [MVP] Nadzorna plošča in analitika zasedenosti pisarne
* **Uporabniška vloga:** Upravitelj pisarne (Office Manager)
* **Acceptance Criteria (Merila sprejemljivosti):**
  1. Nadzorna plošča prikazuje graf zasedenosti po dnevih v tednu (ponedeljek–petek).
  2. Prikaže se odstotek neizkoriščenih rezervacij (`EXPIRED` vs. `CHECKED_IN`) za izbrano časovno obdobje (zadnji teden / zadnji mesec).
  3. Podatki se preračunajo in osvežijo ob odprtju strani ali spremembi časovnega filtra.
  4. Izvoz zbirnih podatkov je omogočen v obliki preprostega CSV poročila.

---

### US-10:   Kot zaposleni želim imeti pregled nad vsemi svojimi aktivnimi in preteklimi rezervacijami ter možnost preklica, če v pisarno ne bom prišel.
* **Povezana funkcionalnost:** Moji zahtevki in pregled zgodovine rezervacij
* **Uporabniška vloga:** Zaposleni / Član
* **Acceptance Criteria (Merila sprejemljivosti):**
  1. Stran "Moje rezervacije" prikazuje ločeni tabeli/seznam za "Aktivne rezervacije" in "Zgodovino" (`COMPLETED`, `CANCELLED`, `EXPIRED`).
  2. Ob vsaki aktivni rezervaciji je viden gumb "Prekliči rezervacijo".
  3. Po potrditvi preklica se status rezervacije spremeni v `CANCELLED`, vir pa postane takoj spet prost za rezervacijo ostalim.
  4. Preklic že zaključene ali potekle rezervacije ni mogoč.

---

### US-11:   Kot zaposleni želim po imenu poiskati sodelavca, da vidim, katero mizo ima rezervirano na določen dan in lahko sedem blizu njega.
* **Povezana funkcionalnost:** Iskanje mesta sodelavca ("Find my colleague")
* **Uporabniška vloga:** Zaposleni / Član
* **Acceptance Criteria (Merila sprejemljivosti):**
  1. Iskalno polje omogoča vnos imena ali priimka sodelavca.
  2. Sistem vrne podatek o lokaciji sodelavca (npr. "Janez Novak – Miza B12, 2. nadstropje") za izbrani datum.
  3. Če sodelavec na izbrani datum nima rezervacije, sistem izpiše "Sodelavec na ta dan nima rezerviranega mesta".
  4. Rezultati iskanja spoštujejo zasebnost in ne prikazujejo osebnih stikov ali pretekle zgodovine izven službenih lokacij.

---

### US-12:   Kot zaposleni želim imeti dnevni in tedenski koledarski pogled zasedenosti sejnih sob, da hitro najdem prost termin za sestanek.
* **Povezana funkcionalnost:** Koledarski pregled zasedenosti sejnih sob
* **Uporabniška vloga:** Zaposleni / Član
* **Acceptance Criteria (Merila sprejemljivosti):**
  1. Koledarski pogled prikazuje časovne bloke zasedenosti po urah za vse sejne sobe.
  2. Zasedeni časovni bloki jasno prikazujejo naziv sestanka ali uporabnika, ki je sobo rezerviral.
  3. Klik na prazen časovni reženj samodejno odpre obrazec za rezervacijo sejne sobe z že predizbranim časom.
  4. Omogočeno je preklapljanje med dnevnim in tedenskim pogledom.

---

### US-13:   Kot zaposleni želim ob rezervaciji delovne mize z enim klikom rezervirati še prosto parkirno mesto za isti časovni interval.
* **Povezana funkcionalnost:** Rezervacija parkirnega mesta vezana na delovno mesto
* **Uporabniška vloga:** Zaposleni / Član
* **Acceptance Criteria (Merila sprejemljivosti):**
  1. Ob potrditvi rezervacije mize sistem prikaže opcijsko stikalo/kljukico "Dodaj rezervacijo parkirnega mesta".
  2. Če je možnost izbrana, sistem poišče prosto parkirno mesto in ustvari vzporedno rezervacijo vira tipa `PARKING_SPOT`.
  3. Če so vsa parkirna mesta zasedena, sistem obvešča "Parkirna mesta so zasedena, rezervacija mize pa je bila uspešna".
  4. Parkirno mesto se na seznamu mojih rezervacij prikazuje kot povezana rezervacija.

---

### US-14:   Kot zaposleni želim prejti e-poštni opomnik 10 minut pred začetkom rezervacije s PIN kodo, da ne pozabim opraviti check-in-a.
* **Povezana funkcionalnost:** Obvestila in opomniki (Email / In-app)
* **Uporabniška vloga:** Zaposleni / Član
* **Acceptance Criteria (Merila sprejemljivosti):**
  1. Sistem 10 minut pred začetkom rezervacije samodejno pošlje e-poštno sporočilo na uporabnikov e-naslov.
  2. Sporočilo vsebuje podrobnosti rezervacije (datum, ura, naziv mize) ter PIN kodo za check-in.
  3. Sporočilo vsebuje direktno povezavo (URL) do aplikacije za hitro potrditev prisotnosti.
  4. Če je bila rezervacija predhodno preklicana, se opomnik ne pošlje.

---

### US-15:   Kot administrator želim spremeniti sistemsko vlogo uporabnika (npr. sprememba iz Zaposleni v Office Manager), da zagotovim ustrezne pravice dostopa.
* **Povezana funkcionalnost:** [MVP] Avtentikacija in profil uporabnika / Administracija
* **Uporabniška vloga:** Administrator (Admin)
* **Acceptance Criteria (Merila sprejemljivosti):**
  1. Admin ima dostop do seznama vseh registriranih uporabnikov v sistemu.
  2. Admin lahko v padajočem meniju izbere novo vlogo (`EMPLOYEE`, `OFFICE_MANAGER`, `ADMIN`) za izbranega uporabnika.
  3. Po shranitvi se uporabnikove pravice v sistemu posodobijo takoj ob njegovi naslednji zahtevi ali osvežitvi žetona.
  4. Admin ne more spremeniti svoje lastne vloge ali izbrisati lastnega računa, da se prepreči zaklenitev sistema.