# Program Counter (PC)

Il **Program Counter** è il registro che contiene l'indirizzo della **prossima
istruzione** da eseguire. A ogni ciclo o si **incrementa** (per passare
all'istruzione successiva) o **carica** un nuovo indirizzo (per i salti).

## Descrizione teorica

Il PC è un caso da manuale di componente sequenziale costruito unendo:

- **stato** — un [registro](register.md) che conserva l'indirizzo;
- **logica combinatoria** — un [sommatore](../README.md) per `PC + 1`, e dei
  [multiplexer](../README.md) per scegliere il prossimo valore.

Priorità del prossimo valore:

```
next = load ? addr : (inc ? PC + 1 : PC)
write = load OR inc
```

Cioè: un salto (`load`) ha la precedenza; altrimenti, se abilitato l'incremento
(`inc`), si avanza; altrimenti il PC mantiene il valore.

## Diagramma ASCII

```
            ┌─────────┐  PC
        ┌──►│ register ├───┬─────────► (uscita: indirizzo corrente)
        │   └─────────┘    │
        │                  ▼
        │            ┌──────────┐ PC+1
        │            │ adder +1 ├───┐
        │            └──────────┘   │
        │     addr ───────────────┐ │
        │                      ┌───┴─┴───┐
        └──────────────────────┤  MUX    │◄── load, inc
                               └─────────┘
```

## Esempio di utilizzo

```go
p := pc.New(8)

// due incrementi
p.Step(addr, bit.Zero, bit.One, bit.Zero); p.Step(addr, bit.Zero, bit.One, bit.One) // -> 1
p.Step(addr, bit.Zero, bit.One, bit.Zero); p.Step(addr, bit.Zero, bit.One, bit.One) // -> 2

// salto a 100
j := bus.FromUint(100, 8)
p.Step(j, bit.One, bit.Zero, bit.Zero); p.Step(j, bit.One, bit.Zero, bit.One) // -> 100
```
