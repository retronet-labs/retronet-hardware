package i8008

import (
	"fmt"
	"math/bits"
	"testing"
)

// --- Riferimento: copia fedele della semantica ALU dell'emulatore 8008 ---
// (retronet-8008/cpu/alu.go: executeALU, add, sub, updateZeroSignParity).
// Serve solo a verificare la conformità senza dipendere dal repo dell'emulatore.

func refUpdateZSP(value byte) (zero, sign, parity bool) {
	return value == 0, value&0x80 != 0, bits.OnesCount8(value)%2 == 0
}

func refAdd(a, value byte, carryIn bool) (byte, Flags) {
	result := uint16(a) + uint16(value)
	if carryIn {
		result++
	}
	out := byte(result)
	z, s, p := refUpdateZSP(out)
	return out, Flags{Carry: result > 0xFF, Zero: z, Sign: s, Parity: p}
}

func refSub(a, value byte, borrowIn bool) (byte, Flags) {
	result := int(a) - int(value)
	if borrowIn {
		result--
	}
	out := byte(result)
	z, s, p := refUpdateZSP(out)
	return out, Flags{Carry: result < 0, Zero: z, Sign: s, Parity: p}
}

func refLogic(out byte) Flags {
	z, s, p := refUpdateZSP(out)
	return Flags{Carry: false, Zero: z, Sign: s, Parity: p}
}

func ref(group byte, a, value byte, carryIn bool) (byte, Flags) {
	switch group & 0x07 {
	case GroupADD:
		return refAdd(a, value, false)
	case GroupADC:
		return refAdd(a, value, carryIn)
	case GroupSUB:
		return refSub(a, value, false)
	case GroupSBB:
		return refSub(a, value, carryIn)
	case GroupAND:
		return a & value, refLogic(a & value)
	case GroupXOR:
		return a ^ value, refLogic(a ^ value)
	case GroupOR:
		return a | value, refLogic(a | value)
	default: // GroupCMP: come SUB (il risultato non viene memorizzato)
		return refSub(a, value, false)
	}
}

// TestALUConformitaEsaustiva verifica che l'adattatore (ALU a porte) produca,
// per ogni gruppo e per ogni combinazione di accumulatore, operando e carry, lo
// stesso risultato e gli stessi flag dell'emulatore 8008.
func TestALUConformitaEsaustiva(t *testing.T) {
	groups := []byte{GroupADD, GroupADC, GroupSUB, GroupSBB, GroupAND, GroupXOR, GroupOR, GroupCMP}
	for _, g := range groups {
		for a := 0; a <= 0xFF; a++ {
			for v := 0; v <= 0xFF; v++ {
				for _, cin := range []bool{false, true} {
					gotR, gotF := ALU(g, byte(a), byte(v), cin)
					wantR, wantF := ref(g, byte(a), byte(v), cin)
					if gotR != wantR || gotF != wantF {
						t.Fatalf("group=%d a=%#02x v=%#02x cin=%v: got (%#02x,%+v), want (%#02x,%+v)",
							g, a, v, cin, gotR, gotF, wantR, wantF)
					}
				}
			}
		}
	}
}

// TestIncrementDecrementConformita verifica che Increment/Decrement (INR/DCR)
// producano value±1 e i flag Zero/Sign/Parity attesi, su tutti i byte.
func TestIncrementDecrementConformita(t *testing.T) {
	for v := 0; v <= 0xFF; v++ {
		gotR, gz, gs, gp := Increment(byte(v))
		wantR := byte(v + 1)
		wz, ws, wp := refUpdateZSP(wantR)
		if gotR != wantR || gz != wz || gs != ws || gp != wp {
			t.Fatalf("Increment(%#02x) = (%#02x,z=%v,s=%v,p=%v), atteso (%#02x,z=%v,s=%v,p=%v)",
				v, gotR, gz, gs, gp, wantR, wz, ws, wp)
		}

		gotR, gz, gs, gp = Decrement(byte(v))
		wantR = byte(v - 1)
		wz, ws, wp = refUpdateZSP(wantR)
		if gotR != wantR || gz != wz || gs != ws || gp != wp {
			t.Fatalf("Decrement(%#02x) = (%#02x,z=%v,s=%v,p=%v), atteso (%#02x,z=%v,s=%v,p=%v)",
				v, gotR, gz, gs, gp, wantR, wz, ws, wp)
		}
	}
}

// TestRotazioniConformita confronta le rotazioni (RLC/RRC/RAL/RAR) con la
// semantica dell'emulatore 8008 (cpu/rotate.go) su tutti i byte.
func TestRotazioniConformita(t *testing.T) {
	for v := 0; v <= 0xFF; v++ {
		b := byte(v)

		if gotR, gotC := RotateLeftCircular(b); gotR != byte((v<<1)|(v>>7)) || gotC != (v&0x80 != 0) {
			t.Fatalf("RLC(%#02x) = %#02x,c=%v", v, gotR, gotC)
		}
		if gotR, gotC := RotateRightCircular(b); gotR != byte((v>>1)|(v<<7)) || gotC != (v&0x01 != 0) {
			t.Fatalf("RRC(%#02x) = %#02x,c=%v", v, gotR, gotC)
		}

		for _, cin := range []bool{false, true} {
			ci := 0
			if cin {
				ci = 1
			}
			if gotR, gotC := RotateLeftThroughCarry(b, cin); gotR != byte((v<<1)|ci) || gotC != (v&0x80 != 0) {
				t.Fatalf("RAL(%#02x,%v) = %#02x,c=%v", v, cin, gotR, gotC)
			}
			cm := 0
			if cin {
				cm = 0x80
			}
			if gotR, gotC := RotateRightThroughCarry(b, cin); gotR != byte((v>>1)|cm) || gotC != (v&0x01 != 0) {
				t.Fatalf("RAR(%#02x,%v) = %#02x,c=%v", v, cin, gotR, gotC)
			}
		}
	}
}

func ExampleALU() {
	// SUB 50 - 20: nessun prestito (Carry=false nella convenzione 8008).
	res, f := ALU(GroupSUB, 50, 20, false)
	fmt.Printf("%d carry=%v zero=%v\n", res, f.Carry, f.Zero)

	// CMP di due valori uguali: Zero=true.
	_, f = ALU(GroupCMP, 42, 42, false)
	fmt.Printf("zero=%v\n", f.Zero)
	// Output:
	// 30 carry=false zero=false
	// zero=true
}
