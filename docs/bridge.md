# Bridge: collegare la ALU a porte agli emulatori 4004/8008

Il pacchetto `bridge` contiene gli **adattatori** che collegano la ALU a porte di
RetroNet Logic ([`alu`](https://github.com/retronet-labs/retronet-logic)) alla
semantica di una CPU specifica. Ogni adattatore:

1. converte gli operandi `byte` ⇄ [`bus.Bus`](../latch);
2. sceglie l'operazione e il riporto entrante secondo l'ISA;
3. rimappa i flag nella convenzione di quella CPU.

L'obiettivo è permettere agli emulatori **esistenti** (`go-4004`,
`retronet-8008`) di **delegare** le operazioni aritmetico-logiche alla ALU
costruita dai gate, senza riscrivere la loro logica e — soprattutto — senza
cambiarne il comportamento osservabile.

## Conformità garantita

Per ciascun adattatore esiste un test di conformità che confronta l'uscita
(risultato + flag) con un **riferimento fedele** alla semantica dell'emulatore,
su **tutte** le combinazioni di ingresso:

- `bridge/i8008`: 8 gruppi ALU × 256 × 256 × 2 (carry) ≈ 1 milione di casi;
- `bridge/i4004`: tutte le coppie di nibble × 2 (carry) per Add/Sub e tutti i
  nibble per Inc/Dec/Complement.

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

## Stato

Per ora gli emulatori restano **intatti**: il bridge è validato in modo
indipendente. Il passo successivo, quando vorrai, è far sì che 4004 e 8008
chiamino davvero questi adattatori al posto della loro aritmetica interna.
