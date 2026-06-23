package latch

import (
	"fmt"
	"testing"

	"github.com/retronet-labs/retronet-logic/bit"
)

func TestSRSequenza(t *testing.T) {
	l := NewSR()

	// Stato iniziale: reset.
	if got := l.Q(); got != bit.Zero {
		t.Fatalf("stato iniziale Q = %v, atteso 0", got)
	}

	passi := []struct {
		nome   string
		s, r   bit.Bit
		wantQ  bit.Bit
		wantQn bit.Bit
	}{
		{"set", bit.One, bit.Zero, bit.One, bit.Zero},
		{"hold dopo set", bit.Zero, bit.Zero, bit.One, bit.Zero},
		{"reset", bit.Zero, bit.One, bit.Zero, bit.One},
		{"hold dopo reset", bit.Zero, bit.Zero, bit.Zero, bit.One},
		{"set di nuovo", bit.One, bit.Zero, bit.One, bit.Zero},
	}
	for _, p := range passi {
		q, qn := l.Step(p.s, p.r)
		if q != p.wantQ || qn != p.wantQn {
			t.Errorf("%s: Step(%v,%v) = (Q=%v,Qn=%v), atteso (Q=%v,Qn=%v)",
				p.nome, p.s, p.r, q, qn, p.wantQ, p.wantQn)
		}
	}
}

// Negli stati validi Qn deve essere il complemento di Q.
func TestSRComplemento(t *testing.T) {
	l := NewSR()
	for _, sr := range [][2]bit.Bit{{1, 0}, {0, 0}, {0, 1}, {0, 0}} {
		q, qn := l.Step(sr[0], sr[1])
		if q == qn {
			t.Errorf("dopo Step(%v,%v): Q=%v e Qn=%v non sono complementari", sr[0], sr[1], q, qn)
		}
	}
}

func ExampleSR() {
	l := NewSR()
	q, _ := l.Step(bit.One, bit.Zero) // set
	fmt.Println(q)
	q, _ = l.Step(bit.Zero, bit.Zero) // hold: mantiene 1
	fmt.Println(q)
	q, _ = l.Step(bit.Zero, bit.One) // reset
	fmt.Println(q)
	// Output:
	// 1
	// 1
	// 0
}
