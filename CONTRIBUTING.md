# Come contribuire a RetroNet Hardware

RetroNet Hardware è un progetto **educativo** costruito sopra
[RetroNet Logic](https://github.com/retronet-labs/retronet-logic): privilegia la
**chiarezza didattica**.

## Principi guida

Ogni componente deve essere **indipendente**, **testabile**, **documentato** e
**riutilizzabile**. Inoltre:

- **niente operatori aritmetici Go** (`+ - * /`) quando esiste l'equivalente
  costruito con full adder o componenti logici (ammessi solo per conversioni di
  interfaccia, indici/cicli e nei test per il valore atteso);
- **nessuna dipendenza esterna** oltre a RetroNet Logic;
- i componenti **sequenziali** sono `struct` con stato e metodo `Step(clk)`,
  clock esplicito, flip-flop edge-triggered in stile master-slave.

## Sviluppo locale (go.work)

Il progetto dipende da RetroNet Logic. Per svilupparli insieme senza pubblicare,
clona i due repo come cartelle sibling e crea un workspace Go (non versionato):

```
work/
├── retronet-logic/
└── retronet-hardware/
```

```sh
cd retronet-hardware
go work init . ../retronet-logic   # crea go.work (ignorato da git)
```

## Verifiche locali

Prima di aprire una PR:

```sh
gofmt -w .
go vet ./...
go test ./...
```

oppure i target del Makefile: `make fmt`, `make vet`, `make test`, `make all`.

## Commit

Commit **piccoli e atomici**, messaggi in stile Conventional Commits in
italiano, ad esempio:

```
feat(flipflop): implementa il D flip-flop edge-triggered (master-slave)
docs(register): aggiungi diagramma di temporizzazione
```

## Documentazione

La documentazione è in **italiano**. Per i componenti sequenziali includi, oltre
all'esempio e al diagramma ASCII, un diagramma degli stati o di temporizzazione
quando aiuta a capire il comportamento rispetto al clock.
