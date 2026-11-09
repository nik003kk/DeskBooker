# Samodejno sproščanje rezervacij (DB-063–DB-067)

Backend ima en Go worker, ki ob zagonu in nato vsako minuto pregleda
rezervacije brez potrjene prisotnosti. Posamezen pregled ima največ
30 sekund in se ne prekriva s prejšnjim. Napaka ne ustavi naslednjih
pregledov. Ob ustavitvi strežnika se worker prekliče in zaključi,
preden se zapre povezovalni bazen PostgreSQL.

DB-063 vzpostavi periodični pregled; DB-064 doda spremembo statusa.
