# RetroNet Hardware

> Simulazione **educativa** di componenti hardware reali (latch, flip-flop,
> registri, fino a una mini-CPU), costruita sopra
> [RetroNet Logic](https://github.com/retronet-labs/retronet-logic).

RetroNet Hardware prende i blocchi combinatori di RetroNet Logic (porte,
sommatori, ALU) e ci costruisce sopra il mondo **sequenziale**: i componenti che
hanno **stato** e sono scanditi da un **clock**. L'obiettivo, come per Logic, è
la chiarezza didattica: niente operatori aritmetici Go dove esiste l'equivalente
costruito con la logica.

## Il confine con RetroNet Logic

| | RetroNet Logic | RetroNet Hardware |
|---|---|---|
| Natura | combinatoria, *senza stato* | sequenziale, *con stato e clock* |
| Componenti | porte, half/full adder, adder N bit, MUX, ALU, tipo `Bus` | latch, flip-flop, registri, register file, PC, CPU |
| Modello | `func(ingressi) uscite` | `struct` con stato + metodo `Step(clk)` |

## Struttura

```
retronet-hardware/
├── latch/        # latch SR (prima cella di memoria)
├── flipflop/     # D latch, D flip-flop edge-triggered
├── register/     # registro a N bit
├── registerfile/ # banco di registri R0-R3
├── pc/           # Program Counter (registro + incremento + salto)
├── memory/       # modello di memoria a byte
├── cpu/          # mini-CPU a 8 bit (datapath + control)
├── bridge/       # adattatori ALU verso gli emulatori RetroNet
│   ├── i4004/
│   ├── i8008/
│   ├── i8080/
│   ├── i8086/    # ALU 8/16 bit + OF/AF, mul/div/shift composti dai gate
│   └── i6502/    # ADC/SBC binari e BCD, compare, BIT, shift/rotate
├── docs/         # guide didattiche in italiano
└── examples/     # programmi eseguibili di dimostrazione
```

## Modello di simulazione

I componenti sequenziali sono `struct` con stato interno e un metodo `Step` che
riceve gli ingressi e il **clock esplicito**. Un ciclo di clock si simula
chiamando `Step` con clock basso e poi alto. I flip-flop sono edge-triggered,
costruiti con lo schema master-slave a partire dai latch.

## Sviluppo locale (multi-repo con go.work)

Questo progetto dipende da RetroNet Logic. In locale i due repo si usano insieme
con un file `go.work` (non versionato). Clona entrambi come cartelle sibling:

```
work/
├── retronet-logic/
└── retronet-hardware/
```

quindi crea il workspace nella cartella di questo repo:

```sh
go work init . ../retronet-logic
go test ./...
```

## Licenza

Distribuito con licenza [MIT](LICENSE).
