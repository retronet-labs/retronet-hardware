# Documentazione di RetroNet Hardware

Guide didattiche in italiano ai componenti **sequenziali** costruiti sopra
[RetroNet Logic](https://github.com/retronet-labs/retronet-logic).

## Indice

| Componente | Pacchetto | Scheda |
|------------|-----------|--------|
| Latch SR | [`latch`](../latch) | [latch.md](latch.md) |
| D latch e D flip-flop | [`flipflop`](../flipflop) | [flip-flop.md](flip-flop.md) |
| Registro a N bit | [`register`](../register) | [register.md](register.md) |
| Bridge ALU 4004/8008 | [`bridge`](../bridge) | [bridge.md](bridge.md) |

## Percorso di lettura consigliato

1. **[Latch SR](latch.md)** — la prima cella di memoria e l'introduzione alla
   retroazione e allo stato.
2. **[D latch e D flip-flop](flip-flop.md)** — la sincronizzazione con il clock,
   dal livello al fronte (master-slave).
3. **[Registro a N bit](register.md)** — la memorizzazione di una parola intera.

In arrivo: register file, contatore/Program Counter e la mini-CPU (vedi
`AGENTS.md` per l'ISA di riferimento).

## Documentazione del codice (godoc)

```sh
go doc ./latch
go doc ./flipflop
go doc ./register
```

## Esempi eseguibili

```sh
go run ./examples/register
```
