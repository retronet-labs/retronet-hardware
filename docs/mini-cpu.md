# Mini-CPU

La **mini-CPU** è il punto d'arrivo di RetroNet: una CPU didattica a 8 bit
costruita assemblando i componenti dello stack — `porte → full adder → adder →
ALU` (RetroNet Logic) e `flip-flop → register → register file / PC` (RetroNet
Hardware). Esegue un piccolo ISA su una memoria a byte.

> A differenza degli emulatori 4004/8008 (modelli *comportamentali* di chip
> reali), questa CPU è *strutturale*: mostra **come** una CPU nasce dai mattoni
> logici. L'ISA è descritto in [cpu-isa.md](cpu-isa.md).

## Componenti del datapath

| Blocco | Da dove viene | Ruolo |
|--------|---------------|-------|
| Register file (R0–R3) | [`registerfile`](register-file.md) | operandi e risultati |
| ALU | `retronet-logic/alu` | aritmetica e logica + flag |
| Program Counter | [`pc`](program-counter.md) | indirizzo della prossima istruzione |
| Memoria (256 B) | `memory` | programma e dati |
| Control/decoder | `cpu` | dall'opcode ai segnali di controllo |

I flag `Zero` e `Carry` (registro di stato) sono prodotti dall'ALU.

## Diagramma del datapath

```
        ┌────────────┐  addr   ┌──────────┐
        │ Program    ├────────►│          │  opcode/operandi
        │ Counter    │         │ Memoria  ├────────────┐
        │ (+1 / load)│◄──jump──┤ (256 B)  │            │
        └─────┬──────┘         └────▲─────┘            ▼
              │                     │ store      ┌────────────┐
              │                     │            │  Decoder/  │
              │                     │            │  Control   │
              │              ┌──────┴─────┐      └─────┬──────┘
              │              │ Register   │  op,Rd,Rs   │
              └──────────────┤ file R0-R3 │◄────────────┘
                             └──┬─────┬───┘
                          Rd ▼  │     │ ▲ risultato
                             ┌──┴─────┴─┐
                             │   ALU    ├──► flag Z, C
                             └──────────┘
```

## Ciclo di esecuzione

Ogni chiamata a `cpu.Step()` esegue un'istruzione completa:

1. **fetch** — legge il byte all'indirizzo del PC; il PC si incrementa
   (tramite il suo sommatore);
2. **decode** — estrae opcode, registro destinazione e sorgente;
3. **execute** — esegue (ALU, memoria, salto) e, al fronte di clock, scrive il
   register file / aggiorna il PC.

Nel percorso dati non si usano operatori aritmetici Go: somma e incremento
passano per l'ALU e per l'incrementatore del PC, entrambi a porte.

## Esempio

```go
m := memory.New(256)
m.Load(programma) // vedi cpu-isa.md
c := cpu.New(m)
c.Run(100)        // esegue fino a HLT
fmt.Println(m.Read(0x20))
```

Demo eseguibile in [`examples/cpu`](../examples/cpu/main.go):

```sh
go run ./examples/cpu
```
