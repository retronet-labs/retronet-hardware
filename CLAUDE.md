# CLAUDE.md — RetroNet Hardware

Simulazione Go **didattica** della logica sequenziale, costruita sopra
RetroNet Logic. Architettura e modello stato/clock sono in [AGENTS.md](AGENTS.md).

## Setup di sviluppo

Questo repo dipende da `retronet-logic`, **pubblicato** su GitHub: con `go.sum`
presente, un clone pulito compila e testa subito (`go build`/`go test`), perché
`retronet-logic v0.3.0` si risolve da GitHub.

Per **co-sviluppare** logic e hardware insieme (modifiche locali a entrambi) usa
un `go.work` (non versionato) che punta al checkout sibling:

```sh
# in C:\work\source\retronet-hardware
go work init . ../retronet-logic
```

Con go.work attivo, `go list -m all` può lamentarsi della versione: `go build`/
`go test` invece risolvono correttamente dal sorgente locale.

## Comandi

- Test: `go test ./...` (richiede `go.work`)
- Formattazione: `gofmt -w .` ; Analisi: `go vet ./...`

## Componenti

- **Sequenziali** (`struct` con stato + `Step(..., clk)`, clock esplicito,
  flip-flop edge-triggered master-slave): `latch` (SR) → `flipflop` (D latch,
  D flip-flop) → `register` (N bit con load) → `registerfile` (banco R0-R3) →
  `pc` (Program Counter: register + adder + mux).
- **`memory`**: modello comportamentale di RAM a byte (scatola nera, non a gate).
- **`cpu`**: mini-CPU a 8 bit *strutturale* che assembla register file + ALU
  (di Logic) + PC + memoria, con ISA didattico (vedi docs/cpu-isa.md). Demo in
  `examples/cpu` (somma 1..5 = 15).
- **`bridge/`**: adattatori che collegano la `alu` di Logic agli emulatori
  esistenti — `bridge/i8008` e `bridge/i4004` — validati da test di conformità
  esaustivi contro un riferimento fedele agli emulatori.

## Stato

Lo stack è completo end-to-end: dai gate fino a una **mini-CPU funzionante**
(register file → PC → CPU, tag locale `v0.2.0`). Gli emulatori `go-4004` e
`retronet-8008` delegano le operazioni ALU ai bridge (`i4004`/`i8008`), con le
loro suite di test verdi.

Possibili prossimi passi: estendere l'ISA della mini-CPU, oppure delegare anche
le rotazioni degli emulatori (servirebbe prima uno shifter combinatorio in
retronet-logic).
