# Mini-CPU — Instruction Set Architecture (ISA)

ISA didattico della mini-CPU di RetroNet Hardware: minimale ma sufficiente a
scrivere programmi reali con calcoli, memoria e salti. Tutto il percorso dati è
costruito dai componenti gate-level (register file, ALU di RetroNet Logic, PC).

## Caratteristiche

- **Dati a 8 bit**, **indirizzi a 8 bit** (memoria di 256 byte).
- **4 registri** general purpose: `R0`–`R3` (2 bit di selezione).
- **Flag**: `Zero` (Z) e `Carry` (C), prodotti dall'ALU.
- Esecuzione **fetch → decode → execute**; l'avvio è all'indirizzo `0x00`.

## Formato delle istruzioni

Ogni istruzione inizia con un **byte di opcode**:

```
 bit:  7 6 5 4 | 3 2 | 1 0
       opcode  |  DD | SS
       (4 bit) | dest| src
```

- `opcode` (nibble alto) seleziona l'operazione;
- `DD` = registro destinazione (0–3); `SS` = registro sorgente (0–3).

Alcune istruzioni hanno un **secondo byte** (immediato o indirizzo). I campi non
usati valgono 0.

## Set di istruzioni

| Opcode | Mnemonico | Byte | Effetto | Flag |
|:------:|-----------|:----:|---------|:----:|
| `0x0` | `NOP`         | 1 | nessuna operazione | — |
| `0x1` | `LDI Rd, imm` | 2 | `Rd = imm` | — |
| `0x2` | `MOV Rd, Rs`  | 1 | `Rd = Rs` | — |
| `0x3` | `ADD Rd, Rs`  | 1 | `Rd = Rd + Rs` | Z, C |
| `0x4` | `SUB Rd, Rs`  | 1 | `Rd = Rd - Rs` | Z, C |
| `0x5` | `AND Rd, Rs`  | 1 | `Rd = Rd & Rs` | Z (C=0) |
| `0x6` | `OR  Rd, Rs`  | 1 | `Rd = Rd \| Rs` | Z (C=0) |
| `0x7` | `LD  Rd, addr`| 2 | `Rd = mem[addr]` | — |
| `0x8` | `ST  Rd, addr`| 2 | `mem[addr] = Rd` | — |
| `0x9` | `JMP addr`    | 2 | `PC = addr` | — |
| `0xA` | `JZ  addr`    | 2 | se `Z`: `PC = addr` | — |
| `0xB` | `JC  addr`    | 2 | se `C`: `PC = addr` | — |
| `0xC` | `SHL Rd`      | 1 | `Rd <<= 1` (entra 0 nel LSB) | Z, C=bit uscito |
| `0xD` | `SHR Rd`      | 1 | `Rd >>= 1` (entra 0 nel MSB) | Z, C=bit uscito |
| `0xE1`| `CALL addr`   | 2 | salva il ritorno sulla pila, poi `PC = addr` | — |
| `0xE0`| `RET`         | 1 | `PC = pop()` (ritorno da subroutine) | — |
| `0xF` | `HLT`         | 1 | ferma la CPU | — |

`CALL`/`RET` usano l'opcode di gruppo `0xE` ("control"): il nibble basso
seleziona la sotto-operazione (`0` = RET, `1` = CALL).

## Pila e subroutine

La CPU ha uno **stack pointer** `SP` a 8 bit; la pila vive **in memoria** e
cresce verso il basso da `0xFF` (valore iniziale di `SP`). `CALL` salva
l'indirizzo di ritorno con `push`, `RET` lo recupera con `pop`; le subroutine si
possono **annidare** fin dove arriva la memoria. Gli aggiornamenti `SP±1` passano
per l'adder a gate. Attenzione a non far collidere la pila con dati/programma.

`SHL`/`SHR` usano lo shifter a gate di RetroNet Logic; il bit che esce
dall'estremità finisce nel flag `Carry`.

Convenzione del Carry (quella nativa dell'ALU): per `ADD` è il riporto uscente;
per `SUB`, `C = 1` significa **nessun prestito** (`Rd >= Rs`).

## Programma di esempio: somma 1+2+3+4+5

```
        LDI R0, 0      ; R0 = somma = 0
        LDI R1, 5      ; R1 = contatore i = 5
        LDI R2, 1      ; R2 = passo di decremento
loop:   ADD R0, R1     ; somma += i
        SUB R1, R2     ; i -= 1   (Z=1 quando i arriva a 0)
        JZ  end        ; se i == 0, esci
        JMP loop
end:    ST  R0, 0x20   ; salva il risultato in mem[0x20]
        HLT
```

Codice macchina (caricato da `0x00`):

```
addr:  00 01 02 03 04 05 06 07 08 09 0A 0B 0C 0D 0E
byte:  10 00 14 05 18 01 31 46 A0 0C 90 06 80 20 F0
```

Al termine `mem[0x20] = 15`.
