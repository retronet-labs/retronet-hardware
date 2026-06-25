# Mini-assembler

`miniasm` è un piccolo assembler per l'[ISA della mini-CPU](cpu-isa.md): traduce
un sorgente testuale in codice macchina. Rende comodo scrivere programmi senza
calcolare a mano opcode e indirizzi.

## Sintassi

- una **istruzione** per riga: `MNEMONICO operandi…` (mnemonici e registri sono
  case-insensitive);
- **registri**: `R0`–`R3`;
- **valori** (immediati/indirizzi): decimale (`42`) o esadecimale (`0x2A`);
- **etichette**: `nome:` a inizio riga; usabili come indirizzo in `JMP/JZ/JC/CALL`
  e in `LD/ST`;
- **commenti**: da `;` a fine riga.

Funziona in **due passate**: la prima calcola gli indirizzi e le etichette, la
seconda emette i byte.

## Esempio

```go
src := `
        LDI R0, 0
        LDI R1, 5
        LDI R2, 1
loop:   ADD R0, R1
        SUB R1, R2
        JZ  end
        JMP loop
end:    ST  R0, 0x20
        HLT
`
code, err := miniasm.Assemble(src)   // []byte pronti per memory.Load
```

Caricando `code` in memoria ed eseguendolo sulla mini-CPU, `mem[0x20]` vale 15.

## Subroutine

```
        LDI R0, 21
        CALL doppio
        ST  R0, 0x30
        HLT
doppio: SHL R0          ; R0 *= 2
        RET
```

Vedi l'elenco completo degli opcode in [cpu-isa.md](cpu-isa.md).
