# D latch e D flip-flop

Dopo il [latch SR](latch.md), questi due componenti aggiungono la
**sincronizzazione con il clock**, fino ad arrivare all'elemento di memoria su
cui si basano i registri: il **D flip-flop sensibile al fronte**.

## D latch (sensibile al livello)

Il latch SR ha due difetti: due ingressi separati e uno stato non valido
(S=R=1). Il **D latch** li risolve usando un solo dato `D`, abilitato dal clock:

```
S = AND(D, clk)
R = AND(NOT D, clk)
```

- con **clk = 1** il latch è **trasparente**: Q segue D;
- con **clk = 0** il latch **mantiene** il valore.

Il limite: finché il clock è alto, ogni variazione di D passa subito in uscita
(sensibilità al *livello*). Per i registri serve catturare il dato in un
**istante preciso**.

### Tabella

```
clk D │ Q
──────┼─────────────
 0  - │ Q  (mantiene)
 1  0 │ 0  (trasparente)
 1  1 │ 1  (trasparente)
```

## D flip-flop (sensibile al fronte) — master-slave

Il **D flip-flop** cattura D **solo sul fronte di salita** del clock. Si
costruisce con due D latch in cascata pilotati da clock opposti:

```
          clk' = NOT clk        clk
   D ──────► [ D latch ] ──────► [ D latch ] ──────► Q
               master              slave
```

- mentre **clk = 0**: il master è trasparente e insegue D; lo slave mantiene;
- sul **fronte di salita** (clk 0→1): il master si congela sul valore di D, lo
  slave diventa trasparente e copia il valore del master in uscita.

Risultato: l'uscita cambia **una sola volta per ciclo**, con il valore di D
presente al fronte di salita. Le variazioni di D mentre il clock è alto vengono
ignorate.

### Diagramma di temporizzazione

```
clk  ▁▁▔▔▁▁▔▔▁▁▔▔
D    ▁▔▔▔▔▔▁▁▁▁▁▁
Q    ▁▁▁▔▔▔▔▔▁▁▁▁    (Q assume D solo sui fronti ▏di salita di clk)
```

## Modello di simulazione

Un ciclo di clock si simula chiamando `Step` con clock basso e poi alto:

```go
ff := flipflop.NewDFlipFlop()
ff.Step(bit.One, bit.Zero) // fase bassa: prepara il dato
q := ff.Step(bit.One, bit.One) // fronte di salita: cattura -> q = 1
```

## Esempio di utilizzo

```go
package main

import (
    "fmt"

    "github.com/retronet-labs/retronet-logic/bit"
    "github.com/retronet-labs/retronet-hardware/flipflop"
)

func main() {
    ff := flipflop.NewDFlipFlop()
    ff.Step(bit.One, bit.Zero)
    fmt.Println(ff.Step(bit.One, bit.One)) // 1 (catturato sul fronte)
}
```
