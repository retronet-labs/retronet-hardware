package register

import (
	"fmt"
	"testing"

	"github.com/retronet-labs/retronet-logic/bit"
	"github.com/retronet-labs/retronet-logic/bus"
)

// ciclo simula un colpo di clock completo (fase bassa + fronte di salita).
func ciclo(r *Register, data bus.Bus, load bit.Bit) bus.Bus {
	r.Step(data, load, bit.Zero)
	return r.Step(data, load, bit.One)
}

func TestRegisterCaricaEMantiene(t *testing.T) {
	const width = 4
	r := New(width)

	if got := r.Value().Uint(); got != 0 {
		t.Fatalf("valore iniziale = %d, atteso 0", got)
	}

	// Carica 10 (1010).
	if got := ciclo(r, bus.FromUint(10, width), bit.One).Uint(); got != 10 {
		t.Errorf("dopo load 10: %d, atteso 10", got)
	}
	// load=0: presenta 5 ma deve mantenere 10.
	if got := ciclo(r, bus.FromUint(5, width), bit.Zero).Uint(); got != 10 {
		t.Errorf("hold con load=0: %d, atteso 10", got)
	}
	// Carica 5.
	if got := ciclo(r, bus.FromUint(5, width), bit.One).Uint(); got != 5 {
		t.Errorf("dopo load 5: %d, atteso 5", got)
	}
	// Value() deve coincidere con l'ultimo contenuto.
	if got := r.Value().Uint(); got != 5 {
		t.Errorf("Value() = %d, atteso 5", got)
	}
}

func TestRegisterLarghezzaErrataPanico(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("atteso panico per dato di larghezza errata")
		}
	}()
	r := New(4)
	r.Step(bus.FromUint(0, 8), bit.One, bit.One)
}

func ExampleRegister() {
	r := New(8)
	// Carica 42, poi mantiene nonostante un nuovo dato con load=0.
	r.Step(bus.FromUint(42, 8), bit.One, bit.Zero)
	r.Step(bus.FromUint(42, 8), bit.One, bit.One)
	fmt.Println(r.Value().Uint())

	r.Step(bus.FromUint(7, 8), bit.Zero, bit.Zero)
	r.Step(bus.FromUint(7, 8), bit.Zero, bit.One)
	fmt.Println(r.Value().Uint())
	// Output:
	// 42
	// 42
}
