# DeskBooker – frontend

Osnovna React SPA, zgrajena z Vite (DB-004).

## Predpogoji
- Node.js različica, združljiva z Vite 7 (20.19+ ali 22.12+)
- npm

## Lokalni zagon

```bash
cd frontend
npm install
npm run dev
```

Nato v brskalniku odpri naslov, ki ga izpiše Vite (običajno `http://localhost:5173`).

## Produkcijska gradnja

```bash
npm run build
npm run preview
```

Z ukazom `npm run build` nastane mapa `dist/`, ki ni vključena v Git. Naslednje naloge bodo dodale navigacijo, avtentikacijo in funkcionalnosti rezervacij.
