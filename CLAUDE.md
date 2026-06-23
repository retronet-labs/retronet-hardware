# CLAUDE.md — RetroNet Hardware

Simulazione Go **didattica** della logica sequenziale, costruita sopra
RetroNet Logic. Architettura e modello stato/clock sono in [AGENTS.md](AGENTS.md).

## ⚠️ Setup di sviluppo (leggere prima di tutto)

Questo repo dipende da `retronet-logic`. In locale serve un file `go.work`
(**non** versionato) che punta al checkout sibling:

```sh
# in C:\work\source\retronet-hardware
go work init . ../retronet-logic
```

Senza `go.work` la build fallisce: il `go.mod` richiede `retronet-logic v0.3.0`,
non pubblicato. (`go list -m all` può lamentarsi della versione anche con go.work
attivo: `go build`/`go test` invece risolvono correttamente dal sorgente locale.)
La CI ricrea il workspace al volo facendo il checkout di entrambi i repo.

## Comandi

- Test: `go test ./...` (richiede `go.work`)
- Formattazione: `gofmt -w .` ; Analisi: `go vet ./...`

## Componenti

- **Sequenziali** (`struct` con stato + `Step(..., clk)`, clock esplicito,
  flip-flop edge-triggered master-slave): `latch` (SR) → `flipflop` (D latch,
  D flip-flop) → `register` (N bit con load).
- **`bridge/`**: adattatori che collegano la `alu` di Logic agli emulatori
  esistenti — `bridge/i8008` e `bridge/i4004` — validati da test di conformità
  esaustivi contro un riferimento fedele agli emulatori.

## Stato e PROSSIMO PASSO

Il bridge ALU è pronto e **conforme** agli emulatori (senza modificarli).

**Prossimo passo (pianificato per il 2026-06-24): delega vera** — far sì che
`go-4004` e `retronet-8008` chiamino i bridge al posto della loro aritmetica
interna. Ricetta dettagliata in [docs/bridge.md](docs/bridge.md). Tocca repo
emulatori funzionanti → procedere solo con via libera esplicito dell'utente.
