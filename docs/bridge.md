# Bridge: collegare la ALU a porte agli emulatori RetroNet

Il pacchetto `bridge` contiene gli **adattatori** che collegano la ALU a porte di
RetroNet Logic ([`alu`](https://github.com/retronet-labs/retronet-logic)) alla
semantica di una CPU specifica. Ogni adattatore:

1. converte gli operandi `byte` ⇄ [`bus.Bus`](../latch);
2. sceglie l'operazione e il riporto entrante secondo l'ISA;
3. rimappa i flag nella convenzione di quella CPU.

L'obiettivo è permettere agli emulatori RetroNet di **delegare** le operazioni
aritmetico-logiche alla ALU costruita dai gate, senza riscrivere la loro logica
e — soprattutto — senza cambiarne il comportamento osservabile.

## Conformità garantita

Per ciascun adattatore esiste un test di conformità che confronta l'uscita
(risultato + flag) con un **riferimento fedele** alla semantica dell'emulatore,
su **tutte** le combinazioni di ingresso:

- `bridge/i8008`: 8 gruppi ALU × 256 × 256 × 2 (carry) ≈ 1 milione di casi;
- `bridge/i4004`: tutte le coppie di nibble × 2 (carry) per Add/Sub e tutti i
  nibble per Inc/Dec/Complement.
- `bridge/i6502`: ADC/SBC × 256 × 256 × 2 × modalità binaria/decimale, più
  test mirati per logica, compare, BIT e shift/rotate.

I riferimenti rispecchiano `retronet-8008/cpu/alu.go` e
`go-4004/cpu/instructions.go`; se quegli emulatori cambiano, i test vanno
risincronizzati.

## Convenzioni di carry/borrow (la differenza chiave)

| | 4004 | 8008 |
|---|---|---|
| Sottrazione | `A + NOT(R) + carry` | `A - value - borrow` |
| Significato del Carry | **come la ALU**: 1 = nessun prestito | **borrow**: 1 = prestito |
| Mappatura dal Carry ALU | identità | inversione: `borrow = NOT(carry_ALU)` |

La ALU di Logic ha un'unica convenzione (Carry = "nessun prestito"); è
l'adattatore a presentarla come la CPU se l'aspetta.

## i8008

```go
result, flags := i8008.ALU(group, a, value, carryIn)
// group: GroupADD..GroupCMP ; flags: Carry, Zero, Sign, Parity (bool)
```

### Delega futura (bozza)

Nell'emulatore, `executeALU` potrebbe diventare:

```go
res, f := i8008.ALU(group, c.A, value, c.Carry)
c.Carry, c.Zero, c.Sign, c.Parity = f.Carry, f.Zero, f.Sign, f.Parity
if group != i8008.GroupCMP {
    c.A = res
}
```

## i4004

```go
result, carry := i4004.Add(a, r, carryIn)
result, carry := i4004.Sub(a, r, carryIn)
result, carry := i4004.Inc(a) // IAC
result, carry := i4004.Dec(a) // DAC
result        := i4004.Complement(a) // CMA
```

### Delega futura (bozza)

```go
// ADD
c.A, c.C = i4004.Add(c.A, c.R[low], c.C)
// SUB
c.A, c.C = i4004.Sub(c.A, c.R[low], c.C)
```

## i6502

```go
result, flags := i6502.ADC(a, value, carryIn, decimalMode)
result, flags := i6502.SBC(a, value, carryIn, decimalMode)
result, flags := i6502.Compare(reg, value)
result, flags := i6502.Logic(i6502.OpEOR, a, value)
```

Il bridge `i6502` segue la convenzione MOS/NMOS 6502:

- `Carry` in sottrazione significa **nessun prestito**;
- `ADC`/`SBC` in decimal mode restituiscono il risultato BCD corretto;
- `Zero`, `Negative` e `Overflow` di `ADC`/`SBC` decimali derivano dal risultato
  binario pre-correzione, coerentemente col comportamento NMOS modellato dagli
  emulatori RetroNet;
- `BIT` produce `Z` da `A & value`, `N` dal bit 7 dell'operando e `V` dal bit 6.

## Stato

**Delega attiva (2026-06-24).** Oltre alla validazione indipendente, gli
emulatori chiamano davvero questi adattatori:

- `retronet-8008/cpu/alu.go` → `executeALU` usa `i8008.ALU(...)`;
- `go-4004/cpu/instructions.go` → ADD/SUB/IAC/DAC/CMA usano `i4004.*`.

Ciascun emulatore usa un `go.work` locale (non versionato) con
`use . ../retronet-hardware ../retronet-logic`. Le suite di test complete dei due
emulatori restano verdi: il comportamento è invariato, ma l'aritmetica passa ora
per la ALU costruita dai gate.
