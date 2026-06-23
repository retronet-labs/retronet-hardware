# Latch SR (Set-Reset)

Il **latch SR** è la più semplice cella di memoria: il primo componente *con
stato* di RetroNet Hardware. È il punto in cui si passa dalla logica
combinatoria a quella **sequenziale**.

## Descrizione teorica

Tutti i componenti visti finora sono combinatori: l'uscita dipende solo dagli
ingressi. Il latch introduce la **retroazione**: le uscite rientrano negli
ingressi, e questo permette al circuito di *ricordare* un valore.

Il latch SR si costruisce con due porte **NOR** a retroazione incrociata:

```
Q  = NOR(R, Qn)
Qn = NOR(S, Q)
```

- **S** (Set) porta l'uscita Q a 1;
- **R** (Reset) porta l'uscita Q a 0;
- con S=R=0 il latch **mantiene** l'ultimo valore (memoria);
- S=R=1 è uno **stato non valido** (Q e Qn risultano entrambi 0).

In simulazione la retroazione non è istantanea come nell'hardware: si memorizza
lo stato (Q, Qn) e lo si rivaluta finché si stabilizza.

## Tabella degli stati

```
S R │ Q (stato successivo)
────┼─────────────────────
0 0 │ Q   (mantiene)
1 0 │ 1   (set)
0 1 │ 0   (reset)
1 1 │ —   (non valido)
```

## Diagramma ASCII

```
      ┌──────┐
R ────┤ NOR  ├──┬──── Q
   ┌──┤      │  │
   │  └──────┘  │
   │  ┌──────┐  │
   └──┤ NOR  ├──┘
S ────┤      ├──────── Qn
      └──────┘
```

## Diagramma di temporizzazione

```
S  ▁▁▔▔▁▁▁▁▁▁▁▁
R  ▁▁▁▁▁▁▔▔▁▁▁▁
Q  ▁▁▁▔▔▔▔▁▁▁▁▁   (sale al Set, scende al Reset, mantiene fra i due)
```

## Esempio di utilizzo

```go
package main

import (
    "fmt"

    "github.com/retronet-labs/retronet-logic/bit"
    "github.com/retronet-labs/retronet-hardware/latch"
)

func main() {
    l := latch.NewSR()
    q, _ := l.Step(bit.One, bit.Zero)  // set  -> Q=1
    fmt.Println(q)
    q, _ = l.Step(bit.Zero, bit.Zero) // hold -> Q=1
    fmt.Println(q)
    q, _ = l.Step(bit.Zero, bit.One)  // reset-> Q=0
    fmt.Println(q)
}
```
