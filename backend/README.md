# DeskBooker – Go backend

REST API za DeskBooker, zgrajen v jeziku Go z ogrodjem Gin.

## Predpogoji

- Go 1.23 ali novejši
- Internetna povezava ob prvem prenosu odvisnosti

## Lokalni zagon

Iz korenske mape repozitorija:

```bash
cd backend
go mod tidy
go run ./cmd/server
```

Privzeti naslov: `http://localhost:8080`.

Health check:

```bash
curl http://localhost:8080/api/health
```

Pričakovani odgovor:

```json
{"service":"deskbooker-api","status":"ok"}
```

## Preizkusi

```bash
cd backend
go test ./...
go vet ./...
```

Za drug port nastavi okoljsko spremenljivko `PORT`, npr. `PORT=8081 go run ./cmd/server`.

## Obseg DB-005

Ta naloga vsebuje zagon Gin REST API, `/api/health` in osnovna HTTP testa. Povezava s PostgreSQL, CORS, domenski endpointi in JWT bodo dodani v naslednjih Issues. Po `go mod tidy` commitaj ustvarjeni `go.sum` v razvojno vejo.
