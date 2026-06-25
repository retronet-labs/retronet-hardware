# Documentazione di RetroNet Hardware

Guide didattiche in italiano ai componenti **sequenziali** costruiti sopra
[RetroNet Logic](https://github.com/retronet-labs/retronet-logic).

## Indice

| Componente | Pacchetto | Scheda |
|------------|-----------|--------|
| Latch SR | [`latch`](../latch) | [latch.md](latch.md) |
| D latch e D flip-flop | [`flipflop`](../flipflop) | [flip-flop.md](flip-flop.md) |
| Registro a N bit | [`register`](../register) | [register.md](register.md) |
| Register file | [`registerfile`](../registerfile) | [register-file.md](register-file.md) |
| Program Counter | [`pc`](../pc) | [program-counter.md](program-counter.md) |
| Mini-CPU | [`cpu`](../cpu) | [mini-cpu.md](mini-cpu.md) · [ISA](cpu-isa.md) |
| Mini-assembler | [`miniasm`](../miniasm) | [miniasm.md](miniasm.md) |
| Bridge ALU 4004/8008 | [`bridge`](../bridge) | [bridge.md](bridge.md) |

## Percorso di lettura consigliato

1. **[Latch SR](latch.md)** — la prima cella di memoria e l'introduzione alla
   retroazione e allo stato.
2. **[D latch e D flip-flop](flip-flop.md)** — la sincronizzazione con il clock,
   dal livello al fronte (master-slave).
3. **[Registro a N bit](register.md)** — la memorizzazione di una parola intera.
4. **[Register file](register-file.md)** — il banco di registri di lavoro.
5. **[Program Counter](program-counter.md)** — registro che si incrementa e salta.
6. **[Mini-CPU](mini-cpu.md)** (+ [ISA](cpu-isa.md)) — l'assemblaggio finale:
   register file + ALU + PC + memoria.

## Documentazione del codice (godoc)

```sh
go doc ./latch
go doc ./flipflop
go doc ./register
go doc ./registerfile
go doc ./pc
go doc ./cpu
```

## Esempi eseguibili

```sh
go run ./examples/register
go run ./examples/cpu
```
