package i4004

import (
	"fmt"
	"testing"
)

// --- Riferimento: copia fedele della semantica aritmetica dell'emulatore 4004 ---
// (go-4004/cpu/instructions.go: ADD, SUB, IAC, DAC, CMA). Tutti i valori sono
// nibble (0-15) e carry = risultato > 0x0F.

func b(cin bool) int {
	if cin {
		return 1
	}
	return 0
}

func refAdd(a, r byte, cin bool) (byte, bool) {
	result := int(a) + int(r) + b(cin)
	return byte(result & 0x0F), result > 0x0F
}

func refSub(a, r byte, cin bool) (byte, bool) {
	sum := int(a) + int((^r)&0x0F) + b(cin)
	return byte(sum & 0x0F), sum > 0x0F
}

func refInc(a byte) (byte, bool) {
	result := int(a) + 1
	return byte(result & 0x0F), result > 0x0F
}

func refDec(a byte) (byte, bool) {
	result := int(a) + 0x0F
	return byte(result & 0x0F), result > 0x0F
}

func TestAddSubConformita(t *testing.T) {
	for a := 0; a <= 0x0F; a++ {
		for r := 0; r <= 0x0F; r++ {
			for _, cin := range []bool{false, true} {
				gotR, gotC := Add(byte(a), byte(r), cin)
				wantR, wantC := refAdd(byte(a), byte(r), cin)
				if gotR != wantR || gotC != wantC {
					t.Fatalf("Add(%d,%d,%v) = (%d,%v), atteso (%d,%v)", a, r, cin, gotR, gotC, wantR, wantC)
				}

				gotR, gotC = Sub(byte(a), byte(r), cin)
				wantR, wantC = refSub(byte(a), byte(r), cin)
				if gotR != wantR || gotC != wantC {
					t.Fatalf("Sub(%d,%d,%v) = (%d,%v), atteso (%d,%v)", a, r, cin, gotR, gotC, wantR, wantC)
				}
			}
		}
	}
}

func TestUnarieConformita(t *testing.T) {
	for a := 0; a <= 0x0F; a++ {
		gotR, gotC := Inc(byte(a))
		wantR, wantC := refInc(byte(a))
		if gotR != wantR || gotC != wantC {
			t.Errorf("Inc(%d) = (%d,%v), atteso (%d,%v)", a, gotR, gotC, wantR, wantC)
		}

		gotR, gotC = Dec(byte(a))
		wantR, wantC = refDec(byte(a))
		if gotR != wantR || gotC != wantC {
			t.Errorf("Dec(%d) = (%d,%v), atteso (%d,%v)", a, gotR, gotC, wantR, wantC)
		}

		if got, want := Complement(byte(a)), (^byte(a))&0x0F; got != want {
			t.Errorf("Complement(%d) = %d, atteso %d", a, got, want)
		}
	}
}

func ExampleSub() {
	// 9 - 5 = 4, nessun prestito (carry = true nella convenzione 4004).
	res, carry := Sub(9, 5, true)
	fmt.Printf("%d carry=%v\n", res, carry)
	// 5 - 9 = -4 -> nibble 12, prestito (carry = false).
	res, carry = Sub(5, 9, true)
	fmt.Printf("%d carry=%v\n", res, carry)
	// Output:
	// 4 carry=true
	// 12 carry=false
}
