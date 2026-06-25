package i8080

import "testing"

func TestALUAddSetsCarryAndAuxiliaryCarry(t *testing.T) {
	result, flags := ALU(GroupADD, 0x8F, 0x81, false)
	if result != 0x10 || !flags.Carry || !flags.AuxiliaryCarry {
		t.Fatalf("ADD result=0x%02X flags=%+v", result, flags)
	}
}

func TestALUSubUsesBorrowConvention(t *testing.T) {
	result, flags := ALU(GroupSUB, 0x00, 0x01, false)
	if result != 0xFF || !flags.Carry || !flags.AuxiliaryCarry || !flags.Sign {
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
