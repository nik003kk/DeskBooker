# GitHub Actions CI (DB-014)

Workflow: `.github/workflows/ci.yml`.

## Kdaj teče?

Ob **vsakem pushu** na katerokoli vejo in ob **Pull Requestu**.
GitHub Actions izvaja dva neodvisna joba na `ubuntu-latest`:

| Job | Okolje | Koraki |
| --- | --- | --- |
| Frontend (React) | Node.js 22 | `npm ci`, `npm test`, `npm run build` |
| Backend (Go) | Go iz `backend/go.mod` | `go mod download`, `go test ./...`, `go vet ./...`, `go build ./...` |

Npm uporablja zavezani `frontend/package-lock.json` (ne `npm install`).
Go in npm odvisnosti se predpomnijo po lockfile oz. go.sum.
Joba se zaženeta ločeno: neuspeh enega ne ustavi izvajanja drugega.

Workflow potrebuje samo bralni dostop do kode (`contents: read`),
brez uporabniških skrivnosti. Trenutni enotni testi ne potrebujejo
PostgreSQL kontejnerja ali datoteke `.env`. Migracijski/seed integracijski
testi s pravo bazo niso del tega osnovnega CI in bodo dodani pozneje.

## Lokalno preverjanje pred push

Iz korena projekta:

```bash
cd frontend
npm ci
npm test
npm run build
cd ../backend
go mod download
go test ./...
go vet ./...
go build ./...
```

## Kako preveriti GitHub Actions

1. Odpri repozitorij GitHub → **Actions** → **CI**.
2. Preveri, da imata joba **Frontend (React)** in **Backend (Go)** zeleno kljukico.
3. Odpri Pull Request → **Checks** in preveri oba rezultata.
4. Če job pade, odpri njegov log. Če se workflow ne zažene, preveri,
   ali so Actions omogočeni v nastavitvah repozitorija.

**Negativni primeri:** okvara enega od `npm test`, `npm run build`,
`go test`, `go vet`, `go build` oziroma namestitve odvisnosti
mora povzročiti rdeč (failed) job. To je pogoj za ustrezno blokiranje
združitve, če nastaviš GitHub *branch protection / required status checks*.
Samo dodajanje workflowa še ne omogoči obveznih statusnih preverjanj.
