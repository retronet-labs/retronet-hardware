package pc

import (
	"fmt"
	"testing"

	"github.com/retronet-labs/retronet-logic/bit"
	"github.com/retronet-labs/retronet-logic/bus"
)

// ciclo esegue un colpo di clock completo (fase bassa + fronte di salita).
func ciclo(p *PC, addr bus.Bus, load, inc bit.Bit) {
	p.Step(addr, load, inc, bit.Zero)
	p.Step(addr, load, inc, bit.One)
}

func TestIncremento(t *testing.T) {
	p := New(8)
	zero := bus.FromUint(0, 8)
	if p.Value().Uint() != 0 {
		t.Fatalf("PC iniziale != 0")
	}
	for want := uint64(1); want <= 5; want++ {
		ciclo(p, zero, bit.Zero, bit.One)
		if got := p.Value().Uint(); got != want {
			t.Fatalf("dopo inc: PC = %d, atteso %d", got, want)
		}
	}
}

func TestCaricamentoEPriorita(t *testing.T) {
	p := New(8)
	// Salto a 0x40.
	ciclo(p, bus.FromUint(0x40, 8), bit.One, bit.Zero)
	if got := p.Value().Uint(); got != 0x40 {
		t.Fatalf("dopo load: PC = %#x, atteso 0x40", got)
	}
	// load ha priorità su inc.
	ciclo(p, bus.FromUint(0x10, 8), bit.One, bit.One)
	if got := p.Value().Uint(); got != 0x10 {
		t.Fatalf("load+inc: PC = %#x, atteso 0x10 (load prevale)", got)
	}
}

func TestMantiene(t *testing.T) {
	p := New(8)
	zero := bus.FromUint(0, 8)
	ciclo(p, zero, bit.Zero, bit.One)  // PC = 1
	ciclo(p, zero, bit.Zero, bit.Zero) // né load né inc: mantiene
	if got := p.Value().Uint(); got != 1 {
		t.Errorf("PC = %d, atteso 1 (mantenuto)", got)
	}
}

func ExamplePC() {
	p := New(8)
	zero := bus.FromUint(0, 8)
	ciclo(p, zero, bit.Zero, bit.One) // inc -> 1
	ciclo(p, zero, bit.Zero, bit.One) // inc -> 2
	fmt.Println(p.Value().Uint())
	ciclo(p, bus.FromUint(100, 8), bit.One, bit.Zero) // salto a 100
	fmt.Println(p.Value().Uint())
	// Output:
	// 2
	// 100
}
