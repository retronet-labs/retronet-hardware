Sei l'architetto software del progetto RetroNet Hardware.

RetroNet Hardware è una libreria Go costruita sopra RetroNet Logic.

L'obiettivo è simulare componenti hardware reali utilizzando i blocchi
elementari forniti da RetroNet Logic.

NON usare operatori aritmetici Go (+ - * /) quando esiste una implementazione
equivalente tramite Full Adder o componenti logici. Sono ammessi gli operatori
solo per il codice di "interfaccia umana" (conversioni da/verso interi, indici,
cicli) e nei test per calcolare il risultato atteso.

Ogni componente deve essere:

- indipendente
- testabile
- documentato
- riutilizzabile

Tutta la documentazione deve essere in italiano.

I commit devono essere piccoli e atomici.


## Architettura: dove finisce Logic e dove inizia Hardware

Il confine tra i due progetti è il confine tra **logica combinatoria** e
**logica sequenziale**:

- **RetroNet Logic** (dipendenza) — tutto ciò che è combinatorio e *senza stato*,
  cioè `uscita = f(ingressi)`: porte, half/full adder, **sommatore a N bit**,
  multiplexer/decoder e **ALU**. Anche il tipo a più bit (`Bus`/word) vive in
  Logic ed è la valuta comune tra i due progetti.
- **RetroNet Hardware** (questo repo) — tutto ciò che ha **stato** e **clock**:
  latch, flip-flop, registri, register file, contatori/Program Counter e infine
  la CPU che assembla il datapath (register file + ALU di Logic) con la control
  unit.

Gerarchia di simulazione (↓ = "costruito sopra"):

```
[RetroNet Logic]  porte → full adder → adder N bit → MUX/decoder → ALU
                                                                     │
[RetroNet Hardware]  latch → flip-flop → registro → register file ───┤
                                              counter / PC ──────────┤
                                                                     ▼
                                                                    CPU
```

## Modello dello stato e del clock (decisione vincolante)

I componenti combinatori restano funzioni pure (`func(ingressi) uscite`).
I componenti **sequenziali** hanno stato e si modellano così:

- sono `struct` che custodiscono lo stato interno (i bit memorizzati);
- espongono un metodo `Step(...)` che riceve gli ingressi e il **clock esplicito**
  e restituisce le uscite, aggiornando lo stato;
- la retroazione (l'uscita che rientra nell'ingresso, es. nel latch SR) è
  realizzata leggendo/aggiornando lo stato memorizzato, iterando finché si
  stabilizza;
- i flip-flop sono **edge-triggered** e si costruiscono con lo schema
  master-slave (due latch con clock opposti), non con scorciatoie.

Un ciclo di clock si simula chiamando `Step` con clock basso e poi alto.

## Dipendenza da RetroNet Logic

- Module path di questo progetto: `github.com/retronet-labs/retronet-hardware`.
- Dipende da `github.com/retronet-labs/retronet-logic` (versione taggata).
- In sviluppo locale i due repo si usano insieme tramite un file `go.work`
  (non versionato) che punta al checkout sibling di retronet-logic. Vedi
  CONTRIBUTING.md.

## Vincoli tecnici (ereditati da RetroNet Logic)

- Go 1.25+.
- Nessuna dipendenza esterna oltre a RetroNet Logic.
- `gofmt`, `go vet` e `go test ./...` devono restare puliti.

## Requisiti didattici

Ogni componente deve avere: descrizione teorica, esempio di utilizzo e diagramma
ASCII. Per i componenti combinatori serve la **tabella di verità**; per i
componenti sequenziali servono inoltre, dove utile, il **diagramma degli stati**
e/o un piccolo **diagramma di temporizzazione (timing)** che mostri il
comportamento rispetto al clock.

## Obiettivo finale: una mini-CPU concreta

Per evitare un finale aperto, il target è una **CPU didattica a 8 bit** con un
ISA minimale (pochi opcode: load immediato, somma, eventualmente AND/OR, salto
condizionato e halt), datapath esplicito (PC, register file, ALU di Logic,
memoria) e control unit. L'ISA esatto va definito e documentato in `docs/` prima
di implementare la CPU.

Non introdurre dipendenze esterne non necessarie.
