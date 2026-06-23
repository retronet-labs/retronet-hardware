package flipflop

import (
	"fmt"
	"testing"

	"github.com/retronet-labs/retronet-logic/bit"
)

// Il D latch è trasparente con clock alto e mantiene con clock basso.
func TestDLatchTrasparenzaEHold(t *testing.T) {
	d := NewDLatch()

	if got := d.Step(bit.One, bit.One); got != bit.One { // trasparente: Q segue D
		t.Errorf("clk=1 d=1: Q=%v, atteso 1", got)
	}
	if got := d.Step(bit.Zero, bit.One); got != bit.Zero { // trasparente: Q segue D
		t.Errorf("clk=1 d=0: Q=%v, atteso 0", got)
	}
	if got := d.Step(bit.One, bit.Zero); got != bit.Zero { // hold: ignora D
		t.Errorf("clk=0 d=1: Q=%v, atteso 0 (hold)", got)
	}
	if got := d.Step(bit.One, bit.One); got != bit.One { // di nuovo trasparente
		t.Errorf("clk=1 d=1: Q=%v, atteso 1", got)
	}
}

// ciclo simula un colpo di clock: presenta d con clock basso, poi alza il clock
// (fronte di salita) e restituisce l'uscita catturata.
func ciclo(ff *DFlipFlop, d bit.Bit) bit.Bit {
	ff.Step(d, bit.Zero)
	return ff.Step(d, bit.One)
}

func TestDFlipFlopFronteDiSalita(t *testing.T) {
	ff := NewDFlipFlop()

	if got := ff.Q(); got != bit.Zero {
		t.Fatalf("stato iniziale Q=%v, atteso 0", got)
	}
	if got := ciclo(ff, bit.One); got != bit.One {
		t.Errorf("dopo ciclo con d=1: Q=%v, atteso 1", got)
	}
	if got := ciclo(ff, bit.One); got != bit.One {
		t.Errorf("ciclo con d=1: Q=%v, atteso 1", got)
	}
	if got := ciclo(ff, bit.Zero); got != bit.Zero {
		t.Errorf("dopo ciclo con d=0: Q=%v, atteso 0", got)
	}
}

// Un cambiamento di D mentre il clock è già alto non deve modificare l'uscita
// (proprietà fondamentale dell'edge-triggered rispetto al latch).
func TestDFlipFlopIgnoraDFuoriDalFronte(t *testing.T) {
	ff := NewDFlipFlop()
	ciclo(ff, bit.One) // Q = 1

	if got := ff.Step(bit.Zero, bit.One); got != bit.One {
		t.Errorf("D cambiato con clk gia' alto: Q=%v, atteso 1 (invariato)", got)
	}
}

func ExampleDFlipFlop() {
	ff := NewDFlipFlop()
	// Un ciclo di clock con D=1: dato presente al fronte di salita.
	ff.Step(bit.One, bit.Zero)
	fmt.Println(ff.Step(bit.One, bit.One)) // fronte di salita -> cattura 1
	// Output: 1
}
