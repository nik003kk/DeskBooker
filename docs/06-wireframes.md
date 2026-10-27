# 06 - Wireframes (Skice uporabniškega vmesnika)

Ta dokument vsebuje opis in povezave do wireframe-ov (skic) ključnih zaslonov aplikacije **DeskBooker**. Wireframe-i prikazujejo postavitev elementov, navigacijo in osnovni izgled odzivne spletne aplikacije.

---

## 1. Nadzorna plošča in iskanje virov (Dashboard & Search)
**Datoteka slike:** `images/wireframe-01-dashboard.png`  
![Wireframe 1 - Nadzorna plošča in iskanje](images/wireframe-01-dashboard.png)

### Opis zaslona:
Zaslon predstavlja vstopno točko za zaposlenega po prijavi v sistem.
* **Zgornja navigacijska vrstica (Header):** Vsebuje logotip aplikacije DeskBooker, povezave na "Iskanje", "Moje rezervacije", "Sodelavci" ter profil prijavljenega uporabnika.
* **Leva stran (Filtri):** Vsebuje obrazec za izbiro datuma, začetnega in končnega časa ter potvrditvena polja za opremo (npr. Dva monitorja, Dvižna miza, Tiha cona).
* **Osrednji del (Rezultati iskanja):** Mreža (grid) kartic razpoložljivih miz in sejnih sob. Vsaka kartica prikaže naziv vira, nadstropje, ikone opreme ter izpostavljen gumb **"Rezerviraj"**.

---

## 2. Pregled in potrditev rezervacije ter Check-in modalno okno
**Datoteka slike:** `images/wireframe-02-reservation-checkin.png`  
![Wireframe 2 - Potrditev in Check-in](images/wireframe-02-reservation-checkin.png)

### Opis zaslona:
Zaslon prikazuje podrobnosti izbranega vira in potek potrditve ter izvedbe check-in-a.
* **Obrazec za potrditev:** Prikaz podrobnosti vira (npr. Miza B12, 2. nadstropje), izbranega časovnega okna ter stikalo za opcijsko rezervacijo parkirnega mesta.
* **Modalno okno za Check-in:** Pop-up, ki se uporabniku odpre ob prihodu na delovno mesto ali kliku na opomnik. Vsebuje polje za vnos 4-mestne PIN kode, prikaz preostalega časa za check-in (odštevalnik 15 minut) ter gumb **"Potrdi prisotnost"**.

---

## 3. Nadzorna plošča upravitelja in analitika zasedenosti (Office Manager Analytics)
**Datoteka slike:** `images/wireframe-03-analytics.png`  
![Wireframe 3 - Analitika zasedenosti](images/wireframe-03-analytics.png)

### Opis zaslona:
Zaslon je namenjen vlogi **Office Manager** za spremljanje učinkovitosti uporabe prostorov.
* **Kanal s ključnimi metrikami (KPI Cards):** Na vrhu so prikazane številke o skupni zasedenosti (npr. 78%), število aktivnih rezervacij in delež neizkoriščenih mest ("No-show rate").
* **Grafični prikazi:** 
  * Stolpčni diagram zasedenosti po dnevih v tednu (Pon–Pet).
  * Pie chart (tortni grafik) razmerja med opravljenimi check-in-i, preklici in poteklimi rezervacijami (`EXPIRED`).
* **Tabela upravljanja virov:** Spodnji del vsebuje tabelo z možnostjo hitrega dodajanja ali deaktivacije posameznih miz in sob.