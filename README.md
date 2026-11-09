# DeskBooker
Sistem za rezervacijo delovnih mest in virov v hibridni pisarni

## Zagon z Dockerjem

React, Go API in PostgreSQL: navodila v [docs/docker.md](docs/docker.md).

Po pripravi korenskega `.env` zaženi `docker compose up -d --build`.
Aplikacija je na http://localhost:5173, migracije se izvedejo samodejno.
