# Registro a N bit

Un **registro** memorizza una **parola** intera a N bit. È il primo componente
che conserva un valore vero e proprio, ed è il mattone dei register file e dei
percorsi dati (datapath) della CPU.

## Descrizione teorica

Un registro a N bit è semplicemente un insieme di N
[D flip-flop](flip-flop.md) che **condividono lo stesso clock**: a ogni fronte
di salita catturano contemporaneamente i loro bit, memorizzando l'intera parola.

Per poter *mantenere* il valore quando non si vuole scrivere, si aggiunge un
ingresso di abilitazione **load**. Per ogni bit un multiplexer 2:1 sceglie cosa
presentare all'ingresso del flip-flop:

```
ingresso_FF = load ? nuovo_dato : valore_attuale
            = OR(AND(valore_attuale, NOT load), AND(nuovo_dato, load))
```

- **load = 1** → al fronte di salita carica il nuovo dato;
- **load = 0** → reimmette il valore attuale, quindi lo mantiene.

Tutto è costruito da flip-flop e porte: nessun operatore aritmetico.

## Diagramma ASCII (registro a 4 bit)

```
          load
           │
 d0 ─[MUX]─┤►DFF─ q0
 d1 ─[MUX]─┤►DFF─ q1
 d2 ─[MUX]─┤►DFF─ q2
 d3 ─[MUX]─┤►DFF─ q3
           │
 clk ──────┴──────  (clock condiviso da tutti i flip-flop)
```

## Modello di simulazione

Come per i flip-flop, un ciclo di clock si simula chiamando `Step` con clock
basso e poi alto:

```go
r := register.New(8)
data := bus.FromUint(42, 8)
r.Step(data, bit.One, bit.Zero) // fase bassa
r.Step(data, bit.One, bit.One)  // fronte di salita -> carica 42
```

## Esempio di utilizzo

```go
package main

import (
    "fmt"

    "github.com/retronet-labs/retronet-logic/bit"
    "github.com/retronet-labs/retronet-logic/bus"
    "github.com/retronet-labs/retronet-hardware/register"
)

func main() {
    r := register.New(8)
    d := bus.FromUint(42, 8)
    r.Step(d, bit.One, bit.Zero)
    r.Step(d, bit.One, bit.One) // carica 42
    fmt.Println(r.Value().Uint()) // 42
}
```

Una dimostrazione è in [`examples/register`](../examples/register/main.go):

```sh
go run ./examples/register
```
