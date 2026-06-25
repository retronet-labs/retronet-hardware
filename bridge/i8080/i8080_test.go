package i8080

import (
	"math/bits"
	"testing"
)

func TestALUAddSetsCarryAndAuxiliaryCarry(t *testing.T) {
	result, flags := ALU(GroupADD, 0x8F, 0x81, false)
	if result != 0x10 || !flags.Carry || !flags.AuxiliaryCarry {
		t.Fatalf("ADD result=0x%02X flags=%+v", result, flags)
	}
}

func TestALUSubUsesBorrowConvention(t *testing.T) {
	// 0x00 - 0x01 = 0xFF: prestito (Carry=true), segno negativo, e Auxiliary
	// Carry = false (nessun mezzo-prestito: 0x0 + 0xE + 1 = 0xF, niente riporto).
	result, flags := ALU(GroupSUB, 0x00, 0x01, false)
	if result != 0xFF || !flags.Carry || flags.AuxiliaryCarry || !flags.Sign {
		t.Fatalf("SUB result=0x%02X flags=%+v", result, flags)
	}
}

func TestLogicalFlagsMatch8080Convention(t *testing.T) {
	result, flags := ALU(GroupANA, 0xF0, 0x0F, false)
	if result != 0 || flags.Carry || !flags.AuxiliaryCarry || !flags.Zero {
		t.Fatalf("ANA result=0x%02X flags=%+v", result, flags)
	}

	result, flags = ALU(GroupXRA, 0x55, 0x55, false)
	if result != 0 || flags.Carry || flags.AuxiliaryCarry || !flags.Zero {
		t.Fatalf("XRA result=0x%02X flags=%+v", result, flags)
	}
}

func TestAnaAuxiliaryCarry(t *testing.T) {
	// Quirk 8080: l'AC di ANA è il bit 3 di (A OR value).
	casi := []struct {
		a, v   byte
		wantAC bool
	}{
		{0xF0, 0x0F, true},  // A|v = 0xFF -> bit3 = 1
		{0x00, 0x00, false}, // A|v = 0x00 -> bit3 = 0
		{0x08, 0x00, true},  // A|v = 0x08 -> bit3 = 1
		{0x04, 0x02, false}, // A|v = 0x06 -> bit3 = 0
	}
	for _, c := range casi {
		if _, flags := ALU(GroupANA, c.a, c.v, false); flags.AuxiliaryCarry != c.wantAC {
			t.Errorf("ANA(0x%02X,0x%02X): AC=%v, atteso %v", c.a, c.v, flags.AuxiliaryCarry, c.wantAC)
		}
	}
}

func TestAdd16CascadesCarry(t *testing.T) {
	result, carry := Add16(0xFFFF, 0x0001)
	if result != 0x0000 || !carry {
		t.Fatalf("Add16 result=0x%04X carry=%v", result, carry)
	}
}

// --- Riferimento 8080 (stile Cringle/8080EXM) per il test differenziale ---

func refParityEven(v byte) bool { return bits.OnesCount8(v)%2 == 0 }

// carryAt indica se l'addizione a+b+cy genera un riporto entrante nel bit dato.
func carryAt(bitNo int, a, b, cy int) bool {
	res := a + b + cy
	return (res^a^b)&(1<<bitNo) != 0
}

func refAdd(a, val byte, cy bool) (byte, Flags) {
	c := 0
	if cy {
		c = 1
	}
	out := byte(int(a) + int(val) + c)
	return out, Flags{
		Carry:          carryAt(8, int(a), int(val), c),
		AuxiliaryCarry: carryAt(4, int(a), int(val), c),
		Zero:           out == 0,
		Sign:           out&0x80 != 0,
		Parity:         refParityEven(out),
	}
}

func refSub(a, val byte, borrow bool) (byte, Flags) {
	// SUB equivale ad ADD(a, ^val, !borrow) con il Carry finale invertito.
	nc := 1
	if borrow {
		nc = 0
	}
	nval := int(^val)
	out := byte(int(a) + nval + nc)
	return out, Flags{
		Carry:          !carryAt(8, int(a), nval, nc),
		AuxiliaryCarry: carryAt(4, int(a), nval, nc),
		Zero:           out == 0,
		Sign:           out&0x80 != 0,
		Parity:         refParityEven(out),
	}
}

func refLogic(a, val, group byte) (byte, Flags) {
	var out byte
	var aux bool
	switch group {
	case GroupANA:
		out = a & val
		aux = (a|val)&0x08 != 0
	case GroupXRA:
		out = a ^ val
	default: // GroupORA
		out = a | val
	}
	return out, Flags{
		Zero:           out == 0,
		Sign:           out&0x80 != 0,
		Parity:         refParityEven(out),
		AuxiliaryCarry: aux,
	}
}

func ref8080(group, a, val byte, carryIn bool) (byte, Flags) {
	switch group {
	case GroupADD:
		return refAdd(a, val, false)
	case GroupADC:
		return refAdd(a, val, carryIn)
	case GroupSUB:
		return refSub(a, val, false)
	case GroupSBB:
		return refSub(a, val, carryIn)
	case GroupANA, GroupXRA, GroupORA:
		return refLogic(a, val, group)
	default: // GroupCMP: come SUB (risultato non memorizzato)
		return refSub(a, val, false)
	}
}

// TestALUDifferentialVs8080Reference confronta la ALU a gate con la semantica
// 8080 di riferimento su TUTTI gli ingressi e tutti i gruppi: risultato e ogni
// flag (compresi Auxiliary Carry e Parity). È la rete esaustiva che cattura i
// bug stanati da 8080EXM.
func TestALUDifferentialVs8080Reference(t *testing.T) {
	groups := []byte{GroupADD, GroupADC, GroupSUB, GroupSBB, GroupANA, GroupXRA, GroupORA, GroupCMP}
	for _, g := range groups {
		for a := 0; a <= 0xFF; a++ {
			for v := 0; v <= 0xFF; v++ {
				for _, cy := range []bool{false, true} {
					gotR, gotF := ALU(g, byte(a), byte(v), cy)
					wantR, wantF := ref8080(g, byte(a), byte(v), cy)
					if gotR != wantR || gotF != wantF {
						t.Fatalf("group=%d a=%#02x v=%#02x cy=%v: got (%#02x,%+v) want (%#02x,%+v)",
							g, a, v, cy, gotR, gotF, wantR, wantF)
					}
				}
			}
		}
	}
}
