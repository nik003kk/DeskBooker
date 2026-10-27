# 01 - Zasnova rešitve: DeskBooker

## Ime rešitve
**DeskBooker** (Sistem za rezervacijo delovnih mest in virov v hibridni pisarni)

## Namen rešitve in problem, ki ga rešitev naslavlja
S prehodom na hibridni način dela (kombinacija dela od doma in iz pisarne) ter uveljavitvijo modela prilagodljivih delovnih mest se podjetja in coworking prostori soočajo z negotovostjo glede zasedenosti svojih prostorov. Zaposleni ob prihodu v pisarno pogosto ugotovijo, da ni proste ustrezne mize, da njihova ekipa sedi drugje ali pa da so sejne sobe zasedene brez predhodne najave. Hkrati pa se dogaja, da so mesta rezervirana, a ostanejo prazna ("ghost bookings").

**DeskBooker** te izzive rešuje z enostavno spletno aplikacijo za rezervacijo delovnih miz, sejnih sob in parkirnih mest glede na želeno opremo (npr. število monitorjev, dvižna miza). Z avtomatskim potrjevanjem prisotnosti ("check-in") in sproščanjem neizkoriščenih rezervacij pa sistem poskrbi, da pisarniške kapacitete ne ostajajo neuporabljene.

## Ciljni uporabniki
1. **Zaposleni / Člani coworking prostora:** Posamezniki, ki želijo rezervirati ustrezno delovno mesto (mizo) ali sejno sobo za določen dan ali časovno okno ter preveriti, kje sedijo njihovi sodelavci.
2. **Upravitelji pisarn (Office Managers):** Odgovorni za upravljanje delovnih prostorov, tlorisov, opreme na mizah, analitiko zasedenosti in določanje pravil rezervacij.
3. **Sistemski administratorji (Admin):** Skrbniki sistema, ki upravljajo uporabniške račune, sistemske vloge, pravice dostopa ter sistemske nastavitve.

## Kratek opis delovanja rešitve
Zaposleni se prijavi v aplikacijo, izbere želeni datum in časovno okno ter na podlagi filtrov (npr. "dva monitorja", "dvižna miza", "tiha cona") preveri razpoložljive mize ali sejne sobe na tlorisu/seznamu. Po potrditvi rezervacije prejme potrdilo ter edinstveno kodo/povezavo za check-in. Ko uporabnik pride v pisarno, mora v 15 minutah od začetka rezervacije opraviti potrditev prisotnosti (check-in). Če tega ne stori, sistem rezervacijo avtomatsko prekliče, sprosti delovno mesto za druge sodelavce ter admina o tem obvesti.