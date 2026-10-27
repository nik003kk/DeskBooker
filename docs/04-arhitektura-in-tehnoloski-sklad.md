# 04 - Arhitektura in tehnološki sklad

## Arhitektura sistema
Sistem **DeskBooker** sledi sodobni tro-nivojski (3-tier) arhitekturi z ločenim sprednjim delom (Frontend) in visokozmogljivim zadnjim delom (Backend REST API), napisanim v jeziku **Go (Golang)**, ter relacijsko podatkovno bazo.

* **Frontend (Odjemalec):** Enostranska spletna aplikacija (SPA - React.js), ki skrbi za uporabniški vmesnik, interaktivno iskanje, filtriranje po opremi ter vmesnik za check-in.
* **Backend API Strežnik (Golang):** Hitra in lahka spletna storitev v jeziku Go, ki skrbi za poslovno logiko (validacija časovnih preklopov rezervacij, upravljanje statusov rezervacij, preverjanje pravic glede na vloge ter generiranje JWT žetonov).
* **(Background Cron Worker v Go):** Vzporedna nit (goroutine / cron task) znotraj Go strežnika, ki vsako minuto preverja potekle rezervacije brez opravljenega check-in-a ter jih avtomatsko sprošča.
* **Podatkovna baza:** Relacijska baza podatkov za zagotavljanje transakcijske integritete (preprečevanje dvojnih rezervacij istega vira ob istem času).
* **Avtentikacija:** Upravljanje sej s spletnimi žetoni JWT (JSON Web Tokens), posredovanimi v HTTP glavah ali varnih piškotkih.



Dodaj, da bodo emaili kot taki v diagramu, da bo videno kam gredo 

---

## Diagram arhitekture (Mermaid)

![tehnološki sklad aplikacije DeskBooker](images/technologies_diagram.png)

```mermaid
graph TD
    User([Uporabnik / Spletni brskalnik]) <-->|HTTPS / REST API / JSON| Frontend[Spletni odjemalec - React.js]
    Frontend <-->|JWT Auth / REST Calls| GoBackend[Backend API Strežnik - Go / Gin]
    GoBackend <-->|SQL / ORM Queries| DB[(Relacijska baza - PostgreSQL)]
    
    subgraph Go Application Process
        GoBackend
        CronWorker[Background Goroutine - Check-in Auto-Release Worker] <-->|Preverjanje in sprostitev rezervacij| DB
    end