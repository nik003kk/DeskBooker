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
