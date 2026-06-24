package registerfile

import (
	"fmt"
	"testing"

	"github.com/retronet-labs/retronet-logic/bit"
	"github.com/retronet-labs/retronet-logic/bus"
)

// scrivi esegue un ciclo di clock completo scrivendo value nel registro sel.
func scrivi(f *File, sel int, value uint64) {
	data := bus.FromUint(value, f.Width())
	f.Step(sel, data, bit.One, bit.Zero)
	f.Step(sel, data, bit.One, bit.One)
}

func TestLetturaScrittura(t *testing.T) {
	f := New(4, 8)
	for i := 0; i < f.Count(); i++ {
		if f.Read(i).Uint() != 0 {
			t.Fatalf("R%d iniziale != 0", i)
		}
	}

	scrivi(f, 2, 0xAB)
	if got := f.Read(2).Uint(); got != 0xAB {
		t.Errorf("R2 = %#x, atteso 0xAB", got)
	}
	for _, i := range []int{0, 1, 3} {
		if f.Read(i).Uint() != 0 {
			t.Errorf("R%d modificato per errore", i)
		}
	}
}

func TestScritturaDisabilitataMantiene(t *testing.T) {
	f := New(4, 8)
	scrivi(f, 1, 0x55)

	// Con write=0 il banco non deve cambiare, anche presentando un dato nuovo.
	data := bus.FromUint(0xFF, 8)
	f.Step(1, data, bit.Zero, bit.Zero)
	f.Step(1, data, bit.Zero, bit.One)

	if got := f.Read(1).Uint(); got != 0x55 {
		t.Errorf("R1 = %#x, atteso 0x55 (mantenuto)", got)
	}
}

func ExampleFile() {
	f := New(4, 8)
	d := bus.FromUint(42, 8)
	f.Step(3, d, bit.One, bit.Zero) // fase bassa
	f.Step(3, d, bit.One, bit.One)  // fronte di salita: scrive R3
	fmt.Println(f.Read(3).Uint())
	// Output: 42
}
