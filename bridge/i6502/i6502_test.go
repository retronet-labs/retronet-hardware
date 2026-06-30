package i6502

import "testing"

func refADC(a, v byte, carryIn bool, decimal bool) (byte, Flags) {
	c := 0
	if carryIn {
		c = 1
	}
	sum := int(a) + int(v) + c
	bin := byte(sum)
	f := Flags{
		Carry:    sum > 0xFF,
		Zero:     bin == 0,
		Negative: bin&0x80 != 0,
		Overflow: (^(a ^ v) & (a ^ bin) & 0x80) != 0,
	}
	if !decimal {
		return bin, f
	}
	result := sum
	if int(a&0x0F)+int(v&0x0F)+c > 9 {
		result += 0x06
	}
	if sum > 0x99 {
		result += 0x60
		f.Carry = true
	} else {
		f.Carry = false
	}
	return byte(result), f
}

func refSBC(a, v byte, carryIn bool, decimal bool) (byte, Flags) {
	borrow := 1
	if carryIn {
		borrow = 0
	}
	diff := int(a) - int(v) - borrow
	bin := byte(diff)
	f := Flags{
		Carry:    diff >= 0,
		Zero:     bin == 0,
		Negative: bin&0x80 != 0,
		Overflow: ((a ^ v) & (a ^ bin) & 0x80) != 0,
	}
	if !decimal {
		return bin, f
	}
	result := diff
	if int(a&0x0F)-borrow < int(v&0x0F) {
		result -= 0x06
	}
	if diff < 0 {
		result -= 0x60
	}
	return byte(result), f
}

func TestADCDifferentialVsReference(t *testing.T) {
	for a := 0; a <= 0xFF; a++ {
		for v := 0; v <= 0xFF; v++ {
			for _, carry := range []bool{false, true} {
				for _, dec := range []bool{false, true} {
					gotR, gotF := ADC(byte(a), byte(v), carry, dec)
					wantR, wantF := refADC(byte(a), byte(v), carry, dec)
					if gotR != wantR || gotF != wantF {
						t.Fatalf("ADC a=%#02x v=%#02x c=%v d=%v: got %#02x %+v want %#02x %+v",
							a, v, carry, dec, gotR, gotF, wantR, wantF)
					}
				}
			}
		}
	}
}

func TestSBCDifferentialVsReference(t *testing.T) {
	for a := 0; a <= 0xFF; a++ {
		for v := 0; v <= 0xFF; v++ {
			for _, carry := range []bool{false, true} {
				for _, dec := range []bool{false, true} {
					gotR, gotF := SBC(byte(a), byte(v), carry, dec)
					wantR, wantF := refSBC(byte(a), byte(v), carry, dec)
					if gotR != wantR || gotF != wantF {
						t.Fatalf("SBC a=%#02x v=%#02x c=%v d=%v: got %#02x %+v want %#02x %+v",
							a, v, carry, dec, gotR, gotF, wantR, wantF)
					}
				}
			}
		}
	}
}

func TestLogicCompareAndBit(t *testing.T) {
	if r, f := Logic(OpAND, 0xF0, 0x0F); r != 0 || !f.Zero || f.Negative {
		t.Fatalf("AND = %#02x %+v", r, f)
	}
	if r, f := Logic(OpORA, 0x80, 0x01); r != 0x81 || f.Zero || !f.Negative {
		t.Fatalf("ORA = %#02x %+v", r, f)
	}
	if r, f := Logic(OpEOR, 0xFF, 0x7F); r != 0x80 || f.Zero || !f.Negative {
		t.Fatalf("EOR = %#02x %+v", r, f)
	}
	if r, f := Compare(0x10, 0x20); r != 0xF0 || f.Carry || f.Zero || !f.Negative {
		t.Fatalf("CMP borrow = %#02x %+v", r, f)
	}
	if f := BIT(0x0F, 0xC0); !f.Zero || !f.Negative || !f.Overflow {
		t.Fatalf("BIT = %+v", f)
	}
}

func TestIncDecShiftRotate(t *testing.T) {
	if r, f := Increment(0xFF); r != 0x00 || !f.Zero || f.Negative {
		t.Fatalf("INC = %#02x %+v", r, f)
	}
	if r, f := Decrement(0x00); r != 0xFF || f.Zero || !f.Negative {
		t.Fatalf("DEC = %#02x %+v", r, f)
	}
	if r, f := ShiftLeft(0x80); r != 0x00 || !f.Carry || !f.Zero {
		t.Fatalf("ASL = %#02x %+v", r, f)
	}
	if r, f := ShiftRight(0x01); r != 0x00 || !f.Carry || !f.Zero {
		t.Fatalf("LSR = %#02x %+v", r, f)
	}
	if r, f := RotateLeft(0x80, true); r != 0x01 || !f.Carry || f.Zero {
		t.Fatalf("ROL = %#02x %+v", r, f)
	}
	if r, f := RotateRight(0x01, true); r != 0x80 || !f.Carry || !f.Negative {
		t.Fatalf("ROR = %#02x %+v", r, f)
	}
}
