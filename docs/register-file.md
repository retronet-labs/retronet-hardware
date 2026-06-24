# Register file

Un **register file** è il banco di registri di lavoro di una CPU: un piccolo
insieme di registri della stessa larghezza, con **porte di lettura combinatorie**
(per leggere gli operandi) e **una porta di scrittura** selezionabile,
sincronizzata dal clock.

## Descrizione teorica

I [registri](register.md) sono già fatti di flip-flop. Un register file li mette
in fila e aggiunge la **selezione**:

- **lettura** — combinatoria: dato un indice, l'uscita è subito il contenuto del
  registro (più porte di lettura = più operandi nello stesso ciclo);
- **scrittura** — al fronte di salita del clock, abilitata su **un solo**
  registro per volta (selezione "one-hot"); gli altri mantengono il valore.

Nel datapath di una CPU il register file fornisce i due operandi all'ALU e
riceve indietro il risultato.

> Scelta di modellazione: la **decodifica dell'indirizzo** (quale registro
> abilitare) è logica di controllo, qui scritta in Go; i registri — il percorso
> dati — restano costruiti dai gate. Un decoder a porte sarebbe un'estensione
> naturale.

## Diagramma ASCII

```
            sel (scrittura)         sel_a   sel_b  (lettura)
               │                      │       │
        ┌──────┴───────┐             ▼       ▼
 data ─►│  R0 R1 ... Rn │──(porte di lettura)──► out_a, out_b
 write ►│  (write one-hot)│
 clk  ─►│  clock condiviso │
        └──────────────────┘
```

## Esempio di utilizzo

```go
f := registerfile.New(4, 8) // 4 registri da 8 bit

data := bus.FromUint(42, 8)
f.Step(3, data, bit.One, bit.Zero) // fase bassa
f.Step(3, data, bit.One, bit.One)  // fronte di salita: scrive R3

fmt.Println(f.Read(3).Uint()) // 42
```
