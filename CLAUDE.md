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

## Stato

Il bridge ALU è pronto, conforme e **collegato**: dal 2026-06-24 gli emulatori
`go-4004` e `retronet-8008` delegano davvero le operazioni ALU ai bridge
(`i4004`/`i8008`), con le loro intere suite di test verdi. Setup via `go.work`
locale in ciascun emulatore (vedi [docs/bridge.md](docs/bridge.md)). Tag locale
`v0.1.0`.

Possibili prossimi passi: register file, Program Counter, una mini-CPU; oppure
estendere la delega ad altre istruzioni (rotazioni, INC/DCR su registro).
