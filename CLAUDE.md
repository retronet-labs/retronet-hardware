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
  esistenti — `bridge/i4004`, `bridge/i8008` e `bridge/i8086` — validati da test
  di conformità esaustivi contro un riferimento fedele agli emulatori.
  - **`bridge/i8086`** (il più recente): ALU parametrica su width **8/16 bit** per
    i gruppi ADD/OR/ADC/SBB/AND/SUB/XOR/CMP(+TEST), `Increment`/`Decrement`,
    `Mul`/`IMul` (shift-and-add), `Div`/`IDiv` (a ripristino) e `Shift`/`Rotate`
    (loop sullo shifter a 1 bit) — tutto composto dai primitivi a gate, senza
    nuovo hardware in Logic. Flag completi 8086: CF/PF/AF/ZF/SF/OF.
    - **AF della sottrazione**: mezzo-prestito = bit 4 di `a ^ b ^ risultato` con
      la **b originale** (non complementata). Bug storico già corretto (v0.7.1).
    È consumato da `retronet-8086` (backend ALU `Gate`).

## Stato

Lo stack è completo end-to-end: dai gate fino a una **mini-CPU funzionante**
(register file → PC → CPU, tag locale `v0.2.0`). Gli emulatori `go-4004`,
`retronet-8008` e `retronet-8086` delegano le operazioni ALU ai bridge
(`i4004`/`i8008`/`i8086`), con le loro suite di test verdi.

Tag rilevanti per il bridge i8086: `v0.6.0`, `v0.7.0`, **`v0.7.1`** (fix AF della
sottrazione) — è la versione richiesta da retronet-8086.

Possibili prossimi passi: estendere l'ISA della mini-CPU, oppure (opzione "gate
completo") barrel shifter / moltiplicatore / divisore dedicati in retronet-logic
al posto della composizione per ripetizione.
